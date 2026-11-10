// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package upgrade

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/util"
)

const (
	latestReleaseUrl   = "https://api.github.com/repos/casdoor/casdoor/releases/latest"
	releaseByTagUrl    = "https://api.github.com/repos/casdoor/casdoor/releases/tags/"
	releaseDownloadUrl = "https://github.com/casdoor/casdoor/releases/download"
	checkCacheTtl      = time.Hour
	checkErrorCacheTtl = 10 * time.Minute

	DeploymentRelease = "release"
	DeploymentDocker  = "docker"
	DeploymentSource  = "source"
	DeploymentBinary  = "binary"

	StateDownloading = "downloading"
	StateInstalling  = "installing"
	StateRestarting  = "restarting"
	StateFailed      = "failed"
)

var releaseTagRegex = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

type ReleaseAsset struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type Release struct {
	TagName     string         `json:"tag_name"`
	HtmlUrl     string         `json:"html_url"`
	PublishedAt string         `json:"published_at"`
	Assets      []ReleaseAsset `json:"assets"`
}

type UpgradeStatus struct {
	State      string `json:"state"`
	Version    string `json:"version"`
	OrderId    string `json:"orderId,omitempty"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Error      string `json:"error"`
}

// VendorOrder is the upgrade the license server asks for, shown to the admin
type VendorOrder struct {
	Id         string `json:"id"`
	Version    string `json:"version"`
	Note       string `json:"note"`
	IssuedTime string `json:"issuedTime"`
	CanUpgrade bool   `json:"canUpgrade"`
}

type UpgradeInfo struct {
	CurrentVersion    string         `json:"currentVersion"`
	CommitOffset      int            `json:"commitOffset"`
	LatestVersion     string         `json:"latestVersion"`
	LatestUrl         string         `json:"latestUrl"`
	PublishedAt       string         `json:"publishedAt"`
	HasNewVersion     bool           `json:"hasNewVersion"`
	CheckError        string         `json:"checkError"`
	DeploymentType    string         `json:"deploymentType"`
	IsCheckDisabled   bool           `json:"isCheckDisabled"`
	IsUpgradeDisabled bool           `json:"isUpgradeDisabled"`
	CanUpgrade        bool           `json:"canUpgrade"`
	Status            *UpgradeStatus `json:"status"`
	// the upgrade the license server asks for, which starts by itself when allowVendorUpgrade is set in app.conf
	VendorOrder        *VendorOrder   `json:"vendorOrder"`
	AllowVendorUpgrade bool           `json:"allowVendorUpgrade"`
	AssetName          string         `json:"assetName"`
	LastResult         *UpgradeResult `json:"lastResult"`
}

var (
	releaseMutex     sync.Mutex
	cachedRelease    *Release
	cachedReleaseErr error
	cachedReleaseAt  time.Time

	statusMutex sync.Mutex
	status      = &UpgradeStatus{}
)

func isCheckDisabled() bool {
	return conf.GetConfigBool("disableVersionCheck")
}

func isUpgradeDisabled() bool {
	return conf.GetConfigBool("disableUpgrade") || conf.IsDemoMode()
}

func isVendorUpgradeAllowed() bool {
	return conf.GetConfigBool("allowVendorUpgrade")
}

func isDocker() bool {
	return util.FileExist("/.dockerenv") || util.FileExist("/run/.containerenv") || os.Getenv("KUBERNETES_SERVICE_HOST") != ""
}

func isSource() bool {
	_, filename, _, _ := runtime.Caller(0)
	rootPath := filepath.Dir(filepath.Dir(filename))
	return util.FileExist(filepath.Join(rootPath, ".git")) || util.FileExist(".git")
}

// GetDeploymentType tells how this Casdoor was installed, only the binaries built by the release workflow
// (util.BuildType set by .goreleaser.yaml) can replace themselves
func GetDeploymentType() string {
	if isDocker() {
		return DeploymentDocker
	}
	if isSource() {
		return DeploymentSource
	}
	if util.BuildType == "release" && releaseTagRegex.MatchString(util.Version) {
		return DeploymentRelease
	}
	return DeploymentBinary
}

func getCurrentVersion() (string, int) {
	versionInfo, err := util.GetVersionInfo()
	if err == nil && versionInfo.Version != "" {
		return versionInfo.Version, versionInfo.CommitOffset
	}
	return util.Version, util.CommitOffset
}

func parseVersion(version string) []int {
	match := releaseTagRegex.FindStringSubmatch(version)
	if match == nil {
		return nil
	}

	res := []int{}
	for _, part := range match[1:] {
		num, err := strconv.Atoi(part)
		if err != nil {
			return nil
		}
		res = append(res, num)
	}
	return res
}

func isNewerVersion(latest string, current string) bool {
	latestParts := parseVersion(latest)
	currentParts := parseVersion(current)
	if latestParts == nil || currentParts == nil {
		return false
	}

	for i := range latestParts {
		if latestParts[i] != currentParts[i] {
			return latestParts[i] > currentParts[i]
		}
	}
	return false
}

func getHttpClient(timeout time.Duration) *http.Client {
	client := &http.Client{Timeout: timeout}
	if proxy.ProxyHttpClient != nil {
		client.Transport = proxy.ProxyHttpClient.Transport
	}
	return client
}

func newRequest(url string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", fmt.Sprintf("casdoor/%s", util.Version))
	return req, nil
}

func fetchLatestRelease() (*Release, error) {
	return fetchRelease(latestReleaseUrl)
}

func fetchRelease(url string) (*Release, error) {
	req, err := newRequest(url)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := getHttpClient(10 * time.Second).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, err
	}
	if !releaseTagRegex.MatchString(release.TagName) {
		return nil, fmt.Errorf("unexpected release tag: %s", release.TagName)
	}
	return &release, nil
}

func getLatestRelease() (*Release, error) {
	releaseMutex.Lock()
	defer releaseMutex.Unlock()

	ttl := checkCacheTtl
	if cachedReleaseErr != nil {
		ttl = checkErrorCacheTtl
	}
	if !cachedReleaseAt.IsZero() && time.Since(cachedReleaseAt) < ttl {
		return cachedRelease, cachedReleaseErr
	}

	cachedRelease, cachedReleaseErr = fetchLatestRelease()
	cachedReleaseAt = time.Now()
	if cachedReleaseErr != nil {
		logs.Warning("failed to check the latest Casdoor release: %s", cachedReleaseErr.Error())
	}
	return cachedRelease, cachedReleaseErr
}

func getAssetName() string {
	osName := map[string]string{"linux": "Linux", "darwin": "Darwin", "windows": "Windows"}[runtime.GOOS]
	arch := map[string]string{"amd64": "x86_64", "arm64": "arm64"}[runtime.GOARCH]
	if osName == "" || arch == "" {
		return ""
	}

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("casdoor_%s_%s.zip", osName, arch)
	}
	return fmt.Sprintf("casdoor_%s_%s.tar.gz", osName, arch)
}

func getChecksumsName(tag string) string {
	return fmt.Sprintf("casdoor_%s_checksums.txt", strings.TrimPrefix(tag, "v"))
}

func (release *Release) getAsset(name string) *ReleaseAsset {
	for i := range release.Assets {
		if release.Assets[i].Name == name {
			return &release.Assets[i]
		}
	}
	return nil
}

func getStatus() *UpgradeStatus {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	res := *status
	return &res
}

func setStatus(update func(status *UpgradeStatus)) {
	statusMutex.Lock()
	defer statusMutex.Unlock()

	update(status)
}

func isUpgrading(state string) bool {
	return state == StateDownloading || state == StateInstalling || state == StateRestarting
}

func canInstallHere() bool {
	return GetDeploymentType() == DeploymentRelease && !isUpgradeDisabled() && getAssetName() != ""
}

// isOrderPending tells whether an order still asks for a version other than the running one
func isOrderPending(order *edition.UpgradeOrder) bool {
	currentVersion, commitOffset := getCurrentVersion()
	return order != nil && releaseTagRegex.MatchString(order.Version) && (order.Version != currentVersion || commitOffset > 0)
}

func GetUpgradeInfo() *UpgradeInfo {
	currentVersion, commitOffset := getCurrentVersion()
	info := &UpgradeInfo{
		CurrentVersion:     currentVersion,
		CommitOffset:       commitOffset,
		DeploymentType:     GetDeploymentType(),
		IsCheckDisabled:    isCheckDisabled(),
		IsUpgradeDisabled:  isUpgradeDisabled(),
		Status:             getStatus(),
		AllowVendorUpgrade: isVendorUpgradeAllowed(),
		AssetName:          getAssetName(),
		LastResult:         GetLastResult(),
	}
	if order := edition.GetUpgradeOrder(); isOrderPending(order) {
		info.VendorOrder = &VendorOrder{
			Id:         order.Id,
			Version:    order.Version,
			Note:       order.Note,
			IssuedTime: order.IssuedTime,
			CanUpgrade: canInstallHere(),
		}
	}
	if info.IsCheckDisabled {
		return info
	}

	release, err := getLatestRelease()
	if err != nil {
		info.CheckError = err.Error()
		return info
	}

	info.LatestVersion = release.TagName
	info.LatestUrl = release.HtmlUrl
	info.PublishedAt = release.PublishedAt
	info.HasNewVersion = isNewerVersion(release.TagName, currentVersion)
	info.CanUpgrade = info.DeploymentType == DeploymentRelease && !info.IsUpgradeDisabled && info.HasNewVersion &&
		getAssetName() != "" && release.getAsset(getAssetName()) != nil && release.getAsset(getChecksumsName(release.TagName)) != nil
	return info
}

// StartUpgrade installs version in the background, which must be the latest release or the one the license server
// asks for; the progress is reported by GetUpgradeInfo()
func StartUpgrade(version string, operator string) error {
	order := edition.GetUpgradeOrder()
	if isOrderPending(order) && version == order.Version {
		return startOrderedUpgrade(order, "", operator)
	}

	info := GetUpgradeInfo()
	if !info.CanUpgrade {
		switch {
		case info.DeploymentType != DeploymentRelease:
			return fmt.Errorf("only an official release binary can upgrade itself, this Casdoor is deployed as: %s", info.DeploymentType)
		case info.IsUpgradeDisabled:
			return fmt.Errorf("upgrading is disabled by disableUpgrade in app.conf")
		case info.IsCheckDisabled:
			return fmt.Errorf("checking the latest version is disabled by disableVersionCheck in app.conf")
		case info.CheckError != "":
			return fmt.Errorf("failed to check the latest version: %s", info.CheckError)
		default:
			return fmt.Errorf("there is no new version to upgrade to")
		}
	}
	if version != info.LatestVersion {
		return fmt.Errorf("the latest version is %s, not %s", info.LatestVersion, version)
	}

	release, err := getLatestRelease()
	if err != nil {
		return err
	}
	return startUpgrade(release, "", "", "", operator)
}

// startOrderedUpgrade installs the release of an upgrade order, downloaded or, for an instance without Internet
// access, from archivePath that the admin uploaded; the checksum in the signed order is trusted over the release's
func startOrderedUpgrade(order *edition.UpgradeOrder, archivePath string, operator string) error {
	switch {
	case GetDeploymentType() != DeploymentRelease:
		return fmt.Errorf("only an official release binary can upgrade itself, this Casdoor is deployed as: %s", GetDeploymentType())
	case isUpgradeDisabled():
		return fmt.Errorf("upgrading is disabled by disableUpgrade in app.conf")
	case getAssetName() == "":
		return fmt.Errorf("there is no release package for %s/%s", runtime.GOOS, runtime.GOARCH)
	case !releaseTagRegex.MatchString(order.Version):
		return fmt.Errorf("the upgrade order asks for an invalid version: %s", order.Version)
	}

	checksum := strings.ToLower(order.Checksums[getAssetName()])
	if archivePath != "" && checksum == "" {
		return fmt.Errorf("the upgrade order has no checksum of %s, ask for an offline upgrade order", getAssetName())
	}

	release := &Release{TagName: order.Version}
	if checksum == "" {
		var err error
		release, err = fetchRelease(releaseByTagUrl + order.Version)
		if err != nil {
			return fmt.Errorf("failed to get the release %s: %s", order.Version, err.Error())
		}
		if release.TagName != order.Version || release.getAsset(getAssetName()) == nil || release.getAsset(getChecksumsName(release.TagName)) == nil {
			return fmt.Errorf("the release %s has no package for %s/%s", order.Version, runtime.GOOS, runtime.GOARCH)
		}
	}
	return startUpgrade(release, checksum, archivePath, order.Id, operator)
}

func startUpgrade(release *Release, checksum string, archivePath string, orderId string, operator string) error {
	statusMutex.Lock()
	if isUpgrading(status.State) {
		statusMutex.Unlock()
		return fmt.Errorf("an upgrade to %s is already in progress", status.Version)
	}
	status = &UpgradeStatus{State: StateDownloading, Version: release.TagName, OrderId: orderId}
	if archivePath != "" {
		status.State = StateInstalling
	}
	statusMutex.Unlock()

	logs.Info("upgrade from %s to %s started by %s", util.Version, release.TagName, operator)
	util.SafeGoroutine(func() { runUpgrade(release, checksum, archivePath, orderId) })
	return nil
}

// InstallUploadedPackage upgrades an instance without Internet access from a release package the admin uploaded,
// with the offline upgrade order the vendor gave
func InstallUploadedPackage(orderText string, archivePath string, operator string) error {
	order, err := edition.ParseUpgradeOrder(orderText)
	if err != nil {
		return err
	}
	return startOrderedUpgrade(order, archivePath, operator)
}

// onUpgradeOrder starts the upgrade a new order asks for when the admin let the vendor upgrade this instance
func onUpgradeOrder(order *edition.UpgradeOrder) {
	if !isVendorUpgradeAllowed() || !isOrderPending(order) || !canInstallHere() || isUpgrading(getStatus().State) {
		return
	}
	err := startOrderedUpgrade(order, "", "the license server")
	if err != nil {
		logs.Warn("upgrade: failed to start the upgrade to %s that the license server asks for: %s", order.Version, err.Error())
		setStatus(func(status *UpgradeStatus) {
			if !isUpgrading(status.State) {
				status.State = StateFailed
				status.Version = order.Version
				status.OrderId = order.Id
				status.Error = err.Error()
			}
		})
	}
}

func init() {
	edition.SetUpgradeOrderHandler(onUpgradeOrder)
}

func runUpgrade(release *Release, checksum string, archivePath string, orderId string) {
	if archivePath != "" {
		defer os.Remove(archivePath)
	}
	exePath, backupDir, err := installRelease(release, checksum, archivePath)
	if err != nil {
		logs.Error("upgrade to %s failed: %s", release.TagName, err.Error())
		setStatus(func(status *UpgradeStatus) {
			status.State = StateFailed
			status.Error = err.Error()
		})
		return
	}

	writePendingUpgrade(&pendingUpgrade{
		From:      util.Version,
		To:        release.TagName,
		OrderId:   orderId,
		BackupDir: backupDir,
		Time:      time.Now().Format(time.RFC3339),
	})

	logs.Info("upgrade to %s installed, restarting", release.TagName)
	setStatus(func(status *UpgradeStatus) { status.State = StateRestarting })
	time.Sleep(2 * time.Second)

	err = restart(exePath)
	if err != nil {
		logs.Error("restart after the upgrade to %s failed: %s", release.TagName, err.Error())
		setStatus(func(status *UpgradeStatus) {
			status.State = StateFailed
			status.Error = fmt.Sprintf("%s is installed but Casdoor failed to restart, please restart it manually: %s", release.TagName, err.Error())
		})
	}
}

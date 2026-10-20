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
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/casdoor/casdoor/util"
)

const (
	maxArchiveSize   = 512 << 20
	maxExtractedSize = 1 << 30
)

type progressWriter struct{}

func (progressWriter) Write(p []byte) (int, error) {
	setStatus(func(status *UpgradeStatus) { status.Downloaded += int64(len(p)) })
	return len(p), nil
}

func getReleaseFileUrl(tag string, name string) string {
	return fmt.Sprintf("%s/%s/%s", releaseDownloadUrl, tag, name)
}

func download(url string, timeout time.Duration) (*http.Response, error) {
	req, err := newRequest(url)
	if err != nil {
		return nil, err
	}

	resp, err := getHttpClient(timeout).Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("downloading %s returned status %d", url, resp.StatusCode)
	}
	return resp, nil
}

func getExpectedChecksum(tag string, assetName string) (string, error) {
	resp, err := download(getReleaseFileUrl(tag, getChecksumsName(tag)), time.Minute)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(io.LimitReader(resp.Body, 1<<20))
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == assetName {
			checksum := strings.ToLower(fields[0])
			if len(checksum) != sha256.Size*2 {
				return "", fmt.Errorf("invalid checksum of %s: %s", assetName, checksum)
			}
			return checksum, nil
		}
	}
	if err = scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("%s has no checksum of %s", getChecksumsName(tag), assetName)
}

func downloadArchive(tag string, assetName string, filePath string, expectedChecksum string) error {
	resp, err := download(getReleaseFileUrl(tag, assetName), 30*time.Minute)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.ContentLength > maxArchiveSize {
		return fmt.Errorf("%s is too large: %d bytes", assetName, resp.ContentLength)
	}
	if resp.ContentLength > 0 {
		setStatus(func(status *UpgradeStatus) { status.Total = resp.ContentLength })
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	written, err := io.Copy(io.MultiWriter(file, hash, progressWriter{}), io.LimitReader(resp.Body, maxArchiveSize+1))
	if err != nil {
		return err
	}
	if written > maxArchiveSize {
		return fmt.Errorf("%s is larger than %d bytes", assetName, maxArchiveSize)
	}

	checksum := hex.EncodeToString(hash.Sum(nil))
	if checksum != expectedChecksum {
		return fmt.Errorf("the SHA-256 of %s is %s, but the release's checksums file says %s", assetName, checksum, expectedChecksum)
	}
	return file.Close()
}

func getBinaryName() string {
	if runtime.GOOS == "windows" {
		return "casdoor.exe"
	}
	return "casdoor"
}

// getArchiveTarget maps an archive entry to the file it is extracted to, only the binary and web/build are wanted,
// conf/app.conf is never overwritten
func getArchiveTarget(destDir string, name string) (string, bool) {
	if strings.Contains(name, "\\") {
		return "", false
	}

	name = path.Clean(strings.TrimPrefix(name, "./"))
	if name != getBinaryName() && !strings.HasPrefix(name, "web/build/") {
		return "", false
	}
	return filepath.Join(destDir, filepath.FromSlash(name)), true
}

func writeFile(target string, reader io.Reader, mode os.FileMode, remaining *int64) error {
	err := os.MkdirAll(filepath.Dir(target), 0o755)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, io.LimitReader(reader, *remaining+1))
	if err != nil {
		return err
	}
	*remaining -= written
	if *remaining < 0 {
		return fmt.Errorf("the release archive extracts to more than %d bytes", maxExtractedSize)
	}
	return file.Close()
}

func getFileMode(target string) os.FileMode {
	if filepath.Base(target) == getBinaryName() {
		return 0o755
	}
	return 0o644
}

func extractTarGz(archivePath string, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	remaining := int64(maxExtractedSize)
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}

		target, ok := getArchiveTarget(destDir, header.Name)
		if !ok {
			continue
		}
		err = writeFile(target, tarReader, getFileMode(target), &remaining)
		if err != nil {
			return err
		}
	}
}

func extractZip(archivePath string, destDir string) error {
	zipReader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	remaining := int64(maxExtractedSize)
	for _, zipFile := range zipReader.File {
		if !zipFile.Mode().IsRegular() {
			continue
		}

		target, ok := getArchiveTarget(destDir, zipFile.Name)
		if !ok {
			continue
		}

		reader, err := zipFile.Open()
		if err != nil {
			return err
		}
		err = writeFile(target, reader, getFileMode(target), &remaining)
		reader.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}
	return dstFile.Close()
}

func getExecutablePath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exePath)
}

// getWebBuildDir returns the frontend folder that comes from the release archive: web/build in the working
// directory or next to the binary, "" when the frontend is served from somewhere else
func getWebBuildDir(exeDir string) string {
	for _, dir := range []string{"web/build", filepath.Join(exeDir, "web", "build")} {
		if util.FileExist(filepath.Join(dir, "index.html")) {
			absDir, err := filepath.Abs(dir)
			if err == nil {
				return absDir
			}
		}
	}
	return ""
}

// renameFile retries for a while because on Windows a file that is being served or scanned by an antivirus cannot be
// renamed until it is closed
func renameFile(oldPath string, newPath string) error {
	var err error
	for i := 0; i < 10; i++ {
		err = os.Rename(oldPath, newPath)
		if err == nil {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return err
}

// replaceExecutable moves the running binary aside instead of overwriting it, Windows allows renaming a running
// exe but not writing to or deleting it
func replaceExecutable(exePath string, newExePath string) error {
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath)

	err := renameFile(exePath, oldPath)
	if err != nil {
		return err
	}

	err = renameFile(newExePath, exePath)
	if err != nil {
		_ = renameFile(oldPath, exePath)
		return err
	}

	_ = os.Remove(oldPath)
	return nil
}

func installRelease(release *Release) (string, error) {
	assetName := getAssetName()
	if assetName == "" {
		return "", fmt.Errorf("there is no release package for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	exePath, err := getExecutablePath()
	if err != nil {
		return "", err
	}
	exeDir := filepath.Dir(exePath)

	leftovers, _ := filepath.Glob(filepath.Join(exeDir, ".upgrade-*"))
	for _, leftover := range leftovers {
		_ = os.RemoveAll(leftover)
	}

	stagingDir, err := os.MkdirTemp(exeDir, ".upgrade-")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(stagingDir)

	expectedChecksum, err := getExpectedChecksum(release.TagName, assetName)
	if err != nil {
		return "", err
	}

	if asset := release.getAsset(assetName); asset != nil {
		setStatus(func(status *UpgradeStatus) { status.Total = asset.Size })
	}
	archivePath := filepath.Join(stagingDir, assetName)
	err = downloadArchive(release.TagName, assetName, archivePath, expectedChecksum)
	if err != nil {
		return "", err
	}

	setStatus(func(status *UpgradeStatus) { status.State = StateInstalling })

	extractDir := filepath.Join(stagingDir, "release")
	if strings.HasSuffix(assetName, ".zip") {
		err = extractZip(archivePath, extractDir)
	} else {
		err = extractTarGz(archivePath, extractDir)
	}
	if err != nil {
		return "", err
	}

	newExePath := filepath.Join(extractDir, getBinaryName())
	if !util.FileExist(newExePath) {
		return "", fmt.Errorf("%s has no %s", assetName, getBinaryName())
	}

	webDir := getWebBuildDir(exeDir)
	newWebDir := filepath.Join(extractDir, "web", "build")
	if webDir != "" && !util.FileExist(filepath.Join(newWebDir, "index.html")) {
		return "", fmt.Errorf("%s has no web/build/index.html", assetName)
	}

	backupDir := filepath.Join(exeDir, "backup", fmt.Sprintf("%s_%s", util.Version, time.Now().Format("20060102150405")))
	err = os.MkdirAll(backupDir, 0o755)
	if err != nil {
		return "", err
	}
	err = copyFile(exePath, filepath.Join(backupDir, filepath.Base(exePath)))
	if err != nil {
		return "", err
	}

	backupWebDir := filepath.Join(backupDir, "web", "build")
	if webDir != "" {
		err = os.MkdirAll(filepath.Dir(backupWebDir), 0o755)
		if err != nil {
			return "", err
		}
		err = renameFile(webDir, backupWebDir)
		if err != nil {
			return "", err
		}
		err = renameFile(newWebDir, webDir)
		if err != nil {
			_ = renameFile(backupWebDir, webDir)
			return "", err
		}
	}

	err = replaceExecutable(exePath, newExePath)
	if err != nil {
		if webDir != "" {
			_ = os.RemoveAll(webDir)
			_ = renameFile(backupWebDir, webDir)
		}
		return "", err
	}

	return exePath, nil
}

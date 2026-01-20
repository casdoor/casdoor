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

package util

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
)

type LatestVersionInfo struct {
	Version     string `json:"version"`
	ReleaseUrl  string `json:"releaseUrl"`
	DownloadUrl string `json:"downloadUrl"`
	HasUpdate   bool   `json:"hasUpdate"`
}

type GitHubRelease struct {
	TagName string `json:"tag_name"`
	HtmlUrl string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadUrl string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetLatestVersion() (*LatestVersionInfo, error) {
	resp, err := http.Get("https://api.github.com/repos/casdoor/casdoor/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API returned status %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// Get current version
	currentVersion, err := GetVersionInfo()
	if err != nil {
		currentVersion = GetBuiltInVersionInfo()
	}

	// Determine the appropriate binary name for the current platform
	var binaryName string
	switch runtime.GOOS {
	case "linux":
		if runtime.GOARCH == "amd64" {
			binaryName = "server_linux_amd64"
		} else if runtime.GOARCH == "arm64" {
			binaryName = "server_linux_arm64"
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			binaryName = "server_windows_amd64.exe"
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			binaryName = "server_darwin_amd64"
		} else if runtime.GOARCH == "arm64" {
			binaryName = "server_darwin_arm64"
		}
	}

	// Find the download URL for the appropriate binary
	downloadUrl := ""
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			downloadUrl = asset.BrowserDownloadUrl
			break
		}
	}

	// Check if there's an update available
	hasUpdate := false
	if currentVersion.Version != "" && release.TagName != "" && release.TagName != currentVersion.Version {
		hasUpdate = true
	}

	return &LatestVersionInfo{
		Version:     release.TagName,
		ReleaseUrl:  release.HtmlUrl,
		DownloadUrl: downloadUrl,
		HasUpdate:   hasUpdate,
	}, nil
}

func PerformUpgrade(downloadUrl string) error {
	if downloadUrl == "" {
		return fmt.Errorf("no download URL available for this platform")
	}

	// Download the new binary
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return fmt.Errorf("failed to download upgrade: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download upgrade: HTTP %d", resp.StatusCode)
	}

	// Note: Actual upgrade implementation would require:
	// 1. Download the binary to a temporary location
	// 2. Verify the binary (checksum, signature)
	// 3. Replace the current binary (may require elevated permissions)
	// 4. Restart the service
	// This is a placeholder implementation that just validates the download is available
	return fmt.Errorf("upgrade functionality requires manual installation - please download from: %s", downloadUrl)
}

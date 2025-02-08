package controllers

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/beego/beego"
	"github.com/casdoor/casdoor/proxy"
	"github.com/casdoor/casdoor/util"
)

const (
	javaCliRepo    = "https://api.github.com/repos/jcasbin/casbin-java-cli/releases/latest"
	goCliRepo      = "https://api.github.com/repos/casbin/casbin-go-cli/releases/latest"
	rustCliRepo    = "https://api.github.com/repos/casbin-rs/casbin-rust-cli/releases/latest"
	downloadFolder = "bin"
)

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

// @Title getBinaryNames
// @Description Get binary names for different platforms and architectures
// @Success 200 {map[string]string} map[string]string "Binary names map"
func getBinaryNames() map[string]string {
	const (
		golang = "go"
		java   = "java"
		rust   = "rust"
	)

	arch := runtime.GOARCH
	archMap := map[string]struct{ goArch, rustArch string }{
		"amd64": {"x86_64", "x86_64"},
		"arm64": {"arm64", "aarch64"},
	}

	archNames, ok := archMap[arch]
	if !ok {
		archNames = struct{ goArch, rustArch string }{arch, arch}
	}

	switch runtime.GOOS {
	case "windows":
		return map[string]string{
			golang: fmt.Sprintf("casbin-go-cli_Windows_%s.zip", archNames.goArch),
			java:   "casbin-java-cli.jar",
			rust:   fmt.Sprintf("casbin-rust-cli-%s-pc-windows-gnu", archNames.rustArch),
		}
	case "darwin":
		return map[string]string{
			golang: fmt.Sprintf("casbin-go-cli_Darwin_%s.tar.gz", archNames.goArch),
			java:   "casbin-java-cli.jar",
			rust:   fmt.Sprintf("casbin-rust-cli-%s-apple-darwin", archNames.rustArch),
		}
	case "linux":
		return map[string]string{
			golang: fmt.Sprintf("casbin-go-cli_Linux_%s.tar.gz", archNames.goArch),
			java:   "casbin-java-cli.jar",
			rust:   fmt.Sprintf("casbin-rust-cli-%s-unknown-linux-gnu", archNames.rustArch),
		}
	default:
		return nil
	}
}

// @Title getFinalBinaryName
// @Description Get final binary name for specific language
// @Param lang string true "Language type (go/java/rust)"
// @Success 200 {string} string "Final binary name"
func getFinalBinaryName(lang string) string {
	switch lang {
	case "go":
		if runtime.GOOS == "windows" {
			return "casbin-go-cli.exe"
		}
		return "casbin-go-cli"
	case "java":
		return "casbin-java-cli.jar"
	case "rust":
		if runtime.GOOS == "windows" {
			return "casbin-rust-cli.exe"
		}
		return "casbin-rust-cli"
	default:
		return ""
	}
}

// @Title getLatestCLIURL
// @Description Get latest CLI download URL from GitHub
// @Param repoURL string true "GitHub repository URL"
// @Param language string true "Language type"
// @Success 200 {string} string "Download URL and version"
func getLatestCLIURL(repoURL string, language string) (string, string, error) {
	client := proxy.GetHttpClient(repoURL)
	resp, err := client.Get(repoURL)
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch release info: %v", err)
	}
	defer resp.Body.Close()

	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}

	binaryNames := getBinaryNames()
	if binaryNames == nil {
		return "", "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	binaryName := binaryNames[language]
	for _, asset := range release.Assets {
		if asset.Name == binaryName {
			return asset.URL, release.TagName, nil
		}
	}

	return "", "", fmt.Errorf("no suitable binary found for OS: %s, language: %s", runtime.GOOS, language)
}

// @Title extractGoCliFile
// @Description Extract the Go CLI file
// @Param filePath string true "The file path"
// @Success 200 {string} string "The extracted file path"
// @router /extractGoCliFile [post]
func extractGoCliFile(filePath string) error {
	tempDir := filepath.Join(downloadFolder, "temp")
	if err := os.MkdirAll(tempDir, 0o755); err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	if runtime.GOOS == "windows" {
		if err := unzipFile(filePath, tempDir); err != nil {
			return err
		}
	} else {
		if err := untarFile(filePath, tempDir); err != nil {
			return err
		}
	}

	execName := "casbin-go-cli"
	if runtime.GOOS == "windows" {
		execName += ".exe"
	}

	var execPath string
	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if info.Name() == execName {
			execPath = path
			return nil
		}
		return nil
	})
	if err != nil {
		return err
	}

	finalPath := filepath.Join(downloadFolder, execName)
	if err := os.Rename(execPath, finalPath); err != nil {
		return err
	}

	return os.Remove(filePath)
}

// @Title unzipFile
// @Description Unzip the file
// @Param zipPath string true "The zip file path"
// @Param destDir string true "The destination directory"
// @Success 200 {string} string "The extracted file path"
// @router /unzipFile [post]
func unzipFile(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// @Title untarFile
// @Description Untar the file
// @Param tarPath string true "The tar file path"
// @Param destDir string true "The destination directory"
// @Success 200 {string} string "The extracted file path"
// @router /untarFile [post]
func untarFile(tarPath, destDir string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(path)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

// @Title createJavaCliWrapper
// @Description Create the Java CLI wrapper
// @Param binPath string true "The binary path"
// @Success 200 {string} string "The created file path"
// @router /createJavaCliWrapper [post]
func createJavaCliWrapper(binPath string) error {
	if runtime.GOOS == "windows" {
		// Create a Windows CMD file
		cmdPath := filepath.Join(binPath, "casbin-java-cli.cmd")
		cmdContent := fmt.Sprintf(`@echo off
java -jar "%s\casbin-java-cli.jar" %%*`, binPath)

		err := os.WriteFile(cmdPath, []byte(cmdContent), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create Java CLI wrapper: %v", err)
		}
	} else {
		// Create Unix shell script
		shPath := filepath.Join(binPath, "casbin-java-cli")
		shContent := fmt.Sprintf(`#!/bin/sh
java -jar "%s/casbin-java-cli.jar" "$@"`, binPath)

		err := os.WriteFile(shPath, []byte(shContent), 0o755)
		if err != nil {
			return fmt.Errorf("failed to create Java CLI wrapper: %v", err)
		}
	}
	return nil
}

// @Title downloadCLI
// @Description Download and setup CLI tools
// @Success 200 {error} error "Error if any"
func downloadCLI() error {
	pathEnv := os.Getenv("PATH")
	binPath, err := filepath.Abs(downloadFolder)
	if err != nil {
		return fmt.Errorf("failed to get absolute path to download directory: %v", err)
	}

	if !strings.Contains(pathEnv, binPath) {
		newPath := fmt.Sprintf("%s%s%s", binPath, string(os.PathListSeparator), pathEnv)
		if err := os.Setenv("PATH", newPath); err != nil {
			return fmt.Errorf("failed to update PATH environment variable: %v", err)
		}
	}

	if err := os.MkdirAll(downloadFolder, 0o755); err != nil {
		return fmt.Errorf("failed to create download directory: %v", err)
	}

	repos := map[string]string{
		"java": javaCliRepo,
		"go":   goCliRepo,
		"rust": rustCliRepo,
	}

	for lang, repo := range repos {
		cliURL, version, err := getLatestCLIURL(repo, lang)
		if err != nil {
			fmt.Printf("failed to get %s CLI URL: %v\n", lang, err)
			continue
		}

		originalPath := filepath.Join(downloadFolder, getBinaryNames()[lang])
		fmt.Printf("downloading %s CLI: %s\n", lang, cliURL)

		client := proxy.GetHttpClient(cliURL)
		resp, err := client.Get(cliURL)
		if err != nil {
			fmt.Printf("failed to download %s CLI: %v\n", lang, err)
			continue
		}

		func() {
			defer resp.Body.Close()

			if err := os.MkdirAll(filepath.Dir(originalPath), 0o755); err != nil {
				fmt.Printf("failed to create directory for %s CLI: %v\n", lang, err)
				return
			}

			tmpFile := originalPath + ".tmp"
			out, err := os.Create(tmpFile)
			if err != nil {
				fmt.Printf("failed to create or write %s CLI: %v\n", lang, err)
				return
			}
			defer func() {
				out.Close()
				os.Remove(tmpFile)
			}()

			if _, err = io.Copy(out, resp.Body); err != nil ||
				out.Close() != nil ||
				os.Rename(tmpFile, originalPath) != nil {
				fmt.Printf("failed to download %s CLI: %v\n", lang, err)
				return
			}
		}()

		if lang == "go" {
			if err := extractGoCliFile(originalPath); err != nil {
				fmt.Printf("failed to extract Go CLI: %v\n", err)
				continue
			}
		} else {
			finalPath := filepath.Join(downloadFolder, getFinalBinaryName(lang))
			if err := os.Rename(originalPath, finalPath); err != nil {
				fmt.Printf("failed to rename %s CLI: %v\n", lang, err)
				continue
			}
		}

		if runtime.GOOS != "windows" {
			execPath := filepath.Join(downloadFolder, getFinalBinaryName(lang))
			if err := os.Chmod(execPath, 0o755); err != nil {
				fmt.Printf("failed to set %s CLI execution permission: %v\n", lang, err)
				continue
			}
		}

		fmt.Printf("downloaded %s CLI version: %s\n", lang, version)

		if lang == "java" {
			if err := createJavaCliWrapper(binPath); err != nil {
				fmt.Printf("failed to create Java CLI wrapper: %v\n", err)
				continue
			}
		}
	}

	return nil
}

// @Title RefreshEngines
// @Tag CLI API
// @Description Refresh all CLI engines
// @Param m query string true "Hash for request validation"
// @Param t query string true "Timestamp for request validation"
// @Success 200 {object} controllers.Response The Response object
// @router /refresh-engines [post]
func (c *ApiController) RefreshEngines() {
	if !beego.AppConfig.DefaultBool("isDemoMode", false) {
		c.ResponseError("refresh engines is only available in demo mode")
		return
	}

	hash := c.Input().Get("m")
	timestamp := c.Input().Get("t")

	if hash == "" || timestamp == "" {
		c.ResponseError("invalid identifier")
		return
	}

	requestTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		c.ResponseError("invalid identifier")
		return
	}

	timeDiff := time.Since(requestTime)
	if timeDiff > 5*time.Minute || timeDiff < -5*time.Minute {
		c.ResponseError("invalid identifier")
		return
	}

	version := "casbin-editor-v1"
	rawString := fmt.Sprintf("%s|%s", version, timestamp)

	hasher := sha256.New()
	hasher.Write([]byte(rawString))
	calculatedHash := strings.ToLower(hex.EncodeToString(hasher.Sum(nil)))

	if calculatedHash != strings.ToLower(hash) {
		c.ResponseError("invalid identifier")
		return
	}

	err = downloadCLI()
	if err != nil {
		c.ResponseError(fmt.Sprintf("failed to refresh engines: %v", err))
		return
	}

	c.ResponseOk(map[string]string{
		"status":  "success",
		"message": "CLI engines updated successfully",
	})
}

// @Title ScheduleCLIUpdater
// @Description Start periodic CLI update scheduler
func ScheduleCLIUpdater() {
	if !beego.AppConfig.DefaultBool("isDemoMode", false) {
		return
	}

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := downloadCLI()
		if err != nil {
			fmt.Printf("failed to update CLI: %v\n", err)
		} else {
			fmt.Println("CLI updated successfully")
		}
	}
}

// @Title DownloadCLI
// @Description Download the CLI
// @Success 200 {string} string "The downloaded file path"
// @router /downloadCLI [post]
func DownloadCLI() error {
	return downloadCLI()
}

// @Title InitCLIDownloader
// @Description Initialize CLI downloader and start update scheduler
func InitCLIDownloader() {
	if !beego.AppConfig.DefaultBool("isDemoMode", false) {
		return
	}

	util.SafeGoroutine(func() {
		err := DownloadCLI()
		if err != nil {
			fmt.Printf("failed to initialize CLI downloader: %v\n", err)
		}

		ScheduleCLIUpdater()
	})
}

package controllers

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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

func getBinaryNames() map[string]string {
	arch := runtime.GOARCH
	var goArch, rustArch string

	switch arch {
	case "amd64":
		goArch = "x86_64"
		rustArch = "x86_64"
	case "arm64":
		goArch = "arm64"
		rustArch = "aarch64"
	default:
		goArch = arch
		rustArch = arch
	}

	switch runtime.GOOS {
	case "windows":
		return map[string]string{
			"go":   fmt.Sprintf("casbin-go-cli_Windows_%s.zip", goArch),
			"java": "casbin-java-cli.jar",
			"rust": fmt.Sprintf("casbin-rust-cli-%s-pc-windows-gnu", rustArch),
		}
	case "darwin":
		result := map[string]string{
			"go":   fmt.Sprintf("casbin-go-cli_Darwin_%s.tar.gz", goArch),
			"java": "casbin-java-cli.jar",
			"rust": fmt.Sprintf("casbin-rust-cli-%s-apple-darwin", rustArch),
		}
		return result
	case "linux":
		return map[string]string{
			"go":   fmt.Sprintf("casbin-go-cli_Linux_%s.tar.gz", goArch),
			"java": "casbin-java-cli.jar",
			"rust": fmt.Sprintf("casbin-rust-cli-%s-unknown-linux-gnu", rustArch),
		}
	default:
		return nil
	}
}

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

func getLatestCLIURL(repoURL string, language string) (string, string, error) {
	resp, err := http.Get(repoURL)
	if err != nil {
		return "", "", err
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

func extractGoCliFile(filePath string) error {
	tempDir := filepath.Join(downloadFolder, "temp")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
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
			if err := os.MkdirAll(path, 0755); err != nil {
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

func createJavaCliWrapper(binPath string) error {
	if runtime.GOOS == "windows" {
		// Create a Windows CMD file
		cmdPath := filepath.Join(binPath, "casbin-java-cli.cmd")
		cmdContent := fmt.Sprintf(`@echo off
java -jar "%s\casbin-java-cli.jar" %%*`, binPath)

		err := os.WriteFile(cmdPath, []byte(cmdContent), 0755)
		if err != nil {
			return fmt.Errorf("创建 Java CLI wrapper 失败: %v", err)
		}
	} else {
		// Create Unix shell script
		shPath := filepath.Join(binPath, "casbin-java-cli")
		shContent := fmt.Sprintf(`#!/bin/sh
java -jar "%s/casbin-java-cli.jar" "$@"`, binPath)

		err := os.WriteFile(shPath, []byte(shContent), 0755)
		if err != nil {
			return fmt.Errorf("创建 Java CLI wrapper 失败: %v", err)
		}
	}
	return nil
}

func downloadCLI() error {
	// 获取系统 PATH 环境变量
	pathEnv := os.Getenv("PATH")
	binPath, err := filepath.Abs(downloadFolder)
	if err != nil {
		return fmt.Errorf("获取下载目录绝对路径失败: %v", err)
	}

	// 检查 bin 目录是否在 PATH 中
	if !strings.Contains(pathEnv, binPath) {
		// 将 bin 目录添加到 PATH
		newPath := fmt.Sprintf("%s%s%s", binPath, string(os.PathListSeparator), pathEnv)
		if err := os.Setenv("PATH", newPath); err != nil {
			return fmt.Errorf("更新 PATH 环境变量失败: %v", err)
		}
	}

	if err := os.MkdirAll(downloadFolder, 0755); err != nil {
		return fmt.Errorf("创建下载目录失败: %v", err)
	}

	repos := map[string]string{
		"java": javaCliRepo,
		"go":   goCliRepo,
		"rust": rustCliRepo,
	}

	for lang, repo := range repos {
		cliURL, version, err := getLatestCLIURL(repo, lang)
		if err != nil {
			fmt.Printf("获取 %s CLI URL 失败: %v\n", lang, err)
			continue
		}

		originalPath := filepath.Join(downloadFolder, getBinaryNames()[lang])
		fmt.Printf("正在下载 %s CLI: %s\n", lang, cliURL)

		resp, err := http.Get(cliURL)
		if err != nil {
			fmt.Printf("下载 %s CLI 失败: %v\n", lang, err)
			continue
		}

		func() {
			defer resp.Body.Close()
			out, err := os.Create(originalPath)
			if err != nil {
				fmt.Printf("创建 %s CLI 文件失败: %v\n", lang, err)
				return
			}
			defer out.Close()

			if _, err = io.Copy(out, resp.Body); err != nil {
				fmt.Printf("保存 %s CLI 失败: %v\n", lang, err)
				return
			}
		}()

		if lang == "go" {
			if err := extractGoCliFile(originalPath); err != nil {
				fmt.Printf("解压 Go CLI 失败: %v\n", err)
				continue
			}
		} else {
			finalPath := filepath.Join(downloadFolder, getFinalBinaryName(lang))
			if err := os.Rename(originalPath, finalPath); err != nil {
				fmt.Printf("重命名 %s CLI 失败: %v\n", lang, err)
				continue
			}
		}

		// 设置可执行权限
		if runtime.GOOS != "windows" {
			execPath := filepath.Join(downloadFolder, getFinalBinaryName(lang))
			if err := os.Chmod(execPath, 0755); err != nil {
				fmt.Printf("设置 %s CLI 执行权限失败: %v\n", lang, err)
				continue
			}
		}

		fmt.Printf("下载 %s CLI 版本: %s\n", lang, version)

		if lang == "java" {
			// 为 Java CLI 创建包装脚本
			if err := createJavaCliWrapper(binPath); err != nil {
				fmt.Printf("创建 Java CLI wrapper 失败: %v\n", err)
				continue
			}
		}
	}

	return nil
}

func RefreshCLIHandler(w http.ResponseWriter, r *http.Request) {
	err := downloadCLI()
	if err != nil {
		http.Error(w, "Failed to refresh CLI: "+err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, "CLI updated successfully")
}

func ScheduleCLIUpdater() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		err := downloadCLI()
		if err != nil {
			fmt.Println("Failed to update CLI:", err)
		}
	}
}

func DownloadCLI() error {
	return downloadCLI()
}

func InitCLIDownloader() {
	// 确保在程序启动时下载并设置 PATH
	err := DownloadCLI()
	if err != nil {
		fmt.Printf("初始 CLI 下载失败: %v\n", err)
	}

	// 启动定时更新任务
	go ScheduleCLIUpdater()
}

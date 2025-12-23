// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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

package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type CLIVersionInfo struct {
	Version    string
	BinaryPath string
	BinaryTime time.Time
}

var (
	cliVersionCache = make(map[string]*CLIVersionInfo)
	cliVersionMutex sync.RWMutex
)

// cleanOldMEIFolders cleans up old _MEIXXX folders from the Casdoor temp directory
// that are older than 24 hours. These folders are created by PyInstaller when
// executing casbin-python-cli and can accumulate over time.
func cleanOldMEIFolders() {
	tempDir := "temp"
	cutoffTime := time.Now().Add(-24 * time.Hour)

	entries, err := os.ReadDir(tempDir)
	if err != nil {
		// Log error but don't fail - cleanup is best-effort
		// This is expected if temp directory doesn't exist yet
		return
	}

	for _, entry := range entries {
		// Check if the entry is a directory and matches the _MEI pattern
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "_MEI") {
			continue
		}

		dirPath := filepath.Join(tempDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if the folder is older than 24 hours
		if info.ModTime().Before(cutoffTime) {
			// Try to remove the directory
			err = os.RemoveAll(dirPath)
			if err != nil {
				// Log but continue with other folders
				fmt.Printf("failed to remove old MEI folder %s: %v\n", dirPath, err)
			} else {
				fmt.Printf("removed old MEI folder: %s\n", dirPath)
			}
		}
	}
}

// getCLIVersion
// @Title getCLIVersion
// @Description Get CLI version with cache mechanism
// @Param language string The language of CLI (go/java/rust etc.)
// @Return string The version string of CLI
// @Return error Error if CLI execution fails
func getCLIVersion(language string) (string, error) {
	binaryName := fmt.Sprintf("casbin-%s-cli", language)

	binaryPath, err := exec.LookPath(binaryName)
	if err != nil {
		return "", fmt.Errorf("executable file not found: %v", err)
	}

	fileInfo, err := os.Stat(binaryPath)
	if err != nil {
		return "", fmt.Errorf("failed to get binary info: %v", err)
	}

	cliVersionMutex.RLock()
	if info, exists := cliVersionCache[language]; exists {
		if info.BinaryPath == binaryPath && info.BinaryTime == fileInfo.ModTime() {
			cliVersionMutex.RUnlock()
			return info.Version, nil
		}
	}
	cliVersionMutex.RUnlock()

	// Clean up old _MEI folders before running the command
	cleanOldMEIFolders()

	cmd := exec.Command(binaryName, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get CLI version: %v", err)
	}

	version := strings.TrimSpace(string(output))

	cliVersionMutex.Lock()
	cliVersionCache[language] = &CLIVersionInfo{
		Version:    version,
		BinaryPath: binaryPath,
		BinaryTime: fileInfo.ModTime(),
	}
	cliVersionMutex.Unlock()

	return version, nil
}

func processArgsToTempFiles(args []string) ([]string, []string, error) {
	tempFiles := []string{}
	newArgs := []string{}
	for i := 0; i < len(args); i++ {
		if (args[i] == "-m" || args[i] == "-p") && i+1 < len(args) {
			pattern := fmt.Sprintf("casbin_temp_%s_*.conf", args[i])
			tempFile, err := os.CreateTemp("", pattern)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create temp file: %v", err)
			}

			_, err = tempFile.WriteString(args[i+1])
			if err != nil {
				tempFile.Close()
				return nil, nil, fmt.Errorf("failed to write to temp file: %v", err)
			}

			tempFile.Close()
			tempFiles = append(tempFiles, tempFile.Name())
			newArgs = append(newArgs, args[i], tempFile.Name())
			i++
		} else {
			newArgs = append(newArgs, args[i])
		}
	}
	return tempFiles, newArgs, nil
}

// RunCasbinCommand
// @Title RunCasbinCommand
// @Tag Enforcer API
// @Description Call Casbin CLI commands
// @Success 200 {object} controllers.Response The Response object
// @router /run-casbin-command [get]
func (c *ApiController) RunCasbinCommand() {
	if err := validateIdentifier(c); err != nil {
		c.ResponseError(err.Error())
		return
	}

	language := c.Ctx.Input.Query("language")
	argString := c.Ctx.Input.Query("args")

	if language == "" {
		language = "go"
	}
	// use "casbin-go-cli" by default, can be also "casbin-java-cli", "casbin-node-cli", etc.
	// the pre-built binary of "casbin-go-cli" can be found at: https://github.com/casbin/casbin-go-cli/releases
	binaryName := fmt.Sprintf("casbin-%s-cli", language)

	_, err := exec.LookPath(binaryName)
	if err != nil {
		c.ResponseError(fmt.Sprintf("executable file: %s not found in PATH", binaryName))
		return
	}

	// RBAC model & policy example:
	// https://door.casdoor.com/api/run-casbin-command?language=go&args=["enforce", "-m", "[request_definition]\nr = sub, obj, act\n\n[policy_definition]\np = sub, obj, act\n\n[role_definition]\ng = _, _\n\n[policy_effect]\ne = some(where (p.eft == allow))\n\n[matchers]\nm = g(r.sub, p.sub) %26%26 r.obj == p.obj %26%26 r.act == p.act", "-p", "p, alice, data1, read\np, bob, data2, write\np, data2_admin, data2, read\np, data2_admin, data2, write\ng, alice, data2_admin", "alice", "data1", "read"]
	// Casbin CLI usage:
	// https://github.com/jcasbin/casbin-java-cli?tab=readme-ov-file#get-started
	var args []string
	err = json.Unmarshal([]byte(argString), &args)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Generate cache key for this command
	cacheKey, err := generateCacheKey(language, args)
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Check if result is cached
	if cachedOutput, found := getCachedCommandResult(cacheKey); found {
		c.ResponseOk(cachedOutput)
		return
	}

	if len(args) > 0 && args[0] == "--version" {
		version, err := getCLIVersion(language)
		if err != nil {
			c.ResponseError(err.Error())
			return
		}
		c.ResponseOk(version)
		return
	}

	tempFiles, processedArgs, err := processArgsToTempFiles(args)
	defer func() {
		for _, file := range tempFiles {
			os.Remove(file)
		}
	}()
	if err != nil {
		c.ResponseError(err.Error())
		return
	}

	// Clean up old _MEI folders before running the command
	// This is especially important for Python CLI which creates these folders
	cleanOldMEIFolders()

	command := exec.Command(binaryName, processedArgs...)
	outputBytes, err := command.CombinedOutput()
	if err != nil {
		errorString := err.Error()
		if outputBytes != nil {
			output := string(outputBytes)
			errorString = fmt.Sprintf("%s, error: %s", output, err.Error())
		}

		c.ResponseError(errorString)
		return
	}

	output := string(outputBytes)
	output = strings.TrimSuffix(output, "\n")

	// Store result in cache
	setCachedCommandResult(cacheKey, output)

	c.ResponseOk(output)
}

// validateIdentifier
// @Title validateIdentifier
// @Description Validate the request hash and timestamp
// @Param hash string The SHA-256 hash string
// @Return error Returns error if validation fails, nil if successful
func validateIdentifier(c *ApiController) error {
	language := c.Ctx.Input.Query("language")
	args := c.Ctx.Input.Query("args")
	hash := c.Ctx.Input.Query("m")
	timestamp := c.Ctx.Input.Query("t")

	if hash == "" || timestamp == "" || language == "" || args == "" {
		return fmt.Errorf("invalid identifier")
	}

	requestTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return fmt.Errorf("invalid identifier")
	}
	timeDiff := time.Since(requestTime)
	if timeDiff > 5*time.Minute || timeDiff < -5*time.Minute {
		return fmt.Errorf("invalid identifier")
	}

	params := map[string]string{
		"language": language,
		"args":     args,
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var paramParts []string
	for _, k := range keys {
		paramParts = append(paramParts, fmt.Sprintf("%s=%s", k, params[k]))
	}
	paramString := strings.Join(paramParts, "&")

	version := "casbin-editor-v1"
	rawString := fmt.Sprintf("%s|%s|%s", version, timestamp, paramString)

	hasher := sha256.New()
	hasher.Write([]byte(rawString))

	calculatedHash := strings.ToLower(hex.EncodeToString(hasher.Sum(nil)))
	if calculatedHash != strings.ToLower(hash) {
		return fmt.Errorf("invalid identifier")
	}

	return nil
}

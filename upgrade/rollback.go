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
	"os"
	"path/filepath"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

const (
	ResultSucceeded = "succeeded"
	// the new version is running but has not run long enough to count as started
	StateVerifying   = "verifying"
	ResultRolledBack = "rolledBack"

	// a new version that fails to start this many times (restarted by systemd or another supervisor) is rolled back
	maxStartAttempts = 3
	// a new version that keeps running this long counts as started
	stableUptime = 2 * time.Minute
)

type pendingUpgrade struct {
	From      string `json:"from"`
	To        string `json:"to"`
	OrderId   string `json:"orderId"`
	BackupDir string `json:"backupDir"`
	Time      string `json:"time"`
	Attempts  int    `json:"attempts"`
}

// UpgradeResult is how the last upgrade ended, kept next to the backups and reported to the license server
type UpgradeResult struct {
	From    string `json:"from"`
	To      string `json:"to"`
	OrderId string `json:"orderId"`
	State   string `json:"state"`
	Error   string `json:"error"`
	Time    string `json:"time"`
}

// taken at startup, after an upgrade replaced the binary the running process can no longer resolve its own path
var backupRoot = func() string {
	exePath, err := getExecutablePath()
	if err != nil {
		return "backup"
	}
	return filepath.Join(filepath.Dir(exePath), "backup")
}()

func getBackupRoot() string {
	return backupRoot
}

func getPendingPath() string {
	return filepath.Join(getBackupRoot(), "upgrade-pending.json")
}

func getResultPath() string {
	return filepath.Join(getBackupRoot(), "upgrade-result.json")
}

func writeJson(path string, v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		err = os.WriteFile(path, data, 0o600)
	}
	if err != nil {
		logs.Warn("upgrade: failed to write %s: %s", path, err.Error())
	}
}

func writePendingUpgrade(pending *pendingUpgrade) {
	writeJson(getPendingPath(), pending)
}

func readPendingUpgrade() *pendingUpgrade {
	data, err := os.ReadFile(getPendingPath())
	if err != nil {
		return nil
	}
	pending := &pendingUpgrade{}
	if json.Unmarshal(data, pending) != nil {
		return nil
	}
	return pending
}

func writeResult(pending *pendingUpgrade, state string, message string) {
	writeJson(getResultPath(), &UpgradeResult{
		From:    pending.From,
		To:      pending.To,
		OrderId: pending.OrderId,
		State:   state,
		Error:   message,
		Time:    time.Now().Format(time.RFC3339),
	})
	_ = os.Remove(getPendingPath())
}

func GetLastResult() *UpgradeResult {
	data, err := os.ReadFile(getResultPath())
	if err != nil {
		return nil
	}
	result := &UpgradeResult{}
	if json.Unmarshal(data, result) != nil {
		return nil
	}
	return result
}

// rollBack puts the binary and frontend from the backup of the upgrade back in place
func rollBack(pending *pendingUpgrade) (string, error) {
	exePath, err := getExecutablePath()
	if err != nil {
		return "", err
	}
	backupExe := filepath.Join(pending.BackupDir, filepath.Base(exePath))
	if !util.FileExist(backupExe) {
		return "", fmt.Errorf("the backup %s does not exist", backupExe)
	}

	backupWebDir := filepath.Join(pending.BackupDir, "web", "build")
	webDir := getWebBuildDir(filepath.Dir(exePath))
	if webDir != "" && util.FileExist(filepath.Join(backupWebDir, "index.html")) {
		failedWebDir := webDir + ".failed"
		_ = os.RemoveAll(failedWebDir)
		if err = renameFile(webDir, failedWebDir); err != nil {
			return "", err
		}
		if err = copyDir(backupWebDir, webDir); err != nil {
			_ = renameFile(failedWebDir, webDir)
			return "", err
		}
		_ = os.RemoveAll(failedWebDir)
	}

	restoredExe := exePath + ".rollback"
	if err = copyFile(backupExe, restoredExe); err != nil {
		return "", err
	}
	if err = replaceExecutable(exePath, restoredExe); err != nil {
		return "", err
	}
	return exePath, nil
}

func copyDir(src string, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

// CheckPendingUpgrade runs first thing at startup: a new version that keeps failing to start is rolled back to the
// backup, one that keeps running is recorded as upgraded
func CheckPendingUpgrade() {
	pending := readPendingUpgrade()
	if pending == nil {
		return
	}

	if util.Version != pending.To {
		state := StateFailed
		if util.Version == pending.From {
			state = ResultRolledBack
		}
		writeResult(pending, state, fmt.Sprintf("%s is running instead of %s", util.Version, pending.To))
		return
	}

	pending.Attempts++
	if pending.Attempts > maxStartAttempts {
		logs.Error("upgrade: %s failed to start %d times, rolling back to %s", pending.To, maxStartAttempts, pending.From)
		exePath, err := rollBack(pending)
		if err != nil {
			logs.Error("upgrade: failed to roll back to %s: %s", pending.From, err.Error())
			writeResult(pending, StateFailed, fmt.Sprintf("%s failed to start and rolling back failed: %s", pending.To, err.Error()))
			return
		}
		writeResult(pending, ResultRolledBack, fmt.Sprintf("%s failed to start %d times", pending.To, maxStartAttempts))
		if err = restart(exePath); err != nil {
			logs.Error("upgrade: failed to start %s after rolling back: %s", pending.From, err.Error())
		}
		return
	}
	writePendingUpgrade(pending)
	setStatus(func(status *UpgradeStatus) {
		status.State = StateVerifying
		status.Version = pending.To
		status.OrderId = pending.OrderId
	})

	util.SafeGoroutine(func() {
		time.Sleep(stableUptime)
		writeResult(pending, ResultSucceeded, "")
		setStatus(func(status *UpgradeStatus) {
			if status.State == StateVerifying {
				*status = UpgradeStatus{}
			}
		})
		logs.Info("upgrade: %s has been running for %s, the upgrade from %s succeeded", pending.To, stableUptime, pending.From)
	})
}

// GetReportState is what the license server learns about upgrades: the one in progress and how the last one ended
func GetReportState() interface{} {
	return map[string]interface{}{
		"status":     getStatus(),
		"lastResult": GetLastResult(),
		"deployment": GetDeploymentType(),
	}
}

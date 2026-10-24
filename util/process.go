// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"encoding/csv"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func getPidByPort(port int) (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("netstat", "-ano")
	case "darwin", "linux":
		cmd = exec.Command("lsof", "-t", "-nP", "-iTCP:"+strconv.Itoa(port), "-sTCP:LISTEN")
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return 0, nil
			}
		}

		return 0, nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			if runtime.GOOS == "windows" {
				// a listening socket has no remote port, unlike the state column this does not depend on the Windows language
				if len(fields) >= 5 && fields[0] == "TCP" && strings.HasSuffix(fields[1], ":"+strconv.Itoa(port)) && strings.HasSuffix(fields[2], ":0") {
					pid, err := strconv.Atoi(fields[len(fields)-1])
					if err != nil {
						return 0, err
					}

					return pid, nil
				}
			} else if len(fields) == 1 {
				// BusyBox lsof ignores all options and lists every open file, only "lsof -t" prints bare PIDs
				pid, err := strconv.Atoi(fields[0])
				if err != nil {
					return 0, err
				}

				return pid, nil
			}
		}
	}

	return 0, nil
}

func getProcessPath(pid int) (string, error) {
	switch runtime.GOOS {
	case "windows":
		output, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
		if err != nil {
			return "", err
		}

		record, err := csv.NewReader(strings.NewReader(string(output))).Read()
		if err != nil {
			return "", err
		}
		if len(record) < 2 {
			return "", fmt.Errorf("process %d not found", pid)
		}

		return record[0], nil
	case "darwin":
		output, err := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)).Output()
		if err != nil {
			return "", err
		}

		return strings.TrimSpace(string(output)), nil
	case "linux":
		path, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
		if err != nil {
			return "", err
		}

		return strings.TrimSuffix(path, " (deleted)"), nil
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func isCasdoorProcess(path string) bool {
	name := strings.ToLower(filepath.Base(path))
	if strings.Contains(name, "casdoor") {
		return true
	}

	exePath, err := os.Executable()
	if err != nil {
		return false
	}

	return name == strings.ToLower(filepath.Base(exePath))
}

func StopOldInstance(port int) error {
	pid, err := getPidByPort(port)
	if err != nil {
		return err
	}
	if pid == 0 {
		return nil
	}

	path, err := getProcessPath(pid)
	if err != nil {
		return fmt.Errorf("port %d is in use by the process with pid: %d, which could not be identified: %v", port, pid, err)
	}
	if !isCasdoorProcess(path) {
		return fmt.Errorf("port %d is in use by another program: %s (pid: %d), please stop it or change httpport", port, path, pid)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	err = process.Kill()
	if err != nil {
		return err
	} else {
		fmt.Printf("The old instance with pid: %d has been stopped\n", pid)
	}

	return nil
}

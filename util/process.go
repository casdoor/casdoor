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
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

func getPidByPort(port int) (int, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "netstat -ano | findstr :"+strconv.Itoa(port))
	case "darwin", "linux":
		cmd = exec.Command("lsof", "-t", "-i", ":"+strconv.Itoa(port))
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return 0, nil
			}
		} else {
			return 0, err
		}
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			if runtime.GOOS == "windows" {
				if fields[1] == "0.0.0.0:"+strconv.Itoa(port) {
					pid, err := strconv.Atoi(fields[len(fields)-1])
					if err != nil {
						return 0, err
					}

					return pid, nil
				}
			} else {
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

func StopOldInstance(port int) error {
	pid, err := getPidByPort(port)
	if err != nil {
		return err
	}
	if pid == 0 {
		return nil
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

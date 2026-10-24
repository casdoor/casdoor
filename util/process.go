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
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

func getListenerPids(port int) ([]int32, error) {
	conns, err := net.Connections("tcp")
	if err != nil {
		return nil, err
	}

	pids := []int32{}
	for _, conn := range conns {
		if conn.Status != "LISTEN" || conn.Laddr.Port != uint32(port) {
			continue
		}

		if !slices.Contains(pids, conn.Pid) {
			pids = append(pids, conn.Pid)
		}
	}

	return pids, nil
}

func getExeName(path string) string {
	name := strings.ToLower(filepath.Base(strings.TrimSuffix(path, " (deleted)")))
	return strings.TrimSuffix(name, ".exe")
}

func isOldInstance(p *process.Process) (bool, string) {
	exe, err := p.Exe()
	if err != nil || exe == "" {
		exe, err = p.Name()
		if err != nil {
			return false, ""
		}
	}

	name := getExeName(exe)
	if strings.Contains(name, "casdoor") {
		return true, exe
	}

	selfExe, err := os.Executable()
	if err != nil {
		return false, exe
	}

	return name == getExeName(selfExe), exe
}

func StopOldInstance(port int) error {
	pids, err := getListenerPids(port)
	if err != nil {
		fmt.Printf("StopOldInstance() error: failed to list the listeners of port %d: %s\n", port, err.Error())
		return nil
	}

	stopped := false
	for _, pid := range pids {
		if pid == int32(os.Getpid()) {
			continue
		}
		if pid == 0 {
			return fmt.Errorf("port %d is in use by another program, please stop it or change httpport", port)
		}

		p, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		ok, exe := isOldInstance(p)
		if !ok {
			return fmt.Errorf("port %d is in use by another program: %s (pid: %d), please stop it or change httpport", port, exe, pid)
		}

		err = p.Kill()
		if err != nil {
			return err
		}

		stopped = true
		fmt.Printf("The old instance with pid: %d has been stopped\n", pid)
	}

	for i := 0; stopped && i < 50; i++ {
		pids, err = getListenerPids(port)
		if err != nil || len(pids) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	return nil
}

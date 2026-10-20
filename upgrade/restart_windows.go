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

//go:build windows

package upgrade

import (
	"os"
	"os/exec"
	"syscall"
)

const createNewConsole = 0x00000010

// restart starts the new binary in its own console, so that it is not killed when the console of this process
// closes, and exits, the new process stops whatever still listens on the port in util.StopOldInstance()
func restart(exePath string) error {
	cmd := exec.Command(exePath, os.Args[1:]...)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNewConsole}
	err := cmd.Start()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

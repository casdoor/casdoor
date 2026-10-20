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

//go:build !windows

package upgrade

import (
	"os"
	"syscall"
)

// restart replaces this process with the new binary in place, keeping the PID so that systemd and other supervisors
// do not notice, Go opens every file and socket with close-on-exec so the port is free for the new process
func restart(exePath string) error {
	return syscall.Exec(exePath, os.Args, os.Environ())
}

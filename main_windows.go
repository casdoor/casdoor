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

package main

import (
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
)

type windowsService struct{}

func (s *windowsService) Execute(args []string, requests <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	changes <- svc.Status{State: svc.StartPending}
	go runServer()
	changes <- svc.Status{State: svc.Running, Accepts: svc.AcceptStop | svc.AcceptShutdown}

	for request := range requests {
		switch request.Cmd {
		case svc.Interrogate:
			changes <- request.CurrentStatus
		case svc.Stop, svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			return false, 0
		}
	}

	return false, 0
}

func runAsService() bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		panic(err)
	}
	if !isService {
		return false
	}

	// Windows services start in C:\Windows\System32, while conf/, web/build/ and logs/ are relative paths
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	err = os.Chdir(filepath.Dir(exePath))
	if err != nil {
		panic(err)
	}

	err = svc.Run("casdoor", &windowsService{})
	if err != nil {
		panic(err)
	}

	return true
}

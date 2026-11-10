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

package hsm

import (
	"fmt"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows"
)

// a C long is 32 bits on Windows, and the PKCS #11 structs are packed to 1 byte
const (
	ulongSize     = 4
	packedStructs = true
)

type library struct {
	path   string
	handle windows.Handle
}

func openLibrary(path string) (*library, error) {
	var flags uintptr
	if filepath.IsAbs(path) {
		// the vendor DLLs next to the module are found too
		flags = windows.LOAD_WITH_ALTERED_SEARCH_PATH
	}
	handle, err := windows.LoadLibraryEx(path, 0, flags)
	if err != nil {
		return nil, fmt.Errorf("failed to load the HSM library %s: %v", path, err)
	}
	return &library{path: path, handle: handle}, nil
}

func (l *library) proc(name string) (uintptr, error) {
	proc, err := windows.GetProcAddress(l.handle, name)
	if err != nil {
		return 0, fmt.Errorf("the HSM library %s has no function %s: %v", l.path, name, err)
	}
	return proc, nil
}

func call(fn uintptr, args ...uintptr) uintptr {
	res, _, _ := syscall.SyscallN(fn, args...)
	return res
}

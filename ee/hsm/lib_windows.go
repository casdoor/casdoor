// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

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

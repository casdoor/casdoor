// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

//go:build !windows && !darwin && !(hsm && linux && (amd64 || arm64 || loong64))

package hsm

import (
	"fmt"
	"runtime"
	"unsafe"
)

const (
	ulongSize     = int(unsafe.Sizeof(uintptr(0)))
	packedStructs = false
)

type library struct{}

// Loading a C library without cgo links the Linux binary to glibc dynamically, so the static builds leave it out
func openLibrary(path string) (*library, error) {
	return nil, fmt.Errorf("this Casdoor build can't load the HSM library %s: HSM support on %s/%s needs a build with \"-tags hsm\" (it links to glibc)", path, runtime.GOOS, runtime.GOARCH)
}

func (l *library) proc(name string) (uintptr, error) {
	return 0, fmt.Errorf("no HSM support in this build")
}

func call(fn uintptr, args ...uintptr) uintptr {
	return 0
}

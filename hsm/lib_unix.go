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

//go:build (hsm && linux && (amd64 || arm64 || loong64)) || darwin

package hsm

import (
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
)

// a C long is as wide as a pointer on the 64-bit Unix systems, and the structs have their natural alignment
const (
	ulongSize     = int(unsafe.Sizeof(uintptr(0)))
	packedStructs = false
)

type library struct {
	path   string
	handle uintptr
}

func openLibrary(path string) (*library, error) {
	handle, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_LOCAL)
	if err != nil {
		return nil, fmt.Errorf("failed to load the HSM library %s: %v", path, err)
	}
	return &library{path: path, handle: handle}, nil
}

func (l *library) proc(name string) (uintptr, error) {
	proc, err := purego.Dlsym(l.handle, name)
	if err != nil {
		return 0, fmt.Errorf("the HSM library %s has no function %s: %v", l.path, name, err)
	}
	return proc, nil
}

func call(fn uintptr, args ...uintptr) uintptr {
	res, _, _ := purego.SyscallN(fn, args...)
	return res
}

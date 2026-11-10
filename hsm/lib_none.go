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

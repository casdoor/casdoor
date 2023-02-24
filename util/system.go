// Copyright 2022 The Casdoor Authors. All Rights Reserved.
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
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// get cpu usage
func GetCpuUsage() ([]float64, error) {
	usage, err := cpu.Percent(time.Second, true)
	return usage, err
}

// get memory usage
func GetMemoryUsage() (uint64, uint64, error) {
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m.TotalAlloc, virtualMem.Total, nil
}

// get github repo release version
func GetGitRepoCommit() (string, error) {
	var fileDate, commit string
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fileInfos, err := ioutil.ReadDir(pwd + "/.git/refs/heads")
	for _, v := range fileInfos {
		if v.Name() == "master" {
			if v.ModTime().String() == fileDate {
				return commit, nil
			} else {
				fileDate = v.ModTime().String()
				break
			}
		}
	}

	content, err := ioutil.ReadFile(pwd + "/.git/refs/heads/master")
	if err != nil {
		return "", err
	}

	// Convert to full length
	temp := string(content)
	commit = strings.ReplaceAll(temp, "\n", "")

	return commit, nil
}

func GetVersionFromCommit(commit string) (string, error) {
	var version string
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	TagFileInfos, err := ioutil.ReadDir(pwd + "/.git/refs/tags")
	if err != nil {
		return "", err
	}

	for _, tagFile := range TagFileInfos {
		content, err := ioutil.ReadFile(pwd + "/.git/refs/tags/" + tagFile.Name())
		if err != nil {
			return "", err
		}
		temp := string(content)
		tagCommit := strings.ReplaceAll(temp, "\n", "")
		if tagCommit == commit {
			version = tagFile.Name()
			return version, nil
		}
	}
	return "", nil
}

func GetCurBranchCommit() (string, error) {
	var fileDate, commit, branchPath string
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fileInfos, err := ioutil.ReadDir(pwd + "/.git")
	for _, v := range fileInfos {
		if v.Name() == "HEAD" {
			if v.ModTime().String() == fileDate {
				return commit, nil
			} else {
				fileDate = v.ModTime().String()
				break
			}
		}
	}

	path, err := ioutil.ReadFile(pwd + "/.git/HEAD")
	if err != nil {
		return "", err
	}

	// Convert to full length
	temp := strings.ReplaceAll(string(path), "\n", "")
	branchPath = temp[5:]

	content, err := ioutil.ReadFile(pwd + "/.git/" + branchPath)
	if err != nil {
		return "", err
	}
	commit = strings.ReplaceAll(string(content), "\n", "")

	return commit, nil
}

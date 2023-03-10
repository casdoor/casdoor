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
	"runtime"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type SystemInfo struct {
	CpuUsage    []float64 `json:"cpuUsage"`
	MemoryUsed  uint64    `json:"memoryUsed"`
	MemoryTotal uint64    `json:"memoryTotal"`
}

type VersionInfo struct {
	Version  string `json:"version"`
	CommitId string `json:"commitId"`
	Desc     string `json:"desc"`
}

// getCpuUsage get cpu usage
func getCpuUsage() ([]float64, error) {
	usage, err := cpu.Percent(time.Second, true)
	return usage, err
}

// getMemoryUsage get memory usage
func getMemoryUsage() (uint64, uint64, error) {
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return m.TotalAlloc, virtualMem.Total, nil
}

func GetSystemInfo() (*SystemInfo, error) {
	cpuUsage, err := getCpuUsage()
	if err != nil {
		return nil, err
	}

	memoryUsed, memoryTotal, err := getMemoryUsage()
	if err != nil {
		return nil, err
	}

	res := &SystemInfo{
		CpuUsage:    cpuUsage,
		MemoryUsed:  memoryUsed,
		MemoryTotal: memoryTotal,
	}
	return res, nil
}

var Commit = ""
var Version = ""
var Desc = ""

// GetVersionInfo get git current commit and repo release version
func GetVersionInfo() (*VersionInfo, error) {
	return &VersionInfo{
		Version:  Version,
		CommitId: Commit,
		Desc:     Desc,
	}, nil
}

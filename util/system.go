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
	"os"
	"path"
	"path/filepath"
	"runtime"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/process"
)

type SystemInfo struct {
	CpuUsage     []float64 `json:"cpuUsage"`
	MemoryUsed   uint64    `json:"memoryUsed"`
	MemoryTotal  uint64    `json:"memoryTotal"`
	DiskUsed     uint64    `json:"diskUsed"`
	DiskTotal    uint64    `json:"diskTotal"`
	NetworkSent  uint64    `json:"networkSent"`
	NetworkRecv  uint64    `json:"networkRecv"`
	NetworkTotal uint64    `json:"networkTotal"`
}

type VersionInfo struct {
	Version      string `json:"version"`
	CommitId     string `json:"commitId"`
	CommitOffset int    `json:"commitOffset"`
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

	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, 0, err
	}

	memInfo, err := proc.MemoryInfo()
	if err != nil {
		return 0, 0, err
	}

	return memInfo.RSS, virtualMem.Total, nil
}

// getDiskUsage gets disk usage for Casdoor's data directory
func getDiskUsage() (uint64, uint64, error) {
	_, filename, _, _ := runtime.Caller(0)
	rootPath := path.Dir(path.Dir(filename))
	dataPath := filepath.Join(rootPath, "data")

	var size uint64
	err := filepath.Walk(dataPath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}

	diskStat, err := disk.Usage(dataPath)
	if err != nil {
		diskStat, err = disk.Usage("/")
		if err != nil {
			return 0, 0, err
		}
	}

	return size, diskStat.Total, nil
}

// getNetworkUsage gets Casdoor process's own I/O usage
func getNetworkUsage() (uint64, uint64, uint64, error) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return 0, 0, 0, err
	}

	ioCounters, err := proc.IOCounters()
	if err != nil {
		return 0, 0, 0, err
	}

	bytesSent := ioCounters.WriteBytes
	bytesRecv := ioCounters.ReadBytes
	bytesTotal := bytesSent + bytesRecv

	return bytesSent, bytesRecv, bytesTotal, nil
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

	diskUsed, diskTotal, err := getDiskUsage()
	if err != nil {
		return nil, err
	}

	networkSent, networkRecv, networkTotal, err := getNetworkUsage()
	if err != nil {
		return nil, err
	}

	res := &SystemInfo{
		CpuUsage:     cpuUsage,
		MemoryUsed:   memoryUsed,
		MemoryTotal:  memoryTotal,
		DiskUsed:     diskUsed,
		DiskTotal:    diskTotal,
		NetworkSent:  networkSent,
		NetworkRecv:  networkRecv,
		NetworkTotal: networkTotal,
	}
	return res, nil
}

// GetVersionInfo get git current commit and repo release version
func GetVersionInfo() (*VersionInfo, error) {
	res := &VersionInfo{
		Version:      "",
		CommitId:     "",
		CommitOffset: -1,
	}

	_, filename, _, _ := runtime.Caller(0)
	rootPath := path.Dir(path.Dir(filename))
	r, err := git.PlainOpen(rootPath)
	if err != nil {
		return res, err
	}
	ref, err := r.Head()
	if err != nil {
		return res, err
	}
	tags, err := r.Tags()
	if err != nil {
		return res, err
	}
	tagMap := make(map[plumbing.Hash]string)
	err = tags.ForEach(func(t *plumbing.Reference) error {
		// This technique should work for both lightweight and annotated tags.
		revHash, err := r.ResolveRevision(plumbing.Revision(t.Name()))
		if err != nil {
			return err
		}
		tagMap[*revHash] = t.Name().Short()
		return nil
	})
	if err != nil {
		return res, err
	}

	cIter, err := r.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return res, err
	}

	commitOffset := 0
	version := ""
	// iterates over the commits
	err = cIter.ForEach(func(c *object.Commit) error {
		tag, ok := tagMap[c.Hash]
		if ok {
			if version == "" {
				version = tag
			}
		}
		if version == "" {
			commitOffset++
		}
		return nil
	})
	if err != nil {
		return res, err
	}

	res = &VersionInfo{
		Version:      version,
		CommitId:     ref.Hash().String(),
		CommitOffset: commitOffset,
	}
	return res, nil
}

func GetBuiltInVersionInfo() *VersionInfo {
	return &VersionInfo{
		Version:      Version,
		CommitId:     CommitId,
		CommitOffset: CommitOffset,
	}
}

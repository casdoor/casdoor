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
	"bufio"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
)

type SystemInfo struct {
	CpuUsage    []float64 `json:"cpuUsage"`
	MemoryUsed  uint64    `json:"memoryUsed"`
	MemoryTotal uint64    `json:"memoryTotal"`
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

func GetVersionInfoFromFile() (*VersionInfo, error) {
	res := &VersionInfo{
		Version:      "",
		CommitId:     "",
		CommitOffset: -1,
	}

	_, filename, _, _ := runtime.Caller(0)
	rootPath := path.Dir(path.Dir(filename))
	file, err := os.Open(filepath.Clean(path.Join(rootPath, "version_info.txt")))
	if err != nil {
		return res, err
	}
	defer file.Close()

	// Read file contents line by line
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Use regular expressions to match strings
		re := regexp.MustCompile(`\{([^{}]+)\}`)
		versionInfo := scanner.Text()
		matches := re.FindStringSubmatch(versionInfo)
		if len(matches) > 1 {
			split := strings.Split(matches[1], " ")
			version := split[0]
			commitId := split[1]
			commitOffset, _ := strconv.Atoi(split[2])
			res = &VersionInfo{
				Version:      version,
				CommitId:     commitId,
				CommitOffset: commitOffset,
			}
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return res, err
	}

	return res, nil
}

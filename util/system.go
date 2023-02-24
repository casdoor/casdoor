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
	"encoding/json"
	"github.com/go-resty/resty/v2"
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

type TagInfo struct {
	TagName string `json:"tag_name"`
	Commit  string `json:"body"`
}

// get github repo release version
func GetVersionInfo() ([]*TagInfo, error) {
	httpClient := resty.New()
	req := httpClient.R()
	req.Method = "GET"
	req.URL = "https://api.github.com/repos/casdoor/casdoor/releases"
	resp, err := req.Execute(req.Method, req.URL)
	if err != nil || resp.StatusCode() != 200 {
		return nil, err
	}

	var tags []*TagInfo
	if err := json.Unmarshal(resp.Body(), &tags); err != nil {
		return nil, err
	}
	for _, tag := range tags {
		if len(tag.Commit) < 50 {
			continue
		}
		tagCommit := tag.Commit[len(tag.Commit)-46 : len(tag.Commit)-6]
		tag.Commit = tagCommit
	}
	return tags, nil
}
func GetRepoVersion() (string, string, string, error) {
	var branchPath, commit string
	pwd, err := os.Getwd()
	if err != nil {
		return "", "", "", err
	}

	path, err := ioutil.ReadFile(pwd + "/.git/HEAD")
	if err != nil {
		return "", "", "", err
	}

	// Convert to full length
	temp := strings.ReplaceAll(string(path), "\n", "")
	branchPath = temp[5:]

	content, err := ioutil.ReadFile(pwd + "/.git/logs/" + branchPath)
	if err != nil {
		return "", "", "", err
	}
	logs := strings.Split(string(content), "\n")

	tags, err := GetVersionInfo()
	if err != nil {
		return "", "", "", err
	}
	version, author, curcommit := "", "", ""
	for i := len(logs) - 2; i >= 0; i-- {
		tmp := strings.Split(logs[i], " ")
		curcommit = tmp[1]
		if commit == "" {
			commit = tmp[1]
		}
		if author == "" {
			author = tmp[2]
		}
		for _, tag := range tags {
			if tag.Commit == curcommit {
				version = tag.TagName
			}
		}
		if version != "" {
			break
		}
	}
	if version == tags[0].TagName || version == "" {
		commit = ""
	}
	return author, commit, version, err
}

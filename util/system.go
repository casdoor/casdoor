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
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"encoding/json"
	"github.com/go-resty/resty/v2"
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
	TagName       string `json:"tag_name"`
	CommitMessage string `json:"body"`
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
	return tags, nil
}
func GetLatestVersion() (string, error) {
	tags, err := GetVersionInfo()
	if err != nil {
		return "", err
	}
	if len(tags) > 0 {
		return tags[0].TagName, nil
	} else {
		return "", nil
	}
}
func GetBasedonVersion() (string, error) {
	materCommit, err := GetMasterCommit()
	if err != nil {
		return "", err
	}

	tags, err := GetVersionInfo()
	if err != nil {
		return "", err
	}

	for _, tag := range tags {
		if len(tag.CommitMessage) < 50 {
			continue
		}
		tagcommit := tag.CommitMessage[len(tag.CommitMessage)-46 : len(tag.CommitMessage)-6]
		log.Println(tagcommit)
		if tagcommit != materCommit {
			continue
		}
		return tag.TagName, nil
	}
	return "", nil
}
func GetMasterCommit() (string, error) {
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
	//在主分支直接返空
	if len(branchPath) > 6 && branchPath[len(branchPath)-6:] == "master" {
		return "", nil
	}
	content, err := ioutil.ReadFile(pwd + "/.git/" + branchPath)
	if err != nil {
		return "", err
	}
	commit = strings.ReplaceAll(string(content), "\n", "")

	return commit, nil
}

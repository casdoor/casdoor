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

//go:build !skipCi
// +build !skipCi

package util

import (
	"path"
	"runtime"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestGetCpuUsage(t *testing.T) {
	usage, err := getCpuUsage()
	assert.Nil(t, err)
	t.Log(usage)
}

func TestGetMemoryUsage(t *testing.T) {
	used, total, err := getMemoryUsage()
	assert.Nil(t, err)
	t.Log(used, total)
}

func TestGetGitRepoVersion(t *testing.T) {
	versionInfo, err := GetVersionInfo()
	assert.Nil(t, err)
	t.Log(versionInfo)
}

func TestGetVersion(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := path.Dir(path.Dir(filename))
	r, err := git.PlainOpen(root)
	if err != nil {
		t.Log(err)
	}
	tags, err := r.Tags()
	if err != nil {
		t.Log(err)
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

	testHash := plumbing.NewHash("f8bc87eb4e5ba3256424cf14aafe0549f812f1cf")
	cIter, err := r.Log(&git.LogOptions{From: testHash})

	aheadCnt := 0
	releaseVersion := ""
	// iterates over the commits
	err = cIter.ForEach(func(c *object.Commit) error {
		tag, ok := tagMap[c.Hash]
		if ok {
			if releaseVersion == "" {
				releaseVersion = tag
			}
		}
		if releaseVersion == "" {
			aheadCnt++
		}
		return nil
	})

	assert.Equal(t, 3, aheadCnt)
	assert.Equal(t, "v1.257.0", releaseVersion)
}

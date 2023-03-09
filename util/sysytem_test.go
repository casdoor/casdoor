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
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestGetCpuUsage(t *testing.T) {
	usage, err := GetCpuUsage()
	assert.Nil(t, err)
	t.Log(usage)
}

func TestGetMemoryUsage(t *testing.T) {
	used, total, err := GetMemoryUsage()
	assert.Nil(t, err)
	t.Log(used, total)
}

func TestGetGitRepoVersion(t *testing.T) {
	commit, version, err := GetGitRepoVersion()
	assert.Nil(t, err)
	t.Log(commit, version)
}

func TestGetVersion(t *testing.T) {
	r, err := git.PlainOpen("..")
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

	testHash := plumbing.NewHash("16b1d0e1f001c1162a263ed50c1a892c947d5783")
	cIter, err := r.Log(&git.LogOptions{From: testHash})

	releaseVersion := ""
	// iterates over the commits
	err = cIter.ForEach(func(c *object.Commit) error {
		tag, ok := tagMap[c.Hash]
		if ok {
			if releaseVersion == "" {
				releaseVersion = tag
			}
		}
		return nil
	})
	assert.Equal(t, "v1.260.0", releaseVersion)
}

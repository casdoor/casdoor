// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/casdoor/oss"
)

// LocalFileSystemProvider file system storage
type LocalFileSystemProvider struct {
	BaseDir string
}

// NewLocalFileSystemStorageProvider initialize the local file system storage
func NewLocalFileSystemStorageProvider() *LocalFileSystemProvider {
	baseFolder := "files"
	absBase, err := filepath.Abs(baseFolder)
	if err != nil {
		panic(err)
	}

	return &LocalFileSystemProvider{BaseDir: absBase}
}

// GetFullPath get full path from absolute/relative path
func (sp LocalFileSystemProvider) GetFullPath(path string) string {
	fullPath := path
	if !strings.HasPrefix(path, sp.BaseDir) {
		fullPath, _ = filepath.Abs(filepath.Join(sp.BaseDir, path))
	}
	return fullPath
}

// Get receive file with given path
func (sp LocalFileSystemProvider) Get(path string) (*os.File, error) {
	return os.Open(sp.GetFullPath(path))
}

// GetStream get file as stream
func (sp LocalFileSystemProvider) GetStream(path string) (io.ReadCloser, error) {
	return os.Open(sp.GetFullPath(path))
}

// Put store a reader into given path
func (sp LocalFileSystemProvider) Put(path string, reader io.Reader) (*oss.Object, error) {
	fullPath := sp.GetFullPath(path)

	err := os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Casdoor fails to create folder: \"%s\" for local file system storage provider: %s. Make sure Casdoor process has correct permission to create/access it, or you can create it manually in advance", filepath.Dir(fullPath), err.Error())
	}

	dst, err := os.Create(filepath.Clean(fullPath))
	if err == nil {
		defer dst.Close()
		if seeker, ok := reader.(io.ReadSeeker); ok {
			seeker.Seek(0, 0)
		}
		_, err = io.Copy(dst, reader)
	}
	return &oss.Object{Path: path, Name: filepath.Base(path), StorageInterface: sp}, err
}

// Delete delete file
func (sp LocalFileSystemProvider) Delete(path string) error {
	return os.Remove(sp.GetFullPath(path))
}

// List list all objects under current path
func (sp LocalFileSystemProvider) List(path string) ([]*oss.Object, error) {
	objects := []*oss.Object{}
	fullPath := sp.GetFullPath(path)

	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if path == fullPath {
			return nil
		}

		if err == nil && !info.IsDir() {
			modTime := info.ModTime()
			objects = append(objects, &oss.Object{
				Path:             strings.TrimPrefix(path, sp.BaseDir),
				Name:             info.Name(),
				LastModified:     &modTime,
				StorageInterface: sp,
			})
		}
		return nil
	})

	return objects, nil
}

// GetEndpoint get endpoint, LocalFileSystemProvider's endpoint is /
func (sp LocalFileSystemProvider) GetEndpoint() string {
	return "/"
}

// GetURL get public accessible URL
func (sp LocalFileSystemProvider) GetURL(path string) (url string, err error) {
	return path, nil
}

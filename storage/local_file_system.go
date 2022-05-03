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
	"github.com/casdoor/oss"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var baseFolder = "files"

// FileSystem file system storage
type FileSystem struct {
	Base string
}

// NewFileSystem initialize the local file system storage
func NewFileSystem(base string) *FileSystem {
	absBase, err := filepath.Abs(base)
	if err != nil {
		panic("local file system storage's base folder is not initialized")
	}

	return &FileSystem{Base: absBase}
}

// GetFullPath get full path from absolute/relative path
func (fileSystem FileSystem) GetFullPath(path string) string {
	fullPath := path
	if !strings.HasPrefix(path, fileSystem.Base) {
		fullPath, _ = filepath.Abs(filepath.Join(fileSystem.Base, path))
	}
	return fullPath
}

// Get receive file with given path
func (fileSystem FileSystem) Get(path string) (*os.File, error) {
	return os.Open(fileSystem.GetFullPath(path))
}

// GetStream get file as stream
func (fileSystem FileSystem) GetStream(path string) (io.ReadCloser, error) {
	return os.Open(fileSystem.GetFullPath(path))
}

// Put store a reader into given path
func (fileSystem FileSystem) Put(path string, reader io.Reader) (*oss.Object, error) {
	var (
		fullPath = fileSystem.GetFullPath(path)
		err      = os.MkdirAll(filepath.Dir(fullPath), os.ModePerm)
	)

	if err != nil {
		return nil, err
	}

	dst, err := os.Create(fullPath)

	if err == nil {
		if seeker, ok := reader.(io.ReadSeeker); ok {
			seeker.Seek(0, 0)
		}
		_, err = io.Copy(dst, reader)
	}

	return &oss.Object{Path: path, Name: filepath.Base(path), StorageInterface: fileSystem}, err
}

// Delete delete file
func (fileSystem FileSystem) Delete(path string) error {
	return os.Remove(fileSystem.GetFullPath(path))
}

// List list all objects under current path
func (fileSystem FileSystem) List(path string) ([]*oss.Object, error) {
	var (
		objects  []*oss.Object
		fullPath = fileSystem.GetFullPath(path)
	)

	filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if path == fullPath {
			return nil
		}

		if err == nil && !info.IsDir() {
			modTime := info.ModTime()
			objects = append(objects, &oss.Object{
				Path:             strings.TrimPrefix(path, fileSystem.Base),
				Name:             info.Name(),
				LastModified:     &modTime,
				StorageInterface: fileSystem,
			})
		}
		return nil
	})

	return objects, nil
}

// GetEndpoint get endpoint, FileSystem's endpoint is /
func (fileSystem FileSystem) GetEndpoint() string {
	return "/"
}

// GetURL get public accessible URL
func (fileSystem FileSystem) GetURL(path string) (url string, err error) {
	return path, nil
}

func NewLocalFileSystemStorageProvider(clientId string, clientSecret string, region string, bucket string, endpoint string) oss.StorageInterface {
	return NewFileSystem(baseFolder)
}

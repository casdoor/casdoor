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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func MergeFolderFiles(folder string) (*bytes.Buffer, error) {
	buf := bytes.NewBuffer(nil)
	filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		fp, fpOpenErr := os.Open(path)
		if fpOpenErr != nil {
			fmt.Printf("Can not open file %v", fpOpenErr)
			return fpOpenErr
		}
		bReader := bufio.NewReader(fp)
		for {
			buffer := make([]byte, 1024)
			readCount, readErr := bReader.Read(buffer)
			if readErr == io.EOF {
				break
			} else {
				buf.Write(buffer[:readCount])
			}
		}
		return err
	})

	return buf, nil
}

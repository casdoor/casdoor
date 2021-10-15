// Copyright 2021 The casbin Authors. All Rights Reserved.
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
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func ParseInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func GetOwnerAndNameFromId(id string) (string, string) {
	tokens := strings.Split(id, "/")
	if len(tokens) != 2 {
		panic(errors.New("GetOwnerAndNameFromId() error, wrong token count for ID: " + id))
	}

	return tokens[0], tokens[1]
}

func GetOwnerAndNameFromIdNoCheck(id string) (string, string) {
	tokens := strings.SplitN(id, "/", 2)
	return tokens[0], tokens[1]
}

func GenerateId() string {
	return uuid.NewString()
}

func GetId(name string) string {
	return fmt.Sprintf("admin/%s", name)
}

func GetMd5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func IsStrsEmpty(strs ...string) bool {
	for _, str := range strs {
		if len(str) == 0 {
			return true
		}
	}
	return false
}

func GetMaxLenStr(strs ...string) string {
	m := 0
	i := 0
	for j, str := range strs {
		l := len(str)
		if l > m {
			m = l
			i = j
		}
	}
	return strs[i]
}

func GetMinLenStr(strs ...string) string {
	m := int(^uint(0) >> 1)
	i := 0
	for j, str := range strs {
		l := len(str)
		if l > m {
			m = l
			i = j
		}
	}
	return strs[i]
}

func ReadStringFromPath(path string) string {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return string(data)
}

func WriteStringToPath(s string, path string) {
	err := ioutil.WriteFile(path, []byte(s), 0644)
	if err != nil {
		panic(err)
	}
}

func ReadBytesFromPath(path string) []byte {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}

	return data
}

func WriteBytesToPath(b []byte, path string) {
	err := ioutil.WriteFile(path, b, 0644)
	if err != nil {
		panic(err)
	}
}

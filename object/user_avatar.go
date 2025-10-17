// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

package object

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/v2/proxy"
)

func downloadImage(client *http.Client, url string) (*bytes.Buffer, string, error) {
	// Download the image
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("downloadImage() error for url [%s]: %s\n", url, err.Error())
		return nil, "", nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("downloadImage() error for url [%s]: %s\n", url, resp.Status)
		return nil, "", nil
	}

	// Get the content type and determine the file extension
	contentType := resp.Header.Get("Content-Type")
	fileExtension := ""

	if strings.Contains(contentType, "text/html") {
		fileExtension = ".html"
	} else if contentType == "image/vnd.microsoft.icon" {
		fileExtension = ".ico"
	} else {
		fileExtensions, err := mime.ExtensionsByType(contentType)
		if err != nil {
			return nil, "", err
		}
		if fileExtensions == nil {
			return nil, "", fmt.Errorf("fileExtensions is nil")
		}

		fileExtension = fileExtensions[0]
	}

	// Save the image to a bytes.Buffer
	buffer := &bytes.Buffer{}
	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, "", err
	}

	return buffer, fileExtension, nil
}

func (user *User) refreshAvatar() (bool, error) {
	var err error
	var fileBuffer *bytes.Buffer
	var ext string

	// Gravatar
	if (user.AvatarType == "Auto" || user.AvatarType == "Gravatar") && user.Email != "" {
		client := proxy.ProxyHttpClient

		has, err := hasGravatar(client, user.Email)
		if err != nil {
			return false, err
		}

		if has {
			fileBuffer, ext, err = getGravatarFileBuffer(client, user.Email)
			if err != nil {
				return false, err
			}

			if fileBuffer != nil {
				user.AvatarType = "Gravatar"
			}
		}
	}

	// Favicon
	if fileBuffer == nil && (user.AvatarType == "Auto" || user.AvatarType == "Favicon") {
		client := proxy.ProxyHttpClient

		fileBuffer, ext, err = getFaviconFileBuffer(client, user.Email)
		if err != nil {
			return false, err
		}

		if fileBuffer != nil {
			user.AvatarType = "Favicon"
		}
	}

	// Identicon
	if fileBuffer == nil && (user.AvatarType == "Auto" || user.AvatarType == "Identicon") {
		fileBuffer, ext, err = getIdenticonFileBuffer(user.Name)
		if err != nil {
			return false, err
		}

		if fileBuffer != nil {
			user.AvatarType = "Identicon"
		}
	}

	if fileBuffer != nil {
		avatarUrl, err := getPermanentAvatarUrlFromBuffer(user.Owner, user.Name, fileBuffer, ext, true)
		if err != nil {
			return false, err
		}
		user.Avatar = avatarUrl
		return true, nil
	}

	return false, nil
}

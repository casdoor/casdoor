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
	"net/http"
	"strings"

	"github.com/casdoor/casdoor/proxy"
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
		if strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "no such host") {
			return nil, "", nil
		} else {
			return nil, "", err
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("downloadImage() error for url [%s]: %s\n", url, resp.Status)
		if resp.StatusCode == 404 {
			return nil, "", nil
		} else {
			return nil, "", fmt.Errorf("failed to download gravatar image: %s", resp.Status)
		}
	}

	// Get the content type and determine the file extension
	contentType := resp.Header.Get("Content-Type")
	fileExtension := ""

	if strings.Contains(contentType, "text/html") {
		fileExtension = ".html"
	} else {
		switch contentType {
		case "image/jpeg":
			fileExtension = ".jpg"
		case "image/png":
			fileExtension = ".png"
		case "image/gif":
			fileExtension = ".gif"
		case "image/vnd.microsoft.icon":
			fileExtension = ".ico"
		case "image/x-icon":
			fileExtension = ".ico"
		default:
			return nil, "", fmt.Errorf("unsupported content type: %s", contentType)
		}
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

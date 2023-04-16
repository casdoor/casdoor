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
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/fogleman/gg"
)

func hasGravatar(client *http.Client, email string) (bool, error) {
	// Clean and lowercase the email
	email = strings.TrimSpace(strings.ToLower(email))

	// Generate MD5 hash of the email
	hash := md5.New()
	io.WriteString(hash, email)
	hashedEmail := fmt.Sprintf("%x", hash.Sum(nil))

	// Create Gravatar URL with d=404 parameter
	gravatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=404", hashedEmail)

	// Send a request to Gravatar
	req, err := http.NewRequest("GET", gravatarURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Check if the user has a custom Gravatar image
	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, fmt.Errorf("failed to fetch gravatar image: %s", resp.Status)
	}
}

func getGravatarFileBuffer(client *http.Client, email string) (*bytes.Buffer, string, error) {
	// Clean and lowercase the email
	email = strings.TrimSpace(strings.ToLower(email))

	// Generate MD5 hash of the email
	hash := md5.New()
	io.WriteString(hash, email)
	hashedEmail := fmt.Sprintf("%x", hash.Sum(nil))

	// Create Gravatar URL
	gravatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%s", hashedEmail)

	// Download the image
	req, err := http.NewRequest("GET", gravatarURL, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download gravatar image: %s", resp.Status)
	}

	// Get the content type and determine the file extension
	contentType := resp.Header.Get("Content-Type")
	fileExtension := ""
	switch contentType {
	case "image/jpeg":
		fileExtension = ".jpg"
	case "image/png":
		fileExtension = ".png"
	case "image/gif":
		fileExtension = ".gif"
	default:
		return nil, "", fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Save the image to a bytes.Buffer
	buffer := &bytes.Buffer{}
	_, err = io.Copy(buffer, resp.Body)
	if err != nil {
		return nil, "", err
	}

	return buffer, fileExtension, nil
}

func getColor(data []byte) color.RGBA {
	r := int(data[0]) % 256
	g := int(data[1]) % 256
	b := int(data[2]) % 256
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}

func getIdenticonFileBuffer(username string) (*bytes.Buffer, string, error) {
	username = strings.TrimSpace(strings.ToLower(username))

	hash := md5.New()
	io.WriteString(hash, username)
	hashedUsername := hash.Sum(nil)

	// Define the size of the image
	const imageSize = 420
	const cellSize = imageSize / 7

	// Create a new image
	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))

	// Create a context
	dc := gg.NewContextForRGBA(img)

	// Set a background color
	dc.SetColor(color.RGBA{240, 240, 240, 255})
	dc.Clear()

	// Get avatar color
	avatarColor := getColor(hashedUsername)

	// Draw cells
	for i := 0; i < 7; i++ {
		for j := 0; j < 7; j++ {
			if (hashedUsername[i] >> uint(j) & 1) == 1 {
				dc.SetColor(avatarColor)
				dc.DrawRectangle(float64(j*cellSize), float64(i*cellSize), float64(cellSize), float64(cellSize))
				dc.Fill()
			}
		}
	}

	// Save image to a bytes.Buffer
	buffer := &bytes.Buffer{}
	err := png.Encode(buffer, img)
	if err != nil {
		return nil, "", fmt.Errorf("failed to save image: %w", err)
	}

	return buffer, ".png", nil
}

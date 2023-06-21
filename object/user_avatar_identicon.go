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
	"strings"

	"github.com/fogleman/gg"
)

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

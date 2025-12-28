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
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

	// Create context with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Send a request to Gravatar
	req, err := http.NewRequestWithContext(ctx, "GET", gravatarURL, nil)
	if err != nil {
		return false, err
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("hasGravatar() error: %v\n", err)
		return false, nil
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
	_, err := io.WriteString(hash, email)
	if err != nil {
		return nil, "", err
	}
	hashedEmail := fmt.Sprintf("%x", hash.Sum(nil))

	// Create Gravatar URL
	gravatarUrl := fmt.Sprintf("https://www.gravatar.com/avatar/%s", hashedEmail)

	return downloadImage(client, gravatarUrl)
}

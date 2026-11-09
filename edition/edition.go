// Copyright 2026 The Casdoor Authors. All Rights Reserved.
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

package edition

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
)

const (
	Community  = "Community"
	Commercial = "Commercial"
	Enterprise = "Enterprise"
)

const PricingUrl = "https://casdoor.cn/pricing"

var ranks = map[string]int{
	Community:  0,
	Commercial: 1,
	Enterprise: 2,
}

var publicKey = ed25519.PublicKey(mustDecodeBase64("yu3Y2GTk/CNjdXfooXzLVUDCo19jTV3VXVDFLdHu+Bg="))

type License struct {
	Id         string `json:"id"`
	Licensee   string `json:"licensee"`
	Edition    string `json:"edition"`
	IssuedTime string `json:"issuedTime"`
	ExpireTime string `json:"expireTime"`
}

type LicenseInfo struct {
	Edition    string `json:"edition"`
	Id         string `json:"id,omitempty"`
	Licensee   string `json:"licensee,omitempty"`
	IssuedTime string `json:"issuedTime,omitempty"`
	ExpireTime string `json:"expireTime,omitempty"`
	Error      string `json:"error,omitempty"`
}

var (
	loadOnce   sync.Once
	current    *License
	currentErr error
)

func mustDecodeBase64(s string) []byte {
	res, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return res
}

// ParseLicense verifies a license of the form "<base64url payload>.<base64url Ed25519 signature of the payload part>"
func ParseLicense(text string) (*License, error) {
	parts := strings.Split(strings.TrimSpace(text), ".")
	if len(parts) != 2 {
		return nil, fmt.Errorf("the license is malformed")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("the license is malformed: %s", err.Error())
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("the license is malformed: %s", err.Error())
	}
	if !ed25519.Verify(publicKey, []byte(parts[0]), signature) {
		return nil, fmt.Errorf("the license signature is invalid")
	}

	license := &License{}
	err = json.Unmarshal(payload, license)
	if err != nil {
		return nil, fmt.Errorf("the license is malformed: %s", err.Error())
	}
	if rank, ok := ranks[license.Edition]; !ok || rank == 0 {
		return nil, fmt.Errorf("the license has an unknown edition: %s", license.Edition)
	}
	if license.ExpireTime != "" {
		if _, err = time.Parse(time.RFC3339, license.ExpireTime); err != nil {
			return nil, fmt.Errorf("the license has an invalid expire time: %s", license.ExpireTime)
		}
	}

	return license, nil
}

// readConfiguredLicense takes the "license" config item (or environment variable), which is either the license
// itself or the path of a file holding it
func readConfiguredLicense() string {
	value := strings.TrimSpace(conf.GetConfigString("license"))
	if value == "" {
		return ""
	}

	if info, err := os.Stat(value); err == nil && !info.IsDir() {
		data, err := os.ReadFile(value)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	return value
}

func load() {
	loadOnce.Do(func() {
		text := readConfiguredLicense()
		if text == "" {
			return
		}

		current, currentErr = ParseLicense(text)
		if currentErr != nil {
			logs.Warn("license: %s, running as the %s edition", currentErr.Error(), Community)
		} else {
			logs.Info("license: %s edition licensed to %s, expires at %s", current.Edition, current.Licensee, current.ExpireTime)
		}
	})
}

func (l *License) isExpired() bool {
	if l.ExpireTime == "" {
		return false
	}
	expireTime, err := time.Parse(time.RFC3339, l.ExpireTime)
	return err != nil || time.Now().After(expireTime)
}

func Get() string {
	load()
	if current == nil || current.isExpired() {
		return Community
	}
	return current.Edition
}

func Has(edition string) bool {
	return ranks[Get()] >= ranks[edition]
}

// Require returns the error to show when a feature needs a higher edition than the license gives
func Require(edition string, feature string) error {
	if Has(edition) {
		return nil
	}
	return fmt.Errorf("%s requires a Casdoor %s edition license or higher, see %s", feature, strings.ToLower(edition), PricingUrl)
}

func GetLicenseInfo() *LicenseInfo {
	load()
	info := &LicenseInfo{Edition: Get()}
	if currentErr != nil {
		info.Error = currentErr.Error()
	}
	if current != nil {
		info.Id = current.Id
		info.Licensee = current.Licensee
		info.IssuedTime = current.IssuedTime
		info.ExpireTime = current.ExpireTime
		if current.isExpired() {
			info.Error = "the license has expired"
		}
	}
	return info
}

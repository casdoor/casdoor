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
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
)

const defaultLicenseServer = "https://admin.casdoor.com"

// Reporter gives what a licensed Casdoor tells the license server about itself
type Reporter struct {
	InstanceId func() string
	Version    func() string
	UserCount  func() int64
}

var (
	reporterMu sync.RWMutex
	reporter   *Reporter
)

// SetReporter is called once the database is ready
func SetReporter(r *Reporter) {
	reporterMu.Lock()
	reporter = r
	reporterMu.Unlock()
}

func getReporter() *Reporter {
	reporterMu.RLock()
	defer reporterMu.RUnlock()
	return reporter
}

func getInstanceId() string {
	r := getReporter()
	if r == nil || r.InstanceId == nil {
		return ""
	}
	return r.InstanceId()
}

var (
	fingerprintOnce sync.Once
	fingerprint     string
)

// getFingerprint identifies the machine (or container): its machine ID when there is one, else its MAC addresses
func getFingerprint() string {
	fingerprintOnce.Do(func() {
		parts := []string{}
		for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
			if data, err := os.ReadFile(path); err == nil && strings.TrimSpace(string(data)) != "" {
				parts = append(parts, strings.TrimSpace(string(data)))
				break
			}
		}
		if len(parts) == 0 {
			interfaces, _ := net.Interfaces()
			for _, i := range interfaces {
				if i.Flags&net.FlagLoopback == 0 && len(i.HardwareAddr) > 0 {
					parts = append(parts, i.HardwareAddr.String())
				}
			}
			sort.Strings(parts)
		}
		hostname, _ := os.Hostname()
		parts = append(parts, hostname)
		sum := sha256.Sum256([]byte(strings.Join(parts, "\n")))
		fingerprint = hex.EncodeToString(sum[:16])
	})
	return fingerprint
}

func getVersion() string {
	r := getReporter()
	if r == nil || r.Version == nil {
		return ""
	}
	return r.Version()
}

// GetActivationCode is what a customer sends us to activate the license of an instance that cannot reach the license server
func GetActivationCode() string {
	instanceId := getInstanceId()
	if instanceId == "" {
		return ""
	}
	hostname, _ := os.Hostname()
	data, _ := json.Marshal(map[string]string{
		"instanceId":  instanceId,
		"fingerprint": getFingerprint(),
		"hostname":    hostname,
		"version":     getVersion(),
	})
	return base64.RawURLEncoding.EncodeToString(data)
}

type reportResponse struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		Status  string `json:"status"`
		License string `json:"license"`
		Message string `json:"message"`
	} `json:"data"`
}

// reportLoop tells the license server every hour that this instance runs the license, so that a license copied to
// more instances than it allows shows up; it also picks up a renewed license and a revocation
func reportLoop() {
	time.Sleep(time.Minute)
	for {
		err := report()
		if err != nil {
			logs.Warn("license: failed to report to the license server: %s", err.Error())
		}
		time.Sleep(time.Hour)
	}
}

func report() error {
	r := getReporter()
	if r == nil {
		return fmt.Errorf("the database is not ready")
	}

	stateMu.RLock()
	text := currentText
	license := current
	stateMu.RUnlock()
	if license == nil {
		return nil
	}

	// the config item overrides the server in the license, e.g. "none" for an instance without Internet access
	server := strings.TrimRight(conf.GetConfigString("licenseServer"), "/")
	if server == "" {
		server = strings.TrimRight(license.Server, "/")
	}
	if server == "" {
		server = defaultLicenseServer
	}
	if server == "none" {
		return nil
	}

	hostname, _ := os.Hostname()
	userCount := int64(0)
	if r.UserCount != nil {
		userCount = r.UserCount()
	}
	body, err := json.Marshal(map[string]interface{}{
		"license":     text,
		"instanceId":  getInstanceId(),
		"fingerprint": getFingerprint(),
		"hostname":    hostname,
		"version":     getVersion(),
		"edition":     Get(),
		"userCount":   userCount,
	})
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Post(server+"/api/report-license", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	res := &reportResponse{}
	err = json.NewDecoder(resp.Body).Decode(res)
	if err != nil {
		return err
	}
	if res.Status != "ok" {
		return fmt.Errorf("%s", res.Msg)
	}

	switch res.Data.Status {
	case "Revoked", "Invalid":
		message := res.Data.Message
		if message == "" {
			message = "the license was rejected by the license server"
		}
		stateMu.Lock()
		if rejectedErr == nil {
			logs.Warn("license: %s, running as the %s edition", message, Community)
		}
		rejectedErr = fmt.Errorf("%s", message)
		stateMu.Unlock()
	default:
		stateMu.Lock()
		if rejectedErr != nil {
			logs.Info("license: the license server accepts the license again")
		}
		rejectedErr = nil
		stateMu.Unlock()
	}

	if res.Data.License != "" && strings.TrimSpace(res.Data.License) != text {
		renewed, err := ParseLicense(res.Data.License)
		if err != nil {
			return fmt.Errorf("the renewed license from the license server is invalid: %s", err.Error())
		}
		if renewed.Id != license.Id {
			return fmt.Errorf("the license server sent a license with another ID: %s", renewed.Id)
		}
		stateMu.Lock()
		current = renewed
		currentText = strings.TrimSpace(res.Data.License)
		stateMu.Unlock()
		logs.Info("license: renewed by the license server, %s edition licensed to %s, expires at %s", renewed.Edition, renewed.Licensee, renewed.ExpireTime)
	}
	return nil
}

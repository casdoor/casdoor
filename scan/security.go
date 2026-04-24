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

package scan

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/object"
)

const dataSourceUrl = "https://casdoor.ai/casdoor-data/data.json"

type CVE struct {
	Name        string   `json:"name"`
	Code        string   `json:"code"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Suggestion  string   `json:"suggestion"`
	Rule        string   `json:"rule"`
	References  []string `json:"references"`
}

type FingerprintHttpInfo struct {
	Method   string               `json:"method"`
	Path     string               `json:"path"`
	Matchers []FingerprintMatcher `json:"matchers"`
}

type FingerprintMatcher struct {
	Pos   string `json:"pos"`
	Value string `json:"value"`
}

type FingerprintVersionInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Part   string `json:"part"`
	Regex  string `json:"regex"`
}

type Fingerprint struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
	Product     string `json:"product"`
	Vendor      string `json:"vendor"`

	HttpInfo    FingerprintHttpInfo    `json:"httpInfo"`
	VersionInfo FingerprintVersionInfo `json:"versionInfo"`
}

type SecurityScanProvider struct {
	Type  string `json:"type"`
	Owner string `json:"owner"`

	OnlineList string `json:"onlineList"`
	TargetURL  string `json:"targetUrl"`
}

type SecurityScanFinding struct {
	Name      string `json:"name"`
	Product   string `json:"product"`
	Vendor    string `json:"vendor"`
	Version   string `json:"version"`
	Severity  string `json:"severity"`
	TargetURL string `json:"targetUrl"`
	CVEs      []CVE  `json:"cves"`
}

type securityScanTarget struct {
	Name    string
	BaseURL string
}

type onlineScanLists struct {
	CVEList         []CVE         `json:"cveList"`
	FingerprintList []Fingerprint `json:"fingerprintList"`
}

func NewScanProviderFromProvider(provider *object.Provider) SecurityScanProvider {
	return SecurityScanProvider{Type: provider.SubType, Owner: provider.Owner, OnlineList: provider.Endpoint, TargetURL: provider.Content}
}

func (v SecurityScanProvider) Scan(target string, command string) (string, error) {
	_ = command

	if !strings.EqualFold(v.Type, "Site") && !strings.EqualFold(v.Type, "Url") {
		return "", fmt.Errorf("scan provider sub type: %s is not supported", v.Type)
	}

	cves, fingerprints, err := getOnlineScanLists(dataSourceUrl)
	if err != nil {
		return "", err
	}

	runtimeCVEList := buildCVEMap(cves)
	runtimeFingerprintList := buildFingerprintList(fingerprints)

	if strings.TrimSpace(v.OnlineList) != "" {
		onlineCVEList, onlineFingerprintList, err := getOnlineScanLists(v.OnlineList)
		if err != nil {
			logs.Warning("scan: failed to load online scan lists, onlineList = %s, err = %v", v.OnlineList, err)
		} else {
			mergeOnlineCVEs(runtimeCVEList, onlineCVEList)
			runtimeFingerprintList = mergeOnlineFingerprints(runtimeFingerprintList, onlineFingerprintList)
		}
	}

	scanTargets, err := v.getScanTargets(strings.TrimSpace(target))
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	findings := make([]SecurityScanFinding, 0)
	findingMap := map[string]int{}

	for _, scanTarget := range scanTargets {
		hasFindingForTarget := false
		for _, fingerprint := range runtimeFingerprintList {
			matched, err := isFingerprintMatched(client, scanTarget.BaseURL, fingerprint)
			if err != nil {
				logs.Warning("scan: fingerprint probe failed, target = %s, fingerprint = %s, err = %v", scanTarget.Name, fingerprint.Name, err)
				continue
			}

			if !matched {
				continue
			}

			hasFindingForTarget = true

			version, err := getFingerprintVersion(client, scanTarget.BaseURL, fingerprint)
			if err != nil {
				logs.Warning("scan: version probe failed, target = %s, fingerprint = %s, err = %v", scanTarget.Name, fingerprint.Name, err)
			}

			cves := filterMatchedCVEs(runtimeCVEList[fingerprint.Name], version)
			finding := SecurityScanFinding{
				Name:      scanTarget.Name,
				Product:   fingerprint.Product,
				Vendor:    fingerprint.Vendor,
				Version:   version,
				Severity:  fingerprint.Severity,
				TargetURL: scanTarget.BaseURL,
				CVEs:      cves,
			}
			finding = normalizeUnknownFinding(finding)

			findingKey := fmt.Sprintf("%s|%s", finding.Name, finding.TargetURL)
			if idx, ok := findingMap[findingKey]; ok {
				findings[idx] = mergeFinding(findings[idx], finding)
				continue
			}

			findingMap[findingKey] = len(findings)
			findings = append(findings, finding)
		}

		if !hasFindingForTarget {
			unknownFinding := normalizeUnknownFinding(SecurityScanFinding{
				Name:      scanTarget.Name,
				TargetURL: scanTarget.BaseURL,
				CVEs:      []CVE{},
			})
			findingMapKey := fmt.Sprintf("%s|%s", unknownFinding.Name, unknownFinding.TargetURL)
			if idx, ok := findingMap[findingMapKey]; ok {
				findings[idx] = mergeFinding(findings[idx], unknownFinding)
			} else {
				findingMap[findingMapKey] = len(findings)
				findings = append(findings, unknownFinding)
			}
		}
	}

	sort.Slice(findings, func(i, j int) bool {
		if findings[i].TargetURL == findings[j].TargetURL {
			return findings[i].Name < findings[j].Name
		}
		return findings[i].TargetURL < findings[j].TargetURL
	})

	resultBytes, err := json.Marshal(findings)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

func (v SecurityScanProvider) ParseResult(rawResult string) (string, error) {
	var findings []SecurityScanFinding
	if err := json.Unmarshal([]byte(rawResult), &findings); err != nil {
		return "", err
	}

	resultBytes, err := json.Marshal(findings)
	if err != nil {
		return "", err
	}

	return string(resultBytes), nil
}

func (v SecurityScanProvider) GetResultSummary(result string) string {
	var findings []SecurityScanFinding
	if err := json.Unmarshal([]byte(result), &findings); err != nil {
		return fmt.Sprintf("invalid result: %v", err)
	}

	targetSet := map[string]struct{}{}
	cveCount := 0
	for _, finding := range findings {
		targetSet[finding.TargetURL] = struct{}{}
		cveCount += len(finding.CVEs)
	}

	return fmt.Sprintf("targets=%d findings=%d cves=%d", len(targetSet), len(findings), cveCount)
}

func (v SecurityScanProvider) getScanTargets(target string) ([]securityScanTarget, error) {
	if strings.EqualFold(v.Type, "Site") {
		sites, err := object.GetSites(v.Owner)
		if err != nil {
			return nil, err
		}

		res := make([]securityScanTarget, 0)
		for _, site := range sites {
			if site == nil {
				continue
			}

			for _, baseURL := range getSiteBaseURLs(site) {
				if target != "" && !strings.Contains(baseURL, target) {
					continue
				}
				res = append(res, securityScanTarget{Name: site.Name, BaseURL: baseURL})
			}
		}
		return res, nil
	}

	targetURL := strings.TrimSpace(target)
	if targetURL == "" {
		targetURL = strings.TrimSpace(v.TargetURL)
	}
	if targetURL == "" {
		return nil, fmt.Errorf("target URL is required for Url scan")
	}

	targetURLs := splitScanTargets(targetURL)
	if len(targetURLs) == 0 {
		return nil, fmt.Errorf("target URL is required for Url scan")
	}

	res := make([]securityScanTarget, 0, len(targetURLs))
	for _, currentTargetURL := range targetURLs {
		targetName, baseURL, err := normalizeScanBaseURL(currentTargetURL)
		if err != nil {
			return nil, err
		}

		res = append(res, securityScanTarget{Name: targetName, BaseURL: baseURL})
	}

	return res, nil
}

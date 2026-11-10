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
	"regexp"
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

// The features that need a license, as passed to Require. The API translates them by these names
const (
	FeatureExclusiveSignin  = "Exclusive sign-in"
	FeatureInactiveTimeout  = "Inactivity timeout"
	FeatureMobileOneClick   = "Mobile one-click login"
	FeatureIdJag            = "ID-JAG"
	FeatureMcpGateway       = "MCP gateway"
	FeatureKeycloakSyncer   = "Keycloak syncer"
	FeatureMultipleSyncers  = "More than one syncer"
	FeatureSecurityPolicy   = "Security policy"
	FeatureKeyRotation      = "Signing key rotation"
	FeaturePushMfa          = "Push notification MFA"
	FeaturePaidIntegrations = "Paid integrations"
	FeatureTracing          = "OpenTelemetry tracing"
	FeatureRbac             = "Role-based access control (RBAC)"
	FeatureAbac             = "Attribute-based access control (ABAC)"
	FeatureGuomi            = "Chinese SM cryptographic algorithms"
	FeatureSiemFormats      = "CEF and LEEF audit log formats"
	FeatureAdminSeparation  = "Three-admin separation"
	FeatureTamperProofLogs  = "Tamper-proof audit records"
	FeatureFieldEncryption  = "Encrypted personal fields"
	FeatureSod              = "Separation of duties"
	FeatureConsentLog       = "Consent log"
	FeatureGovernance       = "Account governance"
	FeatureComplianceReport = "MLPS compliance checklist"
	FeatureSecurityOverview = "Security dashboard and reports"
	FeatureOtpTokenImport   = "Hardware OTP token import"
	FeaturePluginIdp        = "Identity provider plugins"
	FeaturePluginIdv        = "ID verification plugins"
	FeatureFormFill         = "Form autofill"
	FeatureHsm              = "Hardware security module"
	FeatureCertLogin        = "Certificate and UKey login"
	FeaturePam              = "Privileged account management"
	FeatureAccessReview     = "Access reviews"
	FeatureDelegatedAdmin   = "Delegated administration"
	FeatureMcpApproval      = "MCP tool call approval"
)

const KindOnPremise = "OnPremise"

func getKind() string {
	license, _, _ := getState()
	if license == nil {
		return ""
	}
	return license.Kind
}

// The features that a license can also grant on their own, without the edition they belong to, by these keys in its
// "features", e.g. a community edition license with "features": ["rbac", "abac"], or a commercial edition license
// with "features": ["form-fill"]
var featureKeys = map[string]string{
	FeatureRbac:     "rbac",
	FeatureAbac:     "abac",
	FeatureFormFill: "form-fill",
}

var ranks = map[string]int{
	Community:  0,
	Commercial: 1,
	Enterprise: 2,
}

var publicKey = ed25519.PublicKey(mustDecodeBase64("yu3Y2GTk/CNjdXfooXzLVUDCo19jTV3VXVDFLdHu+Bg="))

type License struct {
	Id         string   `json:"id"`
	Licensee   string   `json:"licensee"`
	Edition    string   `json:"edition"`
	IssuedTime string   `json:"issuedTime"`
	ExpireTime string   `json:"expireTime"`
	Features   []string `json:"features,omitempty"`
	// set when the license was activated offline for one Casdoor instance (its database)
	InstanceId string `json:"instanceId,omitempty"`
	// the license server that issued the license, where this instance reports to
	Server string `json:"server,omitempty"`
	// how the license was sold: "SaaS", "OnPremise", "Internal" or "Demo"
	Kind string `json:"kind,omitempty"`
}

type LicenseInfo struct {
	Edition    string   `json:"edition"`
	Features   []string `json:"features,omitempty"`
	Kind       string   `json:"kind,omitempty"`
	Id         string   `json:"id,omitempty"`
	Licensee   string   `json:"licensee,omitempty"`
	IssuedTime string   `json:"issuedTime,omitempty"`
	ExpireTime string   `json:"expireTime,omitempty"`
	Error      string   `json:"error,omitempty"`
	// for a global admin: this instance's ID and the code to send for an offline activation
	InstanceId     string `json:"instanceId,omitempty"`
	ActivationCode string `json:"activationCode,omitempty"`
	// for a global admin: the upgrade the license server asks for
	UpgradeOrder *UpgradeOrder `json:"upgradeOrder,omitempty"`
}

var (
	loadOnce    sync.Once
	stateMu     sync.RWMutex
	current     *License
	currentErr  error
	currentText string
	// set when the license server says the license was revoked or is bound to another instance
	rejectedErr error
)

func init() {
	conf.SetEditionGetter(Get)
	conf.SetHideUpgradesGetter(HidesUpgrades)
}

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
	if rank, ok := ranks[license.Edition]; !ok || (rank == 0 && len(license.Features) == 0) {
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
		if currentErr == nil {
			currentText = strings.TrimSpace(text)
		}
		if currentErr != nil {
			logs.Warn("license: %s, running as the %s edition", currentErr.Error(), Community)
		} else {
			logs.Info("license: %s edition licensed to %s, expires at %s", current.Edition, current.Licensee, current.ExpireTime)
			go reportLoop()
		}
	})
}

// getState returns the license in effect, or nil with the reason it is not
func getState() (*License, string, error) {
	load()
	stateMu.RLock()
	defer stateMu.RUnlock()
	if current == nil {
		return nil, "", currentErr
	}
	if current.isExpired() {
		return nil, "", fmt.Errorf("the license has expired")
	}
	if rejectedErr != nil {
		return nil, "", rejectedErr
	}
	if current.InstanceId != "" {
		instanceId := getInstanceId()
		if instanceId != "" && instanceId != current.InstanceId {
			return nil, "", fmt.Errorf("the license is activated for another Casdoor instance")
		}
	}
	return current, currentText, nil
}

func (l *License) isExpired() bool {
	if l.ExpireTime == "" {
		return false
	}
	expireTime, err := time.Parse(time.RFC3339, l.ExpireTime)
	return err != nil || time.Now().After(expireTime)
}

func Get() string {
	license, _, _ := getState()
	if license == nil {
		return Community
	}
	return license.Edition
}

// GetLicenseText returns the configured license when it is valid and not expired, for the Marketplace to send
// when downloading a paid integration
func GetLicenseText() string {
	_, text, _ := getState()
	return text
}

func Has(edition string) bool {
	return ranks[Get()] >= ranks[edition]
}

// RequiredError tells that a feature needs a higher edition than the license gives
type RequiredError struct {
	Edition string
	Feature string
}

func (e *RequiredError) Error() string {
	return fmt.Sprintf("%s requires a Casdoor %s edition license or higher, see %s", e.Feature, strings.ToLower(e.Edition), PricingUrl)
}

var requiredRegex = regexp.MustCompile(`([A-Za-z0-9][A-Za-z0-9 ()\-]*?) requires a Casdoor (commercial|enterprise) edition license or higher, see ` + regexp.QuoteMeta(PricingUrl))

// FindRequired finds the error of Require in a message, which may have been wrapped or turned into a string,
// and returns it with where it is in the message
func FindRequired(message string) (*RequiredError, int, int) {
	match := requiredRegex.FindStringSubmatchIndex(message)
	if match == nil {
		return nil, 0, 0
	}
	required := &RequiredError{
		Feature: message[match[2]:match[3]],
		Edition: strings.ToUpper(message[match[4]:match[4]+1]) + message[match[4]+1:match[5]],
	}
	return required, match[0], match[1]
}

// HidesUpgrades tells whether the console should leave out the offers of higher editions: an on-premise deployment
// bought its edition by contract and does not upgrade from the console, while a hosted instance can
func HidesUpgrades() bool {
	license, _, _ := getState()
	return license != nil && license.Kind == KindOnPremise
}

// GetFeatures returns the features that the license grants on their own
func GetFeatures() []string {
	license, _, _ := getState()
	if license == nil {
		return nil
	}
	return license.Features
}

// Allows tells whether the license gives the edition, or grants the feature on its own
func Allows(edition string, feature string) bool {
	if Has(edition) {
		return true
	}
	key, ok := featureKeys[feature]
	if !ok {
		return false
	}
	for _, item := range GetFeatures() {
		if item == key {
			return true
		}
	}
	return false
}

// Require returns the error to show when a feature needs a higher edition than the license gives
func Require(edition string, feature string) error {
	if Allows(edition, feature) {
		return nil
	}
	return &RequiredError{Edition: edition, Feature: feature}
}

func GetLicenseInfo() *LicenseInfo {
	_, _, err := getState()
	info := &LicenseInfo{Edition: Get(), Features: GetFeatures(), Kind: getKind(), InstanceId: getInstanceId(), ActivationCode: GetActivationCode(), UpgradeOrder: GetUpgradeOrder()}
	if err != nil {
		info.Error = err.Error()
	}
	stateMu.RLock()
	defer stateMu.RUnlock()
	if current != nil {
		info.Id = current.Id
		info.Licensee = current.Licensee
		info.IssuedTime = current.IssuedTime
		info.ExpireTime = current.ExpireTime
	}
	return info
}

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

package object

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pquerna/otp"
)

// Hardware OTP tokens (such as those of Feitian and Jianshi) come with their seeds in a PSKC file (RFC 6030) or a
// CSV from the vendor. Importing binds each token to a user as the user's TOTP authenticator. A token whose time step
// or length is not the 30 seconds and 6 digits of authenticator apps keeps them after its secret:
// "SECRET?period=60&digits=8".

const otpTokenSerialProperty = "otpTokenSerial"

func parseTotpSecret(value string) (string, uint, otp.Digits) {
	secret, query, found := strings.Cut(value, "?")
	period, digits := uint(MfaTotpPeriodInSeconds), otp.DigitsSix
	if !found {
		return secret, period, digits
	}
	values, err := url.ParseQuery(query)
	if err != nil {
		return secret, period, digits
	}
	if p, err := strconv.Atoi(values.Get("period")); err == nil && p > 0 {
		period = uint(p)
	}
	if values.Get("digits") == "8" {
		digits = otp.DigitsEight
	}
	return secret, period, digits
}

type OtpToken struct {
	Serial string `json:"serial"`
	Secret string `json:"-"`
	Period int    `json:"period"`
	Digits int    `json:"digits"`
	User   string `json:"user"`
}

func (token *OtpToken) getTotpSecret() string {
	secret := token.Secret
	query := url.Values{}
	if token.Period != 0 && token.Period != MfaTotpPeriodInSeconds {
		query.Set("period", strconv.Itoa(token.Period))
	}
	if token.Digits == 8 {
		query.Set("digits", "8")
	}
	if len(query) > 0 {
		secret += "?" + query.Encode()
	}
	return secret
}

// normalizeOtpSecret turns a seed in base32, hex ("hex:" prefix or only hex digits of an even length of 40 and more)
// or base64 ("base64:" prefix) into the base32 that TOTP takes
func normalizeOtpSecret(value string) (string, error) {
	value = strings.TrimSpace(value)
	var raw []byte
	var err error
	switch {
	case strings.HasPrefix(value, "hex:"):
		raw, err = hex.DecodeString(strings.TrimPrefix(value, "hex:"))
	case strings.HasPrefix(value, "base64:"):
		raw, err = base64.StdEncoding.DecodeString(strings.TrimPrefix(value, "base64:"))
	default:
		if decoded, hexErr := hex.DecodeString(value); hexErr == nil && len(value) >= 40 {
			raw = decoded
		} else {
			cleaned := strings.ToUpper(strings.ReplaceAll(value, " ", ""))
			raw, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(cleaned, "="))
		}
	}
	if err != nil {
		return "", fmt.Errorf("the token seed is not base32, hex or base64: %w", err)
	}
	if len(raw) < 10 {
		return "", fmt.Errorf("the token seed is shorter than 80 bits")
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(raw), nil
}

type pskcContainer struct {
	Keys []struct {
		Id             string `xml:"Id,attr"`
		Algorithm      string `xml:"Algorithm,attr"`
		SerialNo       string `xml:"DeviceInfo>SerialNo"`
		ResponseFormat struct {
			Length int `xml:"Length,attr"`
		} `xml:"AlgorithmParameters>ResponseFormat"`
		Secret       string `xml:"Data>Secret>PlainValue"`
		Encrypted    string `xml:"Data>Secret>EncryptedValue>CipherData>CipherValue"`
		TimeInterval int    `xml:"Data>TimeInterval>PlainValue"`
		UserId       string `xml:"UserId"`
	} `xml:"KeyPackage>Key"`
}

// parsePskc reads the TOTP keys of a PSKC file with plaintext secrets
func parsePskc(content string) ([]*OtpToken, error) {
	container := &pskcContainer{}
	err := xml.Unmarshal([]byte(content), container)
	if err != nil {
		return nil, fmt.Errorf("the PSKC file is not valid XML: %w", err)
	}

	tokens := []*OtpToken{}
	for _, key := range container.Keys {
		if key.Encrypted != "" && key.Secret == "" {
			return nil, fmt.Errorf("the seeds in the PSKC file are encrypted, ask the vendor for the file with plaintext seeds or the transport key")
		}
		if !strings.Contains(strings.ToLower(key.Algorithm), "totp") {
			continue
		}
		secret, err := normalizeOtpSecret("base64:" + strings.TrimSpace(key.Secret))
		if err != nil {
			return nil, fmt.Errorf("the token %s: %w", key.Id, err)
		}
		serial := key.SerialNo
		if serial == "" {
			serial = key.Id
		}
		tokens = append(tokens, &OtpToken{Serial: serial, Secret: secret, Period: key.TimeInterval, Digits: key.ResponseFormat.Length, User: key.UserId})
	}
	return tokens, nil
}

// parseOtpTokenCsv reads lines of "serial,secret[,period][,digits]" or, with a header naming them, any order of the
// columns serial, secret, period, digits and user
func parseOtpTokenCsv(content string) ([]*OtpToken, error) {
	reader := csv.NewReader(strings.NewReader(content))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	columns := map[string]int{"serial": 0, "secret": 1, "period": 2, "digits": 3, "user": -1}
	tokens := []*OtpToken{}
	for i, row := range rows {
		if len(row) == 0 || strings.HasPrefix(strings.TrimSpace(row[0]), "#") {
			continue
		}
		if i == 0 && isOtpCsvHeader(row) {
			for key := range columns {
				columns[key] = -1
			}
			for j, name := range row {
				columns[strings.ToLower(strings.TrimSpace(name))] = j
			}
			continue
		}
		get := func(key string) string {
			j := columns[key]
			if j < 0 || j >= len(row) {
				return ""
			}
			return strings.TrimSpace(row[j])
		}

		secret, err := normalizeOtpSecret(get("secret"))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		period, _ := strconv.Atoi(get("period"))
		digits, _ := strconv.Atoi(get("digits"))
		tokens = append(tokens, &OtpToken{Serial: get("serial"), Secret: secret, Period: period, Digits: digits, User: get("user")})
	}
	return tokens, nil
}

func isOtpCsvHeader(row []string) bool {
	for _, name := range row {
		if strings.EqualFold(strings.TrimSpace(name), "secret") {
			return true
		}
	}
	return false
}

// parseOtpAssignments reads lines of "username,serial"
func parseOtpAssignments(content string) (map[string]string, error) {
	res := map[string]string{}
	reader := csv.NewReader(strings.NewReader(content))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if len(row) < 2 || strings.HasPrefix(strings.TrimSpace(row[0]), "#") || strings.EqualFold(strings.TrimSpace(row[0]), "user") {
			continue
		}
		res[strings.TrimSpace(row[1])] = strings.TrimSpace(row[0])
	}
	return res, nil
}

type OtpTokenImportResult struct {
	Imported int      `json:"imported"`
	Skipped  []string `json:"skipped"`
}

// ImportOtpTokens binds hardware OTP tokens to users of the organization: tokens is a PSKC file or a CSV of seeds,
// assignments a CSV of "username,serial" for the tokens that don't name their user
func ImportOtpTokens(owner string, tokensContent string, assignmentsContent string) (*OtpTokenImportResult, error) {
	var tokens []*OtpToken
	var err error
	if strings.HasPrefix(strings.TrimSpace(tokensContent), "<") {
		tokens, err = parsePskc(tokensContent)
	} else {
		tokens, err = parseOtpTokenCsv(tokensContent)
	}
	if err != nil {
		return nil, err
	}
	assignments, err := parseOtpAssignments(assignmentsContent)
	if err != nil {
		return nil, err
	}

	res := &OtpTokenImportResult{Skipped: []string{}}
	for _, token := range tokens {
		if token.User == "" {
			token.User = assignments[token.Serial]
		}
		if token.User == "" {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: no user", token.Serial))
			continue
		}
		if token.Period != 0 && token.Period != 30 && token.Period != 60 {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: the time step of %d seconds is not supported", token.Serial, token.Period))
			continue
		}

		user, err := getUser(owner, token.User)
		if err != nil {
			return nil, err
		}
		if user == nil {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: the user %s does not exist", token.Serial, token.User))
			continue
		}

		user.TotpSecret = token.getTotpSecret()
		if user.PreferredMfaType == "" {
			user.PreferredMfaType = TotpType
		}
		if user.Properties == nil {
			user.Properties = map[string]string{}
		}
		user.Properties[otpTokenSerialProperty] = token.Serial
		_, err = updateUser(user.GetId(), user, []string{"totp_secret", "preferred_mfa_type", "properties"})
		if err != nil {
			return nil, err
		}
		res.Imported++
	}
	return res, nil
}

// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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

package util

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/nyaruka/phonenumbers"
)

func ParseInt(s string) int {
	if s == "" {
		return 0
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}

	return i
}

func ParseIntWithError(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("ParseIntWithError() error, empty string")
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func ParseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}

	return f
}

func ParseBool(s string) bool {
	if s == "\x01" || s == "true" {
		return true
	} else if s == "false" {
		return false
	}

	i := ParseInt(s)
	return i != 0
}

func BoolToString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// CamelToSnakeCase This function transform camelcase in snakecase LoremIpsum in lorem_ipsum
func CamelToSnakeCase(camel string) string {
	var buf bytes.Buffer
	for _, c := range camel {
		if 'A' <= c && c <= 'Z' {
			// just convert [A-Z] to _[a-z]
			if buf.Len() > 0 {
				buf.WriteRune('_')
			}
			buf.WriteRune(c - 'A' + 'a')
			continue
		}
		buf.WriteRune(c)
	}
	return strings.ReplaceAll(buf.String(), " ", "")
}

func SnakeToCamel(snake string) string {
	words := strings.Split(snake, "_")
	for i := range words {
		words[i] = strings.ToLower(words[i])
		if i > 0 {
			words[i] = strings.Title(words[i])
		}
	}
	return strings.Join(words, "")
}

func SpaceToCamel(name string) string {
	words := strings.Split(name, " ")
	for i := range words {
		words[i] = strings.ToLower(words[i])
		words[i] = strings.Title(words[i])
	}
	return strings.Join(words, "")
}

func GetOwnerAndNameFromId(id string) (string, string) {
	tokens := strings.Split(id, "/")
	if len(tokens) != 2 {
		panic(errors.New("GetOwnerAndNameFromId() error, wrong token count for ID: " + id))
	}

	return tokens[0], tokens[1]
}

func GetOwnerAndNameFromIdWithError(id string) (string, string, error) {
	tokens := strings.Split(id, "/")
	if len(tokens) != 2 {
		return "", "", errors.New("GetOwnerAndNameFromId() error, wrong token count for ID: " + id)
	}

	return tokens[0], tokens[1], nil
}

func GetOwnerFromId(id string) string {
	tokens := strings.Split(id, "/")
	if len(tokens) != 2 {
		panic(errors.New("GetOwnerAndNameFromId() error, wrong token count for ID: " + id))
	}

	return tokens[0]
}

func GetOwnerAndNameFromIdNoCheck(id string) (string, string) {
	tokens := strings.SplitN(id, "/", 2)
	return tokens[0], tokens[1]
}

func GetOwnerAndNameAndOtherFromId(id string) (string, string, string) {
	tokens := strings.Split(id, "/")
	if len(tokens) != 3 {
		panic(errors.New("GetOwnerAndNameAndOtherFromId() error, wrong token count for ID: " + id))
	}

	return tokens[0], tokens[1], tokens[2]
}

func GetSharedOrgFromApp(rawName string) (name string, organization string) {
	name = rawName
	splitName := strings.Split(rawName, "-org-")
	if len(splitName) >= 2 {
		organization = splitName[len(splitName)-1]
		name = splitName[0]
	}
	return name, organization
}

func GenerateId() string {
	return uuid.NewString()
}

func GenerateTimeId() string {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	t := tm.Format("20060102_150405")

	random := uuid.NewString()[0:7]

	res := fmt.Sprintf("%s_%s", t, random)
	return res
}

func GenerateSimpleTimeId() string {
	timestamp := time.Now().Unix()
	tm := time.Unix(timestamp, 0)
	t := tm.Format("20060102150405")

	return t
}

func GetRandomName() string {
	rand.Seed(time.Now().UnixNano())
	const charset = "0123456789abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 6)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

func GetId(owner, name string) string {
	return fmt.Sprintf("%s/%s", owner, name)
}

func GetSessionId(owner, name, application string) string {
	return fmt.Sprintf("%s/%s/%s", owner, name, application)
}

func GetMd5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func IsStringsEmpty(strs ...string) bool {
	for _, str := range strs {
		if len(str) == 0 {
			return true
		}
	}
	return false
}

func ReadStringFromPath(path string) string {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		panic(err)
	}

	return string(data)
}

func WriteStringToPath(s string, path string) {
	err := os.WriteFile(path, []byte(s), 0o644)
	if err != nil {
		panic(err)
	}
}

// SnakeString transform XxYy to xx_yy
func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	result := strings.ToLower(string(data[:]))
	return strings.ReplaceAll(result, " ", "")
}

func IsChinese(str string) bool {
	var flag bool
	for _, v := range str {
		if unicode.Is(unicode.Han, v) {
			flag = true
			break
		}
	}
	return flag
}

func GetMaskedPhone(phone string) string {
	return rePhone.ReplaceAllString(phone, "$1****$2")
}

func GetSeperatedPhone(phone string) string {
	if strings.HasPrefix(phone, "+") {
		phoneNumberParsed, err := phonenumbers.Parse(phone, "")
		if err != nil {
			return phone
		}

		phone = fmt.Sprintf("%d", phoneNumberParsed.GetNationalNumber())
	}

	return phone
}

func GetMaskedEmail(email string) string {
	if email == "" {
		return ""
	}

	if !strings.Contains(email, "@") {
		return maskString(email)
	}

	tokens := strings.Split(email, "@")
	username := maskString(tokens[0])
	domain := tokens[1]
	domainTokens := strings.Split(domain, ".")
	domainTokens[len(domainTokens)-2] = maskString(domainTokens[len(domainTokens)-2])
	return fmt.Sprintf("%s@%s", username, strings.Join(domainTokens, "."))
}

func maskString(str string) string {
	if len(str) <= 2 {
		return str
	} else {
		return fmt.Sprintf("%c%s%c", str[0], strings.Repeat("*", len(str)-2), str[len(str)-1])
	}
}

// GetEndPoint remove scheme from url
func GetEndPoint(endpoint string) string {
	for _, prefix := range []string{"https://", "http://"} {
		endpoint = strings.TrimPrefix(endpoint, prefix)
	}
	return endpoint
}

// HasString reports if slice has input string.
func HasString(strs []string, str string) bool {
	for _, i := range strs {
		if i == str {
			return true
		}
	}
	return false
}

func ParseIdToString(input interface{}) (string, error) {
	switch v := input.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	default:
		return "", fmt.Errorf("unsupported id type: %T", input)
	}
}

func GetValueFromDataSourceName(key string, dataSourceName string) string {
	reg := regexp.MustCompile(key + "=([^ ]+)")
	matches := reg.FindStringSubmatch(dataSourceName)
	if len(matches) >= 2 {
		return matches[1]
	}

	return ""
}

func GetUsernameFromEmail(email string) string {
	tokens := strings.Split(email, "@")
	if len(tokens) == 0 {
		return uuid.NewString()
	} else {
		return tokens[0]
	}
}

func StringToInterfaceArray(array []string) []interface{} {
	var (
		interfaceArray []interface{}
		elem           interface{}
	)
	for _, elem = range array {
		jStruct, err := TryJsonToAnonymousStruct(elem.(string))
		if err == nil {
			elem = jStruct
		}
		interfaceArray = append(interfaceArray, elem)
	}
	return interfaceArray
}

func StringToInterfaceArray2d(arrays [][]string) [][]interface{} {
	var interfaceArrays [][]interface{}
	for _, req := range arrays {
		var (
			interfaceArray []interface{}
			elem           interface{}
		)
		for _, elem = range req {
			jStruct, err := TryJsonToAnonymousStruct(elem.(string))
			if err == nil {
				elem = jStruct
			}
			interfaceArray = append(interfaceArray, elem)
		}
		interfaceArrays = append(interfaceArrays, interfaceArray)
	}
	return interfaceArrays
}

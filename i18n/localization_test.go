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

package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/casdoor/casdoor/util"
)

// json.Unmarshal alone silently accepts a later duplicate and would miss a shadowed translation.
func checkCatalogJSON(decoder *json.Decoder, path string) error {
	value, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := value.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		seen := map[string]bool{}
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("invalid object key at %s", path)
			}
			if seen[key] {
				return fmt.Errorf("duplicate key %s/%s", path, key)
			}
			seen[key] = true
			if err = checkCatalogJSON(decoder, path+"/"+key); err != nil {
				return err
			}
		}
	case '[':
		for decoder.More() {
			if err = checkCatalogJSON(decoder, path+"/[]"); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unexpected delimiter at %s", path)
	}
	_, err = decoder.Token()
	return err
}

func loadLocalizationCatalog(t *testing.T, path string) I18nData {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(content))
	if err = checkCatalogJSON(decoder, path); err != nil {
		t.Fatal(err)
	}
	if _, err = decoder.Token(); err != io.EOF {
		t.Fatalf("trailing JSON in %s: %v", path, err)
	}
	catalog := I18nData{}
	if err = json.Unmarshal(content, &catalog); err != nil {
		t.Fatal(err)
	}
	return catalog
}

var localizationPlaceholders = regexp.MustCompile(`\{\{[^{}]+\}\}|%\{[^{}]+\}|\{[A-Za-z][A-Za-z0-9_.]*\}|%(?:expireTime|link|code)|%(?:\[[0-9]+\])?[-+# 0]*[0-9]*(?:\.[0-9]+)?[sdvfqw]`)

func placeholders(value string) []string {
	result := localizationPlaceholders.FindAllString(value, -1)
	sort.Strings(result)
	return result
}

func TestLocalizationCatalogStructure(t *testing.T) {
	for _, root := range []string{"locales", "../web/src/locales"} {
		t.Run(root, func(t *testing.T) {
			english := loadLocalizationCatalog(t, filepath.Join(root, "en/data.json"))
			traditional := loadLocalizationCatalog(t, filepath.Join(root, "zh-TW/data.json"))
			seen := map[string]string{}
			for namespace, pairs := range english {
				for key, source := range pairs {
					fullKey := namespace + ":" + key
					if previous := seen[key]; previous != "" {
						t.Errorf("cross-namespace duplicate: %s and %s", previous, fullKey)
					}
					seen[key] = fullKey
					translation, exists := traditional[namespace][key]
					if !exists || strings.TrimSpace(translation) == "" {
						t.Errorf("missing Traditional Chinese: %s", fullKey)
						continue
					}
					if !reflect.DeepEqual(placeholders(source), placeholders(translation)) {
						t.Errorf("placeholder mismatch: %s: %v vs %v", fullKey, placeholders(source), placeholders(translation))
					}
				}
			}
			for namespace, pairs := range traditional {
				for key := range pairs {
					if _, exists := english[namespace][key]; !exists {
						t.Errorf("unexpected Traditional Chinese key: %s:%s", namespace, key)
					}
				}
			}
		})
	}
}

func TestLocalizationDuplicateDetector(t *testing.T) {
	for _, content := range []string{`{"server":{"Found servers":"first","Found servers":"second"}}`, `{"server":{},"server":{}}`} {
		if err := checkCatalogJSON(json.NewDecoder(strings.NewReader(content)), "sample"); err == nil {
			t.Fatalf("duplicate was accepted: %s", content)
		}
	}
}

func TestLocalizationBackendExtraction(t *testing.T) {
	source := `package example
func render() {
 i18n.Translate(lang, "test:Translated %s")
 i18n.Errorf("test:Formatted %w", err)
 i18n.NewError("test:Literal 100%")
 i18n.WrapError(err, "test:Wrapped %s", name)
 util.MessageErrorf("test:Low-level %q", name)
 util.NewMessageError("test:Low-level literal")
 util.WrapMessageError(err, "test:Low-level wrapped")
 c.T("test:Controller")
 other.NewMessageError("not:a first-party constructor")
 i18n.Errorf(variable)
}
// i18n.Translate(lang, "not:a commented call")
`
	want := []string{"test:Translated %s", "test:Formatted %w", "test:Literal 100%", "test:Wrapped %s", "test:Low-level %q", "test:Low-level literal", "test:Low-level wrapped", "test:Controller"}
	if got := getAllI18nStringsBackend(source, false); !reflect.DeepEqual(got, want) {
		t.Fatalf("backend extraction changed: %v", got)
	}
	for source, want := range map[string]string{
		`package routers; func f() { T(ctx, "test:Router") }`:                "test:Router",
		`package util; func f() { MessageErrorf("test:Utility %s", value) }`: "test:Utility %s",
	} {
		if got := getAllI18nStringsBackend(source, false); !reflect.DeepEqual(got, []string{want}) {
			t.Fatalf("local extraction changed: %v", got)
		}
	}
}

func TestLocalizationBackendSourceKeys(t *testing.T) {
	english := loadLocalizationCatalog(t, "locales/en/data.json")
	for _, folder := range []string{"controllers", "object", "routers", "service", "util", "audit", "scan", "upgrade", "idp", "email", "notification", "captcha", "faceId", "idv", "pp", "storage"} {
		err := filepath.WalkDir("../"+folder, func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			for _, fullKey := range getAllI18nStringsBackend(string(content), false) {
				namespace, key, _ := strings.Cut(fullKey, ":")
				if english[namespace][key] == "" {
					t.Errorf("missing source key: %s: %s", path, fullKey)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestLocalizationFrontendSourceKeys(t *testing.T) {
	english := loadLocalizationCatalog(t, "../web/src/locales/en/data.json")
	namespaces := map[string]bool{}
	for namespace := range english {
		namespaces[namespace] = true
	}
	files := map[string]string{}
	propertyNamespace := regexp.MustCompile(`(?:i18nKey|labelKey|titleKey|descriptionKey|displayNameKey)\s*[:=]\s*"(\w+):`)
	err := filepath.WalkDir("../web/src", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[path] = string(content)
		for _, namespace := range getI18nNamespacesFrontend(string(content)) {
			namespaces[namespace] = true
		}
		for _, match := range propertyNamespace.FindAllStringSubmatch(string(content), -1) {
			namespaces[match[1]] = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for path, content := range files {
		for _, fullKey := range getAllI18nStringsFrontend(content, namespaces) {
			namespace, key, _ := strings.Cut(fullKey, ":")
			if english[namespace][key] == "" {
				t.Errorf("missing source key: %s: %s", path, fullKey)
			}
		}
	}
	// Dynamic stored values must also have captions; parity cannot detect absence in both catalogs.
	for namespace, values := range map[string][]string{"plan": {"Monthly", "Yearly"}, "product": {"Published", "Draft"}, "rule": {"Block"}, "permission": {"Allow"}} {
		for _, key := range values {
			if english[namespace][key] == "" {
				t.Errorf("missing enum caption: %s:%s", namespace, key)
			}
		}
	}
}

func TestLocalizationConcurrentTranslations(t *testing.T) {
	// Test uncached locale loading, not just concurrent reads of an already-warm map.
	langMapMu.Lock()
	langMap = make(map[string]map[string]map[string]string)
	langMapMu.Unlock()
	var workers sync.WaitGroup
	for i := 0; i < 24; i++ {
		workers.Add(1)
		go func(i int) {
			defer workers.Done()
			language, want := "en", "Loading"
			if i%2 == 0 {
				language, want = "zh-TW", "載入中"
			}
			if got := Translate(language, "native:Loading"); got != want {
				t.Errorf("locale bleed: %s: %q", language, got)
			}
		}(i)
	}
	workers.Wait()
}

func TestLocalizationPrimaryMask(t *testing.T) {
	name := "synthetic-mask-subject"
	value := util.MessageErrorf("general:The user: %s doesn't exist", name)
	for _, language := range []string{"en", "zh-TW"} {
		message := LocalizeError(language, value)
		if strings.Contains(message, name) == enableErrorMask {
			t.Fatalf("primary mask changed for %s: %q", language, message)
		}
	}
	if !strings.Contains(value.Error(), name) {
		t.Fatal("mask rewrote the original diagnostic")
	}
}

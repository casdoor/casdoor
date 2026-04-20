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
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

const (
	openClawTranscriptResourceTag                = "openclawTranscript"
	openClawSessionTranscriptPreviewLimit        = 2 * 1024 * 1024
	openClawTranscriptDefaultStorageProviderName = "openclaw-transcript-storage"
)

type OpenClawRawTranscriptRef struct {
	FileName string `json:"fileName,omitempty"`
	FileSize int    `json:"fileSize,omitempty"`
}

type OpenClawSessionTranscriptPreview struct {
	FileName   string `json:"fileName"`
	FileSize   int    `json:"fileSize"`
	LoadedSize int    `json:"loadedSize"`
	Truncated  bool   `json:"truncated"`
	Content    string `json:"content"`
}

func uploadOpenClawRawTranscript(logProvider *Provider, sessionID string, transcriptPath string, fileSize int64) error {
	storageProvider, err := getOpenClawTranscriptStorageProvider(logProvider)
	if err != nil {
		return err
	}
	if storageProvider == nil {
		return fmt.Errorf("no enabled Storage provider found")
	}

	resource, err := getOpenClawRawTranscriptResource(logProvider.Owner, logProvider.Name, sessionID)
	if err != nil {
		return err
	}

	objectKey := ""
	if resource != nil {
		objectKey = strings.TrimSpace(resource.Name)
	}
	if objectKey == "" {
		fullFilePath := getOpenClawTranscriptStoragePath(logProvider.Owner, logProvider.Name, sessionID, util.GenerateUUID())
		objectKey = getOpenClawTranscriptObjectKey(storageProvider, fullFilePath)
	}
	if strings.Contains(objectKey, "..") {
		return fmt.Errorf("the objectKey: %s is not allowed", objectKey)
	}

	if err := putOpenClawRawTranscriptObject(storageProvider, objectKey, transcriptPath); err != nil {
		return err
	}

	resource = buildOpenClawRawTranscriptResource(logProvider, storageProvider, sessionID, transcriptPath, fileSize, objectKey, resource)
	_, err = AddOrUpdateResource(resource)
	return err
}

func getOpenClawTranscriptStorageProvider(logProvider *Provider) (*Provider, error) {
	if logProvider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	providers, err := GetProviders(logProvider.Owner)
	if err != nil {
		return nil, err
	}

	storageProvider, err := selectOpenClawTranscriptStorageProvider(logProvider, providers)
	if err != nil {
		return nil, err
	}
	if storageProvider != nil {
		return storageProvider, nil
	}

	return createOpenClawTranscriptDefaultStorageProvider(logProvider)
}

func selectOpenClawTranscriptStorageProvider(logProvider *Provider, providers []*Provider) (*Provider, error) {
	selectedName := strings.TrimSpace(logProvider.ProviderUrl)
	if selectedName != "" {
		for _, provider := range providers {
			if provider == nil || provider.Name != selectedName {
				continue
			}
			if !isUsableOpenClawTranscriptStorageProviderForOwner(logProvider.Owner, provider) {
				return nil, fmt.Errorf("selected provider %s is not an enabled Storage provider", selectedName)
			}
			return provider, nil
		}

		return nil, fmt.Errorf("selected Storage provider %s is not found", selectedName)
	}

	for _, provider := range providers {
		if !isUsableOpenClawTranscriptStorageProviderForOwner(logProvider.Owner, provider) {
			continue
		}
		return provider, nil
	}

	return nil, nil
}

func isUsableOpenClawTranscriptStorageProvider(provider *Provider) bool {
	return provider != nil && provider.Category == "Storage" && !strings.EqualFold(strings.TrimSpace(provider.State), "Disabled")
}

func isUsableOpenClawTranscriptStorageProviderForOwner(owner string, provider *Provider) bool {
	return isUsableOpenClawTranscriptStorageProvider(provider) && strings.TrimSpace(provider.Owner) == strings.TrimSpace(owner)
}

func createOpenClawTranscriptDefaultStorageProvider(logProvider *Provider) (*Provider, error) {
	if logProvider == nil {
		return nil, fmt.Errorf("provider is nil")
	}

	provider := &Provider{
		Owner:       logProvider.Owner,
		Name:        openClawTranscriptDefaultStorageProviderName,
		CreatedTime: util.GetCurrentTime(),
		DisplayName: "OpenClaw Transcript Storage",
		Category:    "Storage",
		Type:        ProviderTypeLocalFileSystem,
		State:       "Enabled",
		PathPrefix:  "openclaw-transcripts",
		Domain:      getOpenClawDefaultStorageDomain(),
	}

	affected, err := AddProvider(provider)
	if err != nil {
		providers, getErr := GetProviders(logProvider.Owner)
		if getErr != nil {
			return nil, err
		}

		for _, existing := range providers {
			if existing != nil && existing.Name == openClawTranscriptDefaultStorageProviderName && isUsableOpenClawTranscriptStorageProviderForOwner(logProvider.Owner, existing) {
				return existing, nil
			}
		}

		return nil, err
	}
	if !affected {
		return nil, fmt.Errorf("failed to create default OpenClaw transcript Storage provider")
	}

	return provider, nil
}

func getOpenClawDefaultStorageDomain() string {
	if origin := strings.TrimSpace(conf.GetConfigString("origin")); origin != "" {
		return origin
	}
	return strings.TrimSpace(conf.GetConfigString("originFrontend"))
}

func getOpenClawTranscriptStoragePath(owner string, providerName string, sessionID string, nonce string) string {
	return path.Join(
		"openclaw",
		sanitizeOpenClawStoragePathSegment(owner),
		sanitizeOpenClawStoragePathSegment(providerName),
		"sessions",
		sanitizeOpenClawStoragePathSegment(nonce),
		fmt.Sprintf("%s.jsonl", sanitizeOpenClawStoragePathSegment(sessionID)),
	)
}

func getOpenClawTranscriptObjectKey(provider *Provider, fullFilePath string) string {
	escapedPath := util.UrlJoin(provider.PathPrefix, fullFilePath)
	objectKey := util.UrlJoin(util.GetUrlPath(provider.Domain), escapedPath)
	if provider.Type == ProviderTypeLocalFileSystem {
		objectKey = strings.TrimLeft(objectKey, "/")
	}
	if provider.Type == ProviderTypeTencentCloudCOS {
		objectKey = escapePath(objectKey)
	}
	return objectKey
}

func getOpenClawTranscriptResourceParent(owner string, providerName string, sessionID string) string {
	owner = strings.TrimSpace(owner)
	providerName = strings.TrimSpace(providerName)
	sessionID = strings.TrimSpace(sessionID)
	if owner == "" || providerName == "" || sessionID == "" {
		return ""
	}

	return path.Join("openclaw", util.GetMd5Hash(fmt.Sprintf("%s|%s|%s", owner, providerName, sessionID)))
}

func getOpenClawRawTranscriptResource(owner string, providerName string, sessionID string) (*Resource, error) {
	owner = strings.TrimSpace(owner)
	parent := getOpenClawTranscriptResourceParent(owner, providerName, sessionID)
	if owner == "" || parent == "" {
		return nil, nil
	}

	resource := Resource{}
	existed, err := ormer.Engine.
		Where("owner = ? and tag = ? and parent = ?", owner, openClawTranscriptResourceTag, parent).
		Desc("created_time").
		Get(&resource)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}

	return &resource, nil
}

func getOpenClawRawTranscriptRef(owner string, providerName string, sessionID string) (*OpenClawRawTranscriptRef, error) {
	resource, err := getOpenClawRawTranscriptResource(owner, providerName, sessionID)
	if err != nil {
		return nil, err
	}
	return buildOpenClawRawTranscriptRef(resource), nil
}

func buildOpenClawRawTranscriptRef(resource *Resource) *OpenClawRawTranscriptRef {
	if resource == nil {
		return nil
	}

	return &OpenClawRawTranscriptRef{
		FileName: resource.FileName,
		FileSize: resource.FileSize,
	}
}

func GetOpenClawSessionTranscript(id string, lang string) (*OpenClawSessionTranscriptPreview, error) {
	entry, _, payload, err := loadOpenClawSessionEntry(id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	resource, err := getOpenClawRawTranscriptResource(entry.Owner, entry.Provider, payload.SessionID)
	if err != nil {
		return nil, err
	}
	if resource == nil {
		return nil, nil
	}

	provider, err := getOpenClawTranscriptStorageProviderByName(resource.Owner, resource.Provider)
	if err != nil {
		return nil, err
	}
	if provider == nil || provider.Category != "Storage" {
		return nil, fmt.Errorf("storage provider %s is not found", resource.Provider)
	}

	storageProvider, err := getStorageProvider(provider, lang)
	if err != nil {
		return nil, err
	}

	reader, err := storageProvider.GetStream(refineObjectKey(provider, resource.Name))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	fileName := strings.TrimSpace(resource.FileName)
	if fileName == "" {
		fileName = fmt.Sprintf("%s.jsonl", payload.SessionID)
	}

	content, truncated, err := readOpenClawSessionTranscriptPreview(reader, openClawSessionTranscriptPreviewLimit)
	if err != nil {
		return nil, err
	}

	return &OpenClawSessionTranscriptPreview{
		FileName:   filepath.Base(fileName),
		FileSize:   resource.FileSize,
		LoadedSize: len(content),
		Truncated:  truncated,
		Content:    string(content),
	}, nil
}

func putOpenClawRawTranscriptObject(provider *Provider, objectKey string, transcriptPath string) error {
	storageProvider, err := getStorageProvider(provider, "en")
	if err != nil {
		return err
	}

	file, err := os.Open(transcriptPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = storageProvider.Put(refineObjectKey(provider, objectKey), file)
	return err
}

func buildOpenClawRawTranscriptResource(logProvider *Provider, storageProvider *Provider, sessionID string, transcriptPath string, fileSize int64, objectKey string, current *Resource) *Resource {
	now := util.GetCurrentTime()
	createdTime := now
	if current != nil && strings.TrimSpace(current.CreatedTime) != "" {
		createdTime = current.CreatedTime
	}

	return &Resource{
		Owner:       logProvider.Owner,
		Name:        objectKey,
		CreatedTime: createdTime,
		User:        "",
		Provider:    storageProvider.Name,
		Application: CasdoorApplication,
		Tag:         openClawTranscriptResourceTag,
		Parent:      getOpenClawTranscriptResourceParent(logProvider.Owner, logProvider.Name, sessionID),
		FileName:    filepath.Base(transcriptPath),
		FileType:    "text",
		FileFormat:  filepath.Ext(transcriptPath),
		FileSize:    int(fileSize),
		Url:         "",
		Description: "OpenClaw raw session transcript",
	}
}

func getOpenClawTranscriptStorageProviderByName(owner string, name string) (*Provider, error) {
	if strings.TrimSpace(owner) == "" || strings.TrimSpace(name) == "" {
		return nil, nil
	}

	provider := Provider{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&provider)
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, nil
	}

	return &provider, nil
}

func readOpenClawSessionTranscriptPreview(reader io.Reader, limit int) ([]byte, bool, error) {
	if limit < 0 {
		limit = 0
	}

	content, err := io.ReadAll(io.LimitReader(reader, int64(limit)+1))
	if err != nil {
		return nil, false, err
	}
	if len(content) > limit {
		return content[:limit], true, nil
	}

	return content, false, nil
}

func loadOpenClawSessionEntry(id string) (*Entry, *Provider, openClawBehaviorPayload, error) {
	entry, err := GetEntry(id)
	if err != nil {
		return nil, nil, openClawBehaviorPayload{}, err
	}
	if entry == nil {
		return nil, nil, openClawBehaviorPayload{}, nil
	}
	if strings.TrimSpace(entry.Type) != "session" {
		return nil, nil, openClawBehaviorPayload{}, fmt.Errorf("entry %s is not an OpenClaw session entry", id)
	}

	provider, err := GetProvider(util.GetId(entry.Owner, entry.Provider))
	if err != nil {
		return nil, nil, openClawBehaviorPayload{}, err
	}
	if provider != nil && !isOpenClawLogProvider(provider) {
		return nil, nil, openClawBehaviorPayload{}, fmt.Errorf("entry %s is not an OpenClaw session entry", id)
	}

	payload, err := parseOpenClawSessionGraphPayload(entry)
	if err != nil {
		return nil, nil, openClawBehaviorPayload{}, fmt.Errorf("failed to parse anchor entry %s: %w", entry.Name, err)
	}

	return entry, provider, payload, nil
}

func sanitizeOpenClawStoragePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown"
	}

	var builder strings.Builder
	for _, ch := range value {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' {
			builder.WriteRune(ch)
		} else {
			builder.WriteByte('_')
		}
	}

	sanitized := strings.Trim(builder.String(), ". ")
	for strings.Contains(sanitized, "..") {
		sanitized = strings.ReplaceAll(sanitized, "..", "_")
	}
	if sanitized == "" {
		return "unknown"
	}
	return sanitized
}

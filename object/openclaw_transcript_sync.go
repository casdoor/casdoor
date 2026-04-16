package object

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

const openClawTranscriptSyncInterval = 10 * time.Second

const (
	openClawTranscriptSummaryMax        = 100
	openClawTranscriptTaskTextMax       = 2000
	openClawTranscriptAssistantTextMax  = 2000
	openClawTranscriptFinalTextMax      = 2000
	openClawTranscriptToolCommandMax    = 500
	openClawTranscriptToolResultTextMax = 64 * 1024
	openClawTranscriptErrorTextMax      = 500
)

var (
	openClawTranscriptWorkers   = map[string]*openClawTranscriptSyncWorker{}
	openClawTranscriptWorkersMu sync.Mutex

	writeSuccessPathPattern = regexp.MustCompile(`(?i)successfully wrote \d+ bytes to (.+)$`)
)

type openClawTranscriptSyncWorker struct {
	provider      *Provider
	stopCh        chan struct{}
	doneCh        chan struct{}
	fileStates    map[string]openClawTranscriptFileState
	pathErrLogged bool
}

type openClawTranscriptFileState struct {
	ModTimeUnixNano int64
	Size            int64
}

type openClawTranscriptEntry struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	ParentID  string                 `json:"parentId"`
	Timestamp string                 `json:"timestamp"`
	Message   *openClawMessage       `json:"message"`
	Details   map[string]interface{} `json:"details"`
}

type openClawMessage struct {
	Role       string                 `json:"role"`
	Content    json.RawMessage        `json:"content"`
	StopReason string                 `json:"stopReason"`
	ToolCallID string                 `json:"toolCallId"`
	ToolName   string                 `json:"toolName"`
	IsError    bool                   `json:"isError"`
	Timestamp  int64                  `json:"timestamp"`
	Details    map[string]interface{} `json:"details"`
}

type openClawContentItem struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Text      string                 `json:"text"`
	Arguments map[string]interface{} `json:"arguments"`
}

type openClawBehaviorPayload struct {
	Summary       string `json:"summary"`
	Kind          string `json:"kind"`
	SessionID     string `json:"sessionId"`
	EntryID       string `json:"entryId"`
	ToolCallID    string `json:"toolCallId,omitempty"`
	ParentID      string `json:"parentId,omitempty"`
	Timestamp     string `json:"timestamp"`
	Tool          string `json:"tool,omitempty"`
	Query         string `json:"query,omitempty"`
	URL           string `json:"url,omitempty"`
	Path          string `json:"path,omitempty"`
	OK            *bool  `json:"ok,omitempty"`
	Error         string `json:"error,omitempty"`
	AssistantText string `json:"assistantText,omitempty"`
	Text          string `json:"text,omitempty"`
}

type openClawToolContext struct {
	Tool    string
	Query   string
	URL     string
	Path    string
	Command string
}

func startOpenClawTranscriptSync(provider *Provider) {
	if provider == nil || provider.Category != "Log" || provider.Type != "Agent" || provider.SubType != "OpenClaw" {
		return
	}

	id := provider.GetId()
	stopOpenClawTranscriptSync(id)

	worker := &openClawTranscriptSyncWorker{
		provider:   provider,
		stopCh:     make(chan struct{}),
		doneCh:     make(chan struct{}),
		fileStates: map[string]openClawTranscriptFileState{},
	}

	openClawTranscriptWorkersMu.Lock()
	openClawTranscriptWorkers[id] = worker
	openClawTranscriptWorkersMu.Unlock()

	go worker.run()
}

func stopOpenClawTranscriptSync(providerID string) {
	openClawTranscriptWorkersMu.Lock()
	worker, ok := openClawTranscriptWorkers[providerID]
	if ok {
		delete(openClawTranscriptWorkers, providerID)
	}
	openClawTranscriptWorkersMu.Unlock()

	if !ok {
		return
	}

	close(worker.stopCh)
	<-worker.doneCh
}

func (w *openClawTranscriptSyncWorker) run() {
	defer close(w.doneCh)

	w.syncOnce()

	ticker := time.NewTicker(openClawTranscriptSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.syncOnce()
		}
	}
}

func (w *openClawTranscriptSyncWorker) syncOnce() {
	if w.isStopping() {
		return
	}

	if err := w.scanTranscriptDir(); err != nil {
		if os.IsNotExist(err) {
			if !w.pathErrLogged {
				fmt.Printf("OpenClaw transcript sync failed for provider %s: %v\n", w.provider.Name, err)
				w.pathErrLogged = true
			}
		} else {
			fmt.Printf("OpenClaw transcript sync failed for provider %s: %v\n", w.provider.Name, err)
		}
	} else {
		w.pathErrLogged = false
	}
}

func (w *openClawTranscriptSyncWorker) isStopping() bool {
	select {
	case <-w.stopCh:
		return true
	default:
		return false
	}
}

func (w *openClawTranscriptSyncWorker) scanTranscriptDir() error {
	rootDir, err := resolveOpenClawTranscriptDir(w.provider)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return err
	}

	seenPaths := map[string]struct{}{}
	for _, entry := range entries {
		if w.isStopping() {
			return nil
		}

		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".jsonl") || strings.Contains(name, ".reset.") || name == "sessions.json" {
			continue
		}

		path := filepath.Join(rootDir, name)
		seenPaths[path] = struct{}{}

		info, err := entry.Info()
		if err != nil {
			return err
		}
		nextState := openClawTranscriptFileState{
			ModTimeUnixNano: info.ModTime().UnixNano(),
			Size:            info.Size(),
		}
		if w.shouldSkipTranscriptFile(path, nextState) {
			continue
		}

		if err := w.scanTranscriptFile(path); err != nil {
			return err
		}
		w.fileStates[path] = nextState
	}

	for path := range w.fileStates {
		if w.isStopping() {
			return nil
		}

		if _, ok := seenPaths[path]; !ok {
			delete(w.fileStates, path)
		}
	}

	return nil
}

func (w *openClawTranscriptSyncWorker) shouldSkipTranscriptFile(path string, nextState openClawTranscriptFileState) bool {
	currentState, ok := w.fileStates[path]
	if !ok {
		return false
	}
	return currentState.ModTimeUnixNano == nextState.ModTimeUnixNano && currentState.Size == nextState.Size
}

func resolveOpenClawTranscriptDir(provider *Provider) (string, error) {
	if provider == nil {
		return "", fmt.Errorf("provider is nil")
	}

	if endpoint := strings.TrimSpace(provider.Endpoint); endpoint != "" {
		return expandOpenClawPath(endpoint)
	}

	stateDir, err := resolveOpenClawStateDir()
	if err != nil {
		return "", err
	}

	agentID := strings.TrimSpace(provider.Title)
	if agentID == "" {
		agentID = "main"
	}

	return filepath.Join(stateDir, "agents", agentID, "sessions"), nil
}

func fillOpenClawProviderDefaults(provider *Provider) error {
	if !isOpenClawLogProvider(provider) {
		return nil
	}
	if strings.TrimSpace(provider.Title) == "" {
		provider.Title = "main"
	}
	if strings.TrimSpace(provider.Endpoint) != "" {
		resolved, err := expandOpenClawPath(provider.Endpoint)
		if err != nil {
			return err
		}
		provider.Endpoint = resolved
		return nil
	}

	transcriptDir, err := resolveOpenClawTranscriptDir(provider)
	if err != nil {
		return err
	}
	provider.Endpoint = transcriptDir
	return nil
}

func isOpenClawLogProvider(provider *Provider) bool {
	return provider != nil && provider.Category == "Log" && provider.Type == "Agent" && provider.SubType == "OpenClaw"
}

func resolveOpenClawStateDir() (string, error) {
	if override := strings.TrimSpace(os.Getenv("OPENCLAW_STATE_DIR")); override != "" {
		return expandOpenClawPath(override)
	}

	homeDir, err := resolveOpenClawHomeDir()
	if err != nil {
		return "", err
	}

	if profile := strings.TrimSpace(os.Getenv("OPENCLAW_PROFILE")); profile != "" && !strings.EqualFold(profile, "default") {
		return filepath.Join(homeDir, ".openclaw-"+profile), nil
	}

	return filepath.Join(homeDir, ".openclaw"), nil
}

func resolveOpenClawHomeDir() (string, error) {
	if explicitHome := strings.TrimSpace(os.Getenv("OPENCLAW_HOME")); explicitHome != "" {
		return expandOpenClawPath(explicitHome)
	}

	return resolveSystemHomeDir()
}

func resolveSystemHomeDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Clean(homeDir), nil
}

func expandOpenClawPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", nil
	}

	if trimmed == "~" || strings.HasPrefix(trimmed, "~/") || strings.HasPrefix(trimmed, "~\\") {
		homeDir, err := resolveSystemHomeDir()
		if err != nil {
			return "", err
		}
		suffix := strings.TrimPrefix(strings.TrimPrefix(trimmed, "~"), string(filepath.Separator))
		suffix = strings.TrimPrefix(strings.TrimPrefix(suffix, "/"), "\\")
		if suffix == "" {
			return filepath.Clean(homeDir), nil
		}
		return filepath.Clean(filepath.Join(homeDir, suffix)), nil
	}

	if runtime.GOOS == "windows" && len(trimmed) >= 2 && trimmed[0] == '%' {
		if index := strings.Index(trimmed[1:], "%"); index >= 0 {
			end := index + 1
			envKey := trimmed[1:end]
			if envValue := strings.TrimSpace(os.Getenv(envKey)); envValue != "" {
				replaced := envValue + trimmed[end+1:]
				return filepath.Clean(replaced), nil
			}
		}
	}

	return filepath.Clean(trimmed), nil
}

func (w *openClawTranscriptSyncWorker) scanTranscriptFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	sessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
	toolContexts := map[string]openClawToolContext{}
	reader := bufio.NewReader(file)

	for {
		if w.isStopping() {
			return nil
		}

		lineBytes, readErr := reader.ReadBytes('\n')
		if readErr != nil && readErr != io.EOF {
			return readErr
		}

		line := strings.TrimSpace(string(lineBytes))
		if line == "" {
			if readErr == io.EOF {
				break
			}
			continue
		}

		var entry openClawTranscriptEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			if readErr == io.EOF {
				break
			}
			continue
		}

		results := buildOpenClawTranscriptEntries(w.provider, sessionID, entry, toolContexts)
		for _, result := range results {
			if result == nil {
				continue
			}
			if err := addOpenClawTranscriptEntry(result); err != nil {
				return err
			}
		}

		if readErr == io.EOF {
			break
		}
	}

	return nil
}

func buildOpenClawTranscriptEntries(provider *Provider, sessionID string, entry openClawTranscriptEntry, toolContexts map[string]openClawToolContext) []*Entry {
	if entry.Type != "message" || entry.Message == nil {
		return nil
	}

	message := entry.Message
	switch message.Role {
	case "user":
		text := normalizeUserText(extractMessageText(message.Content))
		if text == "" {
			return nil
		}
		if isHeartbeatText(text) {
			return nil
		}

		return []*Entry{newOpenClawTranscriptEntry(provider, sessionID, "task", entry.ID, openClawBehaviorPayload{
			Summary:   truncateText(fmt.Sprintf("task: %s", text), openClawTranscriptSummaryMax),
			Kind:      "task",
			SessionID: sessionID,
			EntryID:   entry.ID,
			ParentID:  entry.ParentID,
			Timestamp: normalizeOpenClawTimestamp(entry.Timestamp, message.Timestamp),
			Text:      truncateText(text, openClawTranscriptTaskTextMax),
		})}
	case "assistant":
		items := parseContentItems(message.Content)
		assistantText := truncateText(extractMessageText(message.Content), openClawTranscriptAssistantTextMax)
		toolEntries := []*Entry{}
		storedAssistantText := false
		for _, item := range items {
			if item.Type != "toolCall" {
				continue
			}
			context := extractOpenClawToolContext(item)
			toolContexts[item.ID] = context
			payload := openClawBehaviorPayload{
				Summary:    truncateText(buildToolCallSummary(context), openClawTranscriptSummaryMax),
				Kind:       "tool_call",
				SessionID:  sessionID,
				EntryID:    entry.ID,
				ToolCallID: item.ID,
				ParentID:   entry.ParentID,
				Timestamp:  normalizeOpenClawTimestamp(entry.Timestamp, message.Timestamp),
				Tool:       context.Tool,
				Query:      context.Query,
				URL:        context.URL,
				Path:       context.Path,
				Text:       truncateText(context.Command, openClawTranscriptToolCommandMax),
			}
			if !storedAssistantText {
				// Avoid duplicating the same assistant text on every tool-call row.
				payload.AssistantText = assistantText
				storedAssistantText = true
			}
			identity := fmt.Sprintf("%s/%s", entry.ID, item.ID)
			toolEntries = append(toolEntries, newOpenClawTranscriptEntry(provider, sessionID, "tool_call", identity, payload))
		}
		if len(toolEntries) > 0 {
			return toolEntries
		}

		if message.StopReason != "stop" {
			return nil
		}
		text := extractMessageText(message.Content)
		if text == "" {
			return nil
		}
		return []*Entry{newOpenClawTranscriptEntry(provider, sessionID, "final", entry.ID, openClawBehaviorPayload{
			Summary:   truncateText(fmt.Sprintf("final: %s", text), openClawTranscriptSummaryMax),
			Kind:      "final",
			SessionID: sessionID,
			EntryID:   entry.ID,
			ParentID:  entry.ParentID,
			Timestamp: normalizeOpenClawTimestamp(entry.Timestamp, message.Timestamp),
			Text:      truncateText(text, openClawTranscriptFinalTextMax),
		})}
	case "toolResult":
		payload, ok := buildToolResultPayload(sessionID, entry, toolContexts[message.ToolCallID])
		if !ok {
			return nil
		}
		return []*Entry{newOpenClawTranscriptEntry(provider, sessionID, "tool_result", entry.ID, payload)}
	default:
		return nil
	}
}

func buildToolResultPayload(sessionID string, entry openClawTranscriptEntry, toolContext openClawToolContext) (openClawBehaviorPayload, bool) {
	message := entry.Message
	if message == nil {
		return openClawBehaviorPayload{}, false
	}

	okValue, errorText := resolveToolResultStatus(entry)
	rawText := extractMessageText(message.Content)
	summaryText := summarizeToolResultText(rawText, okValue)
	toolName := firstNonEmpty(toolContext.Tool, message.ToolName)
	if toolName == "" && rawText == "" && errorText == "" {
		return openClawBehaviorPayload{}, false
	}

	return openClawBehaviorPayload{
		Summary:    truncateText(buildToolResultSummary(toolName, toolContext, okValue, errorText, summaryText), openClawTranscriptSummaryMax),
		Kind:       "tool_result",
		SessionID:  sessionID,
		EntryID:    entry.ID,
		ToolCallID: message.ToolCallID,
		ParentID:   entry.ParentID,
		Timestamp:  normalizeOpenClawTimestamp(entry.Timestamp, message.Timestamp),
		Tool:       toolName,
		Query:      toolContext.Query,
		URL:        toolContext.URL,
		Path:       firstNonEmpty(toolContext.Path, extractWriteSuccessPath(rawText)),
		OK:         &okValue,
		Error:      truncateText(errorText, openClawTranscriptErrorTextMax),
		Text:       truncateText(rawText, openClawTranscriptToolResultTextMax),
	}, true
}

func newOpenClawTranscriptEntry(provider *Provider, sessionID string, entryKind string, identity string, payload openClawBehaviorPayload) *Entry {
	body, _ := json.Marshal(payload)
	nameSource := fmt.Sprintf("%s|%s|%s", provider.Name, sessionID, identity)
	createdTime := payload.Timestamp
	if strings.TrimSpace(createdTime) == "" {
		createdTime = util.GetCurrentTime()
	}

	return &Entry{
		Owner:       CasdoorOrganization,
		Name:        fmt.Sprintf("oc_%s", util.GetMd5Hash(nameSource)),
		CreatedTime: createdTime,
		UpdatedTime: createdTime,
		DisplayName: truncateText(payload.Summary, openClawTranscriptSummaryMax),
		Provider:    provider.Name,
		Type:        "session",
		Message:     string(body),
	}
}

func addOpenClawTranscriptEntry(entry *Entry) error {
	if entry == nil {
		return nil
	}

	_, err := AddEntry(entry)
	if err == nil || isDuplicateEntryError(err) {
		return nil
	}
	return err
}

func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique constraint") || strings.Contains(msg, "already exists")
}

func normalizeOpenClawTimestamp(raw string, fallbackMillis int64) string {
	if trimmed := strings.TrimSpace(raw); trimmed != "" {
		if _, err := time.Parse(time.RFC3339, trimmed); err == nil {
			return trimmed
		}
	}
	if fallbackMillis > 0 {
		return time.UnixMilli(fallbackMillis).UTC().Format(time.RFC3339)
	}
	return util.GetCurrentTime()
}

func parseContentItems(raw json.RawMessage) []openClawContentItem {
	if len(raw) == 0 {
		return nil
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		if strings.TrimSpace(text) == "" {
			return nil
		}
		return []openClawContentItem{{Type: "text", Text: text}}
	}

	var items []openClawContentItem
	if err := json.Unmarshal(raw, &items); err == nil {
		return items
	}

	return nil
}

func extractMessageText(raw json.RawMessage) string {
	items := parseContentItems(raw)
	if len(items) == 0 {
		return ""
	}

	parts := []string{}
	for _, item := range items {
		if item.Type == "text" && strings.TrimSpace(item.Text) != "" {
			parts = append(parts, strings.TrimSpace(item.Text))
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func normalizeUserText(text string) string {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "Sender (untrusted metadata):") {
		return trimmed
	}
	if index := strings.LastIndex(trimmed, "\n\n"); index >= 0 && index+2 < len(trimmed) {
		return strings.TrimSpace(trimmed[index+2:])
	}
	return trimmed
}

func extractOpenClawToolContext(item openClawContentItem) openClawToolContext {
	context := openClawToolContext{Tool: item.Name}
	if item.Arguments == nil {
		return context
	}
	context.Query = stringifyOpenClawArg(item.Arguments["query"])
	context.URL = stringifyOpenClawArg(item.Arguments["url"])
	context.Path = stringifyOpenClawArg(item.Arguments["path"])
	context.Command = stringifyOpenClawArg(item.Arguments["command"])
	return context
}

func stringifyOpenClawArg(value interface{}) string {
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func buildToolCallSummary(context openClawToolContext) string {
	target := firstNonEmpty(context.Query, context.URL, context.Path, context.Command)
	if target == "" {
		return fmt.Sprintf("%s called", context.Tool)
	}
	return fmt.Sprintf("%s: %s", context.Tool, target)
}

func resolveToolResultStatus(entry openClawTranscriptEntry) (bool, string) {
	message := entry.Message
	details := mergeOpenClawTranscriptDetails(entry.Details, message)

	if message != nil && message.IsError {
		return false, stringifyOpenClawArg(details["error"])
	}
	if status, ok := details["status"].(string); ok && strings.EqualFold(status, "error") {
		return false, stringifyOpenClawArg(details["error"])
	}

	text := ""
	if message != nil {
		text = strings.TrimSpace(extractMessageText(message.Content))
	}
	if strings.HasPrefix(text, "{") {
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(text), &payload); err == nil {
			if status, ok := payload["status"].(string); ok && strings.EqualFold(status, "error") {
				return false, stringifyOpenClawArg(payload["error"])
			}
		}
	}

	return true, stringifyOpenClawArg(details["error"])
}

func mergeOpenClawTranscriptDetails(entryDetails map[string]interface{}, message *openClawMessage) map[string]interface{} {
	if message == nil || len(message.Details) == 0 {
		return entryDetails
	}
	if len(entryDetails) == 0 {
		return message.Details
	}

	details := map[string]interface{}{}
	for key, value := range entryDetails {
		details[key] = value
	}
	for key, value := range message.Details {
		details[key] = value
	}
	return details
}

func summarizeToolResultText(text string, okValue bool) string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return ""
	}
	if okValue {
		if path := extractWriteSuccessPath(trimmed); path != "" {
			return fmt.Sprintf("Successfully wrote %s", path)
		}
		return ""
	}
	return trimmed
}

func extractWriteSuccessPath(text string) string {
	matches := writeSuccessPathPattern.FindStringSubmatch(strings.TrimSpace(text))
	if len(matches) < 2 {
		return ""
	}
	return strings.TrimSpace(matches[1])
}

func buildToolResultSummary(tool string, context openClawToolContext, okValue bool, errorText string, text string) string {
	target := firstNonEmpty(context.Query, context.URL, context.Path, context.Command, extractWriteSuccessPath(text))
	status := "ok"
	details := target
	if !okValue {
		status = "failed"
		details = firstNonEmpty(errorText, target)
	}
	if details == "" {
		return fmt.Sprintf("%s %s", tool, status)
	}
	return fmt.Sprintf("%s %s: %s", tool, status, details)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func isHeartbeatText(text string) bool {
	return strings.HasPrefix(strings.TrimSpace(text), "Read HEARTBEAT.md")
}

func truncateText(text string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(text))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max-1]) + "…"
}

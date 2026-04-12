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
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/casdoor/casdoor/util"
)

type OpenClawSessionGraph struct {
	Nodes []*OpenClawSessionGraphNode `json:"nodes"`
	Edges []*OpenClawSessionGraphEdge `json:"edges"`
	Stats OpenClawSessionGraphStats   `json:"stats"`
}

type OpenClawSessionGraphNode struct {
	ID               string `json:"id"`
	ParentID         string `json:"parentId,omitempty"`
	OriginalParentID string `json:"originalParentId,omitempty"`
	EntryID          string `json:"entryId,omitempty"`
	ToolCallID       string `json:"toolCallId,omitempty"`
	Kind             string `json:"kind"`
	Timestamp        string `json:"timestamp"`
	Summary          string `json:"summary"`
	Tool             string `json:"tool,omitempty"`
	Query            string `json:"query,omitempty"`
	URL              string `json:"url,omitempty"`
	Path             string `json:"path,omitempty"`
	OK               *bool  `json:"ok,omitempty"`
	Error            string `json:"error,omitempty"`
	Text             string `json:"text,omitempty"`
	IsAnchor         bool   `json:"isAnchor"`
}

type OpenClawSessionGraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type OpenClawSessionGraphStats struct {
	TotalNodes      int `json:"totalNodes"`
	TaskCount       int `json:"taskCount"`
	ToolCallCount   int `json:"toolCallCount"`
	ToolResultCount int `json:"toolResultCount"`
	FinalCount      int `json:"finalCount"`
	FailedCount     int `json:"failedCount"`
}

type openClawSessionGraphBuilder struct {
	graph    *OpenClawSessionGraph
	nodes    map[string]*OpenClawSessionGraphNode
	edges    []*OpenClawSessionGraphEdge
	edgeKeys map[string]struct{}
}

type openClawSessionGraphRecord struct {
	Entry   *Entry
	Payload openClawBehaviorPayload
}

type openClawAssistantStepGroup struct {
	ParentID          string
	Timestamp         string
	ToolNames         []string
	Text              string
	ToolCallNodeIDs   []string
	ToolResultNodeIDs []string
}

func GetOpenClawSessionGraph(id string) (*OpenClawSessionGraph, error) {
	entry, err := GetEntry(id)
	if err != nil {
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}
	if strings.TrimSpace(entry.Type) != "session" {
		return nil, fmt.Errorf("entry %s is not an OpenClaw session entry", id)
	}

	provider, err := GetProvider(util.GetId(entry.Owner, entry.Provider))
	if err != nil {
		return nil, err
	}
	if provider != nil && !isOpenClawLogProvider(provider) {
		return nil, fmt.Errorf("entry %s is not an OpenClaw session entry", id)
	}

	anchorPayload, err := parseOpenClawSessionGraphPayload(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse anchor entry %s: %w", entry.Name, err)
	}

	records, err := collectOpenClawSessionGraphRecords(entry, anchorPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenClaw session entries from database: %w", err)
	}

	return buildOpenClawSessionGraphFromEntries(anchorPayload, entry.Name, records), nil
}

func parseOpenClawSessionGraphPayload(entry *Entry) (openClawBehaviorPayload, error) {
	if entry == nil {
		return openClawBehaviorPayload{}, fmt.Errorf("entry is nil")
	}

	message := strings.TrimSpace(entry.Message)
	if message == "" {
		return openClawBehaviorPayload{}, fmt.Errorf("message is empty")
	}

	var payload openClawBehaviorPayload
	if err := json.Unmarshal([]byte(message), &payload); err != nil {
		return openClawBehaviorPayload{}, err
	}

	payload.SessionID = strings.TrimSpace(payload.SessionID)
	payload.EntryID = strings.TrimSpace(payload.EntryID)
	payload.ToolCallID = strings.TrimSpace(payload.ToolCallID)
	payload.ParentID = strings.TrimSpace(payload.ParentID)
	payload.Kind = strings.TrimSpace(payload.Kind)
	payload.Summary = strings.TrimSpace(payload.Summary)
	payload.Tool = strings.TrimSpace(payload.Tool)
	payload.Query = strings.TrimSpace(payload.Query)
	payload.URL = strings.TrimSpace(payload.URL)
	payload.Path = strings.TrimSpace(payload.Path)
	payload.Error = strings.TrimSpace(payload.Error)
	payload.AssistantText = strings.TrimSpace(payload.AssistantText)
	payload.Text = strings.TrimSpace(payload.Text)
	payload.Timestamp = strings.TrimSpace(firstNonEmpty(payload.Timestamp, entry.CreatedTime))

	if payload.SessionID == "" {
		return openClawBehaviorPayload{}, fmt.Errorf("sessionId is empty")
	}
	if payload.EntryID == "" {
		return openClawBehaviorPayload{}, fmt.Errorf("entryId is empty")
	}
	if payload.Kind == "" {
		return openClawBehaviorPayload{}, fmt.Errorf("kind is empty")
	}

	return payload, nil
}

func collectOpenClawSessionGraphRecords(anchorEntry *Entry, anchorPayload openClawBehaviorPayload) ([]openClawSessionGraphRecord, error) {
	if anchorEntry == nil {
		return nil, fmt.Errorf("anchor entry is nil")
	}

	entries := []*Entry{}
	query := ormer.Engine.Where("owner = ? and type = ?", anchorEntry.Owner, "session")
	if providerName := strings.TrimSpace(anchorEntry.Provider); providerName != "" {
		query = query.And("provider = ?", providerName)
	}

	if err := query.
		Asc("created_time").
		Asc("name").
		Find(&entries); err != nil {
		return nil, err
	}

	return filterOpenClawSessionGraphRecords(anchorEntry, anchorPayload, entries), nil
}

func filterOpenClawSessionGraphRecords(anchorEntry *Entry, anchorPayload openClawBehaviorPayload, entries []*Entry) []openClawSessionGraphRecord {
	targetSessionID := strings.TrimSpace(anchorPayload.SessionID)
	records := make([]openClawSessionGraphRecord, 0, len(entries)+1)
	hasAnchor := false
	for _, candidate := range entries {
		if candidate == nil {
			continue
		}

		payload, err := parseOpenClawSessionGraphPayload(candidate)
		if err != nil {
			continue
		}
		if payload.SessionID != targetSessionID {
			continue
		}

		records = append(records, openClawSessionGraphRecord{
			Entry:   candidate,
			Payload: payload,
		})
		if candidate.Owner == anchorEntry.Owner && candidate.Name == anchorEntry.Name {
			hasAnchor = true
		}
	}

	if !hasAnchor && anchorEntry != nil {
		records = append(records, openClawSessionGraphRecord{
			Entry:   anchorEntry,
			Payload: anchorPayload,
		})
	}

	sort.SliceStable(records, func(i, j int) bool {
		leftPayload := records[i].Payload
		rightPayload := records[j].Payload
		leftTimestamp := strings.TrimSpace(firstNonEmpty(leftPayload.Timestamp, records[i].Entry.CreatedTime))
		rightTimestamp := strings.TrimSpace(firstNonEmpty(rightPayload.Timestamp, records[j].Entry.CreatedTime))
		if timestampOrder := compareOpenClawGraphTimestamps(leftTimestamp, rightTimestamp); timestampOrder != 0 {
			return timestampOrder < 0
		}
		return records[i].Entry.Name < records[j].Entry.Name
	})

	return records
}

func buildOpenClawSessionGraphFromEntries(anchorPayload openClawBehaviorPayload, anchorEntryName string, records []openClawSessionGraphRecord) *OpenClawSessionGraph {
	builder := newOpenClawSessionGraphBuilder()
	nodeIDsByEntryName := map[string][]string{}
	assistantGroups := map[string]*openClawAssistantStepGroup{}
	toolCallNodesByAssistant := map[string][]*OpenClawSessionGraphNode{}
	toolCallNodeIDByToolCallID := map[string]string{}
	toolCallNodeIDToAssistantID := map[string]string{}
	allToolCallNodes := []*OpenClawSessionGraphNode{}
	toolResultRecords := []openClawSessionGraphRecord{}

	for _, record := range records {
		payload := record.Payload
		switch payload.Kind {
		case "task":
			builder.addNode(&OpenClawSessionGraphNode{
				ID:        payload.EntryID,
				ParentID:  payload.ParentID,
				EntryID:   payload.EntryID,
				Kind:      "task",
				Timestamp: payload.Timestamp,
				Summary:   payload.Summary,
				Text:      payload.Text,
			})
			appendGraphNodeEntryName(nodeIDsByEntryName, record.Entry, payload.EntryID)
		case "tool_call":
			nodeID := buildStoredToolCallNodeID(record.Entry, payload)
			builder.addNode(&OpenClawSessionGraphNode{
				ID:         nodeID,
				ParentID:   payload.EntryID,
				EntryID:    payload.EntryID,
				ToolCallID: payload.ToolCallID,
				Kind:       "tool_call",
				Timestamp:  payload.Timestamp,
				Summary:    payload.Summary,
				Tool:       payload.Tool,
				Query:      payload.Query,
				URL:        payload.URL,
				Path:       payload.Path,
				Text:       payload.Text,
			})
			storedNode := builder.nodes[nodeID]
			appendGraphNodeEntryName(nodeIDsByEntryName, record.Entry, nodeID)
			if storedNode != nil {
				toolCallNodesByAssistant[payload.EntryID] = append(toolCallNodesByAssistant[payload.EntryID], storedNode)
				allToolCallNodes = append(allToolCallNodes, storedNode)
			}
			if payload.ToolCallID != "" && toolCallNodeIDByToolCallID[payload.ToolCallID] == "" {
				toolCallNodeIDByToolCallID[payload.ToolCallID] = nodeID
			}
			if toolCallNodeIDToAssistantID[nodeID] == "" {
				toolCallNodeIDToAssistantID[nodeID] = payload.EntryID
			}

			group := assistantGroups[payload.EntryID]
			if group == nil {
				group = &openClawAssistantStepGroup{
					ParentID:  payload.ParentID,
					Timestamp: payload.Timestamp,
				}
				assistantGroups[payload.EntryID] = group
			}
			group.ParentID = firstNonEmpty(group.ParentID, payload.ParentID)
			group.Timestamp = chooseEarlierTimestamp(group.Timestamp, payload.Timestamp)
			group.ToolNames = append(group.ToolNames, payload.Tool)
			group.Text = firstNonEmpty(group.Text, payload.AssistantText)
			group.ToolCallNodeIDs = appendUniqueString(group.ToolCallNodeIDs, nodeID)
		case "tool_result":
			toolResultRecords = append(toolResultRecords, record)
		case "final":
			builder.addNode(&OpenClawSessionGraphNode{
				ID:        payload.EntryID,
				ParentID:  payload.ParentID,
				EntryID:   payload.EntryID,
				Kind:      "final",
				Timestamp: payload.Timestamp,
				Summary:   payload.Summary,
				Text:      payload.Text,
			})
			appendGraphNodeEntryName(nodeIDsByEntryName, record.Entry, payload.EntryID)
		}
	}

	assistantIDs := make([]string, 0, len(assistantGroups))
	for entryID := range assistantGroups {
		assistantIDs = append(assistantIDs, entryID)
	}
	sort.Strings(assistantIDs)

	for _, assistantID := range assistantIDs {
		group := assistantGroups[assistantID]
		builder.addNode(&OpenClawSessionGraphNode{
			ID:        assistantID,
			ParentID:  strings.TrimSpace(group.ParentID),
			EntryID:   assistantID,
			Kind:      "assistant_step",
			Timestamp: strings.TrimSpace(group.Timestamp),
			Summary:   buildAssistantStepSummary(group.ToolNames),
			Text:      strings.TrimSpace(group.Text),
		})
	}

	for _, record := range toolResultRecords {
		payload := record.Payload
		parentID := strings.TrimSpace(payload.ParentID)
		originalParentID := ""

		if payload.ToolCallID != "" {
			if matchedNodeID := strings.TrimSpace(toolCallNodeIDByToolCallID[payload.ToolCallID]); matchedNodeID != "" {
				originalParentID = parentID
				parentID = matchedNodeID
			}
		}

		if parentID == strings.TrimSpace(payload.ParentID) {
			if matchedNodeID := matchToolResultToolCallNodeID(payload, toolCallNodesByAssistant[payload.ParentID], allToolCallNodes); matchedNodeID != "" && matchedNodeID != parentID {
				originalParentID = parentID
				parentID = matchedNodeID
			}
		}

		builder.addNode(&OpenClawSessionGraphNode{
			ID:               payload.EntryID,
			ParentID:         parentID,
			OriginalParentID: originalParentID,
			EntryID:          payload.EntryID,
			ToolCallID:       payload.ToolCallID,
			Kind:             "tool_result",
			Timestamp:        payload.Timestamp,
			Summary:          payload.Summary,
			Tool:             payload.Tool,
			Query:            payload.Query,
			URL:              payload.URL,
			Path:             payload.Path,
			OK:               cloneBoolPointer(payload.OK),
			Error:            payload.Error,
			Text:             payload.Text,
		})
		appendGraphNodeEntryName(nodeIDsByEntryName, record.Entry, payload.EntryID)
		if assistantID := findAssistantStepIDForToolResult(parentID, toolCallNodeIDToAssistantID); assistantID != "" {
			if group := assistantGroups[assistantID]; group != nil {
				group.ToolResultNodeIDs = appendUniqueString(group.ToolResultNodeIDs, payload.EntryID)
			}
		}
	}

	addOpenClawControlFlowEdges(builder, assistantIDs, assistantGroups)
	markStoredGraphAnchor(builder, anchorPayload, anchorEntryName, nodeIDsByEntryName)
	return builder.finalize()
}

func addOpenClawControlFlowEdges(builder *openClawSessionGraphBuilder, assistantIDs []string, assistantGroups map[string]*openClawAssistantStepGroup) {
	if builder == nil {
		return
	}

	for _, node := range builder.nodes {
		if node == nil || node.Kind != "task" {
			continue
		}
		builder.addEdge(node.ParentID, node.ID)
	}

	for _, assistantID := range assistantIDs {
		group := assistantGroups[assistantID]
		if group == nil {
			continue
		}
		for _, toolCallNodeID := range group.ToolCallNodeIDs {
			builder.addEdge(assistantID, toolCallNodeID)
		}
	}

	for _, node := range builder.nodes {
		if node == nil || node.Kind != "tool_result" {
			continue
		}
		builder.addEdge(node.ParentID, node.ID)
	}

	downstreamRawParents := map[string][]string{}
	for _, node := range builder.nodes {
		if node == nil {
			continue
		}
		if node.Kind != "assistant_step" && node.Kind != "final" {
			continue
		}
		if parentID := strings.TrimSpace(node.ParentID); parentID != "" {
			downstreamRawParents[parentID] = appendUniqueString(downstreamRawParents[parentID], node.ID)
		}
	}

	joinedDownstreamNodeIDs := map[string]struct{}{}
	for _, assistantID := range assistantIDs {
		group := assistantGroups[assistantID]
		if group == nil {
			continue
		}

		toolCallNodeIDs := uniqueStrings(group.ToolCallNodeIDs)
		toolResultNodeIDs := uniqueStrings(group.ToolResultNodeIDs)
		if len(toolCallNodeIDs) <= 1 || len(toolResultNodeIDs) <= 1 {
			continue
		}

		downstreamNodeIDs := []string{}
		for _, toolResultNodeID := range toolResultNodeIDs {
			downstreamNodeIDs = appendUniqueStrings(downstreamNodeIDs, downstreamRawParents[toolResultNodeID])
		}
		if len(downstreamNodeIDs) == 0 {
			continue
		}

		joinID := fmt.Sprintf("join:%s", assistantID)
		joinTimestamp := ""
		for _, toolResultNodeID := range toolResultNodeIDs {
			if node := builder.nodes[toolResultNodeID]; node != nil {
				joinTimestamp = chooseLaterTimestamp(joinTimestamp, node.Timestamp)
			}
		}
		builder.addNode(&OpenClawSessionGraphNode{
			ID:        joinID,
			Kind:      "join",
			Timestamp: joinTimestamp,
			Summary:   "join",
		})
		for _, toolResultNodeID := range toolResultNodeIDs {
			builder.addEdge(toolResultNodeID, joinID)
		}
		for _, downstreamNodeID := range downstreamNodeIDs {
			builder.addEdge(joinID, downstreamNodeID)
			if downstreamNode := builder.nodes[downstreamNodeID]; downstreamNode != nil {
				downstreamNode.ParentID = joinID
			}
			joinedDownstreamNodeIDs[downstreamNodeID] = struct{}{}
		}
	}

	for _, node := range builder.nodes {
		if node == nil {
			continue
		}
		switch node.Kind {
		case "assistant_step", "final":
			if _, ok := joinedDownstreamNodeIDs[node.ID]; ok {
				continue
			}
			builder.addEdge(node.ParentID, node.ID)
		}
	}
}

func appendGraphNodeEntryName(index map[string][]string, entry *Entry, nodeID string) {
	if index == nil || entry == nil {
		return
	}

	entryName := strings.TrimSpace(entry.Name)
	nodeID = strings.TrimSpace(nodeID)
	if entryName == "" || nodeID == "" {
		return
	}

	for _, existingNodeID := range index[entryName] {
		if existingNodeID == nodeID {
			return
		}
	}
	index[entryName] = append(index[entryName], nodeID)
}

func matchToolResultToolCallNodeID(payload openClawBehaviorPayload, assistantToolCalls []*OpenClawSessionGraphNode, allToolCalls []*OpenClawSessionGraphNode) string {
	if matchedNodeID := chooseMatchingToolCallNodeID(payload, assistantToolCalls); matchedNodeID != "" {
		return matchedNodeID
	}

	if len(assistantToolCalls) != len(allToolCalls) {
		return chooseMatchingToolCallNodeID(payload, allToolCalls)
	}

	return ""
}

func chooseMatchingToolCallNodeID(payload openClawBehaviorPayload, candidates []*OpenClawSessionGraphNode) string {
	filtered := make([]*OpenClawSessionGraphNode, 0, len(candidates))
	seenNodeIDs := map[string]struct{}{}
	for _, candidate := range candidates {
		if candidate == nil || candidate.Kind != "tool_call" {
			continue
		}
		if _, ok := seenNodeIDs[candidate.ID]; ok {
			continue
		}
		seenNodeIDs[candidate.ID] = struct{}{}
		filtered = append(filtered, candidate)
	}

	if len(filtered) == 0 {
		return ""
	}
	if len(filtered) == 1 {
		return filtered[0].ID
	}

	filtered = refineToolCallCandidates(filtered, payload.Query, func(node *OpenClawSessionGraphNode) string { return node.Query })
	if len(filtered) == 1 {
		return filtered[0].ID
	}

	filtered = refineToolCallCandidates(filtered, payload.URL, func(node *OpenClawSessionGraphNode) string { return node.URL })
	if len(filtered) == 1 {
		return filtered[0].ID
	}

	filtered = refineToolCallCandidates(filtered, payload.Path, func(node *OpenClawSessionGraphNode) string { return node.Path })
	if len(filtered) == 1 {
		return filtered[0].ID
	}

	filtered = refineToolCallCandidates(filtered, payload.Tool, func(node *OpenClawSessionGraphNode) string { return node.Tool })
	if len(filtered) == 1 {
		return filtered[0].ID
	}

	return ""
}

func refineToolCallCandidates(candidates []*OpenClawSessionGraphNode, expected string, selector func(node *OpenClawSessionGraphNode) string) []*OpenClawSessionGraphNode {
	expected = strings.TrimSpace(expected)
	if expected == "" {
		return candidates
	}

	filtered := make([]*OpenClawSessionGraphNode, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.TrimSpace(selector(candidate)) == expected {
			filtered = append(filtered, candidate)
		}
	}
	if len(filtered) == 0 {
		return candidates
	}
	return filtered
}

func markStoredGraphAnchor(builder *openClawSessionGraphBuilder, anchorPayload openClawBehaviorPayload, anchorEntryName string, nodeIDsByEntryName map[string][]string) {
	anchorNodeID := ""

	if nodeIDs := nodeIDsByEntryName[strings.TrimSpace(anchorEntryName)]; len(nodeIDs) == 1 {
		anchorNodeID = nodeIDs[0]
	}

	if anchorNodeID == "" {
		switch anchorPayload.Kind {
		case "tool_call":
			candidates := []string{}
			for _, node := range builder.nodes {
				if !toolCallPayloadMatchesNode(anchorPayload, node) {
					continue
				}
				candidates = append(candidates, node.ID)
			}

			switch len(candidates) {
			case 1:
				anchorNodeID = candidates[0]
			default:
				anchorNodeID = anchorPayload.EntryID
			}
		default:
			if node := builder.nodes[anchorPayload.EntryID]; node != nil && node.Kind == anchorPayload.Kind {
				anchorNodeID = node.ID
			}
		}
	}

	if anchorNode := builder.nodes[anchorNodeID]; anchorNode != nil {
		anchorNode.IsAnchor = true
	}
}

func buildStoredToolCallNodeID(entry *Entry, payload openClawBehaviorPayload) string {
	if payload.ToolCallID != "" {
		return fmt.Sprintf("tool_call:%s", payload.ToolCallID)
	}
	if entry != nil && strings.TrimSpace(entry.Name) != "" {
		return fmt.Sprintf("tool_call_row:%s", strings.TrimSpace(entry.Name))
	}
	return fmt.Sprintf("tool_call:%s", strings.TrimSpace(payload.EntryID))
}

func newOpenClawSessionGraphBuilder() *openClawSessionGraphBuilder {
	return &openClawSessionGraphBuilder{
		graph: &OpenClawSessionGraph{
			Nodes: []*OpenClawSessionGraphNode{},
			Edges: []*OpenClawSessionGraphEdge{},
		},
		nodes:    map[string]*OpenClawSessionGraphNode{},
		edges:    []*OpenClawSessionGraphEdge{},
		edgeKeys: map[string]struct{}{},
	}
}

func (b *openClawSessionGraphBuilder) addNode(node *OpenClawSessionGraphNode) {
	if b == nil || node == nil {
		return
	}

	node.ID = strings.TrimSpace(node.ID)
	if node.ID == "" {
		return
	}

	if existing := b.nodes[node.ID]; existing != nil {
		mergeOpenClawGraphNode(existing, node)
		return
	}

	cloned := *node
	cloned.ParentID = strings.TrimSpace(cloned.ParentID)
	cloned.OriginalParentID = strings.TrimSpace(cloned.OriginalParentID)
	cloned.EntryID = strings.TrimSpace(cloned.EntryID)
	cloned.ToolCallID = strings.TrimSpace(cloned.ToolCallID)
	cloned.Kind = strings.TrimSpace(cloned.Kind)
	cloned.Timestamp = strings.TrimSpace(cloned.Timestamp)
	cloned.Summary = strings.TrimSpace(cloned.Summary)
	cloned.Tool = strings.TrimSpace(cloned.Tool)
	cloned.Query = strings.TrimSpace(cloned.Query)
	cloned.URL = strings.TrimSpace(cloned.URL)
	cloned.Path = strings.TrimSpace(cloned.Path)
	cloned.Error = strings.TrimSpace(cloned.Error)
	cloned.Text = strings.TrimSpace(cloned.Text)
	cloned.OK = cloneBoolPointer(cloned.OK)
	b.nodes[cloned.ID] = &cloned
}

func (b *openClawSessionGraphBuilder) addEdge(source, target string) {
	if b == nil {
		return
	}

	source = strings.TrimSpace(source)
	target = strings.TrimSpace(target)
	if source == "" || target == "" || source == target {
		return
	}

	key := fmt.Sprintf("%s->%s", source, target)
	if _, ok := b.edgeKeys[key]; ok {
		return
	}
	b.edgeKeys[key] = struct{}{}
	b.edges = append(b.edges, &OpenClawSessionGraphEdge{
		Source: source,
		Target: target,
	})
}

func (b *openClawSessionGraphBuilder) finalize() *OpenClawSessionGraph {
	if b == nil || b.graph == nil {
		return nil
	}

	nodeIDs := make([]string, 0, len(b.nodes))
	for id := range b.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	sort.Slice(nodeIDs, func(i, j int) bool {
		left := b.nodes[nodeIDs[i]]
		right := b.nodes[nodeIDs[j]]
		return compareGraphNodes(left, right) < 0
	})

	b.graph.Nodes = make([]*OpenClawSessionGraphNode, 0, len(nodeIDs))
	b.graph.Stats = OpenClawSessionGraphStats{}
	for _, id := range nodeIDs {
		node := b.nodes[id]
		b.graph.Nodes = append(b.graph.Nodes, node)
		updateOpenClawSessionGraphStats(&b.graph.Stats, node)
	}

	b.graph.Edges = make([]*OpenClawSessionGraphEdge, 0, len(b.edges))
	for _, edge := range b.edges {
		if edge == nil {
			continue
		}
		if b.nodes[edge.Source] == nil || b.nodes[edge.Target] == nil {
			continue
		}
		b.graph.Edges = append(b.graph.Edges, edge)
	}

	sort.Slice(b.graph.Edges, func(i, j int) bool {
		left := b.graph.Edges[i]
		right := b.graph.Edges[j]
		if left.Source != right.Source {
			return left.Source < right.Source
		}
		return left.Target < right.Target
	})

	return b.graph
}

func mergeOpenClawGraphNode(current, next *OpenClawSessionGraphNode) {
	if current == nil || next == nil {
		return
	}

	current.ParentID = firstNonEmpty(current.ParentID, next.ParentID)
	current.OriginalParentID = firstNonEmpty(current.OriginalParentID, next.OriginalParentID)
	current.EntryID = firstNonEmpty(current.EntryID, next.EntryID)
	current.ToolCallID = firstNonEmpty(current.ToolCallID, next.ToolCallID)
	current.Kind = firstNonEmpty(current.Kind, next.Kind)
	current.Timestamp = chooseEarlierTimestamp(current.Timestamp, next.Timestamp)
	current.Summary = firstNonEmpty(current.Summary, next.Summary)
	current.Tool = firstNonEmpty(current.Tool, next.Tool)
	current.Query = firstNonEmpty(current.Query, next.Query)
	current.URL = firstNonEmpty(current.URL, next.URL)
	current.Path = firstNonEmpty(current.Path, next.Path)
	current.Error = firstNonEmpty(current.Error, next.Error)
	current.Text = firstNonEmpty(current.Text, next.Text)
	current.OK = mergeBoolPointers(current.OK, next.OK)
	current.IsAnchor = current.IsAnchor || next.IsAnchor
}

func updateOpenClawSessionGraphStats(stats *OpenClawSessionGraphStats, node *OpenClawSessionGraphNode) {
	if stats == nil || node == nil {
		return
	}

	stats.TotalNodes++
	switch node.Kind {
	case "task":
		stats.TaskCount++
	case "tool_call":
		stats.ToolCallCount++
	case "tool_result":
		stats.ToolResultCount++
		if node.OK != nil && !*node.OK {
			stats.FailedCount++
		}
	case "final":
		stats.FinalCount++
	}
}

func findAssistantStepIDForToolResult(toolCallNodeID string, toolCallNodeIDToAssistantID map[string]string) string {
	toolCallNodeID = strings.TrimSpace(toolCallNodeID)
	if toolCallNodeID == "" {
		return ""
	}
	return toolCallNodeIDToAssistantID[toolCallNodeID]
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if strings.TrimSpace(existing) == value {
			return values
		}
	}
	return append(values, value)
}

func appendUniqueStrings(values []string, additions []string) []string {
	for _, addition := range additions {
		values = appendUniqueString(values, addition)
	}
	return values
}

func uniqueStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		result = appendUniqueString(result, value)
	}
	return result
}

func buildAssistantStepSummary(toolNames []string) string {
	deduped := []string{}
	seen := map[string]struct{}{}
	for _, toolName := range toolNames {
		toolName = strings.TrimSpace(toolName)
		if toolName == "" {
			continue
		}
		if _, ok := seen[toolName]; ok {
			continue
		}
		seen[toolName] = struct{}{}
		deduped = append(deduped, toolName)
	}

	if len(toolNames) == 0 {
		return "assistant step"
	}
	if len(deduped) == 0 {
		return fmt.Sprintf("%d tool calls", len(toolNames))
	}
	if len(deduped) <= 3 {
		return fmt.Sprintf("%d tool calls: %s", len(toolNames), strings.Join(deduped, ", "))
	}
	return fmt.Sprintf("%d tool calls: %s, ...", len(toolNames), strings.Join(deduped[:3], ", "))
}

func toolCallPayloadMatchesNode(payload openClawBehaviorPayload, node *OpenClawSessionGraphNode) bool {
	if node == nil || node.Kind != "tool_call" {
		return false
	}

	if payload.ToolCallID != "" {
		return strings.TrimSpace(node.ToolCallID) == strings.TrimSpace(payload.ToolCallID)
	}
	if strings.TrimSpace(node.EntryID) != strings.TrimSpace(payload.EntryID) {
		return false
	}

	fields := []struct {
		payload string
		node    string
	}{
		{payload.Tool, node.Tool},
		{payload.Query, node.Query},
		{payload.URL, node.URL},
		{payload.Path, node.Path},
		{payload.Text, node.Text},
	}

	matchedField := false
	for _, field := range fields {
		left := strings.TrimSpace(field.payload)
		if left == "" {
			continue
		}
		matchedField = true
		if left != strings.TrimSpace(field.node) {
			return false
		}
	}

	return matchedField
}

func compareGraphNodes(left, right *OpenClawSessionGraphNode) int {
	leftTimestamp := ""
	rightTimestamp := ""
	leftID := ""
	rightID := ""
	if left != nil {
		leftTimestamp = left.Timestamp
		leftID = left.ID
	}
	if right != nil {
		rightTimestamp = right.Timestamp
		rightID = right.ID
	}
	if timestampOrder := compareOpenClawGraphTimestamps(leftTimestamp, rightTimestamp); timestampOrder != 0 {
		return timestampOrder
	}
	if leftID < rightID {
		return -1
	}
	if leftID > rightID {
		return 1
	}
	return 0
}

func chooseEarlierTimestamp(current, next string) string {
	current = strings.TrimSpace(current)
	next = strings.TrimSpace(next)
	if current == "" {
		return next
	}
	if next == "" {
		return current
	}
	if compareOpenClawGraphTimestamps(next, current) < 0 {
		return next
	}
	return current
}

func chooseLaterTimestamp(current, next string) string {
	current = strings.TrimSpace(current)
	next = strings.TrimSpace(next)
	if current == "" {
		return next
	}
	if next == "" {
		return current
	}
	if compareOpenClawGraphTimestamps(next, current) > 0 {
		return next
	}
	return current
}

func compareOpenClawGraphTimestamps(left, right string) int {
	left = strings.TrimSpace(left)
	right = strings.TrimSpace(right)

	leftUnixNano, leftOK := parseOpenClawGraphTimestamp(left)
	rightUnixNano, rightOK := parseOpenClawGraphTimestamp(right)
	if leftOK && rightOK {
		if leftUnixNano < rightUnixNano {
			return -1
		}
		if leftUnixNano > rightUnixNano {
			return 1
		}
		return 0
	}

	if left < right {
		return -1
	}
	if left > right {
		return 1
	}
	return 0
}

func parseOpenClawGraphTimestamp(timestamp string) (_ int64, ok bool) {
	timestamp = strings.TrimSpace(timestamp)
	if timestamp == "" {
		return 0, false
	}

	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	return util.String2Time(timestamp).UnixNano(), true
}

func mergeBoolPointers(current, next *bool) *bool {
	if next == nil {
		return current
	}
	if current == nil {
		return cloneBoolPointer(next)
	}
	value := *current && *next
	return &value
}

func cloneBoolPointer(value *bool) *bool {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

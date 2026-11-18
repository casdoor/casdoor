// Copyright 2026 Casdoor Inc. All Rights Reserved.
//
// This file is part of the commercial and enterprise editions of Casdoor and is licensed under the Casdoor Enterprise
// License, not the Apache License 2.0. You may not use this file except in compliance with that license, see
// ee/LICENSE.

package ee

import (
	"fmt"
	"strings"
	"sync"

	"github.com/casdoor/casdoor/ee/gm"
	"github.com/casdoor/casdoor/object"
)

// The records of an organization with tamper-proof records form a hash chain: each record keeps the SM3 hash of the
// one before it and its own hash over its content and that previous hash. Changing a record breaks its own hash,
// deleting or inserting one breaks the link of the next. The chain is kept per Casdoor process: records written by
// several Casdoor nodes at once into one database interleave into separate chains, which the check reports as breaks.

var (
	recordChainMutex sync.Mutex
	recordChainHeads = map[string]string{}
)

func getRecordHash(record *object.Record, prevHash string) string {
	fields := []string{
		prevHash, record.Name, record.CreatedTime, record.ClientIp, record.User, record.Method, record.RequestUri,
		record.Action, record.Language, record.Object, record.Response, fmt.Sprintf("%d", record.StatusCode), record.Detail,
	}
	var sb strings.Builder
	for _, field := range fields {
		// the length prefix keeps the fields apart, whatever they contain
		sb.WriteString(fmt.Sprintf("%d:%s;", len(field), field))
	}
	return gm.Sm3Hex([]byte(sb.String()))
}

func getRecordChainHead(owner string) (string, error) {
	if head, ok := recordChainHeads[owner]; ok {
		return head, nil
	}
	record := &object.Record{}
	existed, err := object.GetOrmer().Engine.Cols("hash").Where("owner = ? and hash <> ?", owner, "").Desc("id").Get(record)
	if err != nil {
		return "", err
	}
	head := ""
	if existed {
		head = record.Hash
	}
	recordChainHeads[owner] = head
	return head, nil
}

// addChainedRecord inserts the record, chaining it to the organization's previous record when the organization keeps
// tamper-proof records
func addChainedRecord(record *object.Record) (int64, error) {
	policy := GetCachedEnterprisePolicy(record.Owner)
	if policy == nil || !policy.TamperProofRecords {
		return object.InsertRecord(record)
	}

	recordChainMutex.Lock()
	defer recordChainMutex.Unlock()

	prevHash, err := getRecordChainHead(record.Owner)
	if err != nil {
		return 0, err
	}
	record.PrevHash = prevHash
	record.Hash = getRecordHash(record, prevHash)

	affected, err := object.InsertRecord(record)
	if err != nil {
		delete(recordChainHeads, record.Owner)
		return affected, err
	}
	recordChainHeads[record.Owner] = record.Hash
	return affected, nil
}

type RecordChainBreak struct {
	Id          int    `json:"id"`
	CreatedTime string `json:"createdTime"`
	Reason      string `json:"reason"`
}

type RecordChainResult struct {
	Organization string              `json:"organization"`
	Checked      int                 `json:"checked"`
	FirstId      int                 `json:"firstId"`
	LastId       int                 `json:"lastId"`
	FirstTime    string              `json:"firstTime"`
	LastTime     string              `json:"lastTime"`
	Breaks       []*RecordChainBreak `json:"breaks"`
}

// VerifyRecordChain walks the organization's chained records in the order they were written and reports every record
// whose content was changed ("modified") or whose link to the record before it doesn't hold ("unlinked": a record was
// deleted, inserted or reordered there)
func VerifyRecordChain(owner string) (*RecordChainResult, error) {
	res := &RecordChainResult{Organization: owner, Breaks: []*RecordChainBreak{}}

	lastId := 0
	prevHash := ""
	for {
		records := []*object.Record{}
		err := object.GetOrmer().Engine.Where("owner = ? and hash <> ? and id > ?", owner, "", lastId).Asc("id").Limit(1000).Find(&records)
		if err != nil {
			return nil, err
		}
		if len(records) == 0 {
			return res, nil
		}

		for _, record := range records {
			if res.Checked == 0 {
				res.FirstId, res.FirstTime = record.Id, record.CreatedTime
			} else if record.PrevHash != prevHash {
				res.Breaks = append(res.Breaks, &RecordChainBreak{Id: record.Id, CreatedTime: record.CreatedTime, Reason: "unlinked"})
			}
			if getRecordHash(record, record.PrevHash) != record.Hash {
				res.Breaks = append(res.Breaks, &RecordChainBreak{Id: record.Id, CreatedTime: record.CreatedTime, Reason: "modified"})
			}

			prevHash = record.Hash
			res.Checked++
			res.LastId, res.LastTime = record.Id, record.CreatedTime
			lastId = record.Id
		}
	}
}

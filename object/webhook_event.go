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

package object

import (
	"fmt"

	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
	"github.com/xorm-io/core"
)

// WebhookEventStatus represents the delivery status of a webhook event
type WebhookEventStatus string

const (
	WebhookEventStatusPending   WebhookEventStatus = "pending"
	WebhookEventStatusSuccess   WebhookEventStatus = "success"
	WebhookEventStatusFailed    WebhookEventStatus = "failed"
	WebhookEventStatusRetrying  WebhookEventStatus = "retrying"
)

// WebhookEvent represents a webhook delivery event with retry and replay capability
type WebhookEvent struct {
	Owner       string `xorm:"varchar(100) notnull pk" json:"owner"`
	Name        string `xorm:"varchar(100) notnull pk" json:"name"`
	CreatedTime string `xorm:"varchar(100)" json:"createdTime"`
	UpdatedTime string `xorm:"varchar(100)" json:"updatedTime"`

	WebhookName  string             `xorm:"varchar(200) index" json:"webhookName"`
	Organization string             `xorm:"varchar(100) index" json:"organization"`
	EventType    string             `xorm:"varchar(100)" json:"eventType"`
	Status       WebhookEventStatus `xorm:"varchar(50) index" json:"status"`
	
	// Payload stores the event data (Record)
	Payload string `xorm:"mediumtext" json:"payload"`
	
	// Extended user data if applicable
	ExtendedUser string `xorm:"mediumtext" json:"extendedUser"`
	
	// Delivery tracking
	AttemptCount    int    `xorm:"int default 0" json:"attemptCount"`
	MaxRetries      int    `xorm:"int default 3" json:"maxRetries"`
	NextRetryTime   string `xorm:"varchar(100)" json:"nextRetryTime"`
	
	// Last delivery response
	LastStatusCode int    `xorm:"int" json:"lastStatusCode"`
	LastResponse   string `xorm:"mediumtext" json:"lastResponse"`
	LastError      string `xorm:"mediumtext" json:"lastError"`
}

func GetWebhookEvent(id string) (*WebhookEvent, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return nil, err
	}
	return getWebhookEvent(owner, name)
}

func getWebhookEvent(owner string, name string) (*WebhookEvent, error) {
	if owner == "" || name == "" {
		return nil, nil
	}

	event := WebhookEvent{Owner: owner, Name: name}
	existed, err := ormer.Engine.Get(&event)
	if err != nil {
		return &event, err
	}

	if existed {
		return &event, nil
	}
	return nil, nil
}

func GetWebhookEvents(owner, organization, webhookName string, status WebhookEventStatus, offset, limit int) ([]*WebhookEvent, error) {
	events := []*WebhookEvent{}
	session := ormer.Engine.Desc("created_time")
	
	if owner != "" {
		session = session.Where("owner = ?", owner)
	}
	if organization != "" {
		session = session.Where("organization = ?", organization)
	}
	if webhookName != "" {
		session = session.Where("webhook_name = ?", webhookName)
	}
	if status != "" {
		session = session.Where("status = ?", status)
	}
	
	if offset > 0 {
		session = session.Limit(limit, offset)
	} else if limit > 0 {
		session = session.Limit(limit)
	}
	
	err := session.Find(&events)
	if err != nil {
		return nil, err
	}
	
	return events, nil
}

func GetPendingWebhookEvents(limit int) ([]*WebhookEvent, error) {
	events := []*WebhookEvent{}
	currentTime := util.GetCurrentTime()
	
	err := ormer.Engine.
		Where("status = ? OR status = ?", WebhookEventStatusPending, WebhookEventStatusRetrying).
		And("(next_retry_time = '' OR next_retry_time <= ?)", currentTime).
		Asc("created_time").
		Limit(limit).
		Find(&events)
	
	if err != nil {
		return nil, err
	}
	
	return events, nil
}

func AddWebhookEvent(event *WebhookEvent) (bool, error) {
	if event.Name == "" {
		event.Name = util.GenerateId()
	}
	if event.CreatedTime == "" {
		event.CreatedTime = util.GetCurrentTime()
	}
	if event.UpdatedTime == "" {
		event.UpdatedTime = util.GetCurrentTime()
	}
	if event.Status == "" {
		event.Status = WebhookEventStatusPending
	}
	
	affected, err := ormer.Engine.Insert(event)
	if err != nil {
		return false, err
	}
	
	return affected != 0, nil
}

func UpdateWebhookEvent(id string, event *WebhookEvent) (bool, error) {
	owner, name, err := util.GetOwnerAndNameFromIdWithError(id)
	if err != nil {
		return false, err
	}
	
	event.UpdatedTime = util.GetCurrentTime()
	
	affected, err := ormer.Engine.ID(core.PK{owner, name}).AllCols().Update(event)
	if err != nil {
		return false, err
	}
	
	return affected != 0, nil
}

func UpdateWebhookEventStatus(event *WebhookEvent, status WebhookEventStatus, statusCode int, response string, err error) (bool, error) {
	event.Status = status
	event.LastStatusCode = statusCode
	event.LastResponse = response
	event.UpdatedTime = util.GetCurrentTime()
	
	if err != nil {
		event.LastError = err.Error()
	} else {
		event.LastError = ""
	}
	
	affected, dbErr := ormer.Engine.ID(core.PK{event.Owner, event.Name}).
		Cols("status", "last_status_code", "last_response", "last_error", "updated_time", "attempt_count", "next_retry_time").
		Update(event)
	
	if dbErr != nil {
		return false, dbErr
	}
	
	return affected != 0, nil
}

func DeleteWebhookEvent(event *WebhookEvent) (bool, error) {
	affected, err := ormer.Engine.ID(core.PK{event.Owner, event.Name}).Delete(&WebhookEvent{})
	if err != nil {
		return false, err
	}
	
	return affected != 0, nil
}

func (e *WebhookEvent) GetId() string {
	return fmt.Sprintf("%s/%s", e.Owner, e.Name)
}

// CreateWebhookEventFromRecord creates a webhook event from a record
func CreateWebhookEventFromRecord(webhook *Webhook, record *casvisorsdk.Record, extendedUser *User) (*WebhookEvent, error) {
	event := &WebhookEvent{
		Owner:        webhook.Owner,
		Name:         util.GenerateId(),
		CreatedTime:  util.GetCurrentTime(),
		UpdatedTime:  util.GetCurrentTime(),
		WebhookName:  webhook.GetId(),
		Organization: record.Organization,
		EventType:    record.Action,
		Status:       WebhookEventStatusPending,
		Payload:      util.StructToJson(record),
		AttemptCount: 0,
		MaxRetries:   3, // Default max retries
	}
	
	if extendedUser != nil {
		event.ExtendedUser = util.StructToJson(extendedUser)
	}
	
	_, err := AddWebhookEvent(event)
	if err != nil {
		return nil, err
	}
	
	return event, nil
}

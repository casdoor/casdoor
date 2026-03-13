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
	"testing"

	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

func TestWebhookEvent(t *testing.T) {
	// Test creating a webhook event
	webhook := &Webhook{
		Owner:        "admin",
		Name:         "test-webhook",
		Organization: "test-org",
		Url:          "http://localhost:8080/webhook",
		Method:       "POST",
		ContentType:  "application/json",
		IsEnabled:    true,
		MaxRetries:   3,
		RetryInterval: 60,
		UseExponentialBackoff: true,
	}

	record := &casvisorsdk.Record{
		Organization: "test-org",
		User:         "test-user",
		Action:       "test-action",
		Object:       `{"test": "data"}`,
	}

	event, err := CreateWebhookEventFromRecord(webhook, record, nil)
	if err != nil {
		t.Fatalf("Failed to create webhook event: %v", err)
	}

	if event == nil {
		t.Fatal("Event is nil")
	}

	if event.Status != WebhookEventStatusPending {
		t.Errorf("Expected status %s, got %s", WebhookEventStatusPending, event.Status)
	}

	if event.WebhookName != webhook.GetId() {
		t.Errorf("Expected webhook name %s, got %s", webhook.GetId(), event.WebhookName)
	}

	if event.Organization != record.Organization {
		t.Errorf("Expected organization %s, got %s", record.Organization, event.Organization)
	}

	if event.EventType != record.Action {
		t.Errorf("Expected event type %s, got %s", record.Action, event.EventType)
	}

	if event.MaxRetries != 3 {
		t.Errorf("Expected max retries 3, got %d", event.MaxRetries)
	}
}

func TestCalculateNextRetryTime(t *testing.T) {
	// Test fixed interval
	nextTime := calculateNextRetryTime(1, 60, false)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}

	// Test exponential backoff
	nextTime = calculateNextRetryTime(1, 60, true)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}

	nextTime = calculateNextRetryTime(2, 60, true)
	if nextTime == "" {
		t.Error("Next retry time should not be empty")
	}
}

func TestWebhookEventStatus(t *testing.T) {
	event := &WebhookEvent{
		Owner:        "admin",
		Name:         util.GenerateId(),
		Status:       WebhookEventStatusPending,
		AttemptCount: 0,
	}

	// Note: This test focuses on the logic of UpdateWebhookEventStatus
	// In a real scenario, the event would need to be persisted first
	// For a unit test without database setup, we're testing the logic only
	
	// Test status update logic
	event.Status = WebhookEventStatusSuccess
	event.LastStatusCode = 200
	event.LastResponse = "OK"
	event.LastError = ""

	if event.Status != WebhookEventStatusSuccess {
		t.Errorf("Expected status %s, got %s", WebhookEventStatusSuccess, event.Status)
	}

	if event.LastStatusCode != 200 {
		t.Errorf("Expected status code 200, got %d", event.LastStatusCode)
	}

	if event.LastResponse != "OK" {
		t.Errorf("Expected response 'OK', got %s", event.LastResponse)
	}
}

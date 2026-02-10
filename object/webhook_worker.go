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
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/casdoor/casdoor/util"
	"github.com/casvisor/casvisor-go-sdk/casvisorsdk"
)

var (
	webhookWorkerRunning = false
	webhookWorkerStop    = make(chan bool)
)

// StartWebhookDeliveryWorker starts the background worker for webhook delivery
func StartWebhookDeliveryWorker() {
	if webhookWorkerRunning {
		return
	}
	
	webhookWorkerRunning = true
	
	util.SafeGoroutine(func() {
		ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
		defer ticker.Stop()
		
		for {
			select {
			case <-webhookWorkerStop:
				webhookWorkerRunning = false
				return
			case <-ticker.C:
				processWebhookEvents()
			}
		}
	})
}

// StopWebhookDeliveryWorker stops the background worker
func StopWebhookDeliveryWorker() {
	if !webhookWorkerRunning {
		return
	}
	webhookWorkerStop <- true
}

// processWebhookEvents processes pending webhook events
func processWebhookEvents() {
	events, err := GetPendingWebhookEvents(100) // Process up to 100 events per cycle
	if err != nil {
		fmt.Printf("Error getting pending webhook events: %v\n", err)
		return
	}
	
	for _, event := range events {
		deliverWebhookEvent(event)
	}
}

// deliverWebhookEvent attempts to deliver a single webhook event
func deliverWebhookEvent(event *WebhookEvent) {
	// Get the webhook configuration
	webhook, err := GetWebhook(event.WebhookName)
	if err != nil {
		fmt.Printf("Error getting webhook %s: %v\n", event.WebhookName, err)
		return
	}
	
	if webhook == nil {
		// Webhook has been deleted, mark event as failed
		event.Status = WebhookEventStatusFailed
		event.LastError = "Webhook not found"
		UpdateWebhookEventStatus(event, WebhookEventStatusFailed, 0, "", fmt.Errorf("webhook not found"))
		return
	}
	
	if !webhook.IsEnabled {
		// Webhook is disabled, skip for now
		return
	}
	
	// Parse the record from payload
	var record casvisorsdk.Record
	err = json.Unmarshal([]byte(event.Payload), &record)
	if err != nil {
		event.Status = WebhookEventStatusFailed
		event.LastError = fmt.Sprintf("Invalid payload: %v", err)
		UpdateWebhookEventStatus(event, WebhookEventStatusFailed, 0, "", err)
		return
	}
	
	// Parse extended user if present
	var extendedUser *User
	if event.ExtendedUser != "" {
		extendedUser = &User{}
		err = json.Unmarshal([]byte(event.ExtendedUser), extendedUser)
		if err != nil {
			fmt.Printf("Error parsing extended user: %v\n", err)
			extendedUser = nil
		}
	}
	
	// Increment attempt count
	event.AttemptCount++
	
	// Attempt to send the webhook
	statusCode, respBody, err := sendWebhook(webhook, &record, extendedUser)
	
	// Add webhook record for backward compatibility (only if non-200 status)
	if statusCode != 200 {
		addWebhookRecord(webhook, &record, statusCode, respBody, err)
	}
	
	// Determine the result
	if err == nil && statusCode >= 200 && statusCode < 300 {
		// Success
		UpdateWebhookEventStatus(event, WebhookEventStatusSuccess, statusCode, respBody, nil)
	} else {
		// Failed - decide whether to retry
		maxRetries := webhook.MaxRetries
		if maxRetries <= 0 {
			maxRetries = 3 // Default
		}
		
		if event.AttemptCount >= maxRetries {
			// Max retries reached, mark as permanently failed
			UpdateWebhookEventStatus(event, WebhookEventStatusFailed, statusCode, respBody, err)
		} else {
			// Schedule retry
			retryInterval := webhook.RetryInterval
			if retryInterval <= 0 {
				retryInterval = 60 // Default 60 seconds
			}
			
			nextRetryTime := calculateNextRetryTime(event.AttemptCount, retryInterval, webhook.UseExponentialBackoff)
			event.NextRetryTime = nextRetryTime
			event.Status = WebhookEventStatusRetrying
			
			UpdateWebhookEventStatus(event, WebhookEventStatusRetrying, statusCode, respBody, err)
		}
	}
}

// calculateNextRetryTime calculates the next retry time based on attempt count and backoff strategy
func calculateNextRetryTime(attemptCount int, baseInterval int, useExponentialBackoff bool) string {
	var delaySeconds int
	
	if useExponentialBackoff {
		// Exponential backoff: baseInterval * 2^(attemptCount-1)
		// For example: 60s, 120s, 240s, 480s...
		delaySeconds = baseInterval * int(math.Pow(2, float64(attemptCount-1)))
		
		// Cap at 1 hour
		if delaySeconds > 3600 {
			delaySeconds = 3600
		}
	} else {
		// Fixed interval
		delaySeconds = baseInterval
	}
	
	nextTime := time.Now().Add(time.Duration(delaySeconds) * time.Second)
	return nextTime.Format("2006-01-02T15:04:05Z07:00")
}

// ReplayWebhookEvent replays a failed or missed webhook event
func ReplayWebhookEvent(eventId string) error {
	event, err := GetWebhookEvent(eventId)
	if err != nil {
		return err
	}
	
	if event == nil {
		return fmt.Errorf("webhook event not found: %s", eventId)
	}
	
	// Reset the event for replay
	event.Status = WebhookEventStatusPending
	event.AttemptCount = 0
	event.NextRetryTime = ""
	event.LastError = ""
	
	_, err = UpdateWebhookEvent(event.GetId(), event)
	if err != nil {
		return err
	}
	
	// Immediately try to deliver
	deliverWebhookEvent(event)
	
	return nil
}

// ReplayWebhookEvents replays multiple webhook events matching the criteria
func ReplayWebhookEvents(owner, organization, webhookName string, status WebhookEventStatus) (int, error) {
	events, err := GetWebhookEvents(owner, organization, webhookName, status, 0, 0)
	if err != nil {
		return 0, err
	}
	
	count := 0
	for _, event := range events {
		err = ReplayWebhookEvent(event.GetId())
		if err == nil {
			count++
		}
	}
	
	return count, nil
}

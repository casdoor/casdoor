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
	"net/http"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/util"
)

var (
	webhookWorkerMu        sync.Mutex
	webhookWorkerRunning   = false
	webhookWorkerStop      chan struct{}
	webhookPollingInterval = 30 * time.Second // Configurable polling interval
	webhookBatchSize       = 100              // Configurable batch size for processing events
	webhookMaxConcurrency  = 10               // webhooks delivered to at the same time, each one gets its events one by one

	webhookDeliveringMu sync.Mutex
	webhookDelivering   = map[string]bool{}
)

// StartWebhookDeliveryWorker starts the background worker for webhook delivery
func StartWebhookDeliveryWorker() {
	has, err := HasAnyWebhooks()
	if err != nil {
		logs.Error("failed to check webhooks, webhook delivery worker not started: " + err.Error())
		return
	}
	if !has {
		return
	}

	webhookWorkerMu.Lock()
	defer webhookWorkerMu.Unlock()

	if webhookWorkerRunning {
		return
	}

	stopCh := make(chan struct{})
	webhookWorkerStop = stopCh
	webhookWorkerRunning = true

	util.SafeGoroutine(func() {
		ticker := time.NewTicker(webhookPollingInterval)
		defer ticker.Stop()
		defer func() {
			webhookWorkerMu.Lock()
			defer webhookWorkerMu.Unlock()

			if webhookWorkerStop == stopCh {
				webhookWorkerRunning = false
				webhookWorkerStop = nil
			}
		}()

		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				processWebhookEvents()
			}
		}
	})
}

// StopWebhookDeliveryWorker stops the background worker
func StopWebhookDeliveryWorker() {
	webhookWorkerMu.Lock()
	defer webhookWorkerMu.Unlock()

	if !webhookWorkerRunning {
		return
	}

	if webhookWorkerStop == nil {
		webhookWorkerRunning = false
		return
	}

	close(webhookWorkerStop)
	webhookWorkerStop = nil
	webhookWorkerRunning = false
}

// processWebhookEvents processes pending webhook events, a slow or dead endpoint only delays its own events:
// each webhook is delivered to by its own goroutine, and the webhooks still being delivered to are skipped
func processWebhookEvents() {
	webhookDeliveringMu.Lock()
	deliveringWebhooks := []string{}
	for webhook := range webhookDelivering {
		deliveringWebhooks = append(deliveringWebhooks, webhook)
	}
	webhookDeliveringMu.Unlock()

	if len(deliveringWebhooks) >= webhookMaxConcurrency {
		return
	}

	events, err := GetPendingWebhookEvents(webhookBatchSize, deliveringWebhooks)
	if err != nil {
		logs.Error(fmt.Sprintf("failed to get pending webhook events: %v", err))
		return
	}

	webhooks := []string{}
	webhookEvents := map[string][]*WebhookEvent{}
	for _, event := range events {
		if _, ok := webhookEvents[event.Webhook]; !ok {
			webhooks = append(webhooks, event.Webhook)
		}
		webhookEvents[event.Webhook] = append(webhookEvents[event.Webhook], event)
	}

	for _, webhook := range webhooks {
		webhookDeliveringMu.Lock()
		if webhookDelivering[webhook] || len(webhookDelivering) >= webhookMaxConcurrency {
			webhookDeliveringMu.Unlock()
			continue
		}
		webhookDelivering[webhook] = true
		webhookDeliveringMu.Unlock()

		queue := webhookEvents[webhook]
		util.SafeGoroutine(func() {
			defer func() {
				webhookDeliveringMu.Lock()
				delete(webhookDelivering, webhook)
				webhookDeliveringMu.Unlock()
			}()

			// in order, and the rest waits for the next poll once the endpoint is unavailable
			for _, event := range queue {
				if !deliverWebhookEvent(event) {
					break
				}
			}
		})
	}
}

// deliverWebhookEvent attempts to deliver a single webhook event, it returns false if the endpoint is unavailable
func deliverWebhookEvent(event *WebhookEvent) bool {
	// Get the webhook configuration
	webhook, err := GetWebhook(event.Webhook)
	if err != nil {
		logs.Error(fmt.Sprintf("failed to get webhook %s: %v", event.Webhook, err))
		UpdateWebhookEventState(event, WebhookEventStatusFailed, 0, "", fmt.Errorf("get webhook: %w", err))
		return false
	}

	if webhook == nil {
		// Webhook has been deleted, mark event as failed
		event.State = WebhookEventStatusFailed
		event.LastError = "Webhook not found"
		UpdateWebhookEventState(event, WebhookEventStatusFailed, 0, "", fmt.Errorf("webhook not found"))
		return true
	}

	if !webhook.IsEnabled {
		// Disabled webhooks should finalize the event to avoid hot-looping forever.
		UpdateWebhookEventState(event, WebhookEventStatusFailed, 0, "", fmt.Errorf("webhook is disabled"))
		return true
	}

	// Parse the record from payload
	var record Record
	err = json.Unmarshal([]byte(event.Payload), &record)
	if err != nil {
		event.State = WebhookEventStatusFailed
		event.LastError = fmt.Sprintf("Invalid payload: %v", err)
		UpdateWebhookEventState(event, WebhookEventStatusFailed, 0, "", err)
		return true
	}

	// Parse extended user if present
	var extendedUser *User
	if event.ExtendedUser != "" {
		extendedUser = &User{}
		err = json.Unmarshal([]byte(event.ExtendedUser), extendedUser)
		if err != nil {
			logs.Warning(fmt.Sprintf("failed to parse extended user for webhook event %s: %v", event.GetId(), err))
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
		UpdateWebhookEventState(event, WebhookEventStatusSuccess, statusCode, respBody, nil)
	} else {
		// Failed - decide whether to retry
		maxRetries := event.MaxRetries
		if maxRetries <= 0 {
			maxRetries = webhook.MaxRetries
		}
		if maxRetries <= 0 {
			maxRetries = 3 // Default
		}
		event.MaxRetries = maxRetries

		if event.AttemptCount >= maxRetries {
			// Max retries reached, mark as permanently failed
			UpdateWebhookEventState(event, WebhookEventStatusFailed, statusCode, respBody, err)
		} else {
			// Schedule retry
			retryInterval := webhook.RetryInterval
			if retryInterval <= 0 {
				retryInterval = 60 // Default 60 seconds
			}

			nextRetryTime := calculateNextRetryTime(event.AttemptCount, retryInterval, webhook.UseExponentialBackoff)
			event.NextRetryTime = nextRetryTime
			event.State = WebhookEventStatusRetrying

			UpdateWebhookEventState(event, WebhookEventStatusRetrying, statusCode, respBody, err)
		}
	}

	return err == nil && statusCode < 500 && statusCode != http.StatusTooManyRequests
}

// calculateNextRetryTime calculates the next retry time based on attempt count and backoff strategy
func calculateNextRetryTime(attemptCount int, baseInterval int, useExponentialBackoff bool) string {
	var delaySeconds int

	if useExponentialBackoff {
		// Exponential backoff: baseInterval * 2^(attemptCount-1)
		// Cap attemptCount at 10 to prevent overflow
		cappedAttemptCount := attemptCount - 1
		if cappedAttemptCount > 10 {
			cappedAttemptCount = 10
		}

		// Calculate delay with overflow protection
		delaySeconds = baseInterval * (1 << uint(cappedAttemptCount))

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
	event.State = WebhookEventStatusPending
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

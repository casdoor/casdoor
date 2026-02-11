// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/redis/go-redis/v9"
)

const (
	PolicyChangeChannel = "casdoor:policy:change"
)

type PolicyChangeMessage struct {
	EnforcerId string    `json:"enforcerId"`
	Operation  string    `json:"operation"` // "add", "remove", "update", "reload"
	Timestamp  time.Time `json:"timestamp"`
	PodId      string    `json:"podId"` // Unique identifier for this pod instance
}

var (
	redisClient      *redis.Client
	policyPubSubOnce sync.Once
	podId            string
)

// InitPolicySynchronizer initializes the Redis-based policy synchronization mechanism
func InitPolicySynchronizer() error {
	redisEndpoint := conf.GetConfigString("redisEndpoint")
	if redisEndpoint == "" {
		logs.Info("Redis endpoint not configured, policy synchronization disabled")
		return nil
	}

	// Generate a unique pod ID for this instance using hostname or random ID
	podId = generatePodId()

	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: redisEndpoint,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logs.Warning("Failed to connect to Redis for policy sync: %v", err)
		return err
	}

	logs.Info("Policy synchronization initialized with pod ID: %s", podId)

	// Start listening for policy change notifications
	go subscribeToPolicyChanges()

	return nil
}

// subscribeToPolicyChanges listens for policy change notifications from other pods
func subscribeToPolicyChanges() {
	if redisClient == nil {
		return
	}

	ctx := context.Background()
	pubsub := redisClient.Subscribe(ctx, PolicyChangeChannel)
	defer pubsub.Close()

	logs.Info("Subscribed to policy changes on channel: %s", PolicyChangeChannel)

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			logs.Error("Error receiving policy change message: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var changeMsg PolicyChangeMessage
		if err := json.Unmarshal([]byte(msg.Payload), &changeMsg); err != nil {
			logs.Error("Error unmarshaling policy change message: %v", err)
			continue
		}

		// Ignore messages from this pod
		if changeMsg.PodId == podId {
			continue
		}

		logs.Info("Received policy change notification: enforcerId=%s, operation=%s, from pod=%s",
			changeMsg.EnforcerId, changeMsg.Operation, changeMsg.PodId)

		// Reload the enforcer to pick up the changes
		if err := reloadEnforcer(changeMsg.EnforcerId); err != nil {
			logs.Error("Failed to reload enforcer %s: %v", changeMsg.EnforcerId, err)
		}
	}
}

// publishPolicyChange publishes a policy change notification to other pods
func publishPolicyChange(enforcerId string, operation string) error {
	if redisClient == nil {
		// Redis not configured, skip notification
		return nil
	}

	changeMsg := PolicyChangeMessage{
		EnforcerId: enforcerId,
		Operation:  operation,
		Timestamp:  time.Now(),
		PodId:      podId,
	}

	msgBytes, err := json.Marshal(changeMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal policy change message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := redisClient.Publish(ctx, PolicyChangeChannel, msgBytes).Err(); err != nil {
		logs.Warning("Failed to publish policy change notification: %v", err)
		return err
	}

	logs.Info("Published policy change: enforcerId=%s, operation=%s", enforcerId, operation)
	return nil
}

// reloadEnforcer reloads a specific enforcer by its ID
func reloadEnforcer(enforcerId string) error {
	// Special handling for known global enforcers
	switch enforcerId {
	case UserEnforcerId:
		return reloadUserEnforcer()
	case "built-in/api-enforcer-built-in":
		return reloadApiEnforcer()
	default:
		// For other enforcers, we need to reload the enforcer instance
		// The enforcer is loaded on-demand, so the next call will pick up changes
		logs.Info("Enforcer %s will be reloaded on next use", enforcerId)
		return nil
	}
}

// reloadUserEnforcer reloads the global user enforcer
func reloadUserEnforcer() error {
	if userEnforcer == nil {
		return nil
	}

	enforcer, err := GetInitializedEnforcer(UserEnforcerId)
	if err != nil {
		return fmt.Errorf("failed to reload user enforcer: %w", err)
	}

	// Reload the policy from the database
	if err := enforcer.Enforcer.LoadPolicy(); err != nil {
		return fmt.Errorf("failed to load user enforcer policy: %w", err)
	}

	logs.Info("User enforcer reloaded successfully")
	return nil
}

// reloadApiEnforcer reloads the global API enforcer (if needed)
func reloadApiEnforcer() error {
	// API enforcer is in authz package, cannot directly reload from here
	// but it's initialized once and rarely changes
	logs.Info("API enforcer reload skipped (initialized once at startup)")
	return nil
}

// generatePodId generates a unique identifier for this pod instance
// It tries to use the hostname, falling back to a timestamp-based ID
func generatePodId() string {
	hostname := getHostname()
	if hostname != "" {
		// Use hostname without timestamp for stable identification across restarts
		return fmt.Sprintf("pod-%s", hostname)
	}
	// Fallback to timestamp-based ID if hostname is not available
	return fmt.Sprintf("pod-%d", time.Now().UnixNano())
}

// getHostname returns the hostname of the current machine
func getHostname() string {
	hostname, err := getHostnameFromEnv()
	if err != nil || hostname == "" {
		logs.Debug("Could not get hostname from environment: %v", err)
		return ""
	}
	// Clean hostname to make it a valid identifier
	return strings.ReplaceAll(hostname, ".", "-")
}

// getHostnameFromEnv tries to get hostname from Kubernetes environment variables
func getHostnameFromEnv() (string, error) {
	// Try to get pod name from Kubernetes environment variable
	if podName := getEnvVar("HOSTNAME"); podName != "" {
		return podName, nil
	}
	if podName := getEnvVar("POD_NAME"); podName != "" {
		return podName, nil
	}
	// Hostname not available from environment variables
	return "", fmt.Errorf("hostname not available from HOSTNAME or POD_NAME environment variables")
}

// getEnvVar is a helper to get environment variables (can be mocked in tests)
var getEnvVar = os.Getenv

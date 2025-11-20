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
	"strings"
	"testing"
)

// TestSmtpServerWithInvalidConfig tests that TestSmtpServer handles errors gracefully
// when proxy is enabled but not properly configured
func TestSmtpServerWithInvalidConfig(t *testing.T) {
	// Create a provider with invalid SMTP settings and proxy enabled
	provider := &Provider{
		ClientId:     "test@example.com",
		ClientSecret: "password",
		Host:         "invalid-smtp-server.example.com",
		Port:         587,
		Type:         "Default",
		DisableSsl:   false,
		EnableProxy:  true, // Proxy enabled but may not be configured
	}

	// This should not panic, but should return an error
	err := TestSmtpServer(provider)
	if err == nil {
		// It's okay if there's no error (e.g., if the server is somehow reachable)
		// The important thing is that it didn't panic
		t.Log("TestSmtpServer succeeded (server might be reachable)")
	} else {
		// Check that we got a proper error message, not a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Expected a normal error, but got a panic-related error: %v", err)
		} else {
			t.Logf("TestSmtpServer returned expected error: %v", err)
		}
	}
}

// TestSendEmailWithInvalidConfig tests that SendEmail handles errors gracefully
// when proxy is enabled but not properly configured
func TestSendEmailWithInvalidConfig(t *testing.T) {
	// Create a provider with invalid SMTP settings and proxy enabled
	provider := &Provider{
		ClientId:     "test@example.com",
		ClientSecret: "password",
		Host:         "invalid-smtp-server.example.com",
		Port:         587,
		Type:         "Default",
		DisableSsl:   false,
		EnableProxy:  true, // Proxy enabled but may not be configured
	}

	// This should not panic, but should return an error
	err := SendEmail(provider, "Test", "Test content", []string{"recipient@example.com"}, "sender@example.com")
	if err == nil {
		// It's okay if there's no error (e.g., if the server is somehow reachable)
		// The important thing is that it didn't panic
		t.Log("SendEmail succeeded (server might be reachable)")
	} else {
		// Check that we got a proper error message, not a panic
		if strings.Contains(err.Error(), "panic") {
			t.Errorf("Expected a normal error, but got a panic-related error: %v", err)
		} else {
			t.Logf("SendEmail returned expected error: %v", err)
		}
	}
}

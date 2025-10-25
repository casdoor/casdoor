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

package controllers

import (
	"fmt"
	"testing"
	"time"
)

func TestGenerateCacheKey(t *testing.T) {
	tests := []struct {
		name     string
		language string
		args     []string
		wantSame bool
	}{
		{
			name:     "Same language and args should produce same key",
			language: "go",
			args:     []string{"-v"},
			wantSame: true,
		},
		{
			name:     "Different args should produce different key",
			language: "go",
			args:     []string{"-h"},
			wantSame: false,
		},
		{
			name:     "Different language should produce different key",
			language: "java",
			args:     []string{"-v"},
			wantSame: false,
		},
	}

	baseKey, err := generateCacheKey("go", []string{"-v"})
	if err != nil {
		t.Fatalf("Failed to generate base cache key: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := generateCacheKey(tt.language, tt.args)
			if err != nil {
				t.Fatalf("Failed to generate cache key: %v", err)
			}
			if tt.wantSame && key != baseKey {
				t.Errorf("Expected same cache key, got %s vs %s", key, baseKey)
			}
			if !tt.wantSame && key == baseKey {
				t.Errorf("Expected different cache key, got %s", key)
			}
		})
	}
}

func TestCommandCache(t *testing.T) {
	// Clear the cache before testing
	commandCacheMutex.Lock()
	commandCache = make(map[string]*CommandCacheEntry)
	commandCacheMutex.Unlock()

	language := "go"
	args := []string{"-v"}
	cacheKey, err := generateCacheKey(language, args)
	if err != nil {
		t.Fatalf("Failed to generate cache key: %v", err)
	}
	expectedOutput := "test output"

	// Test cache miss
	if output, found := getCachedCommandResult(cacheKey); found {
		t.Errorf("Expected cache miss, got hit with output: %s", output)
	}

	// Test cache set
	setCachedCommandResult(cacheKey, expectedOutput)

	// Test cache hit
	if output, found := getCachedCommandResult(cacheKey); !found {
		t.Error("Expected cache hit, got miss")
	} else if output != expectedOutput {
		t.Errorf("Expected output %s, got %s", expectedOutput, output)
	}

	// Test cache expiration
	oldTTL := cacheTTL
	cacheTTL = 1 * time.Millisecond
	defer func() { cacheTTL = oldTTL }()

	time.Sleep(2 * time.Millisecond)

	if output, found := getCachedCommandResult(cacheKey); found {
		t.Errorf("Expected cache miss after expiration, got hit with output: %s", output)
	}
}

func TestCleanExpiredCacheEntries(t *testing.T) {
	// Clear the cache before testing
	commandCacheMutex.Lock()
	commandCache = make(map[string]*CommandCacheEntry)
	commandCacheMutex.Unlock()

	// Add some entries
	for i := 0; i < 5; i++ {
		key, _ := generateCacheKey("go", []string{fmt.Sprintf("arg%d", i)})
		setCachedCommandResult(key, fmt.Sprintf("output%d", i))
	}

	// Add some expired entries
	oldTTL := cacheTTL
	cacheTTL = 1 * time.Millisecond
	defer func() { cacheTTL = oldTTL }()

	for i := 5; i < 10; i++ {
		key, _ := generateCacheKey("go", []string{fmt.Sprintf("arg%d", i)})
		commandCacheMutex.Lock()
		commandCache[key] = &CommandCacheEntry{
			Output:     fmt.Sprintf("output%d", i),
			CachedTime: time.Now().Add(-2 * time.Minute),
		}
		commandCacheMutex.Unlock()
	}

	// Verify we have 10 entries
	commandCacheMutex.RLock()
	if len(commandCache) != 10 {
		t.Errorf("Expected 10 entries, got %d", len(commandCache))
	}
	commandCacheMutex.RUnlock()

	// Clean expired entries
	cleanExpiredCacheEntries()

	// Verify expired entries are removed
	commandCacheMutex.RLock()
	if len(commandCache) != 5 {
		t.Errorf("Expected 5 entries after cleanup, got %d", len(commandCache))
	}
	commandCacheMutex.RUnlock()
}

func TestConcurrentCacheAccess(t *testing.T) {
	// Clear the cache before testing
	commandCacheMutex.Lock()
	commandCache = make(map[string]*CommandCacheEntry)
	commandCacheMutex.Unlock()

	language := "go"
	args := []string{"-v"}
	cacheKey, err := generateCacheKey(language, args)
	if err != nil {
		t.Fatalf("Failed to generate cache key: %v", err)
	}

	// Test concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			setCachedCommandResult(cacheKey, "output")
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Test concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			getCachedCommandResult(cacheKey)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

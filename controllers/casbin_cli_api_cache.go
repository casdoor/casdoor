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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type CommandCacheEntry struct {
	Output     string
	CachedTime time.Time
}

var (
	commandCache      = make(map[string]*CommandCacheEntry)
	commandCacheMutex sync.RWMutex
	cacheTTL          = 5 * time.Minute
	cleanupInProgress = false
	cleanupMutex      sync.Mutex
)

// generateCacheKey creates a unique cache key based on language and arguments
func generateCacheKey(language string, args []string) (string, error) {
	argsJSON, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("failed to marshal args: %v", err)
	}
	data := fmt.Sprintf("%s:%s", language, string(argsJSON))
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:]), nil
}

// cleanExpiredCacheEntries removes expired entries from the cache
func cleanExpiredCacheEntries() {
	commandCacheMutex.Lock()
	defer commandCacheMutex.Unlock()

	for key, entry := range commandCache {
		if time.Since(entry.CachedTime) >= cacheTTL {
			delete(commandCache, key)
		}
	}

	cleanupMutex.Lock()
	cleanupInProgress = false
	cleanupMutex.Unlock()
}

// getCachedCommandResult retrieves cached command result if available and not expired
func getCachedCommandResult(cacheKey string) (string, bool) {
	commandCacheMutex.RLock()
	defer commandCacheMutex.RUnlock()

	if entry, exists := commandCache[cacheKey]; exists {
		if time.Since(entry.CachedTime) < cacheTTL {
			return entry.Output, true
		}
	}
	return "", false
}

// setCachedCommandResult stores command result in cache and performs periodic cleanup
func setCachedCommandResult(cacheKey string, output string) {
	commandCacheMutex.Lock()
	commandCache[cacheKey] = &CommandCacheEntry{
		Output:     output,
		CachedTime: time.Now(),
	}
	shouldCleanup := len(commandCache)%100 == 0
	commandCacheMutex.Unlock()

	// Periodically clean expired entries (every 100 cache sets)
	if shouldCleanup {
		cleanupMutex.Lock()
		if !cleanupInProgress {
			cleanupInProgress = true
			cleanupMutex.Unlock()
			go cleanExpiredCacheEntries()
		} else {
			cleanupMutex.Unlock()
		}
	}
}

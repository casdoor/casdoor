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

package controllers

import (
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
	"golang.org/x/time/rate"
)

// sendEmailLimiter tracks per-IP rate limits for the /api/send-email endpoint.
// Each IP is allowed at most 1 request per 24 hours to prevent abuse.
var (
	sendEmailLimiterMu sync.Mutex
	sendEmailLimiters  = map[string]*rate.Limiter{}
	sendEmailLastSeen  = map[string]time.Time{}
)

func isSendEmailRateLimitEnabled() bool {
	return conf.GetConfigBool("showGithubCorner")
}

func getSendEmailLimiter(ip string) *rate.Limiter {
	sendEmailLimiterMu.Lock()
	defer sendEmailLimiterMu.Unlock()

	// Evict stale entries (no request for more than 48 hours)
	for k, t := range sendEmailLastSeen {
		if time.Since(t) > 48*time.Hour {
			delete(sendEmailLimiters, k)
			delete(sendEmailLastSeen, k)
		}
	}

	limiter, exists := sendEmailLimiters[ip]
	if !exists {
		// Allow 1 request per 24 hours, burst of 1
		limiter = rate.NewLimiter(rate.Every(24*time.Hour), 1)
		sendEmailLimiters[ip] = limiter
	}
	sendEmailLastSeen[ip] = time.Now()
	return limiter
}

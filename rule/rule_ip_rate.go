// Copyright 2024 The casbin Authors. All Rights Reserved.
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

package rule

import (
	"net/http"
	"sync"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
	"golang.org/x/time/rate"
)

type IpRateRule struct {
	ruleName string
}

type IpRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

var (
	// checkRule() runs for every request in its own goroutine and an unsynchronized
	// concurrent map write is a fatal error that no HTTP handler can recover from,
	// so both package level maps are guarded.
	blackListMu = sync.Mutex{}
	blackList   = map[string]map[string]time.Time{}

	ipRateLimitersMu = sync.Mutex{}
	ipRateLimiters   = map[string]*IpRateLimiter{}
)

// NewIpRateLimiter .
func NewIpRateLimiter(r rate.Limit, b int) *IpRateLimiter {
	i := &IpRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}

	return i
}

// AddIP creates a new rate limiter and adds it to the ips map,
// using the IP address as the key
func (i *IpRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.ips[ip] = limiter

	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists.
// Otherwise, calls AddIP to add IP address to the map
func (i *IpRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}

// setLimit updates the rate limit of the limiter, reporting whether it changed.
func (i *IpRateLimiter) setLimit(r rate.Limit, b int) bool {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.r == r {
		return false
	}

	i.r = r
	i.b = b
	return true
}

// getIpRateLimiter returns the limiter of the rule, creating it when the rule has none yet.
func getIpRateLimiter(ruleName string, r rate.Limit, b int) *IpRateLimiter {
	ipRateLimitersMu.Lock()
	defer ipRateLimitersMu.Unlock()

	ipRateLimiter, found := ipRateLimiters[ruleName]
	if !found {
		ipRateLimiter = NewIpRateLimiter(r, b)
		ipRateLimiters[ruleName] = ipRateLimiter
	}

	return ipRateLimiter
}

// isBlocked reports whether the IP is still within the block time of the rule,
// removing its entry when the block time is over.
func isBlocked(ruleName string, ip string, blockTime time.Duration) bool {
	blackListMu.Lock()
	defer blackListMu.Unlock()

	createAt, found := blackList[ruleName][ip]
	if !found {
		return false
	}
	if time.Since(createAt) < blockTime {
		return true
	}

	delete(blackList[ruleName], ip)
	return false
}

// blockIp adds the IP to the black list of the rule, keeping the IPs that the rule
// already blocks.
func blockIp(ruleName string, ip string) {
	blackListMu.Lock()
	defer blackListMu.Unlock()

	if blackList[ruleName] == nil {
		blackList[ruleName] = map[string]time.Time{}
	}
	blackList[ruleName][ip] = time.Now()
}

func (r *IpRateRule) checkRule(expressions []*object.Expression, req *http.Request) (*RuleResult, error) {
	expression := expressions[0] // IpRate rule should have only one expression
	clientIp := util.GetClientIp(req)

	// If the client IP is in the blacklist, check the block time
	blockTime := time.Duration(util.ParseInt(expression.Value)) * time.Second
	if isBlocked(r.ruleName, clientIp, blockTime) {
		return &RuleResult{
			Action: "Block",
			Reason: "Rate limit exceeded",
		}, nil
	}

	// If the client IP is not in the blacklist, check the rate limit
	parseInt := util.ParseInt(expression.Operator)
	ipRateLimiter := getIpRateLimiter(r.ruleName, rate.Limit(parseInt), parseInt)

	// If the rate limit has changed, update the rate limiter
	limiter := ipRateLimiter.GetLimiter(clientIp)
	if ipRateLimiter.setLimit(rate.Limit(parseInt), parseInt) {
		limiter.SetLimit(rate.Limit(parseInt))
		limiter.SetBurst(parseInt)
		err := limiter.Wait(req.Context())
		if err != nil {
			return nil, err
		}
	} else {
		// If the rate limit is exceeded, add the client IP to the blacklist
		allow := limiter.Allow()
		if !allow {
			blockIp(r.ruleName, clientIp)
			return &RuleResult{
				Action: "Block",
				Reason: "Rate limit exceeded",
			}, nil
		}
	}

	return nil, nil
}

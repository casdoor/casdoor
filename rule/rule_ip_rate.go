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

var blackList = map[string]map[string]time.Time{}

var ipRateLimiters = map[string]*IpRateLimiter{}

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

func (r *IpRateRule) checkRule(expressions []*object.Expression, req *http.Request) (*RuleResult, error) {
	expression := expressions[0] // IpRate rule should have only one expression
	clientIp := util.GetClientIp(req)

	// If the client IP is in the blacklist, check the block time
	createAt, ok := blackList[r.ruleName][clientIp]
	if ok {
		blockTime := util.ParseInt(expression.Value)
		if time.Now().Sub(createAt) < time.Duration(blockTime)*time.Second {
			return &RuleResult{
				Action: "Block",
				Reason: "Rate limit exceeded",
			}, nil
		} else {
			delete(blackList[r.ruleName], clientIp)
		}
	}

	// If the client IP is not in the blacklist, check the rate limit
	ipRateLimiter := ipRateLimiters[r.ruleName]
	parseInt := util.ParseInt(expression.Operator)
	if ipRateLimiter == nil {
		ipRateLimiter = NewIpRateLimiter(rate.Limit(parseInt), parseInt)
		ipRateLimiters[r.ruleName] = ipRateLimiter
	}

	// If the rate limit has changed, update the rate limiter
	limiter := ipRateLimiter.GetLimiter(clientIp)
	if ipRateLimiter.r != rate.Limit(parseInt) {
		ipRateLimiter.r = rate.Limit(parseInt)
		ipRateLimiter.b = parseInt
		limiter.SetLimit(ipRateLimiter.r)
		limiter.SetBurst(ipRateLimiter.b)
		err := limiter.Wait(req.Context())
		if err != nil {
			return nil, err
		}
	} else {
		// If the rate limit is exceeded, add the client IP to the blacklist
		allow := limiter.Allow()
		if !allow {
			blackList[r.ruleName] = map[string]time.Time{}
			blackList[r.ruleName][clientIp] = time.Now()
			return &RuleResult{
				Action: "Block",
				Reason: "Rate limit exceeded",
			}, nil
		}
	}

	return nil, nil
}

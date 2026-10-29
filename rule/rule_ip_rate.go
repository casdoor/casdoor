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

const ipRatePruneInterval = time.Minute

type IpRateRule struct {
	ruleName string
}

type ipRateState struct {
	limit     int
	limiters  map[string]*rate.Limiter
	blackList map[string]time.Time
	pruneTime time.Time
}

var (
	ipRateMu     sync.Mutex
	ipRateStates = map[string]*ipRateState{}
)

func (s *ipRateState) prune(now time.Time, blockTime time.Duration) {
	for ip, createAt := range s.blackList {
		if now.Sub(createAt) >= blockTime {
			delete(s.blackList, ip)
		}
	}

	// A limiter with a full bucket behaves the same as a new one
	for ip, limiter := range s.limiters {
		if limiter.TokensAt(now) >= float64(s.limit) {
			delete(s.limiters, ip)
		}
	}

	s.pruneTime = now
}

func allowIpRate(ruleName string, ip string, limit int, blockTime time.Duration) bool {
	ipRateMu.Lock()
	defer ipRateMu.Unlock()

	now := time.Now()
	state, ok := ipRateStates[ruleName]
	if !ok {
		state = &ipRateState{
			limit:     limit,
			limiters:  map[string]*rate.Limiter{},
			blackList: map[string]time.Time{},
			pruneTime: now,
		}
		ipRateStates[ruleName] = state
	} else if state.limit != limit {
		state.limit = limit
		state.limiters = map[string]*rate.Limiter{}
	}

	if now.Sub(state.pruneTime) >= ipRatePruneInterval {
		state.prune(now, blockTime)
	}

	if createAt, ok := state.blackList[ip]; ok {
		if now.Sub(createAt) < blockTime {
			return false
		}
		delete(state.blackList, ip)
	}

	limiter, ok := state.limiters[ip]
	if !ok {
		limiter = rate.NewLimiter(rate.Limit(limit), limit)
		state.limiters[ip] = limiter
	}

	if !limiter.AllowN(now, 1) {
		state.blackList[ip] = now
		return false
	}

	return true
}

func (r *IpRateRule) checkRule(expressions []*object.Expression, req *http.Request) (*RuleResult, error) {
	expression := expressions[0] // IpRate rule should have only one expression
	clientIp := util.GetClientIp(req)
	limit := util.ParseInt(expression.Operator)
	blockTime := time.Duration(util.ParseInt(expression.Value)) * time.Second

	if !allowIpRate(r.ruleName, clientIp, limit, blockTime) {
		return &RuleResult{
			Action: "Block",
			Reason: "Rate limit exceeded",
		}, nil
	}

	return nil, nil
}

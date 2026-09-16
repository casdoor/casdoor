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
	"context"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/redis/go-redis/v9"
)

// samlAssertionStore remembers the IDs of the SAML assertions already consumed until
// they expire, so that a captured SAMLResponse cannot be replayed to sign in again.
type samlAssertionStore interface {
	// MarkUsed records the assertion ID and reports whether it had been consumed before.
	MarkUsed(id string, ttl time.Duration) (bool, error)
}

const samlAssertionRedisPrefix = "casdoor:saml_assertion:"

var SamlAssertionStore samlAssertionStore = &memorySamlAssertionStore{expiries: map[string]time.Time{}}

// InitSamlAssertionStore switches SamlAssertionStore to Redis when it is configured, so
// that replay detection also holds across multiple Casdoor replicas.
func InitSamlAssertionStore() {
	config := conf.GetRedisConfig()
	if config == nil {
		return
	}

	addrs := strings.Join(config.Addrs, ";")
	client, err := newRedisClient(config)
	if err != nil {
		logs.Warn("saml_assertion_store: failed to connect to Redis (%s), falling back to in-memory store: %v", addrs, err)
		return
	}

	SamlAssertionStore = &redisSamlAssertionStore{client: client}
	logs.Info("saml_assertion_store: using Redis backend at %s", addrs)
}

type memorySamlAssertionStore struct {
	mu       sync.Mutex
	expiries map[string]time.Time
}

func (s *memorySamlAssertionStore) MarkUsed(id string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for usedId, expiry := range s.expiries {
		if !now.Before(expiry) {
			delete(s.expiries, usedId)
		}
	}

	if _, ok := s.expiries[id]; ok {
		return true, nil
	}
	s.expiries[id] = now.Add(ttl)
	return false, nil
}

type redisSamlAssertionStore struct {
	client redis.Cmdable
}

func (s *redisSamlAssertionStore) MarkUsed(id string, ttl time.Duration) (bool, error) {
	isNew, err := s.client.SetNX(context.Background(), samlAssertionRedisPrefix+id, "1", ttl).Result()
	if err != nil {
		return false, err
	}
	return !isNew, nil
}

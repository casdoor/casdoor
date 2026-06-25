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
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/casdoor/casdoor/conf"
	"github.com/redis/go-redis/v9"
)

// DeviceAuthMap stores the state of the OAuth 2.0 Device Authorization Grant (RFC 8628).
//
// It mirrors the subset of sync.Map methods used by the device flow. When "redisEndpoint"
// is configured it is backed by Redis so that the three steps of the device flow (requesting
// the device code, the user approving it in a browser, and the client polling the token
// endpoint) keep working when Casdoor runs as multiple replicas behind a load balancer.
// Otherwise it falls back to an in-process sync.Map for single-instance deployments.
var DeviceAuthMap deviceAuthMapInterface = &memoryDeviceAuthMap{}

const deviceAuthRedisKeyPrefix = "casdoor:device_auth:"

type deviceAuthMapInterface interface {
	Load(key any) (any, bool)
	Store(key, value any)
	Delete(key any)
	LoadAndDelete(key any) (any, bool)
	Range(f func(key, value any) bool)
}

// InitDeviceAuthStore switches the device auth store to Redis when "redisEndpoint" is configured.
// It must be called from main() after the configuration is loaded and before serving requests.
func InitDeviceAuthStore() {
	redisEndpoint := conf.GetConfigString("redisEndpoint")
	if redisEndpoint == "" {
		return
	}

	client, err := newDeviceAuthRedisClient(redisEndpoint)
	if err != nil {
		panic(err)
	}

	DeviceAuthMap = &redisDeviceAuthMap{client: client}
}

func newDeviceAuthRedisClient(redisEndpoint string) (*redis.Client, error) {
	addr := redisEndpoint
	db := 0
	password := ""

	// redisEndpoint follows the same format as the beego redis session provider: "host:port,db"
	// with an optional ",password" suffix.
	if comma := strings.Index(redisEndpoint, ","); comma >= 0 {
		addr = redisEndpoint[:comma]
		rest := redisEndpoint[comma+1:]
		parts := strings.SplitN(rest, ",", 2)
		if d, err := strconv.Atoi(parts[0]); err == nil {
			db = d
		}
		if len(parts) > 1 {
			password = parts[1]
		}
	}

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       db,
		Password: password,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// memoryDeviceAuthMap is the single-instance fallback, wrapping a plain sync.Map.

type memoryDeviceAuthMap struct {
	m sync.Map
}

func (s *memoryDeviceAuthMap) Load(key any) (any, bool)          { return s.m.Load(key) }
func (s *memoryDeviceAuthMap) Store(key, value any)              { s.m.Store(key, value) }
func (s *memoryDeviceAuthMap) Delete(key any)                    { s.m.Delete(key) }
func (s *memoryDeviceAuthMap) LoadAndDelete(key any) (any, bool) { return s.m.LoadAndDelete(key) }
func (s *memoryDeviceAuthMap) Range(f func(key, value any) bool) { s.m.Range(f) }

// redisDeviceAuthMap stores DeviceAuthCache entries as JSON values with a TTL derived from
// ExpiresIn (falling back to DeviceAuthExpiresIn), so entries expire automatically without a
// background cleanup job.

type redisDeviceAuthMap struct {
	client *redis.Client
}

func (s *redisDeviceAuthMap) Load(key any) (any, bool) {
	k, ok := key.(string)
	if !ok {
		return nil, false
	}

	data, err := s.client.Get(context.Background(), deviceAuthRedisKeyPrefix+k).Bytes()
	if err != nil {
		return nil, false
	}

	return decodeDeviceAuthCache(data)
}

func (s *redisDeviceAuthMap) Store(key, value any) {
	k, ok := key.(string)
	if !ok {
		return
	}

	cache, ok := value.(DeviceAuthCache)
	if !ok {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		return
	}

	ttl := cache.ExpiresIn
	if ttl <= 0 {
		ttl = DeviceAuthExpiresIn
	}

	s.client.Set(context.Background(), deviceAuthRedisKeyPrefix+k, data, time.Duration(ttl)*time.Second)
}

func (s *redisDeviceAuthMap) Delete(key any) {
	k, ok := key.(string)
	if !ok {
		return
	}

	s.client.Del(context.Background(), deviceAuthRedisKeyPrefix+k)
}

func (s *redisDeviceAuthMap) LoadAndDelete(key any) (any, bool) {
	k, ok := key.(string)
	if !ok {
		return nil, false
	}

	data, err := s.client.GetDel(context.Background(), deviceAuthRedisKeyPrefix+k).Bytes()
	if err != nil {
		return nil, false
	}

	return decodeDeviceAuthCache(data)
}

// Range is a no-op for the Redis backend: entries expire on their own via the Redis key TTL,
// so the periodic sweep performed by InitCleanupDeviceAuthMap has nothing to do here.
func (s *redisDeviceAuthMap) Range(f func(key, value any) bool) {}

func decodeDeviceAuthCache(data []byte) (DeviceAuthCache, bool) {
	var cache DeviceAuthCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return DeviceAuthCache{}, false
	}

	return cache, true
}

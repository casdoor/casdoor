// Copyright 2025 The Casdoor Authors. All Rights Reserved.
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
	"io"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/casdoor/casdoor/conf"
	"github.com/redis/go-redis/v9"
)

// deviceAuthStore mirrors the sync.Map methods used by the device authorization flow.
// The default implementation is in-memory; when Redis is configured, a Redis-backed
// implementation is used so the flow works correctly across multiple Casdoor replicas.
type deviceAuthStore interface {
	Load(key any) (any, bool)
	Store(key, value any)
	Delete(key any)
	LoadAndDelete(key any) (any, bool)
	Range(f func(key, value any) bool)
}

const deviceAuthRedisPrefix = "casdoor:device_auth:"

// InitDeviceAuthStore switches DeviceAuthMap to a Redis-backed store when Redis is
// configured. It must be called after configuration is loaded and before serving requests.
// On failure it logs a warning and keeps the default in-memory store.
func InitDeviceAuthStore() {
	config := conf.GetRedisConfig()
	if config == nil {
		return
	}

	addrs := strings.Join(config.Addrs, ";")
	client, err := newRedisClient(config)
	if err != nil {
		logs.Warn("device_auth_store: failed to connect to Redis (%s), falling back to in-memory store: %v", addrs, err)
		return
	}

	DeviceAuthMap = &redisDeviceAuthStore{client: client}
	logs.Info("device_auth_store: using Redis backend at %s", addrs)
}

// newRedisClient returns a cluster client when redisEndpoint lists several addresses,
// otherwise a single-node client. Both satisfy redis.Cmdable.
func newRedisClient(config *conf.RedisConfig) (redis.Cmdable, error) {
	var client redis.Cmdable
	if config.IsCluster {
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    config.Addrs,
			Password: config.Password,
			PoolSize: config.PoolSize,
		})
	} else {
		client = redis.NewClient(&redis.Options{
			Addr:     config.Addrs[0],
			Password: config.Password,
			PoolSize: config.PoolSize,
			DB:       config.Db,
		})
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		if closer, ok := client.(io.Closer); ok {
			_ = closer.Close()
		}
		return nil, err
	}

	return client, nil
}

// ── in-memory implementation (default) ──────────────────────────────────────

type memoryDeviceAuthStore struct {
	m sync.Map
}

func (s *memoryDeviceAuthStore) Load(key any) (any, bool)          { return s.m.Load(key) }
func (s *memoryDeviceAuthStore) Store(key, value any)              { s.m.Store(key, value) }
func (s *memoryDeviceAuthStore) Delete(key any)                    { s.m.Delete(key) }
func (s *memoryDeviceAuthStore) LoadAndDelete(key any) (any, bool) { return s.m.LoadAndDelete(key) }
func (s *memoryDeviceAuthStore) Range(f func(key, value any) bool) { s.m.Range(f) }

// ── Redis implementation ─────────────────────────────────────────────────────

type redisDeviceAuthStore struct {
	client redis.Cmdable
}

func (s *redisDeviceAuthStore) redisKey(key any) (string, bool) {
	k, ok := key.(string)
	return deviceAuthRedisPrefix + k, ok
}

func (s *redisDeviceAuthStore) Load(key any) (any, bool) {
	rk, ok := s.redisKey(key)
	if !ok {
		return nil, false
	}

	data, err := s.client.Get(context.Background(), rk).Bytes()
	if err != nil {
		return nil, false
	}

	var cache DeviceAuthCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}
	return cache, true
}

func (s *redisDeviceAuthStore) Store(key, value any) {
	rk, ok := s.redisKey(key)
	if !ok {
		return
	}

	cache, ok := value.(DeviceAuthCache)
	if !ok {
		return
	}

	data, err := json.Marshal(cache)
	if err != nil {
		logs.Warn("device_auth_store: failed to marshal DeviceAuthCache: %v", err)
		return
	}

	ttl := cache.ExpiresIn
	if ttl <= 0 {
		ttl = DeviceAuthExpiresIn
	}

	if err := s.client.Set(context.Background(), rk, data, time.Duration(ttl)*time.Second).Err(); err != nil {
		logs.Warn("device_auth_store: Redis SET failed for key %s: %v", rk, err)
	}
}

func (s *redisDeviceAuthStore) Delete(key any) {
	rk, ok := s.redisKey(key)
	if !ok {
		return
	}

	if err := s.client.Del(context.Background(), rk).Err(); err != nil {
		logs.Warn("device_auth_store: Redis DEL failed for key %s: %v", rk, err)
	}
}

func (s *redisDeviceAuthStore) LoadAndDelete(key any) (any, bool) {
	rk, ok := s.redisKey(key)
	if !ok {
		return nil, false
	}

	// GetDel is atomic and available since Redis 6.2. For older Redis the worst case is a
	// small window between GET and DEL where a concurrent request sees the same entry; that
	// is acceptable given the short device-auth TTL.
	data, err := s.client.GetDel(context.Background(), rk).Bytes()
	if err != nil {
		return nil, false
	}

	var cache DeviceAuthCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}
	return cache, true
}

// Range is a no-op for the Redis backend: entries expire automatically via TTL,
// so the periodic sweep in InitCleanupDeviceAuthMap is not needed.
func (s *redisDeviceAuthStore) Range(_ func(key, value any) bool) {}

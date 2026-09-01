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

package conf

import (
	"strconv"
	"strings"
)

const defaultRedisPoolSize = 1000

// RedisConfig is the parsed "redisEndpoint" config item, shared by the session store and the
// device authorization store.
type RedisConfig struct {
	IsCluster bool
	Addrs     []string
	PoolSize  int
	Password  string
	Db        int

	endpoint string
}

// GetRedisConfig parses "redisEndpoint" and returns nil when Redis is not configured. The
// format is beego's "host:port[,poolSize[,password[,db]]]". Listing several addresses in the
// first field, separated by ";", turns on Redis Cluster mode, e.g.
// "10.0.0.1:6379;10.0.0.2:6379,100,password".
func GetRedisConfig() *RedisConfig {
	endpoint := GetConfigString("redisEndpoint")
	if endpoint == "" {
		return nil
	}

	config := &RedisConfig{PoolSize: defaultRedisPoolSize, endpoint: endpoint}

	parts := strings.Split(endpoint, ",")
	for _, addr := range strings.Split(parts[0], ";") {
		addr = strings.TrimSpace(addr)
		if addr != "" {
			config.Addrs = append(config.Addrs, addr)
		}
	}
	if len(config.Addrs) == 0 {
		return nil
	}
	config.IsCluster = len(config.Addrs) > 1

	if len(parts) > 1 {
		if poolSize, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil && poolSize > 0 {
			config.PoolSize = poolSize
		}
	}
	if len(parts) > 2 {
		config.Password = strings.TrimSpace(parts[2])
	}
	if len(parts) > 3 {
		if db, err := strconv.Atoi(strings.TrimSpace(parts[3])); err == nil && db > 0 {
			config.Db = db
		}
	}

	return config
}

// GetSessionProvider returns the beego session provider name and its config string. Both
// providers accept the same "redisEndpoint" format, so it is passed through unchanged.
func (config *RedisConfig) GetSessionProvider() (string, string) {
	if config.IsCluster {
		return "redis_cluster", config.endpoint
	}
	return "redis", config.endpoint
}

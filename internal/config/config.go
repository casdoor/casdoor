// Copyright 2020 The casbin Authors. All Rights Reserved.
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

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds shared configuration on server.
type Config struct {
	HTTPPort     string
	DBDataSource string
}

// NewConfig reads shared configuration from .env file.
func NewConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("godotenv.Load: %v", err)
	}

	cfg := &Config{
		DBDataSource: os.Getenv("DB_DATABASE_SOURCE"),
		HTTPPort:     os.Getenv("HTTP_PORT"),
	}

	if cfg.DBDataSource == "" {
		return nil, errors.New("DB_DATABASE_SOURCE is empty")
	}

	return cfg, nil
}

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

package mcp

import (
	"context"
	"time"

	"github.com/casdoor/casdoor/util"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"golang.org/x/oauth2"
)

func GetServerTools(owner, name, url, token string) ([]*mcpsdk.Tool, error) {
	var session *mcpsdk.ClientSession
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel()
	client := mcpsdk.NewClient(&mcpsdk.Implementation{Name: util.GetId(owner, name), Version: "1.0.0"}, nil)
	if token != "" {
		httpClient := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
		session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url, HTTPClient: httpClient}, nil)
	} else {
		session, err = client.Connect(ctx, &mcpsdk.StreamableClientTransport{Endpoint: url}, nil)
	}

	if err != nil {
		return nil, err
	}
	defer session.Close()

	toolResult, err := session.ListTools(ctx, nil)
	if err != nil {
		return nil, err
	}

	return toolResult.Tools, nil
}

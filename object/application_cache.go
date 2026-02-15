// Copyright 2021 The Casdoor Authors. All Rights Reserved.
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
	"fmt"
	"strings"
	"sync"

	"github.com/casdoor/casdoor/proxy"
)

var (
	applicationMap      = make(map[string]*Application)
	applicationMapMutex sync.RWMutex
)

func InitApplicationMap() error {
	// Set up the application lookup function for the proxy package
	proxy.SetApplicationLookup(func(domain string) *proxy.Application {
		app := GetApplicationByDomain(domain)
		if app == nil {
			return nil
		}
		return &proxy.Application{
			Owner:        app.Owner,
			Name:         app.Name,
			UpstreamHost: app.UpstreamHost,
		}
	})

	return refreshApplicationMap()
}

func refreshApplicationMap() error {
	applications, err := GetGlobalApplications()
	if err != nil {
		return fmt.Errorf("failed to get global applications: %w", err)
	}

	newApplicationMap := make(map[string]*Application)
	for _, app := range applications {
		if app.Domain != "" {
			newApplicationMap[strings.ToLower(app.Domain)] = app
		}
		for _, domain := range app.OtherDomains {
			if domain != "" {
				newApplicationMap[strings.ToLower(domain)] = app
			}
		}
	}

	applicationMapMutex.Lock()
	applicationMap = newApplicationMap
	applicationMapMutex.Unlock()

	return nil
}

func GetApplicationByDomain(domain string) *Application {
	applicationMapMutex.RLock()
	defer applicationMapMutex.RUnlock()

	domain = strings.ToLower(domain)
	if app, ok := applicationMap[domain]; ok {
		return app
	}
	return nil
}

func RefreshApplicationCache() error {
	return refreshApplicationMap()
}

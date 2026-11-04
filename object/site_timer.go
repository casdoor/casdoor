// Copyright 2023 The casbin Authors. All Rights Reserved.
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
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
)

var (
	siteUpdateMap = map[string]string{}
	lock          = &sync.Mutex{}
)

func monitorSiteCerts() error {
	sites, err := GetGlobalSites()
	if err != nil {
		return err
	}

	for _, site := range sites {
		//updatedTime, ok := siteUpdateMap[site.GetId()]
		//if ok && updatedTime != "" && updatedTime == site.UpdatedTime {
		//	continue
		//}

		lock.Lock()
		err = site.checkCerts()
		lock.Unlock()
		if err != nil {
			return err
		}

		siteUpdateMap[site.GetId()] = site.UpdatedTime
	}

	return err
}

func StartMonitorSitesLoop() {
	logs.Info("StartMonitorSitesLoop() Start!")
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logs.Error("Recovered from StartMonitorSitesLoop() panic: %v", r)
				StartMonitorSitesLoop()
			}
		}()

		for {
			err := refreshSiteMap()
			if err != nil {
				logs.Error(err)
				continue
			}

			err = refreshRuleMap()
			if err != nil {
				logs.Error(err)
				continue
			}

			err = monitorSiteCerts()
			if err != nil {
				logs.Error(err)
				continue
			}

			time.Sleep(5 * time.Second)
		}
	}()
}

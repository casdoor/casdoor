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
	"fmt"
	"sync"
	"time"

	"github.com/casdoor/casdoor/util"
)

var (
	siteUpdateMap    = map[string]string{}
	lock             = &sync.Mutex{}
	monitorLoopOnce  sync.Once
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

func runMonitorSitesLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s] Recovered from StartMonitorSitesLoop() panic: %v\n", util.GetCurrentTime(), r)
			go runMonitorSitesLoop()
		}
	}()

	for {
		err := refreshSiteMap()
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = refreshRuleMap()
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		err = monitorSiteCerts()
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}

		time.Sleep(5 * time.Second)
	}
}

func StartMonitorSitesLoop() {
	monitorLoopOnce.Do(func() {
		fmt.Printf("StartMonitorSitesLoop() Start!\n\n")
		go runMonitorSitesLoop()
	})
}

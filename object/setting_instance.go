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
	"strconv"
	"sync"

	"github.com/casdoor/casdoor/edition"
	"github.com/casdoor/casdoor/util"
	"github.com/xorm-io/core"
)

var (
	instanceIdMu sync.Mutex
	instanceId   string
	versionOnce  sync.Once
	version      string
)

// getInstanceId returns the ID stored in the setting row, creating it on first use. The replicas of a cluster share
// the database and so the ID; only the first writer's ID is kept when they start together.
func getInstanceId() string {
	instanceIdMu.Lock()
	defer instanceIdMu.Unlock()
	if instanceId != "" {
		return instanceId
	}

	setting, err := GetSetting()
	if err != nil || setting == nil {
		return ""
	}
	if setting.InstanceId == "" {
		_, err = ormer.Engine.ID(core.PK{settingOwner, settingName}).Where("instance_id = '' or instance_id is null").Cols("instance_id").Update(&Setting{InstanceId: util.GenerateId()})
		if err != nil {
			return ""
		}
		setting, err = GetSetting()
		if err != nil || setting == nil {
			return ""
		}
	}
	instanceId = setting.InstanceId
	return instanceId
}

func getVersion() string {
	versionOnce.Do(func() {
		info, err := util.GetVersionInfo()
		if err != nil || info.Version == "" {
			info = util.GetBuiltInVersionInfo()
		}
		version = info.Version
		if info.CommitOffset > 0 {
			version = info.Version + "+" + strconv.Itoa(info.CommitOffset)
		}
	})
	return version
}

func initLicenseReporter() {
	getInstanceId()
	edition.SetReporter(&edition.Reporter{
		InstanceId: getInstanceId,
		Version:    getVersion,
		UserCount: func() int64 {
			count, err := GetGlobalUserCount("", "")
			if err != nil {
				return 0
			}
			return count
		},
	})
}

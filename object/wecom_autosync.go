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
	"sync"
	"time"

	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/idp"
	"github.com/casdoor/casdoor/util"
)

type WeComAutoSynchronizer struct {
	sync.Mutex
	wecomIdToStopChan map[string]chan struct{}
}

var globalWeComAutoSynchronizer *WeComAutoSynchronizer

func InitWeComAutoSynchronizer() {
	globalWeComAutoSynchronizer = NewWeComAutoSynchronizer()
	err := globalWeComAutoSynchronizer.WeComAutoSynchronizerStartUpAll()
	if err != nil {
		panic(err)
	}
}

func NewWeComAutoSynchronizer() *WeComAutoSynchronizer {
	return &WeComAutoSynchronizer{
		wecomIdToStopChan: make(map[string]chan struct{}),
	}
}

func GetWeComAutoSynchronizer() *WeComAutoSynchronizer {
	return globalWeComAutoSynchronizer
}

// StartAutoSync starts autosync for specified WeCom, old existing autosync goroutine will be ceased
func (w *WeComAutoSynchronizer) StartAutoSync(wecomId string) error {
	w.Lock()
	defer w.Unlock()

	weCom, err := GetWeCom(wecomId)
	if err != nil {
		return err
	}

	if weCom == nil {
		return fmt.Errorf("WeCom %s doesn't exist", wecomId)
	}

	if res, ok := w.wecomIdToStopChan[wecomId]; ok {
		res <- struct{}{}
		delete(w.wecomIdToStopChan, wecomId)
	}

	stopChan := make(chan struct{})
	w.wecomIdToStopChan[wecomId] = stopChan
	logs.Info(fmt.Sprintf("autoSync started for %s", weCom.Id))
	util.SafeGoroutine(func() {
		err := w.syncRoutine(weCom, stopChan)
		if err != nil {
			panic(err)
		}
	})
	return nil
}

func (w *WeComAutoSynchronizer) StopAutoSync(wecomId string) {
	w.Lock()
	defer w.Unlock()
	if res, ok := w.wecomIdToStopChan[wecomId]; ok {
		res <- struct{}{}
		delete(w.wecomIdToStopChan, wecomId)
	}
}

// syncRoutine is the autosync goroutine
func (w *WeComAutoSynchronizer) syncRoutine(weCom *WeCom, stopChan chan struct{}) error {
	ticker := time.NewTicker(time.Duration(weCom.AutoSync) * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-stopChan:
			logs.Info(fmt.Sprintf("autoSync goroutine for %s stopped", weCom.Id))
			return nil
		case <-ticker.C:
		}

		err := UpdateWeComSyncTime(weCom.Id)
		if err != nil {
			return err
		}

		// Create WeCom syncer
		syncer := idp.NewWeComSyncer(weCom.CorpId, weCom.CorpSecret, weCom.DepartmentId)

		// Fetch all users
		users, err := syncer.GetAllUsers()
		if err != nil {
			logs.Warning(fmt.Sprintf("autoSync failed for %s, error %s", weCom.Id, err))
			continue
		}

		existed, failed, err := SyncWeComUsers(weCom.Owner, users, weCom.Id)
		if err != nil {
			logs.Warning(fmt.Sprintf("autoSync failed for %s, error %s", weCom.Id, err))
			continue
		}

		if len(failed) != 0 {
			logs.Warning(fmt.Sprintf("WeCom autosync, %d new users, but %d user failed during sync: %v", len(users)-len(existed)-len(failed), len(failed), failed))
		} else {
			logs.Info(fmt.Sprintf("WeCom autosync success, %d new users, %d existing users", len(users)-len(existed), len(existed)))
		}
	}
}

// WeComAutoSynchronizerStartUpAll starts all autosync goroutines for existing WeCom servers in each organization
func (w *WeComAutoSynchronizer) WeComAutoSynchronizerStartUpAll() error {
	organizations := []*Organization{}
	err := ormer.Engine.Desc("created_time").Find(&organizations)
	if err != nil {
		logs.Info("failed to start up WeComAutoSynchronizer")
		return err
	}
	for _, org := range organizations {
		weComs, err := GetWeComs(org.Name)
		if err != nil {
			return err
		}

		for _, weCom := range weComs {
			if weCom.AutoSync != 0 {
				err = w.StartAutoSync(weCom.Id)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func UpdateWeComSyncTime(wecomId string) error {
	_, err := ormer.Engine.ID(wecomId).Update(&WeCom{LastSync: util.GetCurrentTime()})
	if err != nil {
		return err
	}

	return nil
}

// SyncWeComUsers syncs users from WeCom to Casdoor
func SyncWeComUsers(owner string, wecomUsers []*idp.UserInfo, wecomId string) (existUsers []string, failedUsers []string, err error) {
	existUsers = []string{}
	failedUsers = []string{}

	for _, wecomUser := range wecomUsers {
		user, err := GetUserByField(owner, "id", wecomUser.Id)
		if err != nil {
			return existUsers, failedUsers, err
		}

		if user != nil {
			existUsers = append(existUsers, wecomUser.Id)
			// Update existing user
			user.DisplayName = wecomUser.DisplayName
			user.Email = wecomUser.Email
			user.Phone = wecomUser.Phone
			user.Avatar = wecomUser.AvatarUrl
			_, err = UpdateUser(user.GetId(), user, []string{"display_name", "email", "phone", "avatar"}, false)
			if err != nil {
				failedUsers = append(failedUsers, wecomUser.Id)
			}
		} else {
			// Create new user
			newUser := &User{
				Owner:       owner,
				Name:        wecomUser.Username,
				CreatedTime: util.GetCurrentTime(),
				Id:          wecomUser.Id,
				Type:        "normal-user",
				DisplayName: wecomUser.DisplayName,
				Email:       wecomUser.Email,
				Phone:       wecomUser.Phone,
				Avatar:      wecomUser.AvatarUrl,
				Properties:  map[string]string{},
			}

			_, err = AddUser(newUser, "en")
			if err != nil {
				failedUsers = append(failedUsers, wecomUser.Id)
			}
		}
	}

	return existUsers, failedUsers, nil
}

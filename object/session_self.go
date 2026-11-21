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
	"slices"
	"time"
)

// UserSessionItem is one signed-in browser of a user, across the applications it signed in to
type UserSessionItem struct {
	SessionId      string   `json:"sessionId"`
	Applications   []string `json:"applications"`
	CreatedTime    string   `json:"createdTime"`
	LastActiveTime string   `json:"lastActiveTime"`
	ExpireTime     string   `json:"expireTime"`
	Ip             string   `json:"ip"`
	UserAgent      string   `json:"userAgent"`
	IsCurrent      bool     `json:"isCurrent"`
}

// GetUserSessionItems lists the live sessions of a user for the user itself, with hashed session ids
func GetUserSessionItems(owner string, name string, currentSessionId string) ([]*UserSessionItem, error) {
	sessions, err := GetUserSessions(owner, name)
	if err != nil {
		return nil, err
	}

	_, expireTimes := getSessionTimes(sessions)
	itemMap := map[string]*UserSessionItem{}
	items := []*UserSessionItem{}
	for _, session := range sessions {
		infos := map[string]*SessionInfo{}
		for _, info := range session.SessionInfos {
			if info != nil {
				infos[info.SessionId] = info
			}
		}

		for _, sessionId := range session.SessionId {
			if sessionId == "" {
				continue
			}

			item, ok := itemMap[sessionId]
			if !ok {
				if sessionId != currentSessionId && !isSessionLive(sessionId, expireTimes[sessionId]) {
					continue
				}
				item = &UserSessionItem{SessionId: GetSessionIdHash(sessionId), IsCurrent: sessionId == currentSessionId}
				itemMap[sessionId] = item
				items = append(items, item)
			}

			if !slices.Contains(item.Applications, session.Application) {
				item.Applications = append(item.Applications, session.Application)
			}

			info := infos[sessionId]
			if info == nil {
				continue
			}
			if item.CreatedTime == "" || info.CreatedTime < item.CreatedTime {
				item.CreatedTime = info.CreatedTime
			}
			if info.LastActiveTime >= item.LastActiveTime {
				item.LastActiveTime = info.LastActiveTime
				item.Ip = info.Ip
				item.UserAgent = info.UserAgent
			}
			if info.ExpireTime > item.ExpireTime {
				item.ExpireTime = info.ExpireTime
			}
		}
	}

	slices.SortStableFunc(items, func(a, b *UserSessionItem) int {
		if a.IsCurrent != b.IsCurrent {
			if a.IsCurrent {
				return -1
			}
			return 1
		}
		return parseSessionTime(b.LastActiveTime).Compare(parseSessionTime(a.LastActiveTime))
	})
	return items, nil
}

func parseSessionTime(value string) time.Time {
	t, _ := time.Parse(time.RFC3339, value)
	return t
}

// LogoutUserSessionByHash signs the user out of one of its sessions, given by its hashed id, in the same
// way as an administrator ending it: the applications signed in with it are told and its tokens expire
func LogoutUserSessionByHash(user *User, sessionIdHash string, host string) (bool, error) {
	sessions, err := GetUserSessions(user.Owner, user.Name)
	if err != nil {
		return false, err
	}

	for _, session := range sessions {
		if sessionId := resolveSessionId(session, sessionIdHash); sessionId != "" {
			return true, logoutUserSession(user, sessionId, host)
		}
	}
	return false, nil
}

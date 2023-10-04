// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

type OriginalSyncer interface {
	GetOriginalUserMap() ([]*OriginalUser, map[string]*OriginalUser, error)
	UpdateUser(oUser *OriginalUser) (bool, error)
	AddUser(oUser *OriginalUser) (bool, error)
	GetOriginalGroupMap() ([]*OriginalGroup, map[string]*OriginalGroup, error)
	UpdateGroup(oGroup *OriginalGroup) (bool, error)
	AddGroup(oGroup *OriginalGroup) (bool, error)
	GetAffiliationMap() ([]*Affiliation, map[int]string, error)
}

func GetOriginalSyncer(syncer *Syncer) (OriginalSyncer, error) {
	if syncer.Type == "Database" || syncer.Type == "Keycloak" {
		return NewDatabaseSyncer(syncer.Type, syncer.DatabaseType, syncer.SslMode, syncer.User, syncer.Password, syncer.Host, syncer.Port, syncer.Database, syncer.Table, isCloudIntranet, syncer.TableColumns, syncer.AvatarBaseUrl, syncer.AffiliationTable)
	} else if syncer.Type == "WeCom" {
		return NewWeComSyncer(syncer.User, syncer.Password, syncer.Secret, syncer.Organization)
	}
	return nil, nil
}

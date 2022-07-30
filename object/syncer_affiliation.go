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

type Affiliation struct {
	Id   int    `xorm:"int notnull pk autoincr" json:"id"`
	Name string `xorm:"varchar(128)" json:"name"`
}

func (syncer *Syncer) getAffiliations() []*Affiliation {
	affiliations := []*Affiliation{}
	err := syncer.Adapter.Engine.Table(syncer.AffiliationTable).Asc("id").Find(&affiliations)
	if err != nil {
		panic(err)
	}

	return affiliations
}

func (syncer *Syncer) getAffiliationMap() ([]*Affiliation, map[int]string) {
	affiliations := syncer.getAffiliations()

	m := map[int]string{}
	for _, affiliation := range affiliations {
		m[affiliation.Id] = affiliation.Name
	}
	return affiliations, m
}

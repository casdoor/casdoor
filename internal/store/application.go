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

package store

import (
	"github.com/casdoor/casdoor/internal/object"
	"github.com/casdoor/casdoor/internal/store/shared"
)

type ApplicationStore struct {
	db *shared.DB
}

func NewApplicationStore(db *shared.DB) *ApplicationStore {
	return &ApplicationStore{
		db: db,
	}
}

func (a *ApplicationStore) Create(app *object.Application) error {
	_, err := a.db.GetEngine().Insert(app)
	return err
}

func (a *ApplicationStore) List(limit, offset int) ([]*object.Application, error) {
	var apps []*object.Application
	err := a.db.GetEngine().Desc("created_time").Limit(limit, offset).Find(&apps)
	return apps, err
}

func (a *ApplicationStore) Get(id string) (*object.Application, error) {
	var app *object.Application
	err := a.db.GetEngine().ID(id).Find(app)
	return app, err
}

func (a *ApplicationStore) Update(app *object.Application) error {
	_, err := a.db.GetEngine().Update(app)
	return err
}

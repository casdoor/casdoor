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

package application

import (
	"net/http"
	"strconv"
	"time"

	"github.com/casdoor/casdoor/internal/object"
	"github.com/casdoor/casdoor/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	applicationStore *store.ApplicationStore
}

func New(applicationStore *store.ApplicationStore) *Handler {
	return &Handler{applicationStore: applicationStore}
}

func (h *Handler) Create(g *gin.Context) {
	app := &object.Application{}
	err := g.BindJSON(app)
	if err != nil {
		_ = g.Error(err)
		return
	}
	app.Id = uuid.New().String()
	app.CreatedTime = time.Now().Format(time.RFC3339)
	app.CreatedBy = "TODO"
	err = h.applicationStore.Create(app)
	if err != nil {
		_ = g.Error(err)
		return
	}
}

func (h *Handler) List(g *gin.Context) {
	limit, _ := strconv.Atoi(g.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(g.DefaultQuery("offset", "0"))
	apps, err := h.applicationStore.List(limit, offset)
	if err != nil {
		_ = g.Error(err)
		return
	}
	g.JSON(http.StatusOK, apps)
}

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

package routers

import (
	stdcontext "context"
	"net/http"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/session"
	"github.com/casdoor/casdoor/util"
)

const FileSessionProvider = "casdoor_file"

const fileSessionGcInterval = time.Hour

// fileSessionProvider wraps beego's "file" provider so that a new session is only written to disk
// once it holds data, instead of one file for every request that comes without a session cookie
type fileSessionProvider struct {
	session.Provider
}

type lazySessionStore struct {
	provider  session.Provider
	sid       string
	lock      sync.RWMutex
	values    map[interface{}]interface{}
	persisted bool
	ephemeral bool
}

func init() {
	provider, err := session.GetProvider("file")
	if err != nil {
		panic(err)
	}
	session.Register(FileSessionProvider, &fileSessionProvider{Provider: provider})
}

func (p *fileSessionProvider) SessionInit(ctx stdcontext.Context, maxLifetime int64, savePath string) error {
	err := p.Provider.SessionInit(ctx, maxLifetime, savePath)
	if err != nil {
		return err
	}

	util.SafeGoroutine(func() {
		for range time.Tick(fileSessionGcInterval) {
			p.Provider.SessionGC(stdcontext.Background())
		}
	})
	return nil
}

func (p *fileSessionProvider) SessionRead(ctx stdcontext.Context, sid string) (session.Store, error) {
	exists, err := p.SessionExist(ctx, sid)
	if err != nil {
		return nil, err
	}
	if exists {
		return p.Provider.SessionRead(ctx, sid)
	}

	return p.newLazySessionStore(sid), nil
}

func (p *fileSessionProvider) SessionRegenerate(ctx stdcontext.Context, oldSid, sid string) (session.Store, error) {
	exists, err := p.SessionExist(ctx, oldSid)
	if err != nil {
		return nil, err
	}
	if exists {
		return p.Provider.SessionRegenerate(ctx, oldSid, sid)
	}

	return p.newLazySessionStore(sid), nil
}

func (p *fileSessionProvider) newLazySessionStore(sid string) *lazySessionStore {
	return &lazySessionStore{provider: p.Provider, sid: sid, values: map[interface{}]interface{}{}}
}

func (s *lazySessionStore) Set(ctx stdcontext.Context, key, value interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.values[key] = value
	return nil
}

func (s *lazySessionStore) Get(ctx stdcontext.Context, key interface{}) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.values[key]
}

func (s *lazySessionStore) Delete(ctx stdcontext.Context, key interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.values, key)
	return nil
}

func (s *lazySessionStore) Flush(ctx stdcontext.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.values = map[interface{}]interface{}{}
	return nil
}

func (s *lazySessionStore) SessionID(ctx stdcontext.Context) string {
	return s.sid
}

func (s *lazySessionStore) SessionRelease(ctx stdcontext.Context, w http.ResponseWriter) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.ephemeral || (!s.persisted && len(s.values) == 0) {
		return
	}

	store, err := s.provider.SessionRead(ctx, s.sid)
	if err != nil {
		logs.Error("SessionRead failed, error: %s", err)
		return
	}
	_ = store.Flush(ctx)
	for key, value := range s.values {
		_ = store.Set(ctx, key, value)
	}
	store.SessionRelease(ctx, w)
	s.persisted = true
}

func (s *lazySessionStore) SessionReleaseIfPresent(ctx stdcontext.Context, w http.ResponseWriter) {
	s.lock.RLock()
	persisted := s.persisted
	s.lock.RUnlock()
	if persisted {
		s.SessionRelease(ctx, w)
	}
}

// skipSessionPersistence keeps the session of a request that authenticates itself (API calls with an
// Authorization header, client ID/secret or access key) in memory only, as the caller won't send the cookie back
func skipSessionPersistence(ctx *context.Context) {
	if store, ok := ctx.Input.CruSession.(*lazySessionStore); ok {
		store.lock.Lock()
		store.ephemeral = true
		store.lock.Unlock()
	}
}

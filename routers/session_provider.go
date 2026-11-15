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
	"hash/fnv"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/session"
	_ "github.com/beego/beego/v2/server/web/session/redis"
	_ "github.com/beego/beego/v2/server/web/session/redis_cluster"
	"github.com/casdoor/casdoor/util"
)

const fileSessionGcInterval = time.Hour

// lazySessionProvider wraps a beego provider ("file", "redis", "redis_cluster") so that a new session is only
// written to the store (and its cookie only sent) once it holds data, and a request only writes back the keys
// it changed, so concurrent requests of the same browser don't overwrite each other's session data
type lazySessionProvider struct {
	session.Provider
	gcInterval time.Duration
}

type deletedSessionValue struct{}

type lazySessionStore struct {
	provider      session.Provider
	sid           string
	lock          sync.RWMutex
	base          session.Store
	changes       map[interface{}]interface{}
	flushed       bool
	ephemeral     bool
	pendingCookie string
	header        http.Header
}

var sessionWriteLocks [64]sync.Mutex

func init() {
	registerLazySessionProvider("file", fileSessionGcInterval)
	registerLazySessionProvider("redis", 0)
	registerLazySessionProvider("redis_cluster", 0)
}

func registerLazySessionProvider(name string, gcInterval time.Duration) {
	provider, err := session.GetProvider(name)
	if err != nil {
		panic(err)
	}
	session.Register(GetLazySessionProvider(name), &lazySessionProvider{Provider: provider, gcInterval: gcInterval})
}

// GetLazySessionProvider returns the name of the lazy wrapper registered for a beego session provider
func GetLazySessionProvider(name string) string {
	return "casdoor_" + name
}

func (p *lazySessionProvider) SessionInit(ctx stdcontext.Context, maxLifetime int64, savePath string) error {
	err := p.Provider.SessionInit(ctx, maxLifetime, savePath)
	if err != nil {
		return err
	}

	if p.gcInterval > 0 {
		util.SafeGoroutine(func() {
			for range time.Tick(p.gcInterval) {
				p.Provider.SessionGC(stdcontext.Background())
			}
		})
	}
	return nil
}

func (p *lazySessionProvider) SessionRead(ctx stdcontext.Context, sid string) (session.Store, error) {
	exists, err := p.SessionExist(ctx, sid)
	if err != nil {
		return nil, err
	}
	if !exists {
		return p.newLazySessionStore(sid, nil), nil
	}

	base, err := p.Provider.SessionRead(ctx, sid)
	if err != nil {
		return nil, err
	}
	return p.newLazySessionStore(sid, base), nil
}

func (p *lazySessionProvider) SessionRegenerate(ctx stdcontext.Context, oldSid, sid string) (session.Store, error) {
	exists, err := p.SessionExist(ctx, oldSid)
	if err != nil {
		return nil, err
	}
	if !exists {
		return p.newLazySessionStore(sid, nil), nil
	}

	base, err := p.Provider.SessionRegenerate(ctx, oldSid, sid)
	if err != nil {
		return nil, err
	}
	return p.newLazySessionStore(sid, base), nil
}

func (p *lazySessionProvider) newLazySessionStore(sid string, base session.Store) *lazySessionStore {
	return &lazySessionStore{provider: p.Provider, sid: sid, base: base, changes: map[interface{}]interface{}{}}
}

func getSessionWriteLock(sid string) *sync.Mutex {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(sid))
	return &sessionWriteLocks[hash.Sum32()%uint32(len(sessionWriteLocks))]
}

func (s *lazySessionStore) Set(ctx stdcontext.Context, key, value interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.changes[key] = value
	if s.pendingCookie != "" && !s.ephemeral {
		s.header.Add("Set-Cookie", s.pendingCookie)
		s.pendingCookie = ""
	}
	return nil
}

func (s *lazySessionStore) Get(ctx stdcontext.Context, key interface{}) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if value, ok := s.changes[key]; ok {
		if _, ok = value.(deletedSessionValue); ok {
			return nil
		}
		return value
	}
	if s.flushed || s.base == nil {
		return nil
	}
	return s.base.Get(ctx, key)
}

func (s *lazySessionStore) Delete(ctx stdcontext.Context, key interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.changes[key] = deletedSessionValue{}
	return nil
}

func (s *lazySessionStore) Flush(ctx stdcontext.Context) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.changes = map[interface{}]interface{}{}
	s.flushed = true
	return nil
}

func (s *lazySessionStore) SessionID(ctx stdcontext.Context) string {
	return s.sid
}

func (s *lazySessionStore) hasValue() bool {
	for _, value := range s.changes {
		if _, ok := value.(deletedSessionValue); !ok {
			return true
		}
	}
	return false
}

// SessionRelease writes only what this request changed on top of what the store holds now, instead of
// the copy read at the start of the request, which may be outdated by then
func (s *lazySessionStore) SessionRelease(ctx stdcontext.Context, w http.ResponseWriter) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.ephemeral || (!s.flushed && len(s.changes) == 0) {
		return
	}
	if s.base == nil && !s.hasValue() {
		return
	}

	writeLock := getSessionWriteLock(s.sid)
	writeLock.Lock()
	defer writeLock.Unlock()

	store, err := s.provider.SessionRead(ctx, s.sid)
	if err != nil {
		logs.Error("SessionRead failed, error: %s", err)
		return
	}
	if s.flushed {
		_ = store.Flush(ctx)
	}
	for key, value := range s.changes {
		if _, ok := value.(deletedSessionValue); ok {
			_ = store.Delete(ctx, key)
		} else {
			_ = store.Set(ctx, key, value)
		}
	}
	store.SessionRelease(ctx, w)

	s.base = store
	s.changes = map[interface{}]interface{}{}
	s.flushed = false
}

func (s *lazySessionStore) SessionReleaseIfPresent(ctx stdcontext.Context, w http.ResponseWriter) {
	s.lock.RLock()
	present := s.base != nil
	s.lock.RUnlock()
	if present {
		s.SessionRelease(ctx, w)
	}
}

// skipSessionPersistence keeps the session of a request that authenticates itself (API calls with an
// Authorization header, client ID/secret or access key) in memory only, as the caller won't send the cookie back
func skipSessionPersistence(ctx *context.Context) {
	if store, ok := ctx.Input.CruSession.(*lazySessionStore); ok {
		store.lock.Lock()
		store.ephemeral = true
		store.pendingCookie = ""
		store.lock.Unlock()
	}
}

// SessionCookieFilter holds back the cookie beego sends for a new, still empty session until the session gets
// data. Otherwise a request that comes with an outdated session ID (e.g. one replaced at sign-in) answers with
// the cookie of a new empty session, which overwrites the cookie of the signed-in session in the browser
func SessionCookieFilter(ctx *context.Context) {
	store, ok := ctx.Input.CruSession.(*lazySessionStore)
	if !ok {
		return
	}

	store.lock.Lock()
	defer store.lock.Unlock()
	if store.base != nil || len(store.changes) != 0 {
		return
	}

	header := ctx.ResponseWriter.Header()
	prefix := web.BConfig.WebConfig.Session.SessionName + "=" + url.QueryEscape(store.sid) + ";"
	cookies := header.Values("Set-Cookie")
	kept := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		if store.pendingCookie == "" && strings.HasPrefix(cookie, prefix) {
			store.pendingCookie = cookie
			continue
		}
		kept = append(kept, cookie)
	}
	if store.pendingCookie == "" {
		return
	}

	header.Del("Set-Cookie")
	for _, cookie := range kept {
		header.Add("Set-Cookie", cookie)
	}
	store.header = header
	if store.ephemeral {
		store.pendingCookie = ""
	}
}

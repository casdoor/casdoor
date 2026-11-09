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

package i18n

import (
	"sync"
	"testing"
)

func TestTranslateConcurrent(t *testing.T) {
	// Before the fix, lazy loading of langMap made concurrent Translate calls
	// hit "fatal error: concurrent map writes" (goroutine crash, not recoverable).
	// Run with -race to verify.
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			_ = Translate("zh", "login:Signed out successfully")
		}()
		go func() {
			defer wg.Done()
			_ = Translate("en", "login:Signed out successfully")
		}()
		go func() {
			defer wg.Done()
			_ = Translate("unsupportedLanguage", "login:Signed out successfully")
		}()
	}
	wg.Wait()
}

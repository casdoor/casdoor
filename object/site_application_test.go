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

import "testing"

func TestSiteMatchesApplication(t *testing.T) {
	tests := []struct {
		name string
		site *Site
		app  *Application
		want bool
	}{
		{
			name: "application in the site organization",
			site: &Site{Owner: "homelab"},
			app:  &Application{Owner: "admin", Organization: "homelab"},
			want: true,
		},
		{
			name: "application owned by the site owner",
			site: &Site{Owner: "admin"},
			app:  &Application{Owner: "admin", Organization: "homelab"},
			want: true,
		},
		{
			name: "shared application",
			site: &Site{Owner: "homelab"},
			app:  &Application{Owner: "admin", Organization: "built-in", IsShared: true},
			want: true,
		},
		{
			name: "application in another organization",
			site: &Site{Owner: "homelab"},
			app:  &Application{Owner: "admin", Organization: "other"},
			want: false,
		},
		{
			name: "nil application",
			site: &Site{Owner: "homelab"},
			app:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.site.MatchesApplication(tt.app); got != tt.want {
				t.Errorf("MatchesApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

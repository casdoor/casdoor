// Copyright 2024 The Casdoor Authors. All Rights Reserved.
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
	"testing"
)

func TestGetSslMode(t *testing.T) {
	tests := []struct {
		name     string
		provider *Provider
		want     string
	}{
		{
			name: "SslMode is set to Auto",
			provider: &Provider{
				SslMode:    "Auto",
				DisableSsl: false,
			},
			want: "Auto",
		},
		{
			name: "SslMode is set to Enable",
			provider: &Provider{
				SslMode:    "Enable",
				DisableSsl: true,
			},
			want: "Enable",
		},
		{
			name: "SslMode is set to Disable",
			provider: &Provider{
				SslMode:    "Disable",
				DisableSsl: false,
			},
			want: "Disable",
		},
		{
			name: "SslMode is empty, DisableSsl is false (backward compatibility)",
			provider: &Provider{
				SslMode:    "",
				DisableSsl: false,
			},
			want: "Auto",
		},
		{
			name: "SslMode is empty, DisableSsl is true (backward compatibility)",
			provider: &Provider{
				SslMode:    "",
				DisableSsl: true,
			},
			want: "Disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getSslMode(tt.provider); got != tt.want {
				t.Errorf("getSslMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

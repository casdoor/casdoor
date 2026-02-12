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

package email

import (
	"testing"
)

func TestNewSmtpEmailProviderSslMode(t *testing.T) {
	tests := []struct {
		name        string
		sslMode     string
		port        int
		wantSsl     bool
		description string
	}{
		{
			name:        "Auto mode on port 465 (should use gomail default SSL=true)",
			sslMode:     "Auto",
			port:        465,
			wantSsl:     true, // gomail sets SSL=true for port 465 by default
			description: "Auto mode should let gomail decide SSL based on port",
		},
		{
			name:        "Auto mode on port 587 (should use gomail default SSL=false)",
			sslMode:     "Auto",
			port:        587,
			wantSsl:     false, // gomail sets SSL=false for port 587 by default
			description: "Auto mode should let gomail decide SSL based on port",
		},
		{
			name:        "Enable mode explicitly sets SSL=true",
			sslMode:     "Enable",
			port:        587,
			wantSsl:     true,
			description: "Enable mode should set SSL=true regardless of port",
		},
		{
			name:        "Disable mode explicitly sets SSL=false",
			sslMode:     "Disable",
			port:        465,
			wantSsl:     false,
			description: "Disable mode should set SSL=false regardless of port",
		},
		{
			name:        "Empty sslMode on port 465 (should use gomail default)",
			sslMode:     "",
			port:        465,
			wantSsl:     true,
			description: "Empty sslMode should let gomail decide SSL based on port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSmtpEmailProvider("test@example.com", "password", "smtp.example.com", tt.port, "Default", tt.sslMode, false)
			if provider.Dialer.SSL != tt.wantSsl {
				t.Errorf("%s: NewSmtpEmailProvider() SSL = %v, want %v. %s", tt.name, provider.Dialer.SSL, tt.wantSsl, tt.description)
			}
		})
	}
}

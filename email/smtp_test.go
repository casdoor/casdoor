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

import "testing"

func TestSmtpEmailProviderSSLMode(t *testing.T) {
	tests := []struct {
		name           string
		disableSslMode string
		port           int
		expectedSSL    bool
	}{
		{
			name:           "DisableSslMode True should disable SSL",
			disableSslMode: "True",
			port:           465,
			expectedSSL:    false,
		},
		{
			name:           "DisableSslMode False should enable SSL",
			disableSslMode: "False",
			port:           587,
			expectedSSL:    true,
		},
		{
			name:           "DisableSslMode Unspecified with port 465 should enable SSL (gomail default)",
			disableSslMode: "",
			port:           465,
			expectedSSL:    true, // gomail defaults to SSL for port 465
		},
		{
			name:           "DisableSslMode Unspecified with port 587 should disable SSL (gomail default)",
			disableSslMode: "",
			port:           587,
			expectedSSL:    false, // gomail defaults to no SSL for other ports (uses STARTTLS)
		},
		{
			name:           "DisableSslMode empty string with port 465 should enable SSL (gomail default)",
			disableSslMode: "",
			port:           465,
			expectedSSL:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewSmtpEmailProvider("test@example.com", "password", "smtp.example.com", tt.port, "Default", tt.disableSslMode, false)
			if provider.Dialer.SSL != tt.expectedSSL {
				t.Errorf("Expected SSL to be %v, but got %v", tt.expectedSSL, provider.Dialer.SSL)
			}
		})
	}
}

package object

import (
	"encoding/json"
	"testing"
)

func TestTokenWrapperOmitsEmptyRefreshToken(t *testing.T) {
	for _, tc := range []struct {
		name         string
		refreshToken string
		wantPresent  bool
	}{
		{name: "no refresh token"},
		{name: "issued refresh token", refreshToken: "refresh-value", wantPresent: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload, err := json.Marshal(TokenWrapper{
				AccessToken: "access-value", RefreshToken: tc.refreshToken,
				TokenType: "Bearer", ExpiresIn: 3600, Scope: "openid",
			})
			if err != nil {
				t.Fatal(err)
			}
			var result map[string]any
			if err := json.Unmarshal(payload, &result); err != nil {
				t.Fatal(err)
			}
			if got, present := result["refresh_token"]; present != tc.wantPresent || (present && got != tc.refreshToken) {
				t.Fatalf("refresh_token = %v, present = %v; want %q, present = %v", got, present, tc.refreshToken, tc.wantPresent)
			}
			if result["access_token"] != "access-value" || result["token_type"] != "Bearer" || result["expires_in"] != float64(3600) || result["scope"] != "openid" {
				t.Fatalf("other token fields changed: %s", payload)
			}
		})
	}
}

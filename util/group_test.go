package util

import "testing"

func TestShouldExtendGroupWithUsers(t *testing.T) {
	tests := []struct {
		name      string
		withUsers string
		want      bool
	}{
		{name: "keeps default behavior", withUsers: "", want: true},
		{name: "keeps explicit expansion", withUsers: "true", want: true},
		{name: "skips explicit expansion opt-out", withUsers: "false", want: false},
		{name: "keeps compatibility for unknown values", withUsers: "invalid", want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldExtendGroupWithUsers(tt.withUsers); got != tt.want {
				t.Fatalf("ShouldExtendGroupWithUsers(%q) = %v, want %v", tt.withUsers, got, tt.want)
			}
		})
	}
}

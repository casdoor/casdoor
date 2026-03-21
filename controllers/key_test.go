package controllers

import "testing"

func TestIsAddKeyEnabledProvided(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		expected bool
	}{
		{name: "missing field", body: `{"name":"k1"}`, expected: false},
		{name: "explicit false", body: `{"name":"k1","isEnabled":false}`, expected: true},
		{name: "explicit true", body: `{"name":"k1","isEnabled":true}`, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := isAddKeyEnabledProvided([]byte(tt.body))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if actual != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

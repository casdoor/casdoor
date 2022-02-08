package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUploadXlsxPath(t *testing.T) {
	scenarios := []struct {
		description string
		input       string
		expected    interface{}
	}{
		{"scenery one", "casdoor", "tmpFiles/casdoor.xlsx"},
		{"scenery two", "casbin", "tmpFiles/casbin.xlsx"},
		{"scenery three", "loremIpsum", "tmpFiles/loremIpsum.xlsx"},
		{"scenery four", "", "tmpFiles/.xlsx"},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			actual := GetUploadXlsxPath(scenery.input)
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}


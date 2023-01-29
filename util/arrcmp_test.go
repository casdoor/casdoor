package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArrcmp(t *testing.T) {

	var s []string
	scenarios := []struct {
		description string
		input       [][]string
		expected    [][]string
	}{
		{"scenery one", [][]string{{"str1", "str2", "str3"}, {"str1", "str2", "str3", "str4"}}, [][]string{{"str4"}, s}},
		{"scenery two", [][]string{{"str1", "str2", "str3"}, {"str1", "str2", "str4"}}, [][]string{{"str4"}, {"str3"}}},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			added, deleted := Arrcmp(scenery.input[0], scenery.input[1])
			actual := [][]string{added, deleted}
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

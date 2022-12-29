package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInArray(t *testing.T) {

	scenarios := []struct {
		description string
		input       []interface{}
		expected    []interface{}
	}{
		{"scenery one", []interface{}{"str1", []string{"str1", "str2", "str3", "str4"}}, []interface{}{true, 0}},
		{"scenery two", []interface{}{"str", []string{"str1", "str2", "str3", "str4"}}, []interface{}{false, -1}},
	}
	for _, scenery := range scenarios {
		t.Run(scenery.description, func(t *testing.T) {
			exists, index := InArray(scenery.input[0], scenery.input[1])
			actual := []interface{}{exists, index}
			assert.Equal(t, scenery.expected, actual, "The returned value not is expected")
		})
	}
}

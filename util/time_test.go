package util

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_GetCurrentTime(t *testing.T) {
	test := GetCurrentTime()
	expected := time.Now().Format(time.RFC3339)

	assert.Equal(t, test, expected, "The times not are equals")

	types := reflect.TypeOf(test).Kind()
	assert.Equal(t, types, reflect.String, "GetCurrentUnixTime should be return string")

}

func Test_GetCurrentUnixTime_Shoud_Return_String(t *testing.T) {
	test := GetCurrentUnixTime()
	types := reflect.TypeOf(test).Kind()
	assert.Equal(t, types, reflect.String, "GetCurrentUnixTime should be return string")
}

func Test_IsTokenExpired(t *testing.T) {

	type input struct {
		createdTime string
		expiresIn   int
	}

	type errorTestCases struct {
		description string
		input       input
		expected    bool
	}

	for _, scenario := range []errorTestCases{
		{
			description: "Token emited now is valid for 60 minutes",
			input: input{
				createdTime: time.Now().Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: false,
		},
		{
			description: "Token emited 60 minutes before now is valid for 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 60).Format(time.RFC3339),
				expiresIn:   61,
			},
			expected: false,
		},
		{
			description: "Token emited 2 hours before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: true,
		},
		{
			description: "Token emited 61 minutes before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 61).Format(time.RFC3339),
				expiresIn:   60,
			},
			expected: true,
		},
		{
			description: "Token emited 2 hours before now  is velid for 120 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Hour * 2).Format(time.RFC3339),
				expiresIn:   121,
			},
			expected: false,
		},
		{
			description: "Token emited 159 minutes before now is Expired after 60 minutes",
			input: input{
				createdTime: time.Now().Add(-time.Minute * 159).Format(time.RFC3339),
				expiresIn:   120,
			},
			expected: true,
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			result := IsTokenExpired(scenario.input.createdTime, scenario.input.expiresIn)
			assert.Equal(t, scenario.expected, result, fmt.Sprintf("Expected %t, but was founded %t", scenario.expected, result))
		})
	}
}

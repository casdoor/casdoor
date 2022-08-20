package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCpuUsage(t *testing.T) {
	usage, err := GetCpuUsage()
	assert.Nil(t, err)
	t.Log(usage)
}

func TestGetMemoryUsage(t *testing.T) {
	used, total, err := GetMemoryUsage()
	assert.Nil(t, err)
	t.Log(used, total)
}

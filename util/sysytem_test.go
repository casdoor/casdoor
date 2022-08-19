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
	usage, err := GetMemoryUsage()
	assert.Nil(t, err)
	t.Log(usage)
}

func TestGetGithubRepoReleaseVersion(t *testing.T) {
	version, err := GetGithubRepoReleaseVersion("casdoor/casdoor")
	assert.Nil(t, err)
	t.Log(version)
}

package util

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// get cpu usage
func GetCpuUsage() (float64, error) {
	usage, err := cpu.Percent(time.Second, false)
	return usage[0], err
}

// get memory usage
func GetMemoryUsage() (float64, error) {
	usage, err := mem.VirtualMemory()
	return usage.UsedPercent, err
}

// get github repo release version
func GetGithubRepoReleaseVersion(repo string) (string, error) {
	// get github repo release version
	resp, err := http.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", err
	}

	// close response body
	defer resp.Body.Close()

	// get github repo release version
	release := struct {
		TagName string `json:"tag_name"`
	}{}
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return "", err
	}

	return release.TagName, nil
}

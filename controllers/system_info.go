package controllers

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type SystemInfo struct {
	MemoryUsed  uint64    `json:"memory_used"`
	MemoryTotal uint64    `json:"memory_total"`
	CpuUsage    []float64 `json:"cpu_usage"`
}

// GetSystemInfo
// @Title GetSystemInfo
// @Tag System API
// @Description get user's system info
// @Param   id    query    string  true        "The id of the user"
// @Success 200 {object} object.SystemInfo The Response object
// @router /get-system-info [get]
func (c *ApiController) GetSystemInfo() {
	id := c.GetString("id")
	if id == "" {
		id = c.GetSessionUsername()
	}

	user := object.GetUser(id)
	if user == nil || !user.IsGlobalAdmin {
		c.ResponseError("You are not authorized to access this resource")
	}

	cpuUsage, err := util.GetCpuUsage()
	if err != nil {
		c.ResponseError(err.Error())
	}

	memoryUsed, memoryTotal, err := util.GetMemoryUsage()
	if err != nil {
		c.ResponseError(err.Error())
	}

	c.Data["json"] = SystemInfo{
		CpuUsage:    cpuUsage,
		MemoryUsed:  memoryUsed,
		MemoryTotal: memoryTotal,
	}
	c.ServeJSON()
}

// GithubLatestVersion
// @Title GithubLatestVersion
// @Tag System API
// @Description get github repo's latest release version info
// @Param   repo    query    string  true        "The GitHub repo"
// @Success 200 {string} latest version of casdoor
// @router /get-release [get]
func (c *ApiController) GithubLatestVersion() {
	version, err := util.GetGithubRepoVersion()
	if err != nil {
		c.ResponseError(err.Error())
	}

	c.Data["json"] = version
	c.ServeJSON()
}

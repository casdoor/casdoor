package controllers

import (
	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type SystemInfo struct {
	CpuUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
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

	memoryUsage, err := util.GetMemoryUsage()
	if err != nil {
		c.ResponseError(err.Error())
	}

	c.Data["json"] = SystemInfo{
		CpuUsage:    cpuUsage,
		MemoryUsage: memoryUsage,
	}
	c.ServeJSON()
}

// GithubLatestReleaseVersion
// @Title GithubLatestReleaseVersion
// @Tag System API
// @Description get github repo's latest release version info
// @Param   repo    query    string  true        "The GitHub repo"
// @Success 200 {string} latest version of casdoor
// @router /get-release [get]
func (c *ApiController) GithubLatestReleaseVersion() {
	repo := c.GetString("repo")
	version, err := util.GetGithubRepoReleaseVersion(repo)
	if err != nil {
		c.ResponseError(err.Error())
	}

	c.Data["json"] = version
	c.ServeJSON()
}

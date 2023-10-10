package controllers

import (
	"strings"

	"github.com/casdoor/casdoor/scim"
)

func (c *RootController) HandleScim() {
	path := c.Ctx.Request.URL.Path
	c.Ctx.Request.URL.Path = strings.TrimPrefix(path, "/scim")
	scim.Server.ServeHTTP(c.Ctx.ResponseWriter, c.Ctx.Request)
}

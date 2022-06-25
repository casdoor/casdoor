package routers

import (
	"net/http"

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/object"
)

const (
	headerOrigin       = "Origin"
	headerAllowOrigin  = "Access-Control-Allow-Origin"
	headerAllowMethods = "Access-Control-Allow-Methods"
	headerAllowHeaders = "Access-Control-Allow-Headers"
)

func CorsFilter(ctx *context.Context) {
	if ctx.Input.Method() == "OPTIONS" {
		origin := ctx.Input.Header(headerOrigin)

		if object.IsAllowOrigin(origin) {
			ctx.Output.Header(headerAllowOrigin, origin)
			ctx.Output.Header(headerAllowMethods, "POST, GET, OPTIONS")
			ctx.Output.Header(headerAllowHeaders, "Content-Type, Authorization")
			ctx.ResponseWriter.WriteHeader(http.StatusOK)
		} else {
			ctx.ResponseWriter.WriteHeader(http.StatusForbidden)
		}
		return
	}
}

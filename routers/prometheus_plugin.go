package routers

import (
	"net/http"
	"strings"
	"time"

	"github.com/astaxie/beego/context"
	"github.com/casdoor/casdoor/object"
)

func PrometheusVisitTimeFilter(ctx *context.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, "/api/metrics") {
		return
	}
	object.VisitTime.Inc()
}

func PrometheusLoginTimeFilter(ctx *context.Context) {
	if strings.HasPrefix(ctx.Request.URL.Path, "/api/metrics") {
		return
	}
	if ctx.Request.URL.Path == "/api/login" {
		object.LoginTimes.Inc()
	}
}

type PrometheusMiddleWareWrapper struct {
	oldHandler http.Handler
}

func (p *PrometheusMiddleWareWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if strings.HasPrefix(req.URL.Path, "/api/metrics") {
		p.oldHandler.ServeHTTP(w, req)
		return
	}
	startTime := time.Now()
	p.oldHandler.ServeHTTP(w, req)
	endTime := time.Now()
	delta := endTime.Sub(startTime).Milliseconds()
	object.ResponseTime.Observe(float64(delta))
}

func PrometheusResponseTimeMiddleWare(h http.Handler) http.Handler {
	return &PrometheusMiddleWareWrapper{
		oldHandler: h,
	}
}

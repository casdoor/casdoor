package routers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/casdoor/casdoor/object"
	"github.com/casdoor/casdoor/util"
)

type PrometheusMiddleWareWrapper struct {
	handler http.Handler
}

func PrometheusMiddleWare(h http.Handler) http.Handler {
	return &PrometheusMiddleWareWrapper{
		handler: h,
	}
}

func (p PrometheusMiddleWareWrapper) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	method := req.Method
	endpoint := req.URL.Path
	if strings.HasPrefix(endpoint, "/api/metrics") {
		systemInfo, err := util.GetSystemInfo()
		if err == nil {
			recordSystemInfo(systemInfo)
		}
		p.handler.ServeHTTP(w, req)
		return
	}

	if strings.HasPrefix(endpoint, "/api") {
		start := time.Now()
		p.handler.ServeHTTP(w, req)
		latency := time.Since(start).Milliseconds()
		object.TotalThroughput.Inc()
		object.APILatency.WithLabelValues(endpoint, method).Observe(float64(latency))
		object.APIThroughput.WithLabelValues(endpoint, method).Inc()
	}
}

func recordSystemInfo(systemInfo *util.SystemInfo) {
	for i, value := range systemInfo.CpuUsage {
		object.CpuUsage.WithLabelValues(fmt.Sprintf("%d", i)).Set(value)
	}
	object.MemoryUsage.WithLabelValues("memoryUsed").Set(float64(systemInfo.MemoryUsed))
	object.MemoryUsage.WithLabelValues("memoryTotal").Set(float64(systemInfo.MemoryTotal))
}

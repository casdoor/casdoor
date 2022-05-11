package object

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var VisitTime = promauto.NewCounter(prometheus.CounterOpts{
	Name: "casdoor_visit_times",
	Help: "The total number of casdoor api visits",
})
var LoginTimes = promauto.NewCounter(prometheus.CounterOpts{
	Name: "casdoor_login_times",
	Help: "The total number of casdoor login",
})

var ResponseTime = promauto.NewSummary(prometheus.SummaryOpts{
	Name: "casdoor_api_responseTime",
	Help: "The summary of casdoor apis' response time in ms",
})

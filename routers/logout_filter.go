package routers

import (
	"fmt"
	"github.com/beego/beego/context"
	"github.com/beego/beego/logs"
	"github.com/casdoor/casdoor/conf"
	"time"
)

var (
	logoutMinutes   = time.Minute * 30
	cookie2LastTime = make(map[string]time.Time, 1)
)

func init() {
	logoutMinutes_int, err := conf.GetConfigInt64("logoutMinutes")
	if err != nil {
		logs.Info(fmt.Sprintf("get logoutMinutes failed, err:%v. use default time duration: 30 minutes", err))
	} else {
		logoutMinutes = time.Minute * time.Duration(logoutMinutes_int)
	}
}

func LogoutFilter(ctx *context.Context) {
	owner, name := getSubject(ctx)
	if owner != "anonymous" && name != "anonymous" {
		sessionId := ctx.Input.CruSession.SessionID()
		currentTime := time.Now()
		if cookieTime, exist := cookie2LastTime[sessionId]; exist {
			if cookieTime.Add(logoutMinutes).Before(currentTime) {
				delete(cookie2LastTime, sessionId)
				ctx.Input.CruSession.Set("username", "")
				ctx.Input.CruSession.Set("accessToken", "")
				ctx.Input.CruSession.Delete("SessionData")
				responseError(ctx, T(ctx, "auth:Long time of no operation"))
			} else {
				cookie2LastTime[sessionId] = currentTime
			}
		} else {
			cookie2LastTime[sessionId] = currentTime
		}
	}
}

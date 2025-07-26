package errorx

import (
	"errors"
	"fmt"

	"github.com/zdzh/errorx/errcode"
)

//  错误码格式建议：
//  采用 10 位数字，前 3 位为所属 app/模块编号，中间 2 位为大类（01=系统，02=认证，03=资源，04=业务，05=外部服务，06=客户端），后 5 位为具体错误编号
//  推荐在展示时使用分隔符格式：AAA-BB-CCCCC，例如：
//  100-01-00001 = 用户中心-系统未知错误
//  100-02-00001 = 用户中心-未授权
//  100-03-00001 = 用户中心-资源未找到
//  100-04-00001 = 用户中心-业务校验失败
//  100-05-00001 = 用户中心-外部服务调用失败
//  200-01-00001 = 订单中心-系统未知错误

const (
	defaultErrCode        = 1010000001
	incorrectPassword     = 1000200001
	loginFailed           = 1000200002
	grantTypeErr          = 1000200003
	UserUnexist           = 1000200004
	LinkUserFailed        = 1000200005
	AuthenticationFailed  = 1000200006
	UnAuthentication      = 1000200007
	Unauthorized          = 1000200008
	AuthenticationExpired = 1000200009
	RequestTimeout        = 1000100001

	InvalidParam    = 1000600001
	NotFoundAppCode = 1000600101
)

func init() {
	errcode.SetDefaultCode(defaultErrCode)
}

var (
	DefaultErr               = errcode.WithCode(errors.New("subscription:Error"), defaultErrCode)
	LoginErr                 = errcode.WithCode(errors.New("check:password or code is incorrect"), incorrectPassword)
	InvalidParamErr          = errcode.WithCode(errors.New("参数错误"), InvalidParam)
	LinkUserErr              = errcode.WithCode(errors.New("关联用户失败"), LinkUserFailed)
	AuthenticationErr        = errcode.WithCode(errors.New("认证失败"), AuthenticationFailed)
	UnAuthenticationErr      = errcode.WithCode(errors.New("认证不通过"), UnAuthentication)
	AuthenticationExpiredErr = errcode.WithCode(errors.New("认证已过期"), AuthenticationExpired)
	UnauthorizedErr          = errcode.WithCode(errors.New("未授权"), Unauthorized)
	RequestTimeoutErr        = errcode.WithCode(errors.New("请求超时"), RequestTimeout)
	NotFoundAppErr           = errcode.WithCode(errors.New("未找到应用"), NotFoundAppCode)
)

var (
	LoginErrFunc = func(message string) error {
		return errcode.WithMessage(errors.New("登录失败"), loginFailed, message)
	}
	GrantTypeErrFunc = func(grant_type string) error {
		return errcode.WithCode(fmt.Errorf("授权方式(%s)错误", grant_type), grantTypeErr)
	}
	UserUnexistErrFunc = func(user string) error {
		return errcode.WithCode(fmt.Errorf("用户(%s)不存在", user), UserUnexist)
	}
	UnauthorizedErrFunc = func(message string) error {
		return errcode.WithMessage(errors.New(message), Unauthorized, "未授权")
	}
)

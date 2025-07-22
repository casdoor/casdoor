package errorx

import (
	"errors"
	"fmt"

	"github.com/zdzh/errorx/errcode"
)

const (
	defaultErrCode    = 101_00_00000
	incorrectPassword = 100_02_00001
	loginFailed       = 100_02_00002
	grantTypeErr      = 100_02_00003
	UserUnexist       = 100_02_00004
	LinkUserFailed       = 100_02_00005

	InvalidParam = 100_06_00001
)

func init() {
	errcode.SetDefaultCode(defaultErrCode)
}

var (
	DefaultErr      = errcode.WithCode(errors.New("subscription:Error"), defaultErrCode)
	LoginErr        = errcode.WithCode(errors.New("check:password or code is incorrect"), incorrectPassword)
	InvalidParamErr = errcode.WithCode(errors.New("参数错误"), InvalidParam)
	LinkUserErr = errcode.WithCode(errors.New("关联用户失败"), LinkUserFailed)

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
	
)

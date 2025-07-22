package errorx

import (
	"errors"

	"github.com/zdzh/errorx/errcode"
)

const defaultErrCode = 101_00_00000

func init() {
	errcode.SetDefaultCode(defaultErrCode)
}

var (
	DefaultErr = errcode.WithCode(errors.New("subscription:Error"), defaultErrCode)
	LoginErr   = errcode.WithCode(errors.New("check:password or code is incorrect"), 100_02_00001)

	InvalidParamErr = errcode.WithCode(errors.New("参数错误"), 100_06_00001)
)

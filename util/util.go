package util

import (
	"fmt"

	"github.com/astaxie/beego/logs"
)

func SafeGoroutine(fn func()) {
	var err error
	go func() {
		defer func() {
			if r := recover(); r != nil {
				var ok bool
				err, ok = r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				logs.Error("goroutine panic: %v", err)
			}
		}()
		fn()
	}()
}
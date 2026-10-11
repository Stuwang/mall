package codeerror

import (
	"fmt"

	"helloworld/base/logger"
)

func (ce CodeError) Error() string {
	return fmt.Sprintf("code %d errormsg %v", ce.Code, ce.Msg)
}

func (ce CodeError) Is(o CodeError) bool {
	return ce.Code == o.Code
}

func (ce CodeError) Errorf(msg string, args ...interface{}) CodeError {
	return CodeError{
		Code: ce.Code,
		Msg:  fmt.Sprintf(msg, args),
	}
}

// As 判断是否为业务 CodeError
func As(err error) (CodeError, bool) {
	if err == nil {
		return CodeError{}, false
	}
	if ce, ok := err.(CodeError); ok {
		return ce, true
	}
	if ce, ok := err.(*CodeError); ok && ce != nil {
		return *ce, true
	}
	return CodeError{}, false
}

// OrSystem 先打错误日志；业务 CodeError 原样返回，未知错误返回 ErrorSystem
func OrSystem(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	logger.Errorf(format+": %v", append(args, err)...)
	if ce, ok := As(err); ok {
		return ce
	}
	return ErrorSystem
}

var (
	_ error = CodeError{}
	_ error = &CodeError{}
)

var (
	// OK = CodeError{0, "OK"} // no use because == nil check
	ErrorParam        = CodeError{Code: 1000, Msg: "param error"}
	ErrorExited       = CodeError{Code: 1001, Msg: "error alreadt exited"}
	ErrorUnimplied    = CodeError{Code: 1002, Msg: "no unimplied results returned"}
	ErrorPanic        = CodeError{Code: 1003, Msg: "panic error"}
	ErrorNotOnline    = CodeError{Code: 1004, Msg: "not online"}
	ErrorUnauthorized = CodeError{Code: 1005, Msg: "unauthorized"}
	ErrorUserExists   = CodeError{Code: 1006, Msg: "user already exists"}
	ErrorUserNotFound = CodeError{Code: 1007, Msg: "user not found"}
	ErrorPassword     = CodeError{Code: 1008, Msg: "password error"}
	ErrorForbidden         = CodeError{Code: 1009, Msg: "forbidden"}
	ErrorStockInsufficient = CodeError{Code: 1010, Msg: "stock insufficient"}
	ErrorAssetInsufficient = CodeError{Code: 1011, Msg: "asset insufficient"}
	ErrorOrderStatus       = CodeError{Code: 1012, Msg: "order status invalid"}
	ErrorPayExpired        = CodeError{Code: 1013, Msg: "pay expired"}
	ErrorUserBanned        = CodeError{Code: 1014, Msg: "user banned"}
	ErrorSystem            = CodeError{Code: 1, Msg: "system error"}
)

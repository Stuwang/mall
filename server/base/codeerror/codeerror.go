package codeerror

import (
	"fmt"
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
	ErrorSystem       = CodeError{Code: 1, Msg: "system error"}
)

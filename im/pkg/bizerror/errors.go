package bizerror

import "mall/kit/codeerror"

var (
	ErrorConvNotFound   = codeerror.CodeError{Code: 1104, Msg: "conversation not found"}
	ErrorSelfPeer       = codeerror.CodeError{Code: 1105, Msg: "cannot chat with self"}
	ErrorSystemToken    = codeerror.CodeError{Code: 1106, Msg: "system token invalid"}
	ErrorBadMessage     = codeerror.CodeError{Code: 1107, Msg: "bad message"}
	ErrorNotImplemented = codeerror.CodeError{Code: 1110, Msg: "not implemented"}
)

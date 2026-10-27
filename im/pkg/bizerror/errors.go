package bizerror

import "mall/kit/codeerror"

var (
	ErrorMsgNotFound    = codeerror.CodeError{Code: 1104, Msg: "message not found"}
	ErrorSelfPeer       = codeerror.CodeError{Code: 1105, Msg: "cannot chat with self"}
	ErrorSystemToken    = codeerror.CodeError{Code: 1106, Msg: "system token invalid"}
	ErrorBadMessage     = codeerror.CodeError{Code: 1107, Msg: "bad message"}
	ErrorIdempotency    = codeerror.CodeError{Code: 1108, Msg: "client message id reused with different body"}
	ErrorNotImplemented = codeerror.CodeError{Code: 1110, Msg: "not implemented"}
	ErrorAlbumPhoto     = codeerror.CodeError{Code: 1120, Msg: "album photo not found"}
	ErrorAlbumFull      = codeerror.CodeError{Code: 1121, Msg: "album photo limit reached"}
)

package bizerror

import "mall/kit/codeerror"

// 商城业务错误。通用错误在 mall/kit/codeerror。
var (
	ErrorStockInsufficient = codeerror.CodeError{Code: 1010, Msg: "stock insufficient"}
	ErrorAssetInsufficient = codeerror.CodeError{Code: 1011, Msg: "asset insufficient"}
	ErrorOrderStatus       = codeerror.CodeError{Code: 1012, Msg: "order status invalid"}
	ErrorPayExpired        = codeerror.CodeError{Code: 1013, Msg: "pay expired"}
)

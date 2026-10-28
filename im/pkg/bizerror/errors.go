package bizerror

import "mall/kit/codeerror"

var (
	ErrorMsgNotFound       = codeerror.CodeError{Code: 1104, Msg: "message not found"}
	ErrorSelfPeer          = codeerror.CodeError{Code: 1105, Msg: "cannot chat with self"}
	ErrorSystemToken       = codeerror.CodeError{Code: 1106, Msg: "system token invalid"}
	ErrorBadMessage        = codeerror.CodeError{Code: 1107, Msg: "bad message"}
	ErrorIdempotency       = codeerror.CodeError{Code: 1108, Msg: "client message id reused with different body"}
	ErrorNotImplemented    = codeerror.CodeError{Code: 1110, Msg: "not implemented"}
	ErrorAlbumPhoto        = codeerror.CodeError{Code: 1120, Msg: "album photo not found"}
	ErrorAlbumFull         = codeerror.CodeError{Code: 1121, Msg: "album photo limit reached"}
	ErrorPostNotFound      = codeerror.CodeError{Code: 1130, Msg: "post not found"}
	ErrorCommentNotFound   = codeerror.CodeError{Code: 1131, Msg: "comment not found"}
	ErrorPostIdempotency   = codeerror.CodeError{Code: 1133, Msg: "client id reused with different body"}
	ErrorCoinInsufficient  = codeerror.CodeError{Code: 1140, Msg: "金币不足"}
	ErrorPayChannelClosed  = codeerror.CodeError{Code: 1141, Msg: "支付渠道未开通"}
	ErrorCoinOrderNotFound = codeerror.CodeError{Code: 1142, Msg: "充值订单不存在"}
	ErrorCoinOrderStatus   = codeerror.CodeError{Code: 1143, Msg: "订单状态不允许"}
	ErrorCoinOrderExpired  = codeerror.CodeError{Code: 1144, Msg: "订单已过期"}
	ErrorGiftNotFound      = codeerror.CodeError{Code: 1145, Msg: "礼物不存在"}
)

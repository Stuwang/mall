package controller

import (
	"context"
	"crypto/subtle"

	"github.com/go-kratos/kratos/v2/transport"

	api "mall/im/api/im"
	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	"mall/kit/auth"
	"mall/kit/logger"
)

const headerSystemToken = "X-Im-System-Token"

var systemToken string

// SetSystemToken 由进程启动时写入内部代发令牌。
func SetSystemToken(token string) {
	systemToken = token
}

// UserController 用户侧 HTTP。业务事务尚未实现。
type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (c *UserController) OpenConv(ctx context.Context, req *api.OpenConvRequest) (*api.Conversation, error) {
	a := auth.MustFromContext(ctx)
	logger.Errorf("OpenConv not implemented user=%d peer=%d", a.UserID, req.GetPeerUserId())
	return nil, biz.NotImplemented()
}

func (c *UserController) ListConv(ctx context.Context, req *api.ListConvRequest) (*api.ListConvReply, error) {
	a := auth.MustFromContext(ctx)
	logger.Errorf("ListConv not implemented user=%d page=%d", a.UserID, req.GetPage())
	return nil, biz.NotImplemented()
}

func (c *UserController) History(ctx context.Context, req *api.HistoryRequest) (*api.HistoryReply, error) {
	a := auth.MustFromContext(ctx)
	logger.Errorf("History not implemented user=%d peer=%d", a.UserID, req.GetPeerUserId())
	return nil, biz.NotImplemented()
}

func (c *UserController) Send(ctx context.Context, req *api.SendRequest) (*api.Msg, error) {
	a := auth.MustFromContext(ctx)
	logger.Errorf("Send not implemented user=%d peer=%d", a.UserID, req.GetPeerUserId())
	return nil, biz.NotImplemented()
}

// InternalController 内部代发。序号分配和预览开关尚未实现。
type InternalController struct{}

func NewInternalController() *InternalController {
	return &InternalController{}
}

func (c *InternalController) Send(ctx context.Context, req *api.InternalSendRequest) (*api.InternalSendReply, error) {
	if err := requireSystemToken(ctx); err != nil {
		return nil, err
	}
	logger.Errorf("internal Send not implemented sender=%d peer=%d visibility=%d senderPreview=%v peerPreview=%v",
		req.GetSenderId(), req.GetPeerUserId(), req.GetVisibility(), req.GetSenderPreview(), req.GetPeerPreview())
	return nil, biz.NotImplemented()
}

func (c *InternalController) Update(ctx context.Context, req *api.InternalUpdateRequest) (*api.InternalSendReply, error) {
	if err := requireSystemToken(ctx); err != nil {
		return nil, err
	}
	logger.Errorf("internal Update not implemented msgID=%d senderPreview=%v peerPreview=%v",
		req.GetMsgId(), req.GetSenderPreview(), req.GetPeerPreview())
	return nil, biz.NotImplemented()
}

func requireSystemToken(ctx context.Context) error {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		logger.Errorf("internal token missing transport")
		return bizerror.ErrorSystemToken
	}
	got := tr.RequestHeader().Get(headerSystemToken)
	if systemToken == "" || subtle.ConstantTimeCompare([]byte(got), []byte(systemToken)) != 1 {
		logger.Errorf("internal token mismatch")
		return bizerror.ErrorSystemToken
	}
	return nil
}

var (
	_ api.ImHTTPServer         = (*UserController)(nil)
	_ api.ImInternalHTTPServer = (*InternalController)(nil)
)

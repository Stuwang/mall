package controller

import (
	"context"
	"crypto/subtle"

	"github.com/go-kratos/kratos/v2/transport"

	api "mall/im/api/im"
	"mall/im/internal/hub"
	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	"mall/kit/auth"
	"mall/kit/logger"
)

const headerSystemToken = "X-Im-System-Token"

var (
	systemToken string
	onlineHub   *hub.Hub
)

// SetSystemToken 由进程启动时写入内部代发令牌。
func SetSystemToken(token string) {
	systemToken = token
}

// SetHub 由进程启动时注入同一份在线连接表，用于提交成功后推送。
func SetHub(h *hub.Hub) {
	onlineHub = h
}

// UserController 用户侧 HTTP。
type UserController struct{}

func NewUserController() *UserController {
	return &UserController{}
}

func (c *UserController) ListConv(ctx context.Context, req *api.ListConvRequest) (*api.ListConvReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.ListConv(a.UserID, req.GetAfterSeq(), req.GetLimit())
	if err != nil {
		logger.Errorf("ListConv user=%d afterSeq=%d err=%v", a.UserID, req.GetAfterSeq(), err)
		return nil, err
	}
	return reply, nil
}

func (c *UserController) History(ctx context.Context, req *api.HistoryRequest) (*api.HistoryReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.History(a.UserID, req.GetPeerUserId(), req.GetAfterSeq(), req.GetLimit())
	if err != nil {
		logger.Errorf("History user=%d peer=%d afterSeq=%d err=%v", a.UserID, req.GetPeerUserId(), req.GetAfterSeq(), err)
		return nil, err
	}
	return reply, nil
}

func (c *UserController) Send(ctx context.Context, req *api.SendRequest) (*api.SendReply, error) {
	a := auth.MustFromContext(ctx)
	reply, pushes, err := biz.Send(a.UserID, req.GetPeerUserId(), req.GetClientMsgId(), req.GetContentType(), req.GetText())
	if err != nil {
		logger.Errorf("Send user=%d peer=%d err=%v", a.UserID, req.GetPeerUserId(), err)
		return nil, err
	}
	pushAll(pushes)
	return reply, nil
}

// InternalController 内部代发。
type InternalController struct{}

func NewInternalController() *InternalController {
	return &InternalController{}
}

func (c *InternalController) Send(ctx context.Context, req *api.InternalSendRequest) (*api.InternalSendReply, error) {
	if err := requireSystemToken(ctx); err != nil {
		return nil, err
	}
	reply, pushes, err := biz.InternalSend(req)
	if err != nil {
		logger.Errorf("internal Send sender=%d peer=%d err=%v", req.GetSenderId(), req.GetPeerUserId(), err)
		return nil, err
	}
	pushAll(pushes)
	return reply, nil
}

func (c *InternalController) Update(ctx context.Context, req *api.InternalUpdateRequest) (*api.InternalSendReply, error) {
	if err := requireSystemToken(ctx); err != nil {
		return nil, err
	}
	reply, pushes, err := biz.InternalUpdate(req)
	if err != nil {
		logger.Errorf("internal Update msgID=%d err=%v", req.GetMsgId(), err)
		return nil, err
	}
	pushAll(pushes)
	return reply, nil
}

func pushAll(pushes []biz.PushItem) {
	if onlineHub == nil {
		return
	}
	for _, p := range pushes {
		if err := onlineHub.PushMsg(p.UserID, p.Msg); err != nil {
			logger.Errorf("push msg user=%d msgID=%d err=%v", p.UserID, p.Msg.GetMsgId(), err)
		}
	}
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

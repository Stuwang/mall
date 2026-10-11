package auth

import (
	"context"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"helloworld/base/codeerror"
)

type authKey struct{}

// AuthInfo 当前登录用户信息
type AuthInfo struct {
	UserID   int64
	Username string
}

type options struct {
	whitelist map[string]struct{}
}

type Option func(*options)

// WithWhitelist 配置免鉴权的 operation（如 /user.v1.User/Login）
func WithWhitelist(ops ...string) Option {
	return func(o *options) {
		for _, op := range ops {
			o.whitelist[op] = struct{}{}
		}
	}
}

// NewContext 将鉴权信息写入 context
func NewContext(ctx context.Context, info *AuthInfo) context.Context {
	return context.WithValue(ctx, authKey{}, info)
}

// FromContext 读取鉴权信息；白名单或可选登录场景使用
func FromContext(ctx context.Context) (*AuthInfo, bool) {
	info, ok := ctx.Value(authKey{}).(*AuthInfo)
	return info, ok && info != nil
}

// MustFromContext 受保护接口推荐写法；无鉴权信息时 panic（视为接线错误）
func MustFromContext(ctx context.Context) *AuthInfo {
	info, ok := FromContext(ctx)
	if !ok {
		panic("auth: AuthInfo missing in context; ensure Middleware protects this operation")
	}
	return info
}

// ValidateToken 校验 token 并返回用户信息（当前为 stub：非空即通过）
func ValidateToken(token string) (*AuthInfo, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, codeerror.ErrorUnauthorized
	}
	return &AuthInfo{
		UserID:   0,
		Username: "",
	}, nil
}

// Middleware 非白名单接口强制 Bearer 鉴权，通过后写入 context
func Middleware(opts ...Option) middleware.Middleware {
	o := &options{whitelist: make(map[string]struct{})}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if _, skip := o.whitelist[tr.Operation()]; skip {
					return handler(ctx, req)
				}
				authHeader := tr.RequestHeader().Get("Authorization")
				token, ok := parseBearer(authHeader)
				if !ok {
					return nil, codeerror.ErrorUnauthorized
				}
				info, err := ValidateToken(token)
				if err != nil {
					return nil, err
				}
				ctx = NewContext(ctx, info)
			} else {
				return nil, codeerror.ErrorUnauthorized
			}
			return handler(ctx, req)
		}
	}
}

func parseBearer(header string) (string, bool) {
	const prefix = "Bearer "
	if header == "" || !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", false
	}
	return token, true
}

package auth

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	khttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/golang-jwt/jwt/v5"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/config"
)

const headerUserID = "UserId"

type authKey struct{}

// AuthInfo 当前登录用户信息
type AuthInfo struct {
	UserID   int64
	Username string
	IsAdmin  bool
}

type jwtClaims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

// UserLoader 测试环境根据 UserId 补全用户信息（含 IsAdmin）
type UserLoader func(userID int64) (*AuthInfo, error)

// ActiveChecker 鉴权通过后校验账号是否可用（如封禁）
type ActiveChecker func(userID int64) error

type options struct {
	whitelist     map[string]struct{}
	pathPrefixes  []string
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

// WithPathPrefixWhitelist 配置免鉴权的 HTTP 路径前缀（如 /files/）
func WithPathPrefixWhitelist(prefixes ...string) Option {
	return func(o *options) {
		o.pathPrefixes = append(o.pathPrefixes, prefixes...)
	}
}

var (
	env            string
	jwtSecret      []byte
	jwtExpireHours int
	userLoader     UserLoader
	activeChecker  ActiveChecker
)

// Init 根据配置初始化鉴权（须在 Middleware 使用前调用）
func Init(cfg config.AuthConfig, envName string) {
	env = envName
	jwtSecret = []byte(cfg.JWTSecret)
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("dev-secret")
	}
	jwtExpireHours = cfg.JWTExpireHours
	if jwtExpireHours <= 0 {
		jwtExpireHours = 168
	}
}

// SetUserLoader 设置用户信息加载器（测试环境 UserId Header 鉴权时使用）
func SetUserLoader(loader UserLoader) {
	userLoader = loader
}

// SetActiveChecker 设置账号可用性检查（封禁等）；鉴权成功后对非白名单请求生效
func SetActiveChecker(checker ActiveChecker) {
	activeChecker = checker
}

func isTestEnv() bool {
	return env == config.EnvTest
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

// RequireAdmin 要求当前用户为管理员
func RequireAdmin(ctx context.Context) (*AuthInfo, error) {
	info, ok := FromContext(ctx)
	if !ok {
		logger.Errorf("RequireAdmin missing auth info")
		return nil, codeerror.ErrorUnauthorized
	}
	if !info.IsAdmin {
		logger.Errorf("RequireAdmin forbidden userID=%d", info.UserID)
		return nil, codeerror.ErrorForbidden
	}
	return info, nil
}

// GenerateToken 使用 JWT 生成登录 token（业务登录成功后调用）
func GenerateToken(userID int64, username string, isAdmin bool) (string, error) {
	if len(jwtSecret) == 0 {
		return "", fmt.Errorf("auth: jwt secret not initialized")
	}
	now := time.Now()
	claims := jwtClaims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(jwtExpireHours) * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken 校验 JWT 并返回用户信息
func ValidateToken(tokenString string) (*AuthInfo, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, codeerror.ErrorUnauthorized
	}
	if len(jwtSecret) == 0 {
		return nil, codeerror.ErrorUnauthorized
	}
	token, err := jwt.ParseWithClaims(tokenString, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, codeerror.ErrorUnauthorized
	}
	claims, ok := token.Claims.(*jwtClaims)
	if !ok {
		return nil, codeerror.ErrorUnauthorized
	}
	return &AuthInfo{
		UserID:   claims.UserID,
		Username: claims.Username,
		IsAdmin:  claims.IsAdmin,
	}, nil
}

// Middleware 非白名单接口强制鉴权，通过后写入 context
func Middleware(opts ...Option) middleware.Middleware {
	o := &options{whitelist: make(map[string]struct{})}
	for _, opt := range opts {
		opt(o)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, codeerror.ErrorUnauthorized
			}
			if _, skip := o.whitelist[tr.Operation()]; skip {
				return handler(ctx, req)
			}
			if len(o.pathPrefixes) > 0 {
				if reqHTTP, ok := khttp.RequestFromServerContext(ctx); ok {
					path := reqHTTP.URL.Path
					for _, p := range o.pathPrefixes {
						if p != "" && strings.HasPrefix(path, p) {
							return handler(ctx, req)
						}
					}
				}
			}

			var (
				info *AuthInfo
				err  error
			)
			if isTestEnv() {
				info, err = authenticateByUserIDHeader(tr)
			} else {
				info, err = authenticateByJWT(tr)
			}
			if err != nil {
				return nil, err
			}
			if activeChecker != nil {
				if err := activeChecker(info.UserID); err != nil {
					logger.Errorf("auth ActiveChecker userID=%d err=%v", info.UserID, err)
					return nil, err
				}
			}
			return handler(NewContext(ctx, info), req)
		}
	}
}

func authenticateByUserIDHeader(tr transport.Transporter) (*AuthInfo, error) {
	raw := strings.TrimSpace(tr.RequestHeader().Get(headerUserID))
	if raw == "" {
		return nil, codeerror.ErrorUnauthorized
	}
	uid, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || uid <= 0 {
		return nil, codeerror.ErrorUnauthorized
	}
	if userLoader != nil {
		info, err := userLoader(uid)
		if err != nil {
			return nil, err
		}
		return info, nil
	}
	return &AuthInfo{UserID: uid}, nil
}

func authenticateByJWT(tr transport.Transporter) (*AuthInfo, error) {
	token, ok := parseBearer(tr.RequestHeader().Get("Authorization"))
	if !ok {
		return nil, codeerror.ErrorUnauthorized
	}
	return ValidateToken(token)
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

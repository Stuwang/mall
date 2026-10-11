package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"helloworld/api/helloworld"
	apiasset "helloworld/api/asset"
	apicart "helloworld/api/cart"
	apimall "helloworld/api/mall"
	apimisc "helloworld/api/misc"
	apiorder "helloworld/api/order"
	apiproduct "helloworld/api/product"
	apiuser "helloworld/api/user"
	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/common/config"
	"helloworld/common/database"
	"helloworld/pkg/assetcontroller"
	"helloworld/pkg/cartcontroller"
	"helloworld/pkg/dao/assetdao"
	"helloworld/pkg/dao/cartdao"
	"helloworld/pkg/dao/miscdao"
	"helloworld/pkg/dao/orderdao"
	"helloworld/pkg/dao/productdao"
	"helloworld/pkg/dao/userdao"
	"helloworld/pkg/mallcontroller"
	"helloworld/pkg/mediacontroller"
	"helloworld/pkg/mediastore"
	"helloworld/pkg/misccontroller"
	"helloworld/pkg/orderbiz"
	"helloworld/pkg/ordercontroller"
	"helloworld/pkg/productcontroller"
	"helloworld/pkg/server"
	"helloworld/pkg/usercontroller"

	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

type greeterServer struct{}

func (s greeterServer) SayHello2(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	a := auth.MustFromContext(ctx)
	return &helloworld.HelloReply{
		Message:     "hello " + request.Name,
		AccessToken: fmt.Sprintf("token %v", a.UserID),
		Type:        helloworld.Type_Type_C,
		Value:       123,
	}, nil
}

func (s greeterServer) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return nil, codeerror.ErrorParam
}

// Init 初始化日志、数据库、鉴权、业务依赖。调用方负责 defer database.Close()。
func Init(cfg *config.Config) {
	logger.InitLogger(cfg.Logger.Filename)
	logger.Infof("begin to start")
	database.InitDb(cfg.Database.Filename)

	if err := userdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate user: %v", err)
	}
	if err := userdao.EnsureAdmins(cfg.Admins); err != nil {
		logger.Fatalf("ensure admins: %v", err)
	}
	if err := assetdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate asset: %v", err)
	}
	if err := productdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate product: %v", err)
	}
	if err := cartdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate cart: %v", err)
	}
	if err := orderdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate order: %v", err)
	}
	if err := miscdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate misc: %v", err)
	}

	auth.Init(cfg.Auth, cfg.Env)
	auth.SetUserLoader(func(userID int64) (*auth.AuthInfo, error) {
		u, err := userdao.NewUserDao(nil).GetByID(userID)
		if err != nil {
			logger.Errorf("UserLoader GetByID userID=%d err=%v", userID, err)
			return nil, err
		}
		login, err := userdao.NewLoginDao(nil).GetPasswordByUserID(userID)
		username := ""
		if err == nil {
			username = login.Identifier
		} else if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorUserNotFound) {
			logger.Errorf("UserLoader GetPasswordByUserID userID=%d err=%v", userID, err)
			return nil, err
		}
		return &auth.AuthInfo{
			UserID:   u.ID,
			Username: username,
			IsAdmin:  u.Admin(),
		}, nil
	})
	auth.SetActiveChecker(func(userID int64) error {
		u, err := userdao.NewUserDao(nil).GetByID(userID)
		if err != nil {
			logger.Errorf("ActiveChecker GetByID userID=%d err=%v", userID, err)
			return err
		}
		if u.Banned() {
			logger.Errorf("ActiveChecker user banned userID=%d banUntil=%d", userID, u.BanUntil)
			return codeerror.ErrorUserBanned
		}
		return nil
	})

	orderbiz.Init(cfg.Order.PayTimeoutSec, cfg.Point.PointsPerYuan)
	orderbiz.StartTimeoutWorker(30 * time.Second)
	if err := mediastore.Init(cfg.Storage); err != nil {
		logger.Fatalf("mediastore init: %v", err)
	}
}

// NewHTTPServer 注册全部 HTTP API 并返回 Kratos HTTP Server。
func NewHTTPServer(cfg *config.Config) *khttp.Server {
	logger.Infof("server starting, env=%s addr=%s", cfg.Env, cfg.Server.HTTPAddr)

	httpSrv := khttp.NewServer(server.HTTPOptions(cfg.Server.HTTPAddr)...)
	helloworld.RegisterGreeterHTTPServer(httpSrv, greeterServer{})
	userCtrl := usercontroller.New()
	apiuser.RegisterUserHTTPServer(httpSrv, userCtrl)
	apiuser.RegisterAdminUserHTTPServer(httpSrv, userCtrl)
	assetCtrl := assetcontroller.New()
	apiasset.RegisterAssetHTTPServer(httpSrv, assetCtrl)
	apiasset.RegisterAdminAssetHTTPServer(httpSrv, assetCtrl)
	productCtrl := productcontroller.New()
	apiproduct.RegisterProductHTTPServer(httpSrv, productCtrl)
	apiproduct.RegisterAdminProductHTTPServer(httpSrv, productCtrl)
	apicart.RegisterCartHTTPServer(httpSrv, cartcontroller.New())
	orderCtrl := ordercontroller.New()
	apiorder.RegisterOrderHTTPServer(httpSrv, orderCtrl)
	apiorder.RegisterAdminOrderHTTPServer(httpSrv, orderCtrl)
	apimall.RegisterMallHTTPServer(httpSrv, mallcontroller.New(cfg.Point.PointsPerYuan, cfg.Mall))
	miscCtrl := misccontroller.New()
	apimisc.RegisterMiscHTTPServer(httpSrv, miscCtrl)
	apimisc.RegisterAdminMiscHTTPServer(httpSrv, miscCtrl)
	mediacontroller.RegisterHTTP(httpSrv)
	return httpSrv
}

// New 组装 Kratos App（仅 API）。ctx 非 nil 时随其取消退出。
func New(cfg *config.Config, ctx context.Context) *kratos.App {
	httpSrv := NewHTTPServer(cfg)
	opts := []kratos.Option{
		kratos.Name("helloworld"),
		kratos.Server(httpSrv),
	}
	if ctx != nil {
		opts = append(opts, kratos.Context(ctx))
	}
	return kratos.New(opts...)
}

// Run 初始化并阻塞运行 API 服务。
func Run(cfg *config.Config) {
	Init(cfg)
	defer database.Close()

	a := New(cfg, nil)
	if err := a.Run(); err != nil {
		logger.Fatalf("%v", err)
	}
}

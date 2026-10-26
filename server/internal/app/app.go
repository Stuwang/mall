package app

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	apiasset "helloworld/api/asset"
	apicart "helloworld/api/cart"
	"helloworld/api/helloworld"
	apimall "helloworld/api/mall"
	apimisc "helloworld/api/misc"
	apiorder "helloworld/api/order"
	apiproduct "helloworld/api/product"
	"helloworld/common/config"
	"helloworld/common/database"
	"helloworld/pkg/assetcontroller"
	"helloworld/pkg/cartcontroller"
	"helloworld/pkg/dao/assetdao"
	"helloworld/pkg/dao/cartdao"
	"helloworld/pkg/dao/miscdao"
	"helloworld/pkg/dao/orderdao"
	"helloworld/pkg/dao/productdao"
	"helloworld/pkg/mallcontroller"
	"helloworld/pkg/misccontroller"
	"helloworld/pkg/orderbiz"
	"helloworld/pkg/ordercontroller"
	"helloworld/pkg/productcontroller"
	"helloworld/pkg/server"
	"mall/file/mediacontroller"
	"mall/file/mediastore"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
	apiuser "mall/user/api/user"
	"mall/user/controller"
	"mall/user/dao"

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
	userdao.SetDB(database.GetDb())

	if err := userdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate user: %v", err)
	}
	admins := make([]userdao.AdminAccount, 0, len(cfg.Admins))
	for _, a := range cfg.Admins {
		admins = append(admins, userdao.AdminAccount{
			Username: a.Username,
			Password: a.Password,
			Nickname: a.Nickname,
		})
	}
	if err := userdao.EnsureAdmins(admins); err != nil {
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

	auth.Init(auth.Config{
		JWTSecret:      cfg.Auth.JWTSecret,
		JWTExpireHours: cfg.Auth.JWTExpireHours,
	}, cfg.Env)
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
	if err := mediastore.Init(mediastore.Config{
		PublicBase:     cfg.Storage.PublicBase,
		MaxUploadBytes: cfg.Storage.MaxUploadBytes,
		RootDir:        cfg.Storage.Local.RootDir,
	}); err != nil {
		logger.Fatalf("mediastore init: %v", err)
	}
	usercontroller.SetAvatarResolver(func(uri string) string {
		return mediastore.Get().PathURL(uri)
	})
	usercontroller.SetAvatarNormalizer(func(raw string) string {
		return mediastore.Get().Normalize(raw)
	})
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
	mediacontroller.RegisterUserUpload(httpSrv)
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

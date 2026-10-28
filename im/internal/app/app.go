package app

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"mall/file/mediacontroller"
	"mall/file/mediastore"
	coinapi "mall/im/api/coin"
	api "mall/im/api/im"
	apipost "mall/im/api/post"
	apiprofile "mall/im/api/profile"
	"mall/im/common/config"
	"mall/im/common/database"
	"mall/im/internal/hub"
	"mall/im/internal/ws"
	"mall/im/pkg/biz"
	"mall/im/pkg/controller"
	imdao "mall/im/pkg/dao"
	"mall/im/pkg/server"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
	userapi "mall/user/api/user"
	usercontroller "mall/user/controller"
	userdao "mall/user/dao"
)

func Init(cfg *config.Config) {
	logger.InitLogger(cfg.Logger.Filename)
	logger.Infof("im begin to start")
	database.InitDb(cfg.Database.Filename)
	userdao.SetDB(database.GetDb())
	if err := userdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate user: %v", err)
	}
	if err := imdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate im: %v", err)
	}

	auth.Init(auth.Config{
		JWTSecret:      cfg.Auth.JWTSecret,
		JWTExpireHours: cfg.Auth.JWTExpireHours,
	}, cfg.Env)
	auth.SetUserLoader(loadAuthUser)
	auth.SetActiveChecker(checkUserActive)
	controller.SetSystemToken(cfg.IM.SystemToken)
	biz.SetTextMaxRunes(cfg.IM.TextMaxRunes)
	loadedGifts := cfg.Gifts()
	gifts := make([]biz.GiftItem, 0, len(loadedGifts))
	for _, g := range loadedGifts {
		gifts = append(gifts, biz.GiftItem{
			ID:     g.ID,
			Name:   g.Name,
			Icon:   g.Icon,
			Effect: g.Effect,
			Price:  g.Price,
		})
	}
	biz.SetCoinConfig(biz.CoinConfig{
		CoinsPerYuan:  cfg.Coin.CoinsPerYuan,
		PayTimeoutSec: cfg.Coin.PayTimeoutSec,
		Gifts:         gifts,
	})

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

func loadAuthUser(userID int64) (*auth.AuthInfo, error) {
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
}

func checkUserActive(userID int64) error {
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
}

func NewHTTPServer(cfg *config.Config) *khttp.Server {
	logger.Infof("im starting, env=%s addr=%s", cfg.Env, cfg.Server.HTTPAddr)
	httpSrv := khttp.NewServer(server.HTTPOptions(cfg.Server.HTTPAddr)...)
	userCtrl := usercontroller.New()
	userapi.RegisterUserHTTPServer(httpSrv, userCtrl)
	userapi.RegisterAdminUserHTTPServer(httpSrv, userCtrl)
	mediacontroller.RegisterFiles(httpSrv)
	mediacontroller.RegisterUserUpload(httpSrv)
	h := hub.New(cfg.IM.MaxConnsPerUser)
	controller.SetHub(h)
	ws.SetActiveChecker(checkUserActive)
	api.RegisterImHTTPServer(httpSrv, controller.NewUserController())
	api.RegisterImInternalHTTPServer(httpSrv, controller.NewInternalController())
	apiprofile.RegisterProfileHTTPServer(httpSrv, controller.NewProfileController())
	apipost.RegisterPostHTTPServer(httpSrv, controller.NewPostController())
	coinapi.RegisterCoinHTTPServer(httpSrv, controller.NewCoinController())
	ws.Register(httpSrv, h, ws.Config{
		IsTest:          cfg.IsTest(),
		HeartbeatSec:    cfg.IM.HeartbeatSec,
		IdleSec:         cfg.IM.IdleSec,
		MaxConnsPerUser: cfg.IM.MaxConnsPerUser,
	})
	return httpSrv
}

func New(cfg *config.Config, ctx context.Context) *kratos.App {
	opts := []kratos.Option{
		kratos.Name("im"),
		kratos.Server(NewHTTPServer(cfg)),
	}
	if ctx != nil {
		opts = append(opts, kratos.Context(ctx))
	}
	return kratos.New(opts...)
}

func Run(cfg *config.Config) {
	Init(cfg)
	defer database.Close()
	if err := New(cfg, nil).Run(); err != nil {
		logger.Fatalf("%v", err)
	}
}

package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	rawgrpc "google.golang.org/grpc"

	"helloworld/api/helloworld"
	apiasset "helloworld/api/asset"
	apiuser "helloworld/api/user"
	"helloworld/base/codeerror"
	"helloworld/base/hash_balancer"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/common/config"
	"helloworld/common/database"
	"helloworld/pkg/assetcontroller"
	"helloworld/pkg/dao/assetdao"
	"helloworld/pkg/dao/userdao"
	"helloworld/pkg/server"
	"helloworld/pkg/usercontroller"

	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

func e() {
	selector.SetGlobalSelector(hash_balancer.NewBuilder(func() selector.Selector {
		builder := selector.DefaultBuilder{}
		return builder.Build()
	}))
	rawgrpc.Dial("")
}

var (
	_ helloworld.GreeterHTTPServer = Server{}
)

type Server struct{}

func (s Server) SayHello2(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	a := auth.MustFromContext(ctx)
	return &helloworld.HelloReply{
		Message:     "hello " + request.Name,
		AccessToken: fmt.Sprintf("token %v", a.UserID),
		Type:        helloworld.Type_Type_C,
		Value:       123,
	}, nil
}

func (s Server) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return nil, codeerror.ErrorParam
}

func main() {
	confPath := flag.String("conf", "configs/config.yaml", "config path")
	flag.Parse()

	cfg, err := config.Load(*confPath)
	if err != nil {
		panic(err)
	}

	logger.InitLogger(cfg.Logger.Filename)
	logger.Infof("begin to start")
	database.InitDb(cfg.Database.Filename)
	defer database.Close()

	if err := userdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate user: %v", err)
	}
	if err := assetdao.AutoMigrate(); err != nil {
		logger.Fatalf("auto migrate asset: %v", err)
	}

	auth.Init(cfg.Auth, cfg.Env)
	logger.Infof("server starting, env=%s addr=%s", cfg.Env, cfg.Server.HTTPAddr)

	httpSrv := http.NewServer(server.HTTPOptions(cfg.Server.HTTPAddr)...)
	helloworld.RegisterGreeterHTTPServer(httpSrv, Server{})
	apiuser.RegisterUserHTTPServer(httpSrv, usercontroller.New())
	apiasset.RegisterAssetHTTPServer(httpSrv, assetcontroller.New())

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(httpSrv),
	)
	if err := app.Run(); err != nil {
		logger.Fatalf("%v", err)
	}
}

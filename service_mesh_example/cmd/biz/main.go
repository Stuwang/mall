package main

import (
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	bizv1 "mall/service_mesh_example/api/biz/v1"
	"mall/service_mesh_example/internal/biz"
	"mall/service_mesh_example/internal/config"
	"mall/service_mesh_example/internal/health"
	"mall/service_mesh_example/internal/peers"

	_ "github.com/go-kratos/kratos/v2/encoding/json"
)

func main() {
	logger := log.With(log.NewStdLogger(os.Stdout), "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	helper := log.NewHelper(logger)

	cfg, err := config.FromEnv()
	if err != nil {
		helper.Errorf("load config: %v", err)
		os.Exit(1)
	}
	helper = log.NewHelper(log.With(logger, "service", cfg.Name))

	pool := peers.New(cfg.Peers, logger)
	httpSrv := khttp.NewServer(
		khttp.Address(cfg.HTTPAddr),
		khttp.Middleware(
			recovery.Recovery(),
			logging.Server(logger),
		),
	)
	bizv1.RegisterBizHTTPServer(httpSrv, biz.New(cfg.Name, pool, logger))

	app := kratos.New(
		kratos.ID(cfg.Name),
		kratos.Name(cfg.Name),
		kratos.Logger(logger),
		kratos.Server(httpSrv, health.NewServer(cfg.HealthAddr)),
	)
	if err := app.Run(); err != nil {
		helper.Errorf("run: %v", err)
		os.Exit(1)
	}
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	stdhttp "net/http"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	rawgrpc "google.golang.org/grpc"

	"helloworld/api/helloworld"
	"helloworld/api/user"
	"helloworld/base/codeerror"
	"helloworld/base/hash_balancer"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/common/config"
	"helloworld/common/database"

	//_ "helloworld/base/pbjson"
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

type Server struct {
}

func (s Server) SayHello2(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{
		Message:     "hello " + request.Name,
		AccessToken: "token",
		Type:        helloworld.Type_Type_C,
		Value:       123,
	}, nil
}

func (s Server) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return nil, codeerror.ErrorParam
}

type CommonResp struct {
	Code     int               `json:"code"`
	Msg      string            `json:"msg"`
	MsgTrans map[string]string `json:"msg_trans"`
	Data     json.RawMessage   `json:"data"`
}

func setIp(handler middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		//headers, _ := metadata2.FromServerContext(ctx)
		//_ = headers
		//if p, ok := req.(base.ParamGetter); ok && p.GetBaseParam() != nil {
		//	p.GetBaseParam().Ip = "123123"
		//}
		return handler(ctx, req)
	}
}

func main() {
	confPath := flag.String("conf", "configs/config.yaml", "config path")
	flag.Parse()

	cfg, err := config.Load(*confPath)
	if err != nil {
		panic(err)
	}

	logger.InitLogger(cfg.Logger.Filename)
	database.InitDb(cfg.Database.Filename)
	defer database.Close()
	auth.Init(cfg.Auth, cfg.Env)

	logger.Infof("server starting, env=%s addr=%s", cfg.Env, cfg.Server.HTTPAddr)

	s := &Server{}
	httpSrv := http.NewServer(
		http.Address(cfg.Server.HTTPAddr),
		http.Middleware(
			metadata.Server(metadata.WithPropagatedPrefix("")),
			setIp,
			auth.Middleware(
				auth.WithWhitelist(
					user.OperationUserRegister,
					user.OperationUserLogin,
				),
			),
		),
		http.ResponseEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, i interface{}) error {
			writer.WriteHeader(stdhttp.StatusOK)
			//bs, err := protojson.MarshalOptions{
			//	EmitUnpopulated: true,
			//	UseProtoNames:   true,
			//	UseEnumNumbers:  true,
			//}.Marshal(i.(proto.Message))
			bs, err := json.Marshal(i)
			if err != nil {
				panic(err)
			}
			var d = CommonResp{
				Code:     0,
				Msg:      "success",
				MsgTrans: nil,
				Data:     json.RawMessage(bs),
			}
			bytes, err := json.Marshal(d)
			if err != nil {
				panic(err)
			}
			writer.Write(bytes)
			return nil
		}),
		http.ErrorEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
			writer.WriteHeader(stdhttp.StatusOK)
			var d = CommonResp{
				Code:     int(-1),
				Msg:      "",
				MsgTrans: nil,
				Data:     nil,
			}

			codeError, ok := err.(codeerror.CodeError)
			if ok {
				d = CommonResp{
					Code:     int(codeError.Code),
					Msg:      codeError.Msg,
					MsgTrans: nil,
					Data:     nil,
				}
			}
			codeErrorPointer, ok := err.(*codeerror.CodeError)
			if ok {
				d = CommonResp{
					Code:     int(codeErrorPointer.Code),
					Msg:      codeErrorPointer.Msg,
					MsgTrans: nil,
					Data:     nil,
				}
			}

			bytes, err := json.Marshal(d)
			if err != nil {
				panic(err)
			}
			writer.Write(bytes)
			return
		}),
	)
	helloworld.RegisterGreeterHTTPServer(httpSrv, s)

	httpSrv.HandleFunc("/hello", func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
		logger.Debugw("this is a log", "a", "b")
		writer.WriteHeader(stdhttp.StatusOK)
		writer.Write([]byte(`hello world`))
	})

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(
			httpSrv,
		),
	)

	if err := app.Run(); err != nil {
		logger.Fatalf("%v", err)
	}
}

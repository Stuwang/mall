package main

import (
	"context"
	"net/url"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	rawgrpc "google.golang.org/grpc"

	"helloworld/api/base"
	"helloworld/api/helloworld"
)

func e() {
	//selector.SetGlobalSelector(hash_balancer.NewBuilder(fun))
	rawgrpc.Dial("")
}

var (
	_ helloworld.GreeterHTTPServer = Server{}
)

type Server struct {
}

func (s Server) SayHello2(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return nil, nil
}

func (s Server) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return nil, base.ErrUserNotFound
}

func NewEndpoint(scheme, host string) *url.URL {
	return &url.URL{Scheme: scheme, Host: host}
}

func main() {
	consulClient, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	r := consul.New(consulClient)

	s := &Server{}
	grpcSrv := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			metadata.Server(),
		),
		grpc.UnaryInterceptor(base.ServerErrorInterceptor),
		grpc.Endpoint(NewEndpoint("grpc", "127.0.0.1:9000")),
	)
	helloworld.RegisterGreeterServer(grpcSrv, s)

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(
			grpcSrv,
		),
		kratos.Registrar(r),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

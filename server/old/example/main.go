package main

import (
	"context"
	"fmt"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/hashicorp/consul/api"
	rawgrpc "google.golang.org/grpc"
	"helloworld/api/helloworld"
	"helloworld/base/hash_balancer"
)

func e() {
	//selector.SetGlobalSelector(hash_balancer.NewBuilder())
	rawgrpc.Dial("")
	selector.SetGlobalSelector(hash_balancer.NewBuilder(func() selector.Selector {
		builder := selector.DefaultBuilder{}
		return builder.Build()
	}))
}

var (
	_ helloworld.GreeterHTTPServer = Server{}
)

type Server struct {
}

func (s Server) SayHello2(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	fmt.Println("call")
	return &helloworld.HelloReply{
		Message:     "aaa",
		AccessToken: "",
		Type:        0,
		Value:       0,
	}, nil
}

func (s Server) SayHello(ctx context.Context, request *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	fmt.Println("call " + request.Name)
	return &helloworld.HelloReply{
		Message:     "aaa",
		AccessToken: "",
		Type:        0,
		Value:       0,
	}, nil
}

func main() {
	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	_ = client
	register := consul.New(client, consul.WithHealthCheckInterval(5))

	s := &Server{}
	grpcSrv := grpc.NewServer(
		grpc.Address(":0"),
		grpc.Middleware(
			metadata.Server(),
		),
		grpc.UnaryInterceptor(),
	)
	httpSrv := http.NewServer(
		http.Address(":0"),
		//http.Address(":8000"),
		http.Middleware(),
		//http.ResponseEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, i interface{}) error {
		//	return nil
		//}),
		//http.ErrorEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
		//}),
	)
	helloworld.RegisterGreeterServer(grpcSrv, s)
	helloworld.RegisterGreeterHTTPServer(httpSrv, s)

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(
			//httpSrv,
			grpcSrv,
		),
		kratos.Registrar(register),
	)

	//go func() {
	//	log.Errorf("exit !!!")
	//	time.Sleep(time.Second * 10)
	//	panic("exit")
	//}()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

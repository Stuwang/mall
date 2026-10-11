package main

import (
	"context"
	zaplog "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/selector/wrr"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	stdhttp "net/http"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	"helloworld/api/base"
	"helloworld/api/helloworld"
)

func t() {
	//c, e := api.NewClient(nil)
	//c.Agent().ServiceRegister()
}

func main() {
	//selector.SetGlobalSelector(hash_balancer.NewBuilder(func() selector.Selector {
	//	builder := selector.DefaultBuilder{}
	//	return builder.Build()
	//}))
	selector.SetGlobalSelector(wrr.NewBuilder())
	callGRPC2()
}

func callHttp() {
	client, err := http.NewClient(context.Background(), http.WithEndpoint("127.0.0.1:9001"), http.WithTimeout(time.Second))
	if err != nil {
		panic(err)
	}
	helloworld.NewGreeterHTTPClient(client)

}

func callGRPC() {
	conn, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithEndpoint("127.0.0.1:9000"),
		grpc.WithUnaryInterceptor(base.ClientErrorInterceptor),
	)
	if err != nil {
		panic(err)
	}

	client := helloworld.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: "kratos"})
	if err != nil {
		if base.ErrUserNotFound.Is(err) {
			log.Info("user not found")
		}
		log.Infof("error %v", err)
	} else {
		log.Infof("[grpc] SayHello %+v\n", reply)
	}
}

func callHTTP2() {
	conn, err := http.NewClient(
		context.Background(),
		http.WithEndpoint("127.0.0.1:8000"),
	)
	if err != nil {
		panic(err)
	}
	client := helloworld.NewGreeterHTTPClient(conn)
	reply, err := client.SayHello(context.Background(), &helloworld.HelloRequest{Name: "empty"})
	if err != nil {
		if errors.Code(err) == stdhttp.StatusInternalServerError {
			log.Info(err)
		}
		//if base.IsUserNotFound(err) {
		//	log.Info("[http] USER_NOT_FOUND_ERROR", err)
		//}
	} else {
		log.Infof("[http] SayHello %s\n", reply.Message)
	}
}

func callGRPC2() {
	zc := zap.NewProductionConfig()
	zc.Level.SetLevel(zapcore.DebugLevel)
	zc.EncoderConfig.FunctionKey = "func"
	zc.Encoding = "console"
	l, err := zc.Build(zap.AddCallerSkip(2))
	if err != nil {
		panic(err)
	}
	log.SetLogger(zaplog.NewLogger(l))

	client, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		panic(err)
	}
	// new dis with consul client
	dis := consul.New(client)

	endpoint := "discovery:///helloworld"
	conn, err := grpc.DialInsecure(context.Background(), grpc.WithEndpoint(endpoint), grpc.WithDiscovery(dis))
	if err != nil {
		panic(err)
	}

	//for i := 0; i < 1000; i++ {
	//	k := i % 10
	//	kv := fmt.Sprintf("%v", k)
	//	client := helloworld.NewGreeterClient(conn)
	//	ctx := context.Background()
	//	ctx = context.WithValue(ctx, hash_balancer.BalancerKey, kv)
	//	reply, err := client.SayHello(ctx,
	//		&helloworld.HelloRequest{Name: kv},
	//	)
	//	if err != nil {
	//		e := errors.FromError(err)
	//		if e.Reason == "USER_NAME_EMPTY" && e.Code == 500 {
	//			log.Infof("[grpc] USER_NAME_EMPTY", err)
	//		}
	//		if base.IsUserNotFound(err) {
	//			log.Infof("[grpc] USER_NOT_FOUND_ERROR", err)
	//		}
	//	} else {
	//		log.Infof("[grpc] SayHello %+v\n", reply)
	//	}
	//}

	for i := 0; true; i++ {
		time.Sleep(time.Millisecond * 100)
		client := helloworld.NewGreeterClient(conn)
		reply, err := client.SayHello(context.Background(),
			&helloworld.HelloRequest{Name: "wrr " + strconv.FormatInt(int64(i), 10)},
		)
		if err != nil {
			e := errors.FromError(err)
			if e.Reason == "USER_NAME_EMPTY" && e.Code == 500 {
				log.Infof("[grpc] USER_NAME_EMPTY", err)
			}
			//if base.IsUserNotFound(err) {
			//	log.Infof("[grpc] USER_NOT_FOUND_ERROR", err)
			//}
		} else {
			_ = reply
			//log.Infof("[grpc] SayHello %+v\n", reply)
		}
	}
}

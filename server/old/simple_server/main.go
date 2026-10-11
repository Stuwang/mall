package main

import (
	"context"
	"encoding/json"
	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"helloworld/api/test_connect"
	stdhttp "net/http"
	"strings"
)

func main() {
	initLog()

}

func createConnectHttpServer() {

}

func startKratosHttpServer() {

	s := &Server{}
	httpSrv := http.NewServer(
		http.Address(":8080"),
		http.Filter(func(handler stdhttp.Handler) stdhttp.Handler {
			return stdhttp.HandlerFunc(func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
				handler.ServeHTTP(writer, request)
			})
		}),
		http.ResponseEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, i interface{}) error {
			writer.WriteHeader(stdhttp.StatusOK)
			bytes, err := json.Marshal(i)
			if err != nil {
				panic(err)
			}
			writer.Write(bytes)
			return nil
		}),
		http.ErrorEncoder(func(writer stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
			writer.WriteHeader(stdhttp.StatusOK)
			e := errors.FromError(err)
			bytes, err := json.Marshal(e)
			if err != nil {
				panic(err)
			}
			writer.Write(bytes)
			return
		}),
	)
	test_connect.RegisterGreeterHTTPServer(httpSrv, s)

	app := kratos.New(
		kratos.Name("helloworld"),
		kratos.Server(
			httpSrv,
		),
	)

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
func initLog() {
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.FunctionKey = "func"
	zapConfig.Encoding = "console"
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	zapConfig.EncoderConfig.ConsoleSeparator = " "
	zapConfig.Level.SetLevel(zapcore.DebugLevel)
	zapConfig.EncoderConfig.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		fullPath := caller.FullPath()
		count := 0
		index := strings.LastIndexFunc(fullPath, func(r rune) bool {
			if r == '/' {
				count++
			}
			if count >= 3 {
				return true
			}
			return false
		})
		encoder.AppendString(fullPath[index+1:])
	}
	l, err := zapConfig.Build()
	if err != nil {
		panic(err)
	}
	log.SetLogger(kzap.NewLogger(l.WithOptions(zap.AddCallerSkip(2))))

	log.Infof("this is a log")

}

var (
	_ test_connect.GreeterServer = Server{}
)

type Server struct {
}

func (s Server) SayHello(ctx context.Context, req *test_connect.Req) (*test_connect.Resp, error) {
	return &test_connect.Resp{Message: "this is " + req.Name}, nil
}

func (s Server) SayHello2(ctx context.Context, req *test_connect.Req) (*test_connect.Resp, error) {
	return &test_connect.Resp{Message: "2 this is " + req.Name}, nil
}

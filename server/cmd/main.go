package main

import (
	"context"
	"encoding/json"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/selector"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	rawgrpc "google.golang.org/grpc"
	"helloworld/base/codeerror"
	stdhttp "net/http"
	"strings"

	"helloworld/api/helloworld"
	"helloworld/base/hash_balancer"

	kzap "github.com/go-kratos/kratos/contrib/log/zap/v2"
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
	zapConfig := zap.NewProductionConfig()
	zapConfig.EncoderConfig.FunctionKey = "func"
	zapConfig.Encoding = "json"
	zapConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
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

	l.With(zap.String("a", "b")).Sugar().Infof("aaaaaa")
	l.Info("this is a msg", zap.String("a", "b"))
	l.Sugar().Debugw("this is a log", "a", "b")

	s := &Server{}
	httpSrv := http.NewServer(
		http.Address(":8080"),
		http.Middleware(
			metadata.Server(metadata.WithPropagatedPrefix("")),
			setIp,
			//jwt.Server(func(token *jwt2.Token) (interface{}, error) {
			//	return []byte("kajshdkajdas"), nil
			//}),
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
		l.Sugar().Debugw("this is a log", "a", "b")
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
		log.Fatal(err)
	}
}

package server

import (
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	api "mall/im/api/im"
	"mall/kit/auth"
	"mall/kit/httpx"
	userapi "mall/user/api/user"
)

func Middlewares() []middleware.Middleware {
	return []middleware.Middleware{
		metadata.Server(metadata.WithPropagatedPrefix("")),
		auth.Middleware(
			auth.WithWhitelist(
				userapi.OperationUserRegister,
				userapi.OperationUserLogin,
				api.OperationImInternalSend,
				api.OperationImInternalUpdate,
			),
		),
	}
}

func HTTPOptions(addr string) []khttp.ServerOption {
	return []khttp.ServerOption{
		khttp.Address(addr),
		khttp.Middleware(Middlewares()...),
		khttp.ResponseEncoder(httpx.ResponseEncoder),
		khttp.ErrorEncoder(httpx.ErrorEncoder),
	}
}

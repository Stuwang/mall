package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"helloworld/api/helloworld"
	"helloworld/api/product"
	"helloworld/api/user"
	"helloworld/common/auth"
)

// Middlewares 返回 HTTP 服务默认中间件链
func Middlewares() []middleware.Middleware {
	return []middleware.Middleware{
		metadata.Server(metadata.WithPropagatedPrefix("")),
		setIP,
		auth.Middleware(
			auth.WithWhitelist(
				user.OperationUserRegister,
				user.OperationUserLogin,
				helloworld.OperationGreeterSayHello,
				// 用户侧商品浏览免登录
				product.OperationProductListProducts,
				product.OperationProductListProductsByCategory,
				product.OperationProductGetProduct,
				product.OperationProductListCategories,
			),
		),
	}
}

func setIP(handler middleware.Handler) middleware.Handler {
	return func(ctx context.Context, req interface{}) (interface{}, error) {
		return handler(ctx, req)
	}
}

// HTTPOptions 组装 HTTP Server 常用选项
func HTTPOptions(addr string) []khttp.ServerOption {
	return []khttp.ServerOption{
		khttp.Address(addr),
		khttp.Middleware(Middlewares()...),
		khttp.ResponseEncoder(ResponseEncoder),
		khttp.ErrorEncoder(ErrorEncoder),
	}
}

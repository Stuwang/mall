package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"helloworld/common/config"
	"helloworld/common/database"
	"helloworld/internal/app"
	"mall/kit/logger"
)

//go:embed all:static/admin
var adminDist embed.FS

//go:embed all:static/web_app
var webDist embed.FS

var (
	adminAPIPrefixes = []string{
		"/user/api",
		"/admin",
		"/files",
		"/product/api",
		"/order/api",
		"/asset/api",
	}
	webAPIPrefixes = []string{
		"/user/api",
		"/product/api",
		"/cart/api",
		"/order/api",
		"/asset/api",
		"/mall/api",
		"/misc/api",
		"/files",
	}
)

func main() {
	confPath := flag.String("conf", "configs/config_prod.yaml", "config path")
	adminAddr := flag.String("admin-addr", ":8081", "admin static server listen address")
	webAddr := flag.String("web-addr", ":8082", "web_app static server listen address")
	flag.Parse()

	cfg, err := config.Load(*confPath)
	if err != nil {
		panic(err)
	}

	app.Init(cfg)
	defer database.Close()

	apiURL, err := listenURL(cfg.Server.HTTPAddr)
	if err != nil {
		logger.Fatalf("resolve api url from %q: %v", cfg.Server.HTTPAddr, err)
	}

	adminFS, err := fs.Sub(adminDist, "static/admin")
	if err != nil {
		logger.Fatalf("admin embed fs: %v", err)
	}
	webFS, err := fs.Sub(webDist, "static/web_app")
	if err != nil {
		logger.Fatalf("web_app embed fs: %v", err)
	}

	adminSrv := newStaticServer(*adminAddr, adminFS, apiURL, adminAPIPrefixes)
	webSrv := newStaticServer(*webAddr, webFS, apiURL, webAPIPrefixes)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	kratosApp := app.New(cfg, ctx)
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		logger.Infof("api listening on %s (proxy target %s)", cfg.Server.HTTPAddr, apiURL.String())
		if err := kratosApp.Run(); err != nil {
			return fmt.Errorf("api: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		return serveHTTP(gctx, "admin", adminSrv)
	})
	g.Go(func() error {
		return serveHTTP(gctx, "web_app", webSrv)
	})
	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = adminSrv.Shutdown(shutdownCtx)
		_ = webSrv.Shutdown(shutdownCtx)
		return nil
	})

	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatalf("%v", err)
	}
}

func serveHTTP(ctx context.Context, name string, srv *http.Server) error {
	logger.Infof("%s listening on %s", name, srv.Addr)
	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("%s: %w", name, err)
			return
		}
		errCh <- nil
	}()
	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

// listenURL 把 ":8080" / "0.0.0.0:8080" 转成本机反代目标 http://127.0.0.1:8080。
func listenURL(addr string) (*url.URL, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// 兼容纯端口写法以外的异常
		if strings.HasPrefix(addr, ":") {
			host, port = "127.0.0.1", strings.TrimPrefix(addr, ":")
		} else {
			return nil, err
		}
	}
	if host == "" || host == "0.0.0.0" || host == "::" {
		host = "127.0.0.1"
	}
	return url.Parse(fmt.Sprintf("http://%s", net.JoinHostPort(host, port)))
}

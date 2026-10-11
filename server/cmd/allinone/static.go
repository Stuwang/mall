package main

import (
	"io"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// spaFileServer 提供静态资源；无扩展名的路径找不到文件时回落 index.html（Vue history 模式）。
func spaFileServer(root fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := root.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		if info, statErr := fs.Stat(root, path); statErr == nil && info.IsDir() {
			fileServer.ServeHTTP(w, r)
			return
		}
		// /assets/xxx.js 等带扩展名资源缺失时不要回落 HTML
		if strings.Contains(filepathBase(path), ".") {
			http.NotFound(w, r)
			return
		}
		serveIndexHTML(w, r, root)
	})
}

func filepathBase(path string) string {
	if i := strings.LastIndex(path, "/"); i >= 0 {
		return path[i+1:]
	}
	return path
}

func serveIndexHTML(w http.ResponseWriter, r *http.Request, root fs.FS) {
	f, err := root.Open("index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	stat, err := f.Stat()
	if err != nil {
		http.NotFound(w, r)
		return
	}
	rs, ok := f.(io.ReadSeeker)
	if !ok {
		http.Error(w, "index.html not seekable", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "index.html", stat.ModTime(), rs)
}

func newAPIProxy(apiURL *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(apiURL)
	original := proxy.Director
	proxy.Director = func(req *http.Request) {
		forwardedHost := req.Host
		original(req)
		req.Host = apiURL.Host
		if forwardedHost != "" {
			req.Header.Set("X-Forwarded-Host", forwardedHost)
		}
		req.Header.Set("X-Forwarded-Proto", "http")
	}
	return proxy
}

// mountAPIProxy 将指定前缀反代到 API（与 vite proxy 对齐）。
func mountAPIProxy(mux *http.ServeMux, proxy *httputil.ReverseProxy, prefixes ...string) {
	for _, p := range prefixes {
		prefix := p
		mux.Handle(prefix+"/", proxy)
		mux.Handle(prefix, proxy)
	}
}

func newStaticServer(addr string, root fs.FS, apiURL *url.URL, apiPrefixes []string) *http.Server {
	mux := http.NewServeMux()
	proxy := newAPIProxy(apiURL)
	mountAPIProxy(mux, proxy, apiPrefixes...)
	mux.Handle("/", spaFileServer(root))
	return &http.Server{
		Addr:    addr,
		Handler: mux,
	}
}

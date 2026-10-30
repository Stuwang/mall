package mediacontroller

import (
	"context"
	"io"
	stdhttp "net/http"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"mall/file/mediastore"
	"mall/kit/auth"
	"mall/kit/codeerror"
	"mall/kit/logger"
)

const (
	adminUploadPath = "/admin/api/v1/media/upload"
	userUploadPath  = "/user/api/v1/media/upload"
	filesPrefix     = mediastore.FilesHTTPPrefix
)

// UploadReply 上传成功返回
type UploadReply struct {
	URI string `json:"uri"`
	URL string `json:"url"`
}

// RegisterHTTP 注册管理端上传与静态访问路由
func RegisterHTTP(srv *khttp.Server) {
	RegisterAdminUpload(srv)
	RegisterFiles(srv)
}

// RegisterAdminUpload 管理端图片上传（需 admin）
func RegisterAdminUpload(srv *khttp.Server) {
	r := srv.Route("/")
	r.POST(adminUploadPath, handleAdminUpload)
}

// RegisterUserUpload 登录用户图片上传（头像等）
func RegisterUserUpload(srv *khttp.Server) {
	r := srv.Route("/")
	r.POST(userUploadPath, handleUserUpload)
}

// RegisterFiles 公开静态文件访问
func RegisterFiles(srv *khttp.Server) {
	srv.HandlePrefix(filesPrefix, stdhttp.HandlerFunc(handleFiles))
}

func handleAdminUpload(ctx khttp.Context) error {
	khttp.SetOperation(ctx, adminUploadPath)
	h := ctx.Middleware(func(c context.Context, _ interface{}) (interface{}, error) {
		if _, err := auth.RequireAdmin(c); err != nil {
			logger.Errorf("media upload RequireAdmin err=%v", err)
			return nil, err
		}
		return saveUpload(ctx, false)
	})
	out, err := h(ctx, nil)
	if err != nil {
		return err
	}
	return ctx.Result(200, out)
}

func handleUserUpload(ctx khttp.Context) error {
	khttp.SetOperation(ctx, userUploadPath)
	h := ctx.Middleware(func(c context.Context, _ interface{}) (interface{}, error) {
		_ = auth.MustFromContext(c)
		return saveUpload(ctx, true)
	})
	out, err := h(ctx, nil)
	if err != nil {
		return err
	}
	return ctx.Result(200, out)
}

func saveUpload(ctx khttp.Context, pathOnly bool) (*UploadReply, error) {
	req := ctx.Request()
	if err := req.ParseMultipartForm(mediastore.Get().MaxUploadBytes() + 1024); err != nil {
		logger.Errorf("media upload ParseMultipartForm err=%v", err)
		return nil, codeerror.ErrorParam
	}
	file, header, err := req.FormFile("file")
	if err != nil {
		logger.Errorf("media upload FormFile err=%v", err)
		return nil, codeerror.ErrorParam
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	uri, err := mediastore.Get().Save(header.Filename, contentType, file, header.Size)
	if err != nil {
		logger.Errorf("media upload Save orig=%s contentType=%s err=%v", header.Filename, contentType, err)
		return nil, err
	}
	url := mediastore.Get().PublicURL(uri)
	if pathOnly {
		url = mediastore.Get().PathURL(uri)
	}
	return &UploadReply{
		URI: uri,
		URL: url,
	}, nil
}

func handleFiles(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet && r.Method != stdhttp.MethodHead {
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, filesPrefix)
	f, contentType, err := mediastore.Get().Open(key)
	if err != nil {
		logger.Errorf("media files Open path=%s err=%v", r.URL.Path, err)
		w.WriteHeader(stdhttp.StatusNotFound)
		return
	}
	defer f.Close()
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=2592000")
	if r.Method == stdhttp.MethodHead {
		return
	}
	if _, err := io.Copy(w, f); err != nil {
		logger.Errorf("media files copy path=%s err=%v", r.URL.Path, err)
	}
}

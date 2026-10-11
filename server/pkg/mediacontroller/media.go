package mediacontroller

import (
	"context"
	"io"
	stdhttp "net/http"
	"strings"

	khttp "github.com/go-kratos/kratos/v2/transport/http"

	"helloworld/base/codeerror"
	"helloworld/base/logger"
	"helloworld/common/auth"
	"helloworld/pkg/mediastore"
)

const (
	uploadPath = "/admin/api/v1/media/upload"
	filesPrefix = mediastore.FilesHTTPPrefix
)

// UploadReply 上传成功返回
type UploadReply struct {
	URI string `json:"uri"`
	URL string `json:"url"`
}

// RegisterHTTP 注册上传与静态访问路由
func RegisterHTTP(srv *khttp.Server) {
	r := srv.Route("/")
	r.POST(uploadPath, handleUpload)
	srv.HandlePrefix(filesPrefix, stdhttp.HandlerFunc(handleFiles))
}

func handleUpload(ctx khttp.Context) error {
	khttp.SetOperation(ctx, uploadPath)
	h := ctx.Middleware(func(c context.Context, _ interface{}) (interface{}, error) {
		if _, err := auth.RequireAdmin(c); err != nil {
			logger.Errorf("media upload RequireAdmin err=%v", err)
			return nil, err
		}
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

		uri, err := mediastore.Get().Save(header.Filename, file, header.Size)
		if err != nil {
			logger.Errorf("media upload Save filename=%s err=%v", header.Filename, err)
			return nil, err
		}
		return &UploadReply{
			URI: uri,
			URL: mediastore.Get().PublicURL(uri),
		}, nil
	})
	out, err := h(ctx, nil)
	if err != nil {
		return err
	}
	return ctx.Result(200, out)
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
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if r.Method == stdhttp.MethodHead {
		return
	}
	if _, err := io.Copy(w, f); err != nil {
		logger.Errorf("media files copy path=%s err=%v", r.URL.Path, err)
	}
}

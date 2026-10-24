package mediastore

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"mall/kit/codeerror"
	"mall/kit/logger"
)

// Config 本地文件存储配置，由宿主服务填入。
type Config struct {
	PublicBase     string
	MaxUploadBytes int64
	RootDir        string
}

// Scheme 逻辑文件 URI 的协议
type Scheme string

const (
	SchemeLocalFile Scheme = "local_file"
	SchemeOSS       Scheme = "oss"

	// FilesHTTPPrefix 公开访问路径前缀
	FilesHTTPPrefix = "/files/"

	defaultMaxUploadBytes = 5 << 20 // 5MB
)

var allowedExt = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".webp": "image/webp",
}

// contentType -> 规范扩展名（统一 .jpg，不用用户原名）
var contentTypeExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var (
	mu    sync.RWMutex
	store *Store
)

// Store 图片存储（本期仅 local）
type Store struct {
	publicBase     string
	rootDir        string
	maxUploadBytes int64
	filesURLPrefix string // publicBase + /files/
}

// Init 初始化全局存储
func Init(cfg Config) error {
	root := strings.TrimSpace(cfg.RootDir)
	if root == "" {
		root = "data/images"
	}
	publicBase := strings.TrimRight(strings.TrimSpace(cfg.PublicBase), "/")
	if publicBase == "" {
		publicBase = "http://localhost:8080"
	}
	maxBytes := cfg.MaxUploadBytes
	if maxBytes <= 0 {
		maxBytes = defaultMaxUploadBytes
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return fmt.Errorf("mediastore mkdir %s: %w", root, err)
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("mediastore abs %s: %w", root, err)
	}
	s := &Store{
		publicBase:     publicBase,
		rootDir:        abs,
		maxUploadBytes: maxBytes,
		filesURLPrefix: publicBase + FilesHTTPPrefix,
	}
	mu.Lock()
	store = s
	mu.Unlock()
	logger.Infof("mediastore init root=%s public_base=%s max_upload=%d", abs, publicBase, maxBytes)
	return nil
}

// Get 返回全局 Store；未 Init 时 panic
func Get() *Store {
	mu.RLock()
	defer mu.RUnlock()
	if store == nil {
		panic("mediastore not initialized")
	}
	return store
}

// MaxUploadBytes 单文件上传上限
func (s *Store) MaxUploadBytes() int64 {
	return s.maxUploadBytes
}

// Save 保存上传文件，返回逻辑 URI（local_file:///key）。
// 磁盘文件名一律随机生成，绝不使用用户原始文件名，避免重名覆盖。
// origName / contentType 仅用于推断扩展名。
func (s *Store) Save(origName, contentType string, r io.Reader, size int64) (string, error) {
	if size > 0 && size > s.maxUploadBytes {
		logger.Errorf("mediastore Save file too large size=%d max=%d", size, s.maxUploadBytes)
		return "", codeerror.ErrorParam
	}
	ext, ok := resolveExt(origName, contentType)
	if !ok {
		logger.Errorf("mediastore Save unsupported type orig=%s contentType=%s", origName, contentType)
		return "", codeerror.ErrorParam
	}

	day := time.Now().Format("2006/01/02")
	// 32 位 hex，不含原文件名
	randomName := strings.ReplaceAll(uuid.NewString(), "-", "") + ext
	key := path.Join(day, randomName)
	full := filepath.Join(s.rootDir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", codeerror.OrSystem(err, "mediastore Save mkdir key=%s", key)
	}

	f, err := os.OpenFile(full, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return "", codeerror.OrSystem(err, "mediastore Save open key=%s", key)
	}
	defer f.Close()

	limited := io.LimitReader(r, s.maxUploadBytes+1)
	written, err := io.Copy(f, limited)
	if err != nil {
		_ = os.Remove(full)
		return "", codeerror.OrSystem(err, "mediastore Save copy key=%s", key)
	}
	if written > s.maxUploadBytes {
		_ = os.Remove(full)
		logger.Errorf("mediastore Save exceeds max written=%d max=%d", written, s.maxUploadBytes)
		return "", codeerror.ErrorParam
	}
	uri := string(SchemeLocalFile) + ":///" + key
	return uri, nil
}

func resolveExt(origName, contentType string) (string, bool) {
	ct := strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	if ext, ok := contentTypeExt[ct]; ok {
		return ext, true
	}
	ext := strings.ToLower(filepath.Ext(origName))
	if ext == ".jpeg" {
		ext = ".jpg"
	}
	if _, ok := allowedExt[ext]; ok {
		return ext, true
	}
	return "", false
}

// Resolve 将逻辑 URI / 外链转为可访问 URL
func (s *Store) Resolve(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		return raw
	}
	if key, ok := parseLocalFileURI(raw); ok {
		return s.filesURLPrefix + key
	}
	if strings.HasPrefix(raw, string(SchemeOSS)+"://") {
		logger.Errorf("mediastore Resolve oss not implemented uri=%s", raw)
		return raw
	}
	logger.Errorf("mediastore Resolve unknown uri=%s", raw)
	return raw
}

// Normalize 写库前将公开 URL 还原为逻辑 URI
func (s *Store) Normalize(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, string(SchemeLocalFile)+"://") || strings.HasPrefix(raw, string(SchemeOSS)+"://") {
		return raw
	}
	if key, ok := s.stripFilesURL(raw); ok {
		return string(SchemeLocalFile) + ":///" + key
	}
	return raw
}

// Open 按 key 打开本地文件（防路径穿越）
func (s *Store) Open(key string) (*os.File, string, error) {
	cleaned, ok := cleanKey(key)
	if !ok {
		logger.Errorf("mediastore Open invalid key=%s", key)
		return nil, "", codeerror.ErrorParam
	}
	full := filepath.Join(s.rootDir, filepath.FromSlash(cleaned))
	rel, err := filepath.Rel(s.rootDir, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		logger.Errorf("mediastore Open path escape key=%s", cleaned)
		return nil, "", codeerror.ErrorParam
	}
	f, err := os.Open(full)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Errorf("mediastore Open not found key=%s", cleaned)
			return nil, "", codeerror.ErrorParam
		}
		return nil, "", codeerror.OrSystem(err, "mediastore Open key=%s", cleaned)
	}
	ct := allowedExt[strings.ToLower(filepath.Ext(cleaned))]
	if ct == "" {
		ct = "application/octet-stream"
	}
	return f, ct, nil
}

// PublicURL 由 URI 得到 url（上传响应用）
func (s *Store) PublicURL(uri string) string {
	return s.Resolve(uri)
}

func (s *Store) stripFilesURL(raw string) (string, bool) {
	if strings.HasPrefix(raw, s.filesURLPrefix) {
		return cleanKey(strings.TrimPrefix(raw, s.filesURLPrefix))
	}
	if strings.HasPrefix(raw, FilesHTTPPrefix) {
		return cleanKey(strings.TrimPrefix(raw, FilesHTTPPrefix))
	}
	return "", false
}

func parseLocalFileURI(raw string) (string, bool) {
	const prefix = string(SchemeLocalFile) + ":///"
	if !strings.HasPrefix(raw, prefix) {
		// 兼容 local_file://key（两斜杠）
		alt := string(SchemeLocalFile) + "://"
		if strings.HasPrefix(raw, alt) {
			return cleanKey(strings.TrimPrefix(raw, alt))
		}
		return "", false
	}
	return cleanKey(strings.TrimPrefix(raw, prefix))
}

func cleanKey(key string) (string, bool) {
	key = strings.TrimSpace(key)
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return "", false
	}
	cleaned := path.Clean("/" + key)
	cleaned = strings.TrimPrefix(cleaned, "/")
	if cleaned == "" || cleaned == "." || strings.HasPrefix(cleaned, "..") || strings.Contains(cleaned, "..") {
		return "", false
	}
	return cleaned, true
}

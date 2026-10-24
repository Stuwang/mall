package mediastore

import (
	"strings"
	"testing"

	"helloworld/common/config"
	"mall/kit/logger"
)

func TestResolveNormalizeRoundTrip(t *testing.T) {
	logger.InitLogger("")
	dir := t.TempDir()
	if err := Init(config.StorageConfig{
		PublicBase:     "http://localhost:8080",
		MaxUploadBytes: 1 << 20,
		Local:          config.LocalStorageCfg{RootDir: dir},
	}); err != nil {
		t.Fatal(err)
	}
	s := Get()
	uri := "local_file:///2026/10/11/demo.jpg"
	url := s.Resolve(uri)
	if url != "http://localhost:8080/files/2026/10/11/demo.jpg" {
		t.Fatalf("Resolve got %q", url)
	}
	back := s.Normalize(url)
	if back != uri {
		t.Fatalf("Normalize got %q want %q", back, uri)
	}
	if s.Normalize(uri) != uri {
		t.Fatalf("Normalize uri should keep")
	}
	ext := "https://cdn.example/a.png"
	if s.Resolve(ext) != ext || s.Normalize(ext) != ext {
		t.Fatalf("external url should pass through")
	}
	rel := s.Normalize("/files/2026/10/11/demo.jpg")
	if !strings.HasPrefix(rel, "local_file:///") {
		t.Fatalf("relative files path Normalize got %q", rel)
	}
}

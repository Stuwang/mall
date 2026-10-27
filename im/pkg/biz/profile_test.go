package biz_test

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"mall/file/mediastore"
	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	"mall/kit/codeerror"
	userdao "mall/user/dao"
)

func setupMedia(t *testing.T) {
	t.Helper()
	if err := mediastore.Init(mediastore.Config{
		PublicBase:     "http://localhost:8090",
		MaxUploadBytes: 1 << 20,
		RootDir:        filepath.Join(t.TempDir(), "images"),
	}); err != nil {
		t.Fatalf("mediastore init: %v", err)
	}
}

func putImage(t *testing.T) string {
	t.Helper()
	uri, err := mediastore.Get().Save("a.jpg", "image/jpeg", bytes.NewReader([]byte("fake-jpeg")), 9)
	if err != nil {
		t.Fatalf("save image: %v", err)
	}
	return uri
}

func TestAlbumAddListDelete(t *testing.T) {
	a, b, _ := setup(t)
	setupMedia(t)

	uri1 := putImage(t)
	uri2 := putImage(t)

	reply, err := biz.AddAlbumPhoto(a, uri1, " 第一张 ")
	if err != nil {
		t.Fatalf("add1: %v", err)
	}
	if reply.Photo.GetCaption() != "第一张" {
		t.Fatalf("caption=%q", reply.Photo.GetCaption())
	}
	if !strings.HasPrefix(reply.Photo.GetImageUrl(), "/files/") {
		t.Fatalf("image_url=%q", reply.Photo.GetImageUrl())
	}
	id1 := reply.Photo.GetId()

	if _, err := biz.AddAlbumPhoto(a, uri2, ""); err != nil {
		t.Fatalf("add2: %v", err)
	}

	list, err := biz.ListAlbum(a, 0, 0, 0)
	if err != nil {
		t.Fatalf("list self: %v", err)
	}
	if len(list.List) != 2 || list.HasMore {
		t.Fatalf("list %#v", list)
	}
	if list.List[0].GetId() <= list.List[1].GetId() {
		t.Fatalf("want id desc got %d %d", list.List[0].GetId(), list.List[1].GetId())
	}

	prof, err := biz.GetProfile(b, a)
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	if prof.GetUserId() != a || prof.GetPhotoCount() != 2 || prof.GetNickname() != "A" {
		t.Fatalf("profile %#v", prof)
	}

	page1, err := biz.ListAlbum(b, a, 0, 1)
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if len(page1.List) != 1 || !page1.HasMore {
		t.Fatalf("page1 %#v", page1)
	}
	page2, err := biz.ListAlbum(b, a, page1.List[0].GetId(), 1)
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(page2.List) != 1 || page2.HasMore || page2.List[0].GetId() != id1 {
		t.Fatalf("page2 %#v", page2)
	}

	if _, err := biz.DeleteAlbumPhoto(b, id1); err == nil {
		t.Fatal("delete others photo should fail")
	} else if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorAlbumPhoto) {
		t.Fatalf("delete others want 1120 got %v", err)
	}

	if _, err := biz.DeleteAlbumPhoto(a, id1); err != nil {
		t.Fatalf("delete own: %v", err)
	}
	if _, err := biz.DeleteAlbumPhoto(a, id1); err == nil {
		t.Fatal("delete again should fail")
	} else if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorAlbumPhoto) {
		t.Fatalf("delete again want 1120 got %v", err)
	}

	prof2, err := biz.GetProfile(a, 0)
	if err != nil {
		t.Fatalf("profile self: %v", err)
	}
	if prof2.GetPhotoCount() != 1 {
		t.Fatalf("photo_count=%d", prof2.GetPhotoCount())
	}
}

func TestAlbumCaptionTooLong(t *testing.T) {
	a, _, _ := setup(t)
	setupMedia(t)
	uri := putImage(t)
	long := strings.Repeat("字", 81)
	_, err := biz.AddAlbumPhoto(a, uri, long)
	if err == nil {
		t.Fatal("want caption error")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorParam) {
		t.Fatalf("want param got %v", err)
	}
}

func TestAlbumFull(t *testing.T) {
	a, _, _ := setup(t)
	setupMedia(t)
	for i := 0; i < 100; i++ {
		uri := putImage(t)
		if _, err := biz.AddAlbumPhoto(a, uri, ""); err != nil {
			t.Fatalf("add %d: %v", i, err)
		}
	}
	uri := putImage(t)
	_, err := biz.AddAlbumPhoto(a, uri, "")
	if err == nil {
		t.Fatal("want album full")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorAlbumFull) {
		t.Fatalf("want 1121 got %v", err)
	}
}

func TestProfileBannedTarget(t *testing.T) {
	a, b, _ := setup(t)
	if _, err := userdao.NewUserDao(nil).SetBanUntil(a, time.Now().Unix()+3600); err != nil {
		t.Fatalf("ban: %v", err)
	}

	_, err := biz.GetProfile(b, a)
	if err == nil {
		t.Fatal("want banned")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorUserBanned) {
		t.Fatalf("want 1014 got %v", err)
	}

	if _, err := biz.GetProfile(a, 0); err != nil {
		t.Fatalf("self profile while banned: %v", err)
	}
	setupMedia(t)
	uri := putImage(t)
	_, err = biz.AddAlbumPhoto(a, uri, "")
	if err == nil {
		t.Fatal("banned user add should fail")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorUserBanned) {
		t.Fatalf("add want 1014 got %v", err)
	}
}

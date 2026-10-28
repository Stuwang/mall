package biz_test

import (
	"strings"
	"testing"
	"time"

	"mall/im/pkg/biz"
	"mall/im/pkg/bizerror"
	"mall/kit/codeerror"
	userdao "mall/user/dao"
)

func TestPostCreateRecommendLikeComment(t *testing.T) {
	a, b, _ := setup(t)

	p1, err := biz.CreatePost(a, "post0001", "第一条", nil)
	if err != nil {
		t.Fatalf("create1: %v", err)
	}
	if p1.Post.GetPostId() == 0 || p1.Post.GetLikeCount() != 0 {
		t.Fatalf("post1 %#v", p1.Post)
	}

	// 幂等同正文
	again, err := biz.CreatePost(a, "post0001", "第一条", nil)
	if err != nil {
		t.Fatalf("idempotent: %v", err)
	}
	if again.Post.GetPostId() != p1.Post.GetPostId() {
		t.Fatalf("idempotent id changed")
	}

	// 幂等不同正文
	_, err = biz.CreatePost(a, "post0001", "改了", nil)
	if err == nil {
		t.Fatal("want idempotency conflict")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorPostIdempotency) {
		t.Fatalf("want 1133 got %v", err)
	}

	p2, err := biz.CreatePost(b, "post0002", "第二条", nil)
	if err != nil {
		t.Fatalf("create2: %v", err)
	}

	liked, err := biz.LikePost(a, p2.Post.GetPostId())
	if err != nil {
		t.Fatalf("like: %v", err)
	}
	if !liked.Post.GetLiked() || liked.Post.GetLikeCount() != 1 || liked.Post.GetRecommendScore() != 1 {
		t.Fatalf("after like %#v", liked.Post)
	}
	liked2, err := biz.LikePost(a, p2.Post.GetPostId())
	if err != nil {
		t.Fatalf("like again: %v", err)
	}
	if liked2.Post.GetLikeCount() != 1 {
		t.Fatalf("double like count=%d", liked2.Post.GetLikeCount())
	}

	cmt, err := biz.CreateComment(a, p2.Post.GetPostId(), "cmt00001", "好")
	if err != nil {
		t.Fatalf("comment: %v", err)
	}
	if cmt.Post.GetCommentCount() != 1 || cmt.Post.GetRecommendScore() != 2 {
		t.Fatalf("after comment %#v", cmt.Post)
	}

	rec, err := biz.Recommend(a, false, 0, 0, 10)
	if err != nil {
		t.Fatalf("recommend: %v", err)
	}
	if len(rec.List) < 2 {
		t.Fatalf("recommend len=%d", len(rec.List))
	}
	if rec.List[0].GetPostId() != p2.Post.GetPostId() {
		t.Fatalf("want p2 first by score got %#v", rec.List[0])
	}

	unliked, err := biz.UnlikePost(a, p2.Post.GetPostId())
	if err != nil {
		t.Fatalf("unlike: %v", err)
	}
	if unliked.Post.GetLiked() || unliked.Post.GetLikeCount() != 0 {
		t.Fatalf("after unlike %#v", unliked.Post)
	}
	// 复活
	reliked, err := biz.LikePost(a, p2.Post.GetPostId())
	if err != nil {
		t.Fatalf("relike: %v", err)
	}
	if !reliked.Post.GetLiked() || reliked.Post.GetLikeCount() != 1 {
		t.Fatalf("after relike %#v", reliked.Post)
	}

	_, err = biz.DeleteComment(b, cmt.Comment.GetCommentId())
	if err == nil {
		t.Fatal("delete others comment should fail")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorCommentNotFound) {
		t.Fatalf("want 1131 got %v", err)
	}

	delCmt, err := biz.DeleteComment(a, cmt.Comment.GetCommentId())
	if err != nil {
		t.Fatalf("delete own comment: %v", err)
	}
	if delCmt.Post.GetCommentCount() != 0 {
		t.Fatalf("comment_count=%d", delCmt.Post.GetCommentCount())
	}

	_, err = biz.DeletePost(b, p1.Post.GetPostId())
	if err == nil {
		t.Fatal("delete others post should fail")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorPostNotFound) {
		t.Fatalf("want 1130 got %v", err)
	}
	if _, err := biz.DeletePost(a, p1.Post.GetPostId()); err != nil {
		t.Fatalf("delete own post: %v", err)
	}
	_, err = biz.GetPost(a, p1.Post.GetPostId())
	if err == nil {
		t.Fatal("deleted post detail should fail")
	}
}

func TestPostRecommendCursorAndBan(t *testing.T) {
	a, b, c := setup(t)
	for i, text := range []string{"a1", "a2", "a3"} {
		cid := "postcur0" + string(rune('1'+i))
		if _, err := biz.CreatePost(a, cid, text, nil); err != nil {
			t.Fatalf("create %s: %v", cid, err)
		}
	}
	_ = b
	page1, err := biz.Recommend(c, false, 0, 0, 2)
	if err != nil {
		t.Fatalf("page1: %v", err)
	}
	if len(page1.List) != 2 || !page1.HasMore {
		t.Fatalf("page1 %#v", page1)
	}
	last := page1.List[len(page1.List)-1]
	page2, err := biz.Recommend(c, true, last.GetRecommendScore(), last.GetPostId(), 2)
	if err != nil {
		t.Fatalf("page2: %v", err)
	}
	if len(page2.List) < 1 {
		t.Fatalf("page2 empty")
	}
	if page2.List[0].GetPostId() == page1.List[0].GetPostId() {
		t.Fatalf("cursor overlap")
	}

	if _, err := userdao.NewUserDao(nil).SetBanUntil(a, time.Now().Unix()+3600); err != nil {
		t.Fatalf("ban: %v", err)
	}
	rec, err := biz.Recommend(c, false, 0, 0, 20)
	if err != nil {
		t.Fatalf("recommend after ban: %v", err)
	}
	for _, p := range rec.List {
		if p.GetUserId() == a {
			t.Fatalf("banned author still listed")
		}
	}
	if len(page1.List) > 0 {
		_, err = biz.GetPost(c, page1.List[0].GetPostId())
		if err == nil {
			t.Fatal("detail of banned author should 1130")
		}
		if ce, ok := codeerror.As(err); !ok || !ce.Is(bizerror.ErrorPostNotFound) {
			t.Fatalf("want 1130 got %v", err)
		}
	}

	_, err = biz.CreatePost(a, "postban01", "不能发", nil)
	if err == nil {
		t.Fatal("banned create should fail")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorUserBanned) {
		t.Fatalf("want 1014 got %v", err)
	}
}

func TestPostTextLimits(t *testing.T) {
	a, _, _ := setup(t)
	_, err := biz.CreatePost(a, "postlen01", "", nil)
	if err == nil {
		t.Fatal("empty text and images")
	}
	if ce, ok := codeerror.As(err); !ok || !ce.Is(codeerror.ErrorParam) {
		t.Fatalf("want param got %v", err)
	}
	_, err = biz.CreatePost(a, "postlen02", strings.Repeat("字", 1001), nil)
	if err == nil {
		t.Fatal("too long")
	}
}

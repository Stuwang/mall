package controller

import (
	"context"

	apipost "mall/im/api/post"
	"mall/im/pkg/biz"
	"mall/kit/auth"
	"mall/kit/logger"
)

// PostController 帖子、评论、点赞。
type PostController struct{}

func NewPostController() *PostController {
	return &PostController{}
}

func (c *PostController) Recommend(ctx context.Context, req *apipost.RecommendRequest) (*apipost.RecommendReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.Recommend(a.UserID, req.GetHasCursor(), req.GetCursorScore(), req.GetCursorId(), req.GetLimit())
	if err != nil {
		logger.Errorf("Recommend user=%d err=%v", a.UserID, err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) ListUserPost(ctx context.Context, req *apipost.ListUserPostRequest) (*apipost.ListUserPostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.ListUserPost(a.UserID, req.GetUserId(), req.GetBeforeId(), req.GetLimit())
	if err != nil {
		logger.Errorf("ListUserPost viewer=%d target=%d err=%v", a.UserID, req.GetUserId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) CreatePost(ctx context.Context, req *apipost.CreatePostRequest) (*apipost.CreatePostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.CreatePost(a.UserID, req.GetClientPostId(), req.GetText(), req.GetImageUris())
	if err != nil {
		logger.Errorf("CreatePost user=%d err=%v", a.UserID, err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) DeletePost(ctx context.Context, req *apipost.DeletePostRequest) (*apipost.DeletePostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.DeletePost(a.UserID, req.GetPostId())
	if err != nil {
		logger.Errorf("DeletePost user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) GetPost(ctx context.Context, req *apipost.GetPostRequest) (*apipost.GetPostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.GetPost(a.UserID, req.GetPostId())
	if err != nil {
		logger.Errorf("GetPost user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) LikePost(ctx context.Context, req *apipost.LikePostRequest) (*apipost.LikePostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.LikePost(a.UserID, req.GetPostId())
	if err != nil {
		logger.Errorf("LikePost user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) UnlikePost(ctx context.Context, req *apipost.UnlikePostRequest) (*apipost.UnlikePostReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.UnlikePost(a.UserID, req.GetPostId())
	if err != nil {
		logger.Errorf("UnlikePost user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) ListComment(ctx context.Context, req *apipost.ListCommentRequest) (*apipost.ListCommentReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.ListComment(a.UserID, req.GetPostId(), req.GetAfterId(), req.GetLimit())
	if err != nil {
		logger.Errorf("ListComment user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) CreateComment(ctx context.Context, req *apipost.CreateCommentRequest) (*apipost.CreateCommentReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.CreateComment(a.UserID, req.GetPostId(), req.GetClientCommentId(), req.GetText())
	if err != nil {
		logger.Errorf("CreateComment user=%d postID=%d err=%v", a.UserID, req.GetPostId(), err)
		return nil, err
	}
	return reply, nil
}

func (c *PostController) DeleteComment(ctx context.Context, req *apipost.DeleteCommentRequest) (*apipost.DeleteCommentReply, error) {
	a := auth.MustFromContext(ctx)
	reply, err := biz.DeleteComment(a.UserID, req.GetCommentId())
	if err != nil {
		logger.Errorf("DeleteComment user=%d commentID=%d err=%v", a.UserID, req.GetCommentId(), err)
		return nil, err
	}
	return reply, nil
}

var _ apipost.PostHTTPServer = (*PostController)(nil)

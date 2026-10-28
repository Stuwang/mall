package biz

import (
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	apipost "mall/im/api/post"
	"mall/im/common/database"
	"mall/im/pkg/bizerror"
	imdao "mall/im/pkg/dao"
	"mall/im/pkg/model"
	"mall/kit/codeerror"
	"mall/kit/logger"
	kitcommon "mall/kit/model/common"
	userdao "mall/user/dao"

	"gorm.io/gorm"
)

const (
	postTextMax        = 1000
	commentTextMax     = 300
	postListDefault    = 20
	postListMax        = 50
	commentListDefault = 30
	commentListMax     = 50
	postImageMax       = 9
)

// Recommend 推荐列表。
func Recommend(viewerID int64, hasCursor bool, cursorScore, cursorID int64, limit int32) (*apipost.RecommendReply, error) {
	lim := clampLimit(limit, postListDefault, postListMax)
	rows, err := imdao.NewPostDao(nil).ListRecommend(hasCursor, cursorScore, cursorID, lim+1)
	if err != nil {
		logger.Errorf("Recommend ListRecommend viewer=%d err=%v", viewerID, err)
		return nil, err
	}
	hasMore := len(rows) > lim
	if hasMore {
		rows = rows[:lim]
	}

	authors := uniqueUserIDs(rows)
	banned, err := bannedUserSet(authors)
	if err != nil {
		logger.Errorf("Recommend bannedUserSet viewer=%d err=%v", viewerID, err)
		return nil, err
	}
	visible := make([]*model.Post, 0, len(rows))
	for _, p := range rows {
		if banned[p.UserID] {
			continue
		}
		visible = append(visible, p)
	}
	liked, err := imdao.NewPostLikeDao(nil).ListLikedPostIDs(viewerID, postIDsOf(visible))
	if err != nil {
		logger.Errorf("Recommend ListLikedPostIDs viewer=%d err=%v", viewerID, err)
		return nil, err
	}
	profiles, err := userProfiles(postAuthors(visible))
	if err != nil {
		logger.Errorf("Recommend userProfiles viewer=%d err=%v", viewerID, err)
		return nil, err
	}
	list := make([]*apipost.PostItem, 0, len(visible))
	for _, p := range visible {
		list = append(list, toPostItem(p, profiles[p.UserID], liked[p.ID]))
	}
	return &apipost.RecommendReply{List: list, HasMore: hasMore}, nil
}

// CreatePost 发帖。正文与图片至少有一个。
func CreatePost(userID int64, clientPostID, text string, imageURIs []string) (*apipost.CreatePostReply, error) {
	if err := requireActiveUser(userID); err != nil {
		return nil, err
	}
	if err := validateIdempotencyKey(clientPostID); err != nil {
		logger.Errorf("CreatePost client_post_id invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	text = strings.TrimSpace(text)
	n := utf8.RuneCountInString(text)
	if n > postTextMax {
		logger.Errorf("CreatePost text length user=%d n=%d", userID, n)
		return nil, codeerror.ErrorParam
	}
	uris, err := normalizePostImages(imageURIs)
	if err != nil {
		return nil, err
	}
	if n < 1 && len(uris) == 0 {
		logger.Errorf("CreatePost empty text and images user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	imagesJSON, err := encodePostImages(uris)
	if err != nil {
		logger.Errorf("CreatePost encode images user=%d err=%v", userID, err)
		return nil, codeerror.OrSystem(err, "CreatePost encode images")
	}
	dao := imdao.NewPostDao(nil)
	exist, err := dao.GetByClientID(userID, clientPostID)
	if err != nil {
		logger.Errorf("CreatePost GetByClientID user=%d err=%v", userID, err)
		return nil, err
	}
	if exist != nil {
		if exist.Text != text || exist.Images != imagesJSON {
			logger.Errorf("CreatePost idempotency conflict user=%d postID=%d", userID, exist.ID)
			return nil, bizerror.ErrorPostIdempotency
		}
		item, err := buildPostItem(userID, exist)
		if err != nil {
			return nil, err
		}
		return &apipost.CreatePostReply{Post: item}, nil
	}
	now := time.Now().Unix()
	p := &model.Post{
		UserID:         userID,
		ClientPostID:   clientPostID,
		Text:           text,
		Images:         imagesJSON,
		LikeCount:      0,
		CommentCount:   0,
		RecommendScore: 0,
		Deleted:        kitcommon.DeletedNo,
		Ct:             now,
		Ut:             now,
	}
	if err := dao.Create(p); err != nil {
		logger.Errorf("CreatePost Create user=%d err=%v", userID, err)
		return nil, err
	}
	item, err := buildPostItem(userID, p)
	if err != nil {
		return nil, err
	}
	return &apipost.CreatePostReply{Post: item}, nil
}

// DeletePost 软删除自己的帖子。
func DeletePost(userID, postID int64) (*apipost.DeletePostReply, error) {
	if postID <= 0 {
		logger.Errorf("DeletePost post_id invalid user=%d postID=%d", userID, postID)
		return nil, codeerror.ErrorParam
	}
	if err := requireActiveUser(userID); err != nil {
		return nil, err
	}
	ok, err := imdao.NewPostDao(nil).SoftDelete(userID, postID)
	if err != nil {
		logger.Errorf("DeletePost SoftDelete user=%d postID=%d err=%v", userID, postID, err)
		return nil, err
	}
	if !ok {
		logger.Errorf("DeletePost not found user=%d postID=%d", userID, postID)
		return nil, bizerror.ErrorPostNotFound
	}
	return &apipost.DeletePostReply{}, nil
}

// GetPost 详情。
func GetPost(viewerID, postID int64) (*apipost.GetPostReply, error) {
	p, err := loadVisiblePost(postID)
	if err != nil {
		return nil, err
	}
	item, err := buildPostItem(viewerID, p)
	if err != nil {
		return nil, err
	}
	return &apipost.GetPostReply{Post: item}, nil
}

// LikePost 点赞。已赞则幂等成功。
func LikePost(userID, postID int64) (*apipost.LikePostReply, error) {
	item, err := mutateLike(userID, postID, true)
	if err != nil {
		return nil, err
	}
	return &apipost.LikePostReply{Post: item}, nil
}

// UnlikePost 取消点赞。未赞则幂等成功。
func UnlikePost(userID, postID int64) (*apipost.UnlikePostReply, error) {
	item, err := mutateLike(userID, postID, false)
	if err != nil {
		return nil, err
	}
	return &apipost.UnlikePostReply{Post: item}, nil
}

func mutateLike(userID, postID int64, wantLiked bool) (*apipost.PostItem, error) {
	if postID <= 0 {
		logger.Errorf("mutateLike post_id invalid user=%d postID=%d", userID, postID)
		return nil, codeerror.ErrorParam
	}
	if err := requireActiveUser(userID); err != nil {
		return nil, err
	}
	var out *model.Post
	err := database.GetDb().Transaction(func(tx *gorm.DB) error {
		postDao := imdao.NewPostDao(tx)
		likeDao := imdao.NewPostLikeDao(tx)
		p, err := postDao.GetByIDForUpdate(postID)
		if err != nil {
			return err
		}
		if err := ensurePostVisible(p); err != nil {
			return err
		}
		row, err := likeDao.Get(postID, userID)
		if err != nil {
			return err
		}
		already := row != nil && row.Deleted == kitcommon.DeletedNo
		if already == wantLiked {
			out = p
			return nil
		}
		now := time.Now().Unix()
		if wantLiked {
			if row == nil {
				if err := likeDao.Create(&model.PostLike{
					PostID:  postID,
					UserID:  userID,
					Deleted: kitcommon.DeletedNo,
					Ct:      now,
					Ut:      now,
				}); err != nil {
					return err
				}
			} else {
				if err := likeDao.Revive(postID, userID); err != nil {
					return err
				}
			}
			if err := postDao.IncLike(postID); err != nil {
				return err
			}
		} else {
			if err := likeDao.SoftDelete(postID, userID); err != nil {
				return err
			}
			if err := postDao.DecLike(postID); err != nil {
				return err
			}
		}
		p2, err := postDao.GetByID(postID)
		if err != nil {
			return err
		}
		out = p2
		return nil
	})
	if err != nil {
		logger.Errorf("mutateLike user=%d postID=%d wantLiked=%v err=%v", userID, postID, wantLiked, err)
		return nil, err
	}
	return buildPostItem(userID, out)
}

// ListComment 评论列表。
func ListComment(viewerID, postID, afterID int64, limit int32) (*apipost.ListCommentReply, error) {
	if _, err := loadVisiblePost(postID); err != nil {
		return nil, err
	}
	lim := clampLimit(limit, commentListDefault, commentListMax)
	rows, err := imdao.NewPostCommentDao(nil).ListActive(postID, afterID, lim+1)
	if err != nil {
		logger.Errorf("ListComment ListActive postID=%d err=%v", postID, err)
		return nil, err
	}
	hasMore := false
	if len(rows) > lim {
		hasMore = true
		rows = rows[:lim]
	}
	profiles, err := userProfiles(commentAuthors(rows))
	if err != nil {
		logger.Errorf("ListComment userProfiles postID=%d err=%v", postID, err)
		return nil, err
	}
	list := make([]*apipost.CommentItem, 0, len(rows))
	for _, c := range rows {
		list = append(list, toCommentItem(c, profiles[c.UserID]))
	}
	_ = viewerID
	return &apipost.ListCommentReply{List: list, HasMore: hasMore}, nil
}

// CreateComment 发表评论。
func CreateComment(userID, postID int64, clientCommentID, text string) (*apipost.CreateCommentReply, error) {
	if postID <= 0 {
		logger.Errorf("CreateComment post_id invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	if err := requireActiveUser(userID); err != nil {
		return nil, err
	}
	if err := validateIdempotencyKey(clientCommentID); err != nil {
		logger.Errorf("CreateComment client_comment_id invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	text = strings.TrimSpace(text)
	if n := utf8.RuneCountInString(text); n < 1 || n > commentTextMax {
		logger.Errorf("CreateComment text length user=%d n=%d", userID, n)
		return nil, codeerror.ErrorParam
	}

	cmtDao := imdao.NewPostCommentDao(nil)
	exist, err := cmtDao.GetByClientID(postID, userID, clientCommentID)
	if err != nil {
		return nil, err
	}
	if exist != nil {
		if exist.Text != text || exist.Deleted != kitcommon.DeletedNo {
			if exist.Text != text {
				logger.Errorf("CreateComment idempotency conflict user=%d commentID=%d", userID, exist.ID)
				return nil, bizerror.ErrorPostIdempotency
			}
			// 已删除同行同正文：按新评论处理会撞唯一键，当作冲突
			if exist.Deleted != kitcommon.DeletedNo {
				logger.Errorf("CreateComment idempotency on deleted user=%d commentID=%d", userID, exist.ID)
				return nil, bizerror.ErrorPostIdempotency
			}
		}
		p, err := loadVisiblePost(postID)
		if err != nil {
			return nil, err
		}
		postItem, err := buildPostItem(userID, p)
		if err != nil {
			return nil, err
		}
		profiles, err := userProfiles([]int64{userID})
		if err != nil {
			return nil, err
		}
		return &apipost.CreateCommentReply{
			Comment: toCommentItem(exist, profiles[userID]),
			Post:    postItem,
		}, nil
	}

	var created *model.PostComment
	var outPost *model.Post
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		postDao := imdao.NewPostDao(tx)
		commentDao := imdao.NewPostCommentDao(tx)
		p, err := postDao.GetByIDForUpdate(postID)
		if err != nil {
			return err
		}
		if err := ensurePostVisible(p); err != nil {
			return err
		}
		now := time.Now().Unix()
		c := &model.PostComment{
			PostID:          postID,
			UserID:          userID,
			ClientCommentID: clientCommentID,
			Text:            text,
			Deleted:         kitcommon.DeletedNo,
			Ct:              now,
			Ut:              now,
		}
		if err := commentDao.Create(c); err != nil {
			return err
		}
		if err := postDao.IncComment(postID); err != nil {
			return err
		}
		p2, err := postDao.GetByID(postID)
		if err != nil {
			return err
		}
		created = c
		outPost = p2
		return nil
	})
	if err != nil {
		logger.Errorf("CreateComment user=%d postID=%d err=%v", userID, postID, err)
		return nil, err
	}
	postItem, err := buildPostItem(userID, outPost)
	if err != nil {
		return nil, err
	}
	profiles, err := userProfiles([]int64{userID})
	if err != nil {
		return nil, err
	}
	return &apipost.CreateCommentReply{
		Comment: toCommentItem(created, profiles[userID]),
		Post:    postItem,
	}, nil
}

// DeleteComment 删除自己的评论。
func DeleteComment(userID, commentID int64) (*apipost.DeleteCommentReply, error) {
	if commentID <= 0 {
		logger.Errorf("DeleteComment comment_id invalid user=%d", userID)
		return nil, codeerror.ErrorParam
	}
	if err := requireActiveUser(userID); err != nil {
		return nil, err
	}
	cmt, err := imdao.NewPostCommentDao(nil).GetByID(commentID)
	if err != nil {
		return nil, err
	}
	if cmt == nil || cmt.Deleted != kitcommon.DeletedNo {
		logger.Errorf("DeleteComment not found user=%d commentID=%d", userID, commentID)
		return nil, bizerror.ErrorCommentNotFound
	}
	if _, err := loadVisiblePost(cmt.PostID); err != nil {
		return nil, err
	}
	if cmt.UserID != userID {
		logger.Errorf("DeleteComment not owner user=%d commentID=%d", userID, commentID)
		return nil, bizerror.ErrorCommentNotFound
	}

	var outPost *model.Post
	err = database.GetDb().Transaction(func(tx *gorm.DB) error {
		postDao := imdao.NewPostDao(tx)
		commentDao := imdao.NewPostCommentDao(tx)
		p, err := postDao.GetByIDForUpdate(cmt.PostID)
		if err != nil {
			return err
		}
		if err := ensurePostVisible(p); err != nil {
			return err
		}
		ok, err := commentDao.SoftDelete(userID, commentID)
		if err != nil {
			return err
		}
		if !ok {
			return bizerror.ErrorCommentNotFound
		}
		if err := postDao.DecComment(cmt.PostID); err != nil {
			return err
		}
		p2, err := postDao.GetByID(cmt.PostID)
		if err != nil {
			return err
		}
		outPost = p2
		return nil
	})
	if err != nil {
		logger.Errorf("DeleteComment user=%d commentID=%d err=%v", userID, commentID, err)
		return nil, err
	}
	item, err := buildPostItem(userID, outPost)
	if err != nil {
		return nil, err
	}
	return &apipost.DeleteCommentReply{Post: item}, nil
}

func requireActiveUser(userID int64) error {
	u, err := userdao.NewUserDao(nil).GetByID(userID)
	if err != nil {
		logger.Errorf("requireActiveUser GetByID user=%d err=%v", userID, err)
		return err
	}
	if u.Banned() {
		logger.Errorf("requireActiveUser banned user=%d", userID)
		return codeerror.ErrorUserBanned
	}
	return nil
}

func loadVisiblePost(postID int64) (*model.Post, error) {
	if postID <= 0 {
		logger.Errorf("loadVisiblePost post_id invalid postID=%d", postID)
		return nil, codeerror.ErrorParam
	}
	p, err := imdao.NewPostDao(nil).GetByID(postID)
	if err != nil {
		logger.Errorf("loadVisiblePost GetByID postID=%d err=%v", postID, err)
		return nil, err
	}
	if err := ensurePostVisible(p); err != nil {
		return nil, err
	}
	return p, nil
}

func ensurePostVisible(p *model.Post) error {
	if p == nil || p.Deleted != kitcommon.DeletedNo {
		logger.Errorf("ensurePostVisible missing or deleted")
		return bizerror.ErrorPostNotFound
	}
	u, err := userdao.NewUserDao(nil).GetByID(p.UserID)
	if err != nil {
		if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorUserNotFound) {
			logger.Errorf("ensurePostVisible author missing postID=%d userID=%d", p.ID, p.UserID)
			return bizerror.ErrorPostNotFound
		}
		logger.Errorf("ensurePostVisible GetByID postID=%d userID=%d err=%v", p.ID, p.UserID, err)
		return err
	}
	if u.Banned() {
		logger.Errorf("ensurePostVisible author banned postID=%d userID=%d", p.ID, p.UserID)
		return bizerror.ErrorPostNotFound
	}
	return nil
}

func buildPostItem(viewerID int64, p *model.Post) (*apipost.PostItem, error) {
	if p == nil {
		return nil, bizerror.ErrorPostNotFound
	}
	liked := false
	row, err := imdao.NewPostLikeDao(nil).Get(p.ID, viewerID)
	if err != nil {
		return nil, err
	}
	if row != nil && row.Deleted == kitcommon.DeletedNo {
		liked = true
	}
	profiles, err := userProfiles([]int64{p.UserID})
	if err != nil {
		return nil, err
	}
	return toPostItem(p, profiles[p.UserID], liked), nil
}

type userBrief struct {
	Nickname string
	Avatar   string
}

func toPostItem(p *model.Post, brief userBrief, liked bool) *apipost.PostItem {
	return &apipost.PostItem{
		PostId:         p.ID,
		UserId:         p.UserID,
		Nickname:       brief.Nickname,
		Avatar:         brief.Avatar,
		Text:           p.Text,
		LikeCount:      p.LikeCount,
		CommentCount:   p.CommentCount,
		Liked:          liked,
		Ct:             p.Ct,
		RecommendScore: p.RecommendScore,
		ImageUrls:      decodePostImageURLs(p.Images),
	}
}

func normalizePostImages(raw []string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	if len(raw) > postImageMax {
		logger.Errorf("normalizePostImages too many n=%d max=%d", len(raw), postImageMax)
		return nil, codeerror.ErrorParam
	}
	out := make([]string, 0, len(raw))
	for _, r := range raw {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}
		uri, err := normalizeImageURI(r)
		if err != nil {
			logger.Errorf("normalizePostImages invalid uri=%s err=%v", r, err)
			return nil, codeerror.ErrorParam
		}
		out = append(out, uri)
	}
	if len(out) > postImageMax {
		logger.Errorf("normalizePostImages too many after trim n=%d", len(out))
		return nil, codeerror.ErrorParam
	}
	return out, nil
}

func encodePostImages(uris []string) (string, error) {
	if len(uris) == 0 {
		return "", nil
	}
	bs, err := json.Marshal(uris)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func decodePostImageURLs(raw string) []string {
	uris := decodePostImageURIs(raw)
	if len(uris) == 0 {
		return nil
	}
	out := make([]string, 0, len(uris))
	for _, uri := range uris {
		out = append(out, avatarURL(uri))
	}
	return out
}

func decodePostImageURIs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var uris []string
	if err := json.Unmarshal([]byte(raw), &uris); err != nil {
		logger.Errorf("decodePostImageURIs err=%v", err)
		return nil
	}
	return uris
}

func toCommentItem(c *model.PostComment, brief userBrief) *apipost.CommentItem {
	return &apipost.CommentItem{
		CommentId: c.ID,
		PostId:    c.PostID,
		UserId:    c.UserID,
		Nickname:  brief.Nickname,
		Avatar:    brief.Avatar,
		Text:      c.Text,
		Ct:        c.Ct,
	}
}

func userProfiles(ids []int64) (map[int64]userBrief, error) {
	out := make(map[int64]userBrief, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := out[id]; ok {
			continue
		}
		u, err := userdao.NewUserDao(nil).GetByID(id)
		if err != nil {
			if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorUserNotFound) {
				out[id] = userBrief{}
				continue
			}
			return nil, err
		}
		out[id] = userBrief{Nickname: u.Nickname, Avatar: avatarURL(u.Avatar)}
	}
	return out, nil
}

func bannedUserSet(ids []int64) (map[int64]bool, error) {
	out := make(map[int64]bool)
	for _, id := range ids {
		if id <= 0 || out[id] {
			continue
		}
		u, err := userdao.NewUserDao(nil).GetByID(id)
		if err != nil {
			if ce, ok := codeerror.As(err); ok && ce.Is(codeerror.ErrorUserNotFound) {
				out[id] = true
				continue
			}
			return nil, err
		}
		if u.Banned() {
			out[id] = true
		}
	}
	return out, nil
}

func uniqueUserIDs(posts []*model.Post) []int64 {
	seen := map[int64]struct{}{}
	var ids []int64
	for _, p := range posts {
		if _, ok := seen[p.UserID]; ok {
			continue
		}
		seen[p.UserID] = struct{}{}
		ids = append(ids, p.UserID)
	}
	return ids
}

func postAuthors(posts []*model.Post) []int64 {
	return uniqueUserIDs(posts)
}

func postIDsOf(posts []*model.Post) []int64 {
	ids := make([]int64, 0, len(posts))
	for _, p := range posts {
		ids = append(ids, p.ID)
	}
	return ids
}

func commentAuthors(comments []*model.PostComment) []int64 {
	seen := map[int64]struct{}{}
	var ids []int64
	for _, c := range comments {
		if _, ok := seen[c.UserID]; ok {
			continue
		}
		seen[c.UserID] = struct{}{}
		ids = append(ids, c.UserID)
	}
	return ids
}

func clampLimit(limit int32, def, max int) int {
	n := int(limit)
	if n <= 0 {
		return def
	}
	if n > max {
		return max
	}
	return n
}

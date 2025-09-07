// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
	"fmt"
	"strings"
	"time"

	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/rocboss/paopao-ce/pkg/debug"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	_ core.TweetService       = (*tweetSrv)(nil)
	_ core.TweetManageService = (*tweetManageSrv)(nil)
	_ core.TweetHelpService   = (*tweetHelpSrv)(nil)

	_ core.TweetServantA       = (*tweetSrvA)(nil)
	_ core.TweetManageServantA = (*tweetManageSrvA)(nil)
	_ core.TweetHelpServantA   = (*tweetHelpSrvA)(nil)
)

type tweetSrv struct {
	db *gorm.DB
}

type tweetManageSrv struct {
	cacheIndex core.CacheIndexService
	db         *gorm.DB
}

type tweetHelpSrv struct {
	db *gorm.DB
}

type tweetSrvA struct {
	db *gorm.DB
}

type tweetManageSrvA struct {
	db *gorm.DB
}

type tweetHelpSrvA struct {
	db *gorm.DB
}

func newTweetService(db *gorm.DB) core.TweetService {
	return &tweetSrv{
		db: db,
	}
}

func newTweetManageService(db *gorm.DB, cacheIndex core.CacheIndexService) core.TweetManageService {
	return &tweetManageSrv{
		cacheIndex: cacheIndex,
		db:         db,
	}
}

func newTweetHelpService(db *gorm.DB) core.TweetHelpService {
	return &tweetHelpSrv{
		db: db,
	}
}

func newTweetServantA(db *gorm.DB) core.TweetServantA {
	return &tweetSrvA{
		db: db,
	}
}

func newTweetManageServantA(db *gorm.DB) core.TweetManageServantA {
	return &tweetManageSrvA{
		db: db,
	}
}

func newTweetHelpServantA(db *gorm.DB) core.TweetHelpServantA {
	return &tweetHelpSrvA{
		db: db,
	}
}

// MergePosts post数据整合
func (s *tweetHelpSrv) MergePosts(posts []*ms.Post) ([]*ms.PostFormated, error) {
	postIds := make([]int64, 0, len(posts))
	userIds := make([]int64, 0, len(posts)*2) // Double capacity for both host and visitor IDs
	for _, post := range posts {
		postIds = append(postIds, post.ID)
		userIds = append(userIds, post.GetHostID())
		if visitorID := post.GetVisitorID(); visitorID > 0 {
			userIds = append(userIds, visitorID)
		}
	}

	postContents, err := s.getPostContentsByIDs(postIds)
	if err != nil {
		return nil, err
	}

	users, err := s.getUsersByIDs(userIds)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]*dbr.UserFormated, len(users))
	for _, user := range users {
		userMap[user.ID] = user.Format()
	}

	contentMap := make(map[int64][]*dbr.PostContentFormated, len(postContents))
	for _, content := range postContents {
		contentMap[content.PostID] = append(contentMap[content.PostID], content.Format())
	}

	// 数据整合
	postsFormated := make([]*dbr.PostFormated, 0, len(posts))
	for _, post := range posts {
		postFormated := post.Format()
		postFormated.User = userMap[post.GetHostID()]
		if visitorID := post.GetVisitorID(); visitorID > 0 {
			postFormated.Visitor = userMap[visitorID]
		}
		postFormated.Contents = contentMap[post.ID]
		postsFormated = append(postsFormated, postFormated)
	}
	return postsFormated, nil
}

// RevampPosts post数据整形修复
func (s *tweetHelpSrv) RevampPosts(posts []*ms.PostFormated) ([]*ms.PostFormated, error) {
	postIds := make([]int64, 0, len(posts))
	userIds := make([]int64, 0, len(posts)*2) // Double capacity for both host and visitor IDs
	for _, post := range posts {
		postIds = append(postIds, post.ID)
		userIds = append(userIds, post.GetHostID())
		if visitorID := post.GetVisitorID(); visitorID > 0 {
			userIds = append(userIds, visitorID)
		}
	}

	postContents, err := s.getPostContentsByIDs(postIds)
	if err != nil {
		return nil, err
	}

	users, err := s.getUsersByIDs(userIds)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int64]*dbr.UserFormated, len(users))
	for _, user := range users {
		userMap[user.ID] = user.Format()
	}

	contentMap := make(map[int64][]*dbr.PostContentFormated, len(postContents))
	for _, content := range postContents {
		contentMap[content.PostID] = append(contentMap[content.PostID], content.Format())
	}

	// 数据整合
	for _, post := range posts {
		post.User = userMap[post.GetHostID()]
		if visitorID := post.GetVisitorID(); visitorID > 0 {
			post.Visitor = userMap[visitorID]
		}
		post.Contents = contentMap[post.ID]
	}
	return posts, nil
}

func (s *tweetHelpSrv) getPostContentsByIDs(ids []int64) ([]*dbr.PostContent, error) {
	return (&dbr.PostContent{}).List(s.db, &dbr.ConditionsT{
		"post_id IN ?": ids,
		"ORDER":        "sort ASC",
	}, 0, 0)
}

func (s *tweetHelpSrv) getUsersByIDs(ids []int64) ([]*dbr.User, error) {
	user := &dbr.User{}

	return user.List(s.db, &dbr.ConditionsT{
		"id IN ?": ids,
	}, 0, 0)
}

func (s *tweetManageSrv) CreatePostCollection(postID, userID int64) (*ms.PostCollection, error) {
	collection := &dbr.PostCollection{
		PostID: postID,
		UserID: userID,
	}

	return collection.Create(s.db)
}

func (s *tweetManageSrv) DeletePostCollection(p *ms.PostCollection) error {
	return p.Delete(s.db)
}

func (s *tweetManageSrv) CreatePostContent(content *ms.PostContent) (*ms.PostContent, error) {
	return content.Create(s.db)
}

func (s *tweetManageSrv) CreateAttachment(obj *ms.Attachment) (int64, error) {
	attachment, err := obj.Create(s.db)
	return attachment.ID, err
}

func (s *tweetManageSrv) UpdatePostContent(content *ms.PostContent) error {
	postContent := &dbr.PostContent{
		Model: &dbr.Model{
			ID: content.ID,
		},
		Content:  content.Content,
		Duration: content.Duration,
		Size:     content.Size,
	}
	return postContent.UpdateAudioContentByRoomId(s.db, content.RoomID, content.Content, content.Duration, content.Size)
}

func (s *tweetManageSrv) CreatePost(post *ms.Post) (*ms.Post, error) {
	logrus.Debugf("DEBUG CreatePost: Starting with post=%+v", post)
	logrus.Debugf("DEBUG CreatePost: Model=%+v", post.Model)
	logrus.Debugf("DEBUG CreatePost: UserID=%v (type: %T)", post.UserID, post.UserID)

	post.LatestRepliedOn = time.Now().Unix()
	logrus.Debugf("DEBUG CreatePost: About to call post.Create with db=%+v", s.db)
	
	p, err := post.Create(s.db)
	if err != nil {
		logrus.Errorf("DEBUG CreatePost Error: %v", err)
		return nil, err
	}
	
	logrus.Debugf("DEBUG CreatePost: Successfully created post=%+v", p)
	s.cacheIndex.SendAction(core.IdxActCreatePost, post)
	return p, nil
}

func (s *tweetManageSrv) DeletePost(post *ms.Post) ([]string, error) {
	var mediaContents []string
	postId := post.ID
	postContent := &dbr.PostContent{}
	err := s.db.Transaction(
		func(tx *gorm.DB) error {
			if contents, err := postContent.MediaContentsByPostId(tx, postId); err == nil {
				mediaContents = contents
			} else {
				return err
			}

			// 删推文
			if err := post.Delete(tx); err != nil {
				return err
			}

			// 删内容
			if err := postContent.DeleteByPostId(tx, postId); err != nil {
				return err
			}

			// 删评论
			if contents, err := s.deleteCommentByPostId(tx, postId); err == nil {
				mediaContents = append(mediaContents, contents...)
			} else {
				return err
			}

			if tags := strings.Split(post.Tags, ","); len(tags) > 0 {
				// 删tag，宽松处理错误，有错误不会回滚
				deleteTags(tx, tags)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	s.cacheIndex.SendAction(core.IdxActDeletePost, post)
	return mediaContents, nil
}

func (s *tweetManageSrv) deleteCommentByPostId(db *gorm.DB, postId int64) ([]string, error) {
	comment := &dbr.Comment{}
	commentContent := &dbr.CommentContent{}

	// 获取推文的所有评论id
	commentIds, err := comment.CommentIdsByPostId(db, postId)
	if err != nil {
		return nil, err
	}

	// 获取评论的媒体内容
	mediaContents, err := commentContent.MediaContentsByCommentId(db, commentIds)
	if err != nil {
		return nil, err
	}

	// 删评论
	if err = comment.DeleteByPostId(db, postId); err != nil {
		return nil, err
	}

	// 删评论内容
	if err = commentContent.DeleteByCommentIds(db, commentIds); err != nil {
		return nil, err
	}

	// 删评论的评论
	if err = (&dbr.CommentReply{}).DeleteByCommentIds(db, commentIds); err != nil {
		return nil, err
	}

	return mediaContents, nil
}

func (s *tweetManageSrv) LockPost(post *ms.Post) error {
	post.IsLock = 1 - post.IsLock
	return post.Update(s.db)
}

func (s *tweetManageSrv) StickPost(post *ms.Post) error {
	post.IsTop = 1 - post.IsTop
	if err := post.Update(s.db); err != nil {
		return err
	}
	s.cacheIndex.SendAction(core.IdxActStickPost, post)
	return nil
}

func (s *tweetManageSrv) HighlightPost(userId int64, postId int64) (res int, err error) {
	var post dbr.Post
	tx := s.db.Begin()
	defer tx.Rollback()
	post.Get(tx)
	if err = tx.Where("id = ? AND is_del = 0", postId).First(&post).Error; err != nil {
		return
	}
	if post.GetHostID() != userId {
		return 0, cs.ErrNoPermission
	}
	post.IsEssence = 1 - post.IsEssence
	if err = post.Update(tx); err != nil {
		return
	}
	tx.Commit()
	return post.IsEssence, nil
}

func (s *tweetManageSrv) VisiblePost(post *ms.Post, visibility cs.TweetVisibleType) (err error) {
	oldVisibility := post.Visibility
	post.Visibility = ms.PostVisibleT(visibility)
	// TODO: 这个判断是否可以不要呢
	if oldVisibility == ms.PostVisibleT(visibility) {
		return nil
	}
	// 私密推文 特殊处理
	if visibility == cs.TweetVisitPrivate {
		// 强制取消置顶
		// TODO: 置顶推文用户是否有权设置成私密？ 后续完善
		post.IsTop = 0
	}
	tx := s.db.Begin()
	defer tx.Rollback()
	if err = post.Update(tx); err != nil {
		return
	}
	// tag处理
	tags := strings.Split(post.Tags, ",")
	// TODO: 暂时宽松不处理错误，这里或许可以有优化，后续完善
	if oldVisibility == dbr.PostVisitPrivate {
		// 从私密转为非私密才需要重新创建tag
		createTags(tx, post.GetHostID(), tags)
	} else if visibility == cs.TweetVisitPrivate {
		// 从非私密转为私密才需要删除tag
		deleteTags(tx, tags)
	}
	tx.Commit()
	s.cacheIndex.SendAction(core.IdxActVisiblePost, post)
	return
}

func (s *tweetManageSrv) UpdatePost(post *ms.Post) (err error) {
	if err = post.Update(s.db); err != nil {
		return
	}
	s.cacheIndex.SendAction(core.IdxActUpdatePost, post)
	return
}

func (s *tweetManageSrv) CreatePostStar(postID, userID int64) (*ms.PostStar, error) {
	star := &dbr.PostStar{
		PostID: postID,
		UserID: userID,
	}
	return star.Create(s.db)
}

func (s *tweetManageSrv) DeletePostStar(p *ms.PostStar) error {
	return p.Delete(s.db)
}

func (s *tweetSrv) GetPostByID(id int64) (*ms.Post, error) {
	post := &dbr.Post{
		Model: &dbr.Model{
			ID: id,
		},
	}
	return post.Get(s.db)
}

func (s *tweetSrv) GetPosts(conditions ms.ConditionsT, offset, limit int) ([]*ms.Post, error) {
	return (&dbr.Post{}).List(s.db, conditions, offset, limit)
}

func (s *tweetSrv) ListUserTweets(userId int64, style uint8, justEssence bool, limit, offset int) (res []*ms.Post, total int64, err error) {
	logrus.Debugf("DEBUG ListUserTweets: Starting with userId=%d, style=%d, justEssence=%v, limit=%d, offset=%d", userId, style, justEssence, limit, offset)
	
	// Include empty content filter in initial query setup
	db := s.db.Model(&dbr.Post{}).Where("CAST(user_id->0 AS bigint) = ? AND EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')", userId)
	logrus.Debugf("DEBUG ListUserTweets: Initial query condition user_id->0 = %d with empty content filter", userId)
	
	switch style {
	case cs.StyleUserTweetsAdmin:
		fallthrough
	case cs.StyleUserTweetsSelf:
		db = db.Where("visibility >= ?", cs.TweetVisitPrivate)
		logrus.Debugf("DEBUG ListUserTweets: Added admin/self visibility condition >= %d", cs.TweetVisitPrivate)
	case cs.StyleUserTweetsFriend:
		db = db.Where("visibility >= ?", cs.TweetVisitFriend)
		logrus.Debugf("DEBUG ListUserTweets: Added friend visibility condition >= %d", cs.TweetVisitFriend)
	case cs.StyleUserTweetsFollowing:
		db = db.Where("visibility >= ?", cs.TweetVisitFollowing)
		logrus.Debugf("DEBUG ListUserTweets: Added following visibility condition >= %d", cs.TweetVisitFollowing)
	case cs.StyleUserTweetsGuest:
		fallthrough
	default:
		db = db.Where("visibility >= ?", cs.TweetVisitPublic)
		logrus.Debugf("DEBUG ListUserTweets: Added public visibility condition >= %d", cs.TweetVisitPublic)
	}
	if justEssence {
		db = db.Where("is_essence=1")
		logrus.Debug("DEBUG ListUserTweets: Added essence condition")
	}
	
	if err = db.Count(&total).Error; err != nil {
		logrus.Errorf("DEBUG ListUserTweets: Error counting total: %v", err)
		return
	}
	logrus.Debugf("DEBUG ListUserTweets: Total count = %d", total)
	
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
		logrus.Debugf("DEBUG ListUserTweets: Added offset=%d and limit=%d", offset, limit)
	}
	if err = db.Order("is_top DESC, latest_replied_on DESC").Find(&res).Error; err != nil {
		logrus.Errorf("DEBUG ListUserTweets: Error fetching results: %v", err)
		return
	}
	logrus.Debugf("DEBUG ListUserTweets: Successfully fetched %d posts", len(res))
	return
}

func (s *tweetSrv) ListIndexNewestTweets(limit, offset int) (res []*ms.Post, total int64, err error) {
	// Include empty content filter in initial query setup
	db := s.db.Table(_post_).Where("visibility >= ? AND EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')", cs.TweetVisitPublic)
	
	if err = db.Count(&total).Error; err != nil {
		return
	}
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if err = db.Order("is_top DESC, latest_replied_on DESC").Find(&res).Error; err != nil {
		return
	}
	return
}

func (s *tweetSrv) ListIndexHotsTweets(limit, offset int) (res []*ms.Post, total int64, err error) {
	// Include empty content filter in initial query setup
	db := s.db.Table(_post_).Joins(fmt.Sprintf("LEFT JOIN %s metric ON %s.id=metric.post_id", _post_metric_, _post_)).Where(fmt.Sprintf("visibility >= ? AND %s.is_del=0 AND metric.is_del=0 AND EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')", _post_), cs.TweetVisitPublic)
	
	if err = db.Count(&total).Error; err != nil {
		return
	}
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if err = db.Order("is_top DESC, metric.rank_score DESC, latest_replied_on DESC").Find(&res).Error; err != nil {
		return
	}
	return
}

func (s *tweetSrv) ListSyncSearchTweets(limit, offset int) (res []*ms.Post, total int64, err error) {
	// Include empty content filter in initial query setup
	db := s.db.Table(_post_).Where("visibility >= ? AND EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')", cs.TweetVisitFriend)
	if err = db.Count(&total).Error; err != nil {
		return
	}
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if err = db.Find(&res).Error; err != nil {
		return
	}
	return
}

func (s *tweetSrv) ListFollowingTweets(userId int64, limit, offset int) (res []*ms.Post, total int64, err error) {
	logrus.Debugf("DEBUG ListFollowingTweets: Starting with userId=%d, limit=%d, offset=%d", userId, limit, offset)
	
	beFriendIds, beFollowIds, xerr := s.getUserRelation(userId)
	if xerr != nil {
		logrus.Errorf("DEBUG ListFollowingTweets: Error getting user relations: %v", xerr)
		return nil, 0, xerr
	}
	logrus.Debugf("DEBUG ListFollowingTweets: Got friend IDs=%v, follow IDs=%v", beFriendIds, beFollowIds)
	
	beFriendCount, beFollowCount := len(beFriendIds), len(beFollowIds)
	// Include empty content filter in initial query setup
	db := s.db.Model(&dbr.Post{}).Where("EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')")
	//可见性: 0私密 10充电可见 20订阅可见 30保留 40保留 50好友可见 60关注可见 70保留 80保留 90公开',
	switch {
	case beFriendCount > 0 && beFollowCount > 0:
		db = db.Where("CAST(user_id->0 AS bigint)=? OR (visibility>=50 AND CAST(user_id->0 AS bigint) IN(?)) OR (visibility>=60 AND CAST(user_id->0 AS bigint) IN(?))", userId, beFriendIds, beFollowIds)
		logrus.Debug("DEBUG ListFollowingTweets: Using condition for both friends and followers")
	case beFriendCount > 0 && beFollowCount == 0:
		db = db.Where("CAST(user_id->0 AS bigint)=? OR (visibility>=50 AND CAST(user_id->0 AS bigint) IN(?))", userId, beFriendIds)
		logrus.Debug("DEBUG ListFollowingTweets: Using condition for friends only")
	case beFriendCount == 0 && beFollowCount > 0:
		db = db.Where("CAST(user_id->0 AS bigint)=? OR (visibility>=60 AND CAST(user_id->0 AS bigint) IN(?))", userId, beFollowIds)
		logrus.Debug("DEBUG ListFollowingTweets: Using condition for followers only")
	case beFriendCount == 0 && beFollowCount == 0:
		db = db.Where("CAST(user_id->0 AS bigint) = ?", userId)
		logrus.Debug("DEBUG ListFollowingTweets: Using condition for user only")
	}
	
	if err = db.Count(&total).Error; err != nil {
		logrus.Errorf("DEBUG ListFollowingTweets: Error counting total: %v", err)
		return
	}
	logrus.Debugf("DEBUG ListFollowingTweets: Total count = %d", total)
	
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
		logrus.Debugf("DEBUG ListFollowingTweets: Added offset=%d and limit=%d", offset, limit)
	}
	if err = db.Order("is_top DESC, latest_replied_on DESC").Find(&res).Error; err != nil {
		logrus.Errorf("DEBUG ListFollowingTweets: Error fetching results: %v", err)
		return
	}
	logrus.Debugf("DEBUG ListFollowingTweets: Successfully fetched %d posts", len(res))
	return
}

func (s *tweetSrv) getUserRelation(userId int64) (beFriendIds []int64, beFollowIds []int64, err error) {
	if err = s.db.Table(_contact_).Where("friend_id=? AND status=2 AND is_del=0", userId).Select("user_id").Find(&beFriendIds).Error; err != nil {
		return
	}
	if err = s.db.Table(_following_).Where("user_id=? AND is_del=0", userId).Select("follow_id").Find(&beFollowIds).Error; err != nil {
		return
	}
	// 即是好友又是关注者，保留好友去除关注者
	for _, id := range beFriendIds {
		for i := 0; i < len(beFollowIds); i++ {
			// 找到item即删，数据库已经保证唯一性
			if beFollowIds[i] == id {
				lastIdx := len(beFollowIds) - 1
				beFollowIds[i] = beFollowIds[lastIdx]
				beFollowIds = beFollowIds[:lastIdx]
				break
			}
		}
	}
	return
}

func (s *tweetSrv) GetPostCount(conditions ms.ConditionsT) (int64, error) {
	return (&dbr.Post{}).Count(s.db, conditions)
}

func (s *tweetSrv) GetUserPostStar(postID, userID int64) (*ms.PostStar, error) {
	star := &dbr.PostStar{
		PostID: postID,
		UserID: userID,
	}
	return star.Get(s.db)
}

func (s *tweetSrv) GetUserPostStars(userID int64, limit int, offset int) ([]*ms.PostStar, error) {
	star := &dbr.PostStar{
		UserID: userID,
	}
	return star.List(s.db, &dbr.ConditionsT{
		"ORDER": s.db.NamingStrategy.TableName("PostStar") + ".id DESC",
	}, cs.RelationSelf, limit, offset)
}

func (s *tweetSrv) ListUserStarTweets(user *cs.VistUser, limit int, offset int) (res []*ms.PostStar, total int64, err error) {
	star := &dbr.PostStar{
		UserID: user.UserId,
	}
	if total, err = star.Count(s.db, user.RelTyp, &dbr.ConditionsT{}); err != nil {
		return
	}
	res, err = star.List(s.db, &dbr.ConditionsT{
		"ORDER": s.db.NamingStrategy.TableName("PostStar") + ".id DESC",
	}, user.RelTyp, limit, offset)
	return
}

func (s *tweetSrv) getUserTweets(db *gorm.DB, user *cs.VistUser, limit int, offset int) (res []*ms.Post, total int64, err error) {
	logrus.Debugf("DEBUG getUserTweets: Starting with userId=%d, relationType=%d, limit=%d, offset=%d", user.UserId, user.RelTyp, limit, offset)
	
	visibilities := []core.PostVisibleT{core.PostVisitPublic}
	switch user.RelTyp {
	case cs.RelationAdmin, cs.RelationSelf:
		visibilities = append(visibilities, core.PostVisitPrivate, core.PostVisitFriend)
		logrus.Debug("DEBUG getUserTweets: Added private and friend visibilities for admin/self")
	case cs.RelationFriend:
		visibilities = append(visibilities, core.PostVisitFriend)
		logrus.Debug("DEBUG getUserTweets: Added friend visibility for friend")
	case cs.RelationGuest:
		fallthrough
	default:
		logrus.Debug("DEBUG getUserTweets: Using only public visibility for guest")
	}
	
	// Include empty content filter in initial query setup
	db = db.Where("visibility IN ? AND is_del=0 AND EXISTS (SELECT 1 FROM p_post_content WHERE post_id = p_post.id AND content != '')", visibilities)
	logrus.Debugf("DEBUG getUserTweets: Using visibilities=%v with empty content filter", visibilities)
	
	err = db.Count(&total).Error
	if err != nil {
		logrus.Errorf("DEBUG getUserTweets: Error counting total: %v", err)
		return
	}
	logrus.Debugf("DEBUG getUserTweets: Total count = %d", total)
	
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
		logrus.Debugf("DEBUG getUserTweets: Added offset=%d and limit=%d", offset, limit)
	}
	
	err = db.Order("is_top DESC, latest_replied_on DESC").Find(&res).Error
	if err != nil {
		logrus.Errorf("DEBUG getUserTweets: Error fetching results: %v", err)
		return
	}
	logrus.Debugf("DEBUG getUserTweets: Successfully fetched %d posts", len(res))
	return
}

func (s *tweetSrv) ListUserMediaTweets(user *cs.VistUser, limit int, offset int) ([]*ms.Post, int64, error) {
	logrus.Debugf("DEBUG ListUserMediaTweets: Starting with userId=%d, limit=%d, offset=%d", user.UserId, limit, offset)
	db := s.db.Table(_post_by_media_).Where("CAST(user_id->0 AS bigint) = ?", user.UserId)
	logrus.Debugf("DEBUG ListUserMediaTweets: Using query condition CAST(user_id->0 AS bigint) = %d", user.UserId)
	return s.getUserTweets(db, user, limit, offset)
}

func (s *tweetSrv) ListUserCommentTweets(user *cs.VistUser, limit int, offset int) ([]*ms.Post, int64, error) {
	logrus.Debugf("DEBUG ListUserCommentTweets: Starting with userId=%d, limit=%d, offset=%d", user.UserId, limit, offset)
	db := s.db.Table(_post_by_comment_).Where("CAST(user_id->0 AS bigint) = ?", user.UserId)
	logrus.Debugf("DEBUG ListUserCommentTweets: Using query condition CAST(user_id->0 AS bigint) = %d", user.UserId)
	return s.getUserTweets(db, user, limit, offset)
}

func (s *tweetSrv) GetUserPostStarCount(userID int64) (int64, error) {
	star := &dbr.PostStar{
		UserID: userID,
	}
	return star.Count(s.db, cs.RelationSelf, &dbr.ConditionsT{})
}

func (s *tweetSrv) GetUserPostCollection(postID, userID int64) (*ms.PostCollection, error) {
	star := &dbr.PostCollection{
		PostID: postID,
		UserID: userID,
	}
	return star.Get(s.db)
}

func (s *tweetSrv) GetUserPostCollections(userID int64, offset, limit int) ([]*ms.PostCollection, error) {
	collection := &dbr.PostCollection{
		UserID: userID,
	}

	return collection.List(s.db, &dbr.ConditionsT{
		"ORDER": s.db.NamingStrategy.TableName("PostCollection") + ".id DESC",
	}, offset, limit)
}

func (s *tweetSrv) GetUserPostCollectionCount(userID int64) (int64, error) {
	collection := &dbr.PostCollection{
		UserID: userID,
	}
	return collection.Count(s.db, &dbr.ConditionsT{})
}

func (s *tweetSrv) GetUserWalletBills(userID int64, offset, limit int) ([]*ms.WalletStatement, error) {
	statement := &dbr.WalletStatement{
		UserID: userID,
	}

	return statement.List(s.db, &dbr.ConditionsT{
		"ORDER": "id DESC",
	}, offset, limit)
}

func (s *tweetSrv) GetUserWalletBillCount(userID int64) (int64, error) {
	statement := &dbr.WalletStatement{
		UserID: userID,
	}
	return statement.Count(s.db, &dbr.ConditionsT{})
}

func (s *tweetSrv) GetPostAttatchmentBill(postID, userID int64) (*ms.PostAttachmentBill, error) {
	bill := &dbr.PostAttachmentBill{
		PostID: postID,
		UserID: userID,
	}

	return bill.Get(s.db)
}

func (s *tweetSrv) GetPostContentsByIDs(ids []int64) ([]*ms.PostContent, error) {
	return (&dbr.PostContent{}).List(s.db, &dbr.ConditionsT{
		"post_id IN ?": ids,
		"ORDER":        "sort ASC",
	}, 0, 0)
}

func (s *tweetSrv) GetPostContentByID(id int64) (*ms.PostContent, error) {
	return (&dbr.PostContent{
		Model: &dbr.Model{
			ID: id,
		},
	}).Get(s.db)
}

func (s *tweetSrv) GetPostContentByRoomID(roomID string) (*ms.PostContent, error) {
	return (&dbr.PostContent{}).GetPostContentByRoomID(s.db, roomID)
}

func (s *tweetSrv) GetPostContentBySessionID(roomID string, sessionID string) (*ms.PostContent, error) {
	return (&dbr.PostContent{}).GetPostContentBySessionID(s.db, roomID, sessionID)
}

func (s *tweetSrv) GetPostBySessionID(sessionID string) (*ms.Post, error) {
	return (&dbr.Post{}).GetPostBySessionID(s.db, sessionID)
}

func (s *tweetSrv) GetAudioContentByPostID(postID int64) (*ms.PostContent, error) {
	return (&dbr.PostContent{}).GetAudioContentByPostID(s.db, postID)
}

func (s *tweetSrv) UpdateContentByRoomId(roomID string, content string, duration float64, size int64) error {
	postContent := &dbr.PostContent{}
	return postContent.UpdateAudioContentByRoomId(s.db, roomID, content, fmt.Sprintf("%.1f", duration), fmt.Sprintf("%d", size))
}

func (s *tweetSrv) UpdateContentByPostID(postID int64, content string, duration float64, size int64) error {
	postContent := &dbr.PostContent{}
	return postContent.UpdateAudioContentByPostID(s.db, postID, content, fmt.Sprintf("%.1f", duration), fmt.Sprintf("%d", size))
}

func (s *tweetSrvA) TweetInfoById(id int64) (*cs.TweetInfo, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) TweetItemById(id int64) (*cs.TweetItem, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) UserTweets(visitorId, userId int64) (cs.TweetList, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) ReactionByTweetId(userId int64, tweetId int64) (*cs.ReactionItem, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) UserReactions(userId int64, offset int, limit int) (cs.ReactionList, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) FavoriteByTweetId(userId int64, tweetId int64) (*cs.FavoriteItem, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) UserFavorites(userId int64, offset int, limit int) (cs.FavoriteList, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrvA) AttachmentByTweetId(userId int64, tweetId int64) (*cs.AttachmentBill, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetManageSrvA) CreateAttachment(obj *cs.Attachment) (int64, error) {
	// TODO
	return 0, debug.ErrNotImplemented
}

func (s *tweetManageSrvA) CreateTweet(userId int64, req *cs.NewTweetReq) (*cs.TweetItem, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetManageSrvA) DeleteTweet(userId int64, tweetId int64) ([]string, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetManageSrvA) LockTweet(userId int64, tweetId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) StickTweet(userId int64, tweetId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) VisibleTweet(userId int64, visibility cs.TweetVisibleType) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) CreateReaction(userId int64, tweetId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) DeleteReaction(userId int64, reactionId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) CreateFavorite(userId int64, tweetId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetManageSrvA) DeleteFavorite(userId int64, favoriteId int64) error {
	// TODO
	return debug.ErrNotImplemented
}

func (s *tweetHelpSrvA) RevampTweets(tweets cs.TweetList) (cs.TweetList, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetHelpSrvA) MergeTweets(tweets cs.TweetInfo) (cs.TweetList, error) {
	// TODO
	return nil, debug.ErrNotImplemented
}

func (s *tweetSrv) GetPostLocation(postID int64, userID int64, pageSize int) (page int, position int, totalPosts int64, err error) {
	return (&dbr.Post{}).GetPostLocation(s.db, postID, userID, pageSize)
}
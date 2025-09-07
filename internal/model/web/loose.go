// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"

	"github.com/alimy/mir/v4"
	"github.com/gin-gonic/gin"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/model/joint"
	"github.com/rocboss/paopao-ce/internal/servants/base"
	"github.com/rocboss/paopao-ce/pkg/app"
)

const (
	TagTypeHot       = cs.TagTypeHot
	TagTypeNew       = cs.TagTypeNew
	TagTypeFollow    = cs.TagTypeFollow
	TagTypePin       = cs.TagTypePin
	TagTypeHotExtral = cs.TagTypeHotExtral
)

const (
	UserPostsStylePost      = "post"
	UserPostsStyleComment   = "comment"
	UserPostsStyleHighlight = "highlight"
	UserPostsStyleMedia     = "media"
	UserPostsStyleStar      = "star"

	StyleTweetsNewest    = "newest"
	StyleTweetsHots      = "hots"
	StyleTweetsFollowing = "following"
)

type TagType = cs.TagType

type CommentStyleType string

type TweetCommentsReq struct {
	SimpleInfo `form:"-" binding:"-"`
	TweetId    int64            `form:"id" binding:"required"`
	Style      CommentStyleType `form:"style"`
	Page       int              `form:"-" binding:"-"`
	PageSize   int              `form:"-" binding:"-"`
}

type TweetCommentsResp struct {
	joint.CachePageResp
}

type TimelineReq struct {
	BaseInfo   `form:"-"  binding:"-"`
	Query      string              `form:"query"`
	Visibility []core.PostVisibleT `form:"query"`
	Type       string              `form:"type"`
	Style      string              `form:"style"`
	Page       int                 `form:"-"  binding:"-"`
	PageSize   int                 `form:"-"  binding:"-"`
}

type TimelineResp struct {
	joint.CachePageResp
}

type GetUserTweetsReq struct {
	BaseInfo `form:"-" binding:"-"`
	Username string `form:"username" binding:"required"`
	Style    string `form:"style"`
	Page     int    `form:"-" binding:"-"`
	PageSize int    `form:"-" binding:"-"`
}

type GetUserTweetsResp struct {
	joint.CachePageResp
}

type GetUserProfileReq struct {
	BaseInfo `form:"-" binding:"-"`
	Username string `form:"username" binding:"required"`
}

type GetUserProfileResp struct {
	ID          int64            `json:"id"`
	Nickname    string           `json:"nickname"`
	Username    string           `json:"username"`
	Status      int              `json:"status"`
	Avatar      string           `json:"avatar"`
	IsAdmin     bool             `json:"is_admin"`
	IsFriend    bool             `json:"is_friend"`
	IsFollowing bool             `json:"is_following"`
	CreatedOn   int64            `json:"created_on"`
	Follows     int64            `json:"follows"`
	Followings  int64            `json:"followings"`
	TweetsCount int              `json:"tweets_count"`
	Categories  []int64          `json:"categories,omitempty"`
	ReactionCounts map[int64]int64 `json:"reaction_counts"` // reaction_type_id -> count (reactions received)
 	IsOnline    bool             `json:"is_online,omitempty" gorm:"-"` // User's online status (optional, not in DB)
}

type TopicListReq struct {
	SimpleInfo `form:"-"  binding:"-"`
	Type       TagType `json:"type" form:"type" binding:"required"`
	Num        int     `json:"num" form:"num" binding:"required"`
	ExtralNum  int     `json:"extral_num" form:"extral_num"`
}

// TopicListResp 主题返回值
// TODO: 优化内容定义
type TopicListResp struct {
	Topics       cs.TagList `json:"topics"`
	ExtralTopics cs.TagList `json:"extral_topics,omitempty"`
}

type TweetDetailReq struct {
	BaseInfo `form:"-"  binding:"-"`
	TweetId  int64 `form:"id"`
}

type TweetDetailResp ms.PostFormated

// PostLocationReq represents the request for getting post location
type PostLocationReq struct {
	BaseInfo `form:"-" binding:"-"`
	PostID   int64  `uri:"postId" binding:"required"`
	Username string `form:"username" binding:"required"`
}

// Bind implements custom binding for path parameters
func (r *PostLocationReq) Bind(c *gin.Context) mir.Error {
	user, _ := base.UserFrom(c)
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get postId from path parameter
	if postIdStr := c.Param("postId"); postIdStr != "" {
		if postId, err := strconv.ParseInt(postIdStr, 10, 64); err == nil {
			r.PostID = postId
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid post ID")
		}
	}
	
	// Get username from query parameter
	r.Username = c.Query("username")
	
	return nil
}

// PostLocationResp represents the response for post location
type PostLocationResp struct {
	PostID     int64 `json:"post_id"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalPosts int64 `json:"total_posts"`
	Position   int   `json:"position"` // Position within the page (1-based)
}

func (r *GetUserTweetsReq) SetPageInfo(page int, pageSize int) {
	r.Page, r.PageSize = page, pageSize
}

func (r *TweetCommentsReq) SetPageInfo(page int, pageSize int) {
	r.Page, r.PageSize = page, pageSize
}

func (r *TimelineReq) Bind(c *gin.Context) mir.Error {
	user, _ := base.UserFrom(c)
	r.BaseInfo = BaseInfo{
		User: user,
	}
	r.Page, r.PageSize = app.GetPageInfo(c)
	r.Query, r.Type, r.Style = c.Query("query"), "search", c.Query("style")
	return nil
}

func (s CommentStyleType) ToInnerValue() (res cs.StyleCommentType) {
	switch s {
	case "hots":
		res = cs.StyleCommentHots
	case "newest":
		res = cs.StyleCommentNewest
	case "default":
		fallthrough
	default:
		res = cs.StyleCommentDefault
	}
	return
}

func (s CommentStyleType) String() (res string) {
	switch s {
	case "default":
		res = conf.InfixCommentDefault
	case "hots":
		res = conf.InfixCommentHots
	case "newest":
		res = conf.InfixCommentNewest
	default:
		res = "_"
	}
	return
}
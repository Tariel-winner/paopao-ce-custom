// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"net/http"
	"strconv"
	"github.com/alimy/mir/v4"
	"github.com/gin-gonic/gin"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/rocboss/paopao-ce/internal/model/joint"
	"github.com/rocboss/paopao-ce/internal/servants/base"
	"github.com/rocboss/paopao-ce/pkg/app"
	"github.com/rocboss/paopao-ce/pkg/convert"
	"github.com/rocboss/paopao-ce/pkg/xerror"
)





type MessageStyle = cs.MessageStyle
type UserProfile = cs.UserProfile
type UserProfileWithFollow = cs.UserProfileWithFollow
type UserFormated = ms.UserFormated
type UserReactionWithUser = ms.UserReactionWithUser
type LocationData = ms.LocationData

type ChangeAvatarReq struct {
	BaseInfo `json:"-" binding:"-"`
	Avatar   string `json:"avatar" form:"avatar" binding:"required"`
}

type SyncSearchIndexReq struct {
	BaseInfo `json:"-" binding:"-"`
}

type UserInfoReq struct {
	BaseInfo `json:"-" binding:"-"`
	Username string `json:"username" form:"username" binding:"required"`
}

type UserInfoResp struct {
	Id          int64  `json:"id"`
	Nickname    string `json:"nickname"`
	Username    string `json:"username"`
	Status      int    `json:"status"`
	Avatar      string `json:"avatar"`
	Balance     int64  `json:"balance"`
	Phone       string `json:"phone"`
	IsAdmin     bool   `json:"is_admin"`
	CreatedOn   int64  `json:"created_on"`
	Follows     int64  `json:"follows"`
	Followings  int64  `json:"followings"`
	TweetsCount int    `json:"tweets_count"`
	IsOnline    bool   `json:"is_online,omitempty" gorm:"-"` // User's online status (optional, not in DB)
}

type GetMessagesReq struct {
	SimpleInfo `json:"-" binding:"-"`
	joint.BasePageInfo
	Style MessageStyle `form:"style" binding:"required"`
}

type GetMessagesResp struct {
	joint.CachePageResp
}

type ReadMessageReq struct {
	SimpleInfo `json:"-" binding:"-"`
	ID         int64 `json:"id" binding:"required"`
}

type ReadAllMessageReq struct {
	SimpleInfo `json:"-" binding:"-"`
}

type SendWhisperReq struct {
	SimpleInfo `json:"-" binding:"-"`
	UserID     int64  `json:"user_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

type GetCollectionsReq BasePageReq
type GetCollectionsResp base.PageResp

type GetStarsReq BasePageReq
type GetStarsResp base.PageResp

type UserPhoneBindReq struct {
	BaseInfo `json:"-" binding:"-"`
	Phone    string `json:"phone" form:"phone" binding:"required"`
	Captcha  string `json:"captcha" form:"captcha" binding:"required"`
}

type ChangePasswordReq struct {
	BaseInfo    `json:"-" binding:"-"`
	Password    string `json:"password" form:"password" binding:"required"`
	OldPassword string `json:"old_password" form:"old_password" binding:"required"`
}

type ChangeNicknameReq struct {
	BaseInfo `json:"-" binding:"-"`
	Nickname string `json:"nickname" form:"nickname" binding:"required"`
}

type SuggestUsersReq struct {
	BaseInfo `json:"-" binding:"-"`
	Keyword  string
}

type SuggestUsersResp struct {
	Suggests []UserProfileWithFollow `json:"suggests"`
}

type SuggestTagsReq struct {
	Keyword string
}

type SuggestTagsResp struct {
	Suggests []string `json:"suggest"`
}

type TweetStarStatusReq struct {
	SimpleInfo `json:"-" binding:"-"`
	TweetId    int64 `form:"id"`
}

type TweetStarStatusResp struct {
	Status bool `json:"status"`
}

type TweetCollectionStatusReq struct {
	SimpleInfo `json:"-" binding:"-"`
	TweetId    int64 `form:"id"`
}

type TweetCollectionStatusResp struct {
	Status bool `json:"status"`
}

func (r *UserInfoReq) Bind(c *gin.Context) mir.Error {
	username, exist := base.UserNameFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.Username = username
	return nil
}

func (r *GetCollectionsReq) Bind(c *gin.Context) mir.Error {
	return (*BasePageReq)(r).Bind(c)
}

func (r *GetStarsReq) Bind(c *gin.Context) mir.Error {
	return (*BasePageReq)(r).Bind(c)
}

func (r *SuggestTagsReq) Bind(c *gin.Context) mir.Error {
	r.Keyword = c.Query("k")
	return nil
}

func (r *SuggestUsersReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: &ms.User{
			Model: &ms.Model{ID: userId},
		},
	}
	r.Keyword = c.Query("k")
	return nil
}

func (r *TweetCollectionStatusReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	r.TweetId = convert.StrTo(c.Query("id")).MustInt64()
	return nil
}

func (r *TweetStarStatusReq) Bind(c *gin.Context) mir.Error {
	UserId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: UserId,
	}
	r.TweetId = convert.StrTo(c.Query("id")).MustInt64()
	return nil
}

// Room request structs
type RoomListReq struct {
	BaseInfo `json:"-" binding:"-"`
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,min=1,max=50"`
}

type CreateRoomReq struct {
	BaseInfo   `json:"-" binding:"-"`
	HMSRoomID  string   `json:"hms_room_id,omitempty"`
	Topics     []string `json:"topics,omitempty"`
	Categories dbr.Int64Array `json:"categories,omitempty"`
}

type UpdateRoomReq struct {
	BaseInfo          `json:"-" binding:"-"`
	RoomID            int64     `json:"room_id" binding:"required"`
	HMSRoomID         string    `json:"hms_room_id,omitempty"`
	SpeakerIDs        []int64   `json:"speaker_ids,omitempty"`
	StartTime         int64     `json:"start_time,omitempty"`
	Queue             *Queue    `json:"queue,omitempty"`
	IsBlockedFromSpace *int16   `json:"is_blocked_from_space,omitempty"`
	Topics            []string  `json:"topics,omitempty"`
	Categories        dbr.Int64Array `json:"categories,omitempty"`
}

type GetRoomByIDReq struct {
	BaseInfo `json:"-" binding:"-"`
	RoomID   int64 `json:"-" binding:"-"`
}

type GetRoomByHostIDReq struct {
	BaseInfo `json:"-" binding:"-"`
	HostID   int64 `json:"-" binding:"-"`
}

type GetUserRoomReq struct {
	BaseInfo `json:"-" binding:"-"`
}

// User Reaction API Models (User-to-User reactions) - These belong in core service
type CreateUserReactionReq struct {
	SimpleInfo      `json:"-" binding:"-"`
	TargetUserID    int64 `json:"target_user_id" binding:"required"`   // User being reacted to
	ReactionTypeID  int64 `json:"reaction_type_id" binding:"required"` // Type of reaction (1=like, 2=love, etc.)
}

func (r *CreateUserReactionReq) Bind(c *gin.Context) mir.Error {
	// Bind JSON body
	if err := c.ShouldBindJSON(r); err != nil {
		return mir.NewError(xerror.InvalidParams.StatusCode(), xerror.InvalidParams.WithDetails(err.Error()))
	}
	
	// Set user ID from JWT token
	uid, ok := base.UserIdFrom(c)
	if !ok {
		return xerror.UnauthorizedTokenError
	}
	r.SetUserId(uid)
	
	return nil
}

type CreateUserReactionResp struct {
	Status         bool   `json:"status"`
	ReactionTypeID int64  `json:"reaction_type_id"`
	ReactionName   string `json:"reaction_name"`
	ReactionIcon   string `json:"reaction_icon"`
}



type GetUserReactionsReq struct {
	SimpleInfo `json:"-" binding:"-"`
	UserID     int64 `json:"user_id" binding:"required"` // User who receives reactions
}

type GetUserReactionsResp struct {
	ReactionCounts map[int64]int64 `json:"reaction_counts"` // reaction_type_id -> count (reactions received)
}

type GetUserReactionUsersReq struct {
	BaseInfo       `json:"-" binding:"-"`
	UserID         int64 `json:"user_id" binding:"required"`         // User who receives reactions
	ReactionTypeID int64 `json:"reaction_type_id" binding:"required"` // Specific reaction type
	Limit          int   `json:"limit" binding:"required"`
	Offset         int   `json:"offset"`
}

type GetUserReactionUsersResp struct {
	Users []*UserFormated `json:"users"` // Users who reacted to the target user
	Total int64              `json:"total"`
}

type GetUserGivenReactionsReq struct {
	SimpleInfo `json:"-" binding:"-"`
	UserID     int64 `json:"user_id" binding:"required"` // User who gives reactions
}

type GetUserGivenReactionsResp struct {
	ReactionCounts map[int64]int64 `json:"reaction_counts"` // reaction_type_id -> count (reactions given)
}

type GetUserGivenReactionUsersReq struct {
	BaseInfo       `json:"-" binding:"-"`
	UserID         int64 `json:"user_id" binding:"required"`         // User who gives reactions
	ReactionTypeID int64 `json:"reaction_type_id" binding:"required"` // Specific reaction type
	Limit          int   `json:"limit" binding:"required"`
	Offset         int   `json:"offset"`
}

type GetUserGivenReactionUsersResp struct {
	Users []*UserFormated `json:"users"` // Users that the reactor has reacted to
	Total int64              `json:"total"`
}

// GetReactionsToTwoUsersReq gets reactions TO two users with pagination
type GetReactionsToTwoUsersReq struct {
	SimpleInfo `json:"-" binding:"-"`
	User1ID    int64 `json:"user1_id" binding:"required"` // First user ID
	User2ID    int64 `json:"user2_id" binding:"required"` // Second user ID
	Page       int   `json:"page" binding:"required,min=1"`      // Page number
	PageSize   int   `json:"page_size" binding:"required,min=1,max=50"` // Page size
}

// GetReactionsToTwoUsersResp response for reactions to two users with pagination
type GetReactionsToTwoUsersResp struct {
	List  []*UserReactionWithUser `json:"list"`  // Reactions with user data
	Pager base.Pager                 `json:"pager"` // Pagination info
}

// Global reaction timeline request/response
type GetGlobalReactionTimelineReq struct {
	SimpleInfo `json:"-" binding:"-"`
	Page       int `json:"page" binding:"required,min=1"`      // Page number
	PageSize   int `json:"page_size" binding:"required,min=1,max=50"` // Page size
}

type GetGlobalReactionTimelineResp struct {
	List  []*cs.UserReactionWithBothUsers `json:"list"`  // Reactions with both reactor and target user data
	Pager base.Pager                         `json:"pager"` // Pagination info
}

// User-specific reaction timeline request/response
type GetUserReactionTimelineReq struct {
	SimpleInfo `json:"-" binding:"-"`
	UserID     int64 `json:"user_id" binding:"required"` // User whose reactions to get
	Page       int   `json:"page" binding:"required,min=1"`      // Page number
	PageSize   int   `json:"page_size" binding:"required,min=1,max=50"` // Page size
}

type GetUserReactionTimelineResp struct {
	List  []*cs.UserReactionWithBothUsers `json:"list"`  // Reactions with both reactor and target user data
	Pager base.Pager                         `json:"pager"` // Pagination info
}

// SetUserCategoriesReq sets categories for a user
type SetUserCategoriesReq struct {
	BaseInfo    `json:"-" binding:"-"`
	CategoryIDs dbr.Int64Array `json:"category_ids" binding:"required"`
}

// SetUserCategoriesResp response for setting user categories
type SetUserCategoriesResp struct {
	Success bool `json:"success"`
}

// Room Bind methods
func (r *RoomListReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	r.Page, r.PageSize = app.GetPageInfo(c)
	return nil
}

func (r *CreateRoomReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	return nil
}

func (r *UpdateRoomReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get room ID from path parameter
	if roomIdStr := c.Param("id"); roomIdStr != "" {
		if roomId, err := strconv.ParseInt(roomIdStr, 10, 64); err == nil {
			r.RoomID = roomId
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid room ID")
		}
	}
	
	return nil
}

func (r *GetRoomByIDReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get room ID from path parameter
	if roomIdStr := c.Param("id"); roomIdStr != "" {
		if roomId, err := strconv.ParseInt(roomIdStr, 10, 64); err == nil {
			r.RoomID = roomId
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid room ID")
		}
	}
	
	return nil
}

func (r *GetRoomByHostIDReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get host ID from path parameter
	if hostIdStr := c.Param("hostId"); hostIdStr != "" {
		if hostId, err := strconv.ParseInt(hostIdStr, 10, 64); err == nil {
			r.HostID = hostId
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid host ID")
		}
	}
	
	return nil
}

func (r *GetUserRoomReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	return nil
}

// GetReactionsToTwoUsersReq Bind method
func (r *GetReactionsToTwoUsersReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	
	// Get user IDs from query parameters
	if user1IdStr := c.Query("user1_id"); user1IdStr != "" {
		if user1Id, err := strconv.ParseInt(user1IdStr, 10, 64); err == nil {
			r.User1ID = user1Id
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user1_id")
		}
	}
	
	if user2IdStr := c.Query("user2_id"); user2IdStr != "" {
		if user2Id, err := strconv.ParseInt(user2IdStr, 10, 64); err == nil {
			r.User2ID = user2Id
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user2_id")
		}
	}
	
	// Get pagination from query parameters using the same logic as other endpoints
	r.Page, r.PageSize = app.GetPageInfo(c)
	
	return nil
}

// GetGlobalReactionTimelineReq Bind method
func (r *GetGlobalReactionTimelineReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	
	// Get pagination from query parameters using the same logic as other endpoints
	r.Page, r.PageSize = app.GetPageInfo(c)
	
	return nil
}

// GetUserReactionTimelineReq Bind method
func (r *GetUserReactionTimelineReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	
	// Get user ID from query parameter
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	}
	
	// Get pagination from query parameters using the same logic as other endpoints
	r.Page, r.PageSize = app.GetPageInfo(c)
	
	return nil
}

// GetUserReactionsReq Bind method
func (r *GetUserReactionsReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	
	// Get user ID from query parameter, if not provided use authenticated user ID
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	} else {
		// If no user_id provided, use authenticated user ID
		r.UserID = userId
	}
	
	return nil
}

// GetUserReactionUsersReq Bind method
func (r *GetUserReactionUsersReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get user ID from query parameter, if not provided use authenticated user ID
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	} else {
		// If no user_id provided, use authenticated user ID
		r.UserID = user.ID
	}
	
	// Get reaction type ID from query parameter
	if reactionTypeIdStr := c.Query("reaction_type_id"); reactionTypeIdStr != "" {
		if reactionTypeID, err := strconv.ParseInt(reactionTypeIdStr, 10, 64); err == nil {
			r.ReactionTypeID = reactionTypeID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid reaction_type_id")
		}
	}
	
	// Get limit and offset from query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			r.Limit = limit
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid limit")
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			r.Offset = offset
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid offset")
		}
	}
	
	return nil
}

// GetUserGivenReactionsReq Bind method
func (r *GetUserGivenReactionsReq) Bind(c *gin.Context) mir.Error {
	userId, exist := base.UserIdFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.SimpleInfo = SimpleInfo{
		Uid: userId,
	}
	
	// Get user ID from query parameter, if not provided use authenticated user ID
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	} else {
		// If no user_id provided, use authenticated user ID
		r.UserID = userId
	}
	
	return nil
}

// GetUserGivenReactionUsersReq Bind method
func (r *GetUserGivenReactionUsersReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	
	// Get user ID from query parameter, if not provided use authenticated user ID
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	} else {
		// If no user_id provided, use authenticated user ID
		r.UserID = user.ID
	}
	
	// Get reaction type ID from query parameter
	if reactionTypeIdStr := c.Query("reaction_type_id"); reactionTypeIdStr != "" {
		if reactionTypeID, err := strconv.ParseInt(reactionTypeIdStr, 10, 64); err == nil {
			r.ReactionTypeID = reactionTypeID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid reaction_type_id")
		}
	}
	
	// Get limit and offset from query parameters
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			r.Limit = limit
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid limit")
		}
	}
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			r.Offset = offset
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid offset")
		}
	}
	
	return nil
}

// UserOnlineStatusReq request for checking user online status
type UserOnlineStatusReq struct {
	UserID int64 `json:"user_id" form:"user_id" binding:"required"`
}

// UserOnlineStatusResp response for user online status
type UserOnlineStatusResp struct {
	UserID   int64 `json:"user_id"`
	IsOnline bool  `json:"is_online"`
}

// UserOnlineStatusReq Bind method
func (r *UserOnlineStatusReq) Bind(c *gin.Context) mir.Error {
	// Get user_id from query parameter
	if userIdStr := c.Query("user_id"); userIdStr != "" {
		if userID, err := strconv.ParseInt(userIdStr, 10, 64); err == nil {
			r.UserID = userID
		} else {
			return mir.Errorln(http.StatusBadRequest, "Invalid user_id")
		}
	} else {
		return mir.Errorln(http.StatusBadRequest, "user_id is required")
	}
	return nil
}

// CentrifugoTokenReq request for getting Centrifugo token
type CentrifugoTokenReq struct {
	BaseInfo `json:"-" binding:"-"`
}

// CentrifugoTokenResp response for Centrifugo token
type CentrifugoTokenResp struct {
	Token string `json:"token"`
}

// Bind method for CentrifugoTokenReq
func (r *CentrifugoTokenReq) Bind(c *gin.Context) mir.Error {
	user, exist := base.UserFrom(c)
	if !exist {
		return xerror.UnauthorizedAuthNotExist
	}
	r.BaseInfo = BaseInfo{
		User: user,
	}
	return nil
}

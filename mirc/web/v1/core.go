package v1

import (
	. "github.com/alimy/mir/v4"
	. "github.com/alimy/mir/v4/engine"
	"github.com/rocboss/paopao-ce/internal/model/web"
)

func init() {
	Entry[Core]()
}

// Core 核心服务，需要授权访问
type Core struct {
	Chain `mir:"-"`
	Group `mir:"v1"`

	// SyncSearchIndex 同步索引
	SyncSearchIndex func(Get, web.SyncSearchIndexReq) `mir:"/sync/index"`

	// GetUserInfo 获取当前用户信息
	GetUserInfo func(Get, web.UserInfoReq) web.UserInfoResp `mir:"/user/info"`

	// GetUserOnlineStatus 获取用户在线状态
	GetUserOnlineStatus func(Get, web.UserOnlineStatusReq) web.UserOnlineStatusResp `mir:"/user/online-status"`

	// GetCentrifugoToken 获取Centrifugo连接令牌
	GetCentrifugoToken func(Get, web.CentrifugoTokenReq) web.CentrifugoTokenResp `mir:"/centrifugo/token"`

	// GetMessages 获取消息列表
	GetMessages func(Get, web.GetMessagesReq) web.GetMessagesResp `mir:"/user/messages"`

	// ReadMessage 标记未读消息已读
	ReadMessage func(Post, web.ReadMessageReq) `mir:"/user/message/read"`

	// ReadAllMessage 标记所有未读消息已读
	ReadAllMessage func(Post, web.ReadAllMessageReq) `mir:"/user/message/readall"`

	// SendUserWhisper 发送用户私信
	SendUserWhisper func(Post, web.SendWhisperReq) `mir:"/user/whisper"`

	// GetCollections 获取用户收藏列表
	GetCollections func(Get, web.GetCollectionsReq) web.GetCollectionsResp `mir:"/user/collections"`

	// GetStars 获取用户点赞列表
	GetStars func(Get, web.GetStarsReq) web.GetStarsResp `mir:"/user/stars"`

	// UserPhoneBind 绑定用户手机号
	UserPhoneBind func(Post, web.UserPhoneBindReq) `mir:"/user/phone"`

	// ChangePassword 修改密码
	ChangePassword func(Post, web.ChangePasswordReq) `mir:"/user/password"`

	// ChangeNickname 修改昵称
	ChangeNickname func(Post, web.ChangeNicknameReq) `mir:"/user/nickname"`

	// ChangeAvatar 修改头像
	ChangeAvatar func(Post, web.ChangeAvatarReq) `mir:"/user/avatar"`

	// SetUserCategories 设置用户分类
	SetUserCategories func(Post, web.SetUserCategoriesReq) web.SetUserCategoriesResp `mir:"/user/categories"`

	// SuggestUsers 检索用户
	SuggestUsers func(Get, web.SuggestUsersReq) web.SuggestUsersResp `mir:"/suggest/users"`

	// SuggestTags 检索标签
	SuggestTags func(Get, web.SuggestTagsReq) web.SuggestTagsResp `mir:"/suggest/tags"`

	// TweetStarStatus 获取动态点赞状态
	TweetStarStatus func(Get, web.TweetStarStatusReq) web.TweetStarStatusResp `mir:"/post/star"`

	// TweetCollectionStatus 获取动态收藏状态
	TweetCollectionStatus func(Get, web.TweetCollectionStatusReq) web.TweetCollectionStatusResp `mir:"/post/collection"`

	// Room endpoints
	// ListRooms returns a paginated list of rooms
	ListRooms func(Get, web.RoomListReq) (*web.RoomListResp, mir.Error) `mir:"/rooms"`
	
	// CreateRoom creates a new room
	CreateRoom func(Post, web.CreateRoomReq) (*web.Room, mir.Error) `mir:"/rooms"`
	
	// UpdateRoom updates a room's information
	UpdateRoom func(Put, web.UpdateRoomReq) mir.Error `mir:"/rooms/:id"`
	
	// GetUserRoom returns a user's room
	GetUserRoom func(Get, web.GetUserRoomReq) (*web.Room, mir.Error) `mir:"/rooms/user"`
	
	// GetRoomByID returns a room by its ID
	GetRoomByID func(Get, web.GetRoomByIDReq) (*web.Room, mir.Error) `mir:"/rooms/:id"`
	
	// GetRoomByHostID returns a room by its host ID
	GetRoomByHostID func(Get, web.GetRoomByHostIDReq) (*web.Room, mir.Error) `mir:"/rooms/host/:hostId"`

	// Category endpoints
	// GetAllCategories gets all available categories
	GetAllCategories func(Get) web.CategoryListResp `mir:"/categories"`

	// User Reaction endpoints - These belong in core service, not private service
	// CreateUserReaction 创建用户对用户的反应
	CreateUserReaction func(Post, web.CreateUserReactionReq) (*web.CreateUserReactionResp, mir.Error) `mir:"/user/reaction"`



	// GetUserReactions 获取用户收到的反应统计 (别人对用户的反应)
	GetUserReactionsCounts func(Get, web.GetUserReactionsReq) (*web.GetUserReactionsResp, mir.Error) `mir:"/user/reactions"`

	// GetUserReactionUsers 获取对用户有特定反应的用户列表 (谁对用户有反应)
	GetUserReactionUsers func(Get, web.GetUserReactionUsersReq) (*web.GetUserReactionUsersResp, mir.Error) `mir:"/user/reaction/users"`

	// GetUserGivenReactions 获取用户发出的反应统计 (用户对别人的反应)
	GetUserGivenReactionsCounts func(Get, web.GetUserGivenReactionsReq) (*web.GetUserGivenReactionsResp, mir.Error) `mir:"/user/given-reactions"`

	// GetUserGivenReactionUsers 获取用户对谁有特定反应的列表 (用户对谁有反应)
	GetUserGivenReactionUsers func(Get, web.GetUserGivenReactionUsersReq) (*web.GetUserGivenReactionUsersResp, mir.Error) `mir:"/user/given-reaction/users"`

	// GetReactionsToTwoUsers 获取对两个用户的反应 (带分页，按时间倒序)
	GetReactionsToTwoUsers func(Get, web.GetReactionsToTwoUsersReq) (*web.GetReactionsToTwoUsersResp, mir.Error) `mir:"/user/reactions/to-two-users"`

	// GetGlobalReactionTimeline 获取全局反应时间线 (所有用户的反应)
	GetGlobalReactionTimeline func(Get, web.GetGlobalReactionTimelineReq) (*web.GetGlobalReactionTimelineResp, mir.Error) `mir:"/user/reactions/timeline/global"`

	// GetUserReactionTimeline 获取特定用户的反应时间线 (用户对谁有反应)
	GetUserReactionTimeline func(Get, web.GetUserReactionTimelineReq) (*web.GetUserReactionTimelineResp, mir.Error) `mir:"/user/reactions/timeline/user"`
}

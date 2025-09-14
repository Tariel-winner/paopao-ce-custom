// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"context"
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/alimy/mir/v4"
	"github.com/gin-gonic/gin"
	api "github.com/rocboss/paopao-ce/auto/api/v1"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/model/joint"
	"github.com/rocboss/paopao-ce/internal/model/web"
	"github.com/rocboss/paopao-ce/internal/servants/base"
	"github.com/rocboss/paopao-ce/internal/servants/chain"
	"github.com/rocboss/paopao-ce/pkg/xerror"
	"github.com/sirupsen/logrus"
	"github.com/golang-jwt/jwt/v5"
)

var (
	// _MaxWhisperNumDaily 当日单用户私信总数限制（TODO 配置化、积分兑换等）
	_maxWhisperNumDaily int64 = 200
	_maxCaptchaTimes    int   = 2
)

var (
	_ api.Core = (*coreSrv)(nil)
)

type coreSrv struct {
	api.UnimplementedCoreServant
	*base.DaoServant
	oss            core.ObjectStorageService
	wc             core.WebCache
	messagesExpire int64
	prefixMessages string
}

func (s *coreSrv) Chain() gin.HandlersChain {
	return gin.HandlersChain{chain.JWT()}
}

func (s *coreSrv) SyncSearchIndex(req *web.SyncSearchIndexReq) mir.Error {
	if req.User != nil && req.User.IsAdmin {
		s.PushAllPostToSearch()
	} else {
		logrus.Warnf("sync search index need admin permision user: %#v", req.User)
	}
	return nil
}

func (s *coreSrv) GetUserInfo(req *web.UserInfoReq) (*web.UserInfoResp, mir.Error) {
	user, err := s.Ds.UserProfileByName(req.Username)
	if err != nil {
		logrus.Errorf("coreSrv.GetUserInfo occurs error[1]: %s", err)
		return nil, xerror.UnauthorizedAuthNotExist
	}
	follows, followings, err := s.Ds.GetFollowCount(user.ID)
	if err != nil {
		return nil, web.ErrGetFollowCountFailed
	}
	resp := &web.UserInfoResp{
		Id:          user.ID,
		Nickname:    user.Nickname,
		Username:    user.Username,
		Status:      user.Status,
		Avatar:      user.Avatar,
		Balance:     user.Balance,
		IsAdmin:     user.IsAdmin,
		CreatedOn:   user.CreatedOn,
		Follows:     follows,
		Followings:  followings,
		TweetsCount: user.TweetsCount,
		IsOnline:   s.Ds.IsUserOnline(user.ID), // Add online status
	}
	if user.Phone != "" && len(user.Phone) == 11 {
		resp.Phone = user.Phone[0:3] + "****" + user.Phone[7:]
	}
	return resp, nil
}

func (s *coreSrv) GetMessages(req *web.GetMessagesReq) (res *web.GetMessagesResp, _ mir.Error) {
	limit, offset := req.PageSize, (req.Page-1)*req.PageSize
	// 尝试直接从缓存中获取数据
	key, ok := "", false
	if res, key, ok = s.messagesFromCache(req, limit, offset); ok {
		// logrus.Debugf("coreSrv.GetMessages from cache key:%s", key)
		return
	}
	messages, totalRows, err := s.Ds.GetMessages(req.Uid, req.Style, limit, offset)
	if err != nil {
		logrus.Errorf("Ds.GetMessages err[1]: %s", err)
		return nil, web.ErrGetMessagesFailed
	}
	for _, mf := range messages {
		// TODO: 优化处理这里的user获取逻辑以及错误处理
		if mf.SenderUserID > 0 {
			if user, err := s.Ds.GetUserByID(mf.SenderUserID); err == nil {
				mf.SenderUser = user.Format()
			}
		}
		if mf.Type == ms.MsgTypeWhisper && mf.ReceiverUserID != req.Uid {
			if user, err := s.Ds.GetUserByID(mf.ReceiverUserID); err == nil {
				mf.ReceiverUser = user.Format()
			}
		}
		// 好友申请消息不需要获取其他信息
		if mf.Type == ms.MsgTypeRequestingFriend {
			continue
		}
		if mf.PostID > 0 {
			post, err := s.GetTweetBy(mf.PostID)
			if err == nil {
				mf.Post = post
				if mf.CommentID > 0 {
					comment, err := s.Ds.GetCommentByID(mf.CommentID)
					if err == nil {
						mf.Comment = comment
						if mf.ReplyID > 0 {
							reply, err := s.Ds.GetCommentReplyByID(mf.ReplyID)
							if err == nil {
								mf.Reply = reply
							}
						}
					}
				}
			}
		}
	}
	if err != nil {
		logrus.Errorf("Ds.GetMessages err[2]: %s", err)
		return nil, web.ErrGetMessagesFailed
	}
	if err = s.PrepareMessages(req.Uid, messages); err != nil {
		logrus.Errorf("get messages err[3]: %s", err)
		return nil, web.ErrGetMessagesFailed
	}
	resp := joint.PageRespFrom(messages, req.Page, req.PageSize, totalRows)
	// 缓存处理
	base.OnCacheRespEvent(s.wc, key, resp, s.messagesExpire)
	return &web.GetMessagesResp{
		CachePageResp: joint.CachePageResp{
			Data: resp,
		},
	}, nil
}

func (s *coreSrv) ReadMessage(req *web.ReadMessageReq) mir.Error {
	message, err := s.Ds.GetMessageByID(req.ID)
	if err != nil {
		return web.ErrReadMessageFailed
	}
	if message.ReceiverUserID != req.Uid {
		return web.ErrNoPermission
	}
	if err = s.Ds.ReadMessage(message); err != nil {
		logrus.Errorf("Ds.ReadMessage err: %s", err)
		return web.ErrReadMessageFailed
	}
	// 缓存处理
	onMessageActionEvent(_messageActionRead, req.Uid)
	return nil
}

func (s *coreSrv) ReadAllMessage(req *web.ReadAllMessageReq) mir.Error {
	if err := s.Ds.ReadAllMessage(req.Uid); err != nil {
		logrus.Errorf("coreSrv.Ds.ReadAllMessage err: %s", err)
		return web.ErrReadMessageFailed
	}
	// 缓存处理
	onMessageActionEvent(_messageActionRead, req.Uid)
	return nil
}

func (s *coreSrv) SendUserWhisper(req *web.SendWhisperReq) mir.Error {
	// 不允许发送私信给自己
	if req.Uid == req.UserID {
		return web.ErrNoWhisperToSelf
	}
	// 今日频次限制
	ctx := context.Background()
	if count, _ := s.Redis.GetCountWhisper(ctx, req.Uid); count >= _maxWhisperNumDaily {
		return web.ErrTooManyWhisperNum
	}
	// 创建私信
	_, err := s.Ds.CreateMessage(&ms.Message{
		SenderUserID:   req.Uid,
		ReceiverUserID: req.UserID,
		Type:           ms.MsgTypeWhisper,
		Brief:          "给你发送新私信了",
		Content:        req.Content,
	})
	if err != nil {
		logrus.Errorf("Ds.CreateWhisper err: %s", err)
		return web.ErrSendWhisperFailed
	}
	// 缓存处理, 不需要处理错误
	onMessageActionEvent(_messageActionSendWhisper, req.Uid, req.UserID)
	// 写入当日（自然日）计数缓存
	s.Redis.IncrCountWhisper(ctx, req.Uid)

	return nil
}

func (s *coreSrv) GetCollections(req *web.GetCollectionsReq) (*web.GetCollectionsResp, mir.Error) {
	collections, err := s.Ds.GetUserPostCollections(req.UserId, (req.Page-1)*req.PageSize, req.PageSize)
	if err != nil {
		logrus.Errorf("Ds.GetUserPostCollections err: %s", err)
		return nil, web.ErrGetCollectionsFailed
	}
	totalRows, err := s.Ds.GetUserPostCollectionCount(req.UserId)
	if err != nil {
		logrus.Errorf("Ds.GetUserPostCollectionCount err: %s", err)
		return nil, web.ErrGetCollectionsFailed
	}
	var posts []*ms.Post
	for _, collection := range collections {
		posts = append(posts, collection.Post)
	}
	postsFormated, err := s.Ds.MergePosts(posts)
	if err != nil {
		logrus.Errorf("Ds.MergePosts err: %s", err)
		return nil, web.ErrGetCollectionsFailed
	}
	if err = s.PrepareTweets(req.UserId, postsFormated); err != nil {
		logrus.Errorf("get collections prepare tweets err: %s", err)
		return nil, web.ErrGetCollectionsFailed
	}
	resp := base.PageRespFrom(postsFormated, req.Page, req.PageSize, totalRows)
	return (*web.GetCollectionsResp)(resp), nil
}

func (s *coreSrv) UserPhoneBind(req *web.UserPhoneBindReq) mir.Error {
	// 手机重复性检查
	u, err := s.Ds.GetUserByPhone(req.Phone)
	if err == nil && u.Model != nil && u.ID != 0 && u.ID != req.User.ID {
		return web.ErrExistedUserPhone
	}

	// 如果禁止phone verify 则允许通过任意验证码
	if _enablePhoneVerify {
		c, err := s.Ds.GetLatestPhoneCaptcha(req.Phone)
		if err != nil {
			return web.ErrErrorPhoneCaptcha
		}
		if c.Captcha != req.Captcha {
			return web.ErrErrorPhoneCaptcha
		}
		if c.ExpiredOn < time.Now().Unix() {
			return web.ErrErrorPhoneCaptcha
		}
		if c.UseTimes >= _maxCaptchaTimes {
			return web.ErrMaxPhoneCaptchaUseTimes
		}
		// 更新检测次数
		s.Ds.UsePhoneCaptcha(c)
	}

	// 执行绑定
	user := req.User
	user.Phone = req.Phone
	if err := s.Ds.UpdateUser(user); err != nil {
		// TODO: 优化错误处理逻辑，失败后上面的逻辑也应该回退
		logrus.Errorf("Ds.UpdateUser err: %s", err)
		return xerror.ServerError
	}
	return nil
}

func (s *coreSrv) GetStars(req *web.GetStarsReq) (*web.GetStarsResp, mir.Error) {
	stars, err := s.Ds.GetUserPostStars(req.UserId, req.PageSize, (req.Page-1)*req.PageSize)
	if err != nil {
		logrus.Errorf("Ds.GetUserPostStars err: %s", err)
		return nil, web.ErrGetStarsFailed
	}
	totalRows, err := s.Ds.GetUserPostStarCount(req.UserId)
	if err != nil {
		logrus.Errorf("Ds.GetUserPostStars err: %s", err)
		return nil, web.ErrGetStarsFailed
	}
	var posts []*ms.Post
	for _, star := range stars {
		posts = append(posts, star.Post)
	}
	postsFormated, err := s.Ds.MergePosts(posts)
	if err != nil {
		logrus.Errorf("Ds.MergePosts err: %s", err)
		return nil, web.ErrGetStarsFailed
	}
	resp := base.PageRespFrom(postsFormated, req.Page, req.PageSize, totalRows)
	return (*web.GetStarsResp)(resp), nil
}

func (s *coreSrv) ChangePassword(req *web.ChangePasswordReq) mir.Error {
	// 密码检查
	if err := checkPassword(req.Password); err != nil {
		return err
	}
	// 旧密码校验
	user := req.User
	if !validPassword(user.Password, req.OldPassword, req.User.Salt) {
		return web.ErrErrorOldPassword
	}
	// 更新入库
	user.Password, user.Salt = encryptPasswordAndSalt(req.Password)
	if err := s.Ds.UpdateUser(user); err != nil {
		logrus.Errorf("Ds.UpdateUser err: %s", err)
		return xerror.ServerError
	}
	return nil
}

func (s *coreSrv) SuggestTags(req *web.SuggestTagsReq) (*web.SuggestTagsResp, mir.Error) {
	tags, err := s.Ds.TagsByKeyword(req.Keyword)
	if err != nil {
		logrus.Errorf("Ds.GetTagsByKeyword err: %s", err)
		return nil, xerror.ServerError
	}
	resp := &web.SuggestTagsResp{}
	for _, t := range tags {
		resp.Suggests = append(resp.Suggests, t.Tag)
	}
	return resp, nil
}

func (s *coreSrv) SuggestUsers(req *web.SuggestUsersReq) (*web.SuggestUsersResp, mir.Error) {
	// Get users from user service
	users, err := s.Ds.GetUsersByKeyword(req.Keyword)
	if err != nil {
		logrus.Errorf("Ds.GetUsersByKeyword err: %s", err)
		return nil, xerror.ServerError
	}
	
	resp := &web.SuggestUsersResp{
		Suggests: make([]web.UserProfileWithFollow, 0, len(users)),
	}
	
	// If we have users and an authenticated user, get follow status
	var followStatus map[int64]bool
	if len(users) > 0 && req.User != nil && req.User.Model != nil {
		// Collect user IDs
		var userIds []int64
		for _, user := range users {
			userIds = append(userIds, user.ID)
		}
		
		// Get follow status from relation service
		followStatus, err = s.Ds.IsMyFollow(req.User.Model.ID, userIds...)
		if err != nil {
			logrus.Errorf("Ds.IsMyFollow err: %s", err)
			// Continue without follow status
		}
	}
	
	// Convert from []*cs.UserProfile to []web.UserProfileWithFollow
	for _, user := range users {
		if user != nil {
			profile := web.UserProfileWithFollow{
				ID:          user.ID,
				Nickname:    user.Nickname,
				Username:    user.Username,
				Phone:      user.Phone,
				Status:     user.Status,
				Avatar:     user.Avatar,
				Balance:    user.Balance,
				IsAdmin:    user.IsAdmin,
				CreatedOn:  user.CreatedOn,
				TweetsCount: user.TweetsCount,
				IsOnline:   s.Ds.IsUserOnline(user.ID), // Add online status
			}
			if followStatus != nil {
				profile.IsFollowing = followStatus[user.ID]
			}
			resp.Suggests = append(resp.Suggests, profile)
		}
	}
	
	return resp, nil
}

func (s *coreSrv) ChangeNickname(req *web.ChangeNicknameReq) mir.Error {
	if utf8.RuneCountInString(req.Nickname) < 2 || utf8.RuneCountInString(req.Nickname) > 12 {
		return web.ErrNicknameLengthLimit
	}
	user := req.User
	user.Nickname = req.Nickname
	if err := s.Ds.UpdateUser(user); err != nil {
		logrus.Errorf("Ds.UpdateUser err: %s", err)
		return xerror.ServerError
	}
	// 缓存处理
	onChangeUsernameEvent(user.ID, user.Username)
	return nil
}

func (s *coreSrv) ChangeAvatar(req *web.ChangeAvatarReq) (xerr mir.Error) {
	defer func() {
		if xerr != nil {
			deleteOssObjects(s.oss, []string{req.Avatar})
		}
	}()

	if err := s.Ds.CheckAttachment(req.Avatar); err != nil {
		logrus.Errorf("Ds.CheckAttachment failed: %s", err)
		return xerror.InvalidParams
	}
	if err := s.oss.PersistObject(s.oss.ObjectKey(req.Avatar)); err != nil {
		logrus.Errorf("Ds.ChangeUserAvatar persist object failed: %s", err)
		return xerror.ServerError
	}
	user := req.User
	user.Avatar = req.Avatar
	if err := s.Ds.UpdateUser(user); err != nil {
		logrus.Errorf("Ds.UpdateUser failed: %s", err)
		return xerror.ServerError
	}
	// 缓存处理
	onChangeUsernameEvent(user.ID, user.Username)
	return nil
}

func (s *coreSrv) TweetCollectionStatus(req *web.TweetCollectionStatusReq) (*web.TweetCollectionStatusResp, mir.Error) {
	resp := &web.TweetCollectionStatusResp{
		Status: true,
	}
	if _, err := s.Ds.GetUserPostCollection(req.TweetId, req.Uid); err != nil {
		resp.Status = false
		return resp, nil
	}
	return resp, nil
}

func (s *coreSrv) TweetStarStatus(req *web.TweetStarStatusReq) (*web.TweetStarStatusResp, mir.Error) {
	resp := &web.TweetStarStatusResp{
		Status: true,
	}
	if _, err := s.Ds.GetUserPostStar(req.TweetId, req.Uid); err != nil {
		resp.Status = false
		return resp, nil
	}
	return resp, nil
}

func (s *coreSrv) messagesFromCache(req *web.GetMessagesReq, limit int, offset int) (res *web.GetMessagesResp, key string, ok bool) {
	key = fmt.Sprintf("%s%d:%s:%d:%d", s.prefixMessages, req.Uid, req.Style, limit, offset)
	if data, err := s.wc.Get(key); err == nil {
		ok, res = true, &web.GetMessagesResp{
			CachePageResp: joint.CachePageResp{
				JsonResp: data,
			},
		}
	}
	return
}

// Room methods
func (s *coreSrv) ListRooms(req *web.RoomListReq) (*web.RoomListResp, mir.Error) {
	logrus.WithFields(logrus.Fields{
		"page": req.Page,
		"page_size": req.PageSize,
	}).Info("Listing rooms")
	
	limit, offset := req.PageSize, (req.Page-1)*req.PageSize
	
	// Get user context for category prioritization
	var userID int64 = 0 // Default to no user context
	if req.User != nil {
		userID = req.User.ID
		logrus.WithField("user_id", userID).Debug("Using user context for category prioritization")
	}
	
	// Get paginated online user IDs and their locations using cursor-based pagination (prevents duplication)
	// Convert page/offset to cursor for Redis SCAN
	cursor := uint64(offset) // Use offset as initial cursor
	onlineUserIDs, userLocations, _, err := s.wc.GetOnlineUsersWithCursor(cursor, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get online user IDs from Redis, returning empty room list")
		// Return empty list since we only show rooms with online hosts
		resp := joint.PageRespFrom([]*web.Room{}, req.Page, req.PageSize, 0)
		return &web.RoomListResp{CachePageResp: joint.CachePageResp{Data: resp}}, nil
	}
	
	// Get total count of online users for pagination
	totalOnlineUsers, err := s.wc.GetOnlineUsersCount()
	if err != nil {
		logrus.WithError(err).Error("Failed to get online users count, using fallback")
		totalOnlineUsers = int64(len(onlineUserIDs)) // Fallback to current page size
	}
	
	logrus.WithField("online_user_count", len(onlineUserIDs)).Info("Retrieved online user IDs from Redis")
	
	// Handle edge cases for pagination
	if len(onlineUserIDs) == 0 {
		// No online users - return empty result
		resp := joint.PageRespFrom([]*web.Room{}, req.Page, req.PageSize, 0)
		return &web.RoomListResp{
			CachePageResp: joint.CachePageResp{
				Data: resp,
			},
		}, nil
	}
	
	// Edge case: User requesting beyond available data
	if int64(offset) >= totalOnlineUsers {
		// User went beyond available data, return empty with correct total
		resp := joint.PageRespFrom([]*web.Room{}, req.Page, req.PageSize, totalOnlineUsers)
		return &web.RoomListResp{
			CachePageResp: joint.CachePageResp{
				Data: resp,
			},
		}, nil
	}
	
	// Get rooms filtered by online user IDs
	rooms, _, err := s.Ds.ListRooms(limit, offset, userID, onlineUserIDs)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"limit": limit,
			"offset": offset,
			"user_id": userID,
			"online_user_count": len(onlineUserIDs),
		}).Error("Failed to list rooms by online host IDs")
		return nil, web.ErrGetRoomsFailed
	}


	// Enrich room data with following status for ListRooms
	enrichedRooms := make([]*web.Room, 0, len(rooms))
	for _, room := range rooms {
		if enrichedRoom, err := s.enrichRoomDataWithFollowing(room, userID); err == nil {
			// Add host location if available (we already have it from GetOnlineUsersWithCursor)
			if location, exists := userLocations[room.HostID]; exists {
				enrichedRoom.HostLocation = location
			}
			enrichedRooms = append(enrichedRooms, enrichedRoom)
		} else {
			logrus.WithError(err).WithField("room_id", room.ID).Error("Failed to enrich room data")
		}
	}

	// Create paginated response using totalOnlineUsers for consistent pagination
	resp := joint.PageRespFrom(enrichedRooms, req.Page, req.PageSize, totalOnlineUsers)
	
	logrus.WithFields(logrus.Fields{
		"total": totalOnlineUsers,
		"page": req.Page,
		"page_size": req.PageSize,
		"online_user_count": len(onlineUserIDs),
	}).Info("Successfully listed rooms with Redis filtering")
	
	return &web.RoomListResp{
		CachePageResp: joint.CachePageResp{
			Data: resp,
		},
	}, nil
}

func (s *coreSrv) CreateRoom(req *web.CreateRoomReq) (*web.Room, mir.Error) {
	logrus.WithFields(logrus.Fields{
		"hms_room_id": req.HMSRoomID,
		"topics": req.Topics,
		"categories": req.Categories,
	}).Info("Attempting to create room")

	// Create new room
	room := &ms.Room{
		Model: &ms.Model{},
		HostID:            req.User.ID,
		HMSRoomID:         req.HMSRoomID,
		SpeakerIDs:        []int64{},
		StartTime:         time.Now().Unix(),
		Queue:             &ms.Queue{},
		IsBlockedFromSpace: 0,
		Topics:            req.Topics,
		Categories:        []int64(req.Categories), // Convert from dbr.Int64Array to []int64
	}

	logrus.WithFields(logrus.Fields{
		"room": room,
	}).Info("Creating new room")

	if err := s.Ds.CreateRoom(room); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"room": room,
		}).Error("Failed to create room")
		return nil, web.ErrCreateRoomFailed
	}

	logrus.WithField("room_id", room.ID).Info("Successfully created room")

	return s.enrichRoomData(room)
}

func (s *coreSrv) UpdateRoom(req *web.UpdateRoomReq) mir.Error {
	logrus.WithFields(logrus.Fields{
		"room_id": req.RoomID,
		"hms_room_id": req.HMSRoomID,
		"speaker_ids": req.SpeakerIDs,
		"topics": req.Topics,
		"categories": req.Categories,
		"queue": req.Queue,
		"is_blocked_from_space": req.IsBlockedFromSpace,
	}).Info("Attempting to update room")

	// Verify room ownership
	room, err := s.Ds.GetRoomByID(req.RoomID)
	if err != nil {
		logrus.WithError(err).WithField("room_id", req.RoomID).Error("Room not found")
		return xerror.NotFound
	}
	if room.HostID != req.User.ID {
		logrus.WithFields(logrus.Fields{
			"room_host_id": room.HostID,
			"request_user_id": req.User.ID,
		}).Error("User is not the room host")
		return xerror.UnauthorizedAuthFailed
	}

	// Prepare updates map
	updates := make(map[string]interface{})
	if req.HMSRoomID != "" {
		updates["hms_room_id"] = req.HMSRoomID
	}
	if req.SpeakerIDs != nil {
		// Always include the host in speaker_ids
		speakerIDs := make([]int64, 0, len(req.SpeakerIDs)+1)
		speakerIDs = append(speakerIDs, req.User.ID) // Add host first
		
		// Add other speakers (excluding host if they included it)
		for _, id := range req.SpeakerIDs {
			if id != req.User.ID {
				speakerIDs = append(speakerIDs, id)
			}
		}
		updates["speaker_ids"] = speakerIDs
	}
	if req.Queue != nil {
		updates["queue"] = req.Queue
	}
	if req.IsBlockedFromSpace != nil {
		updates["is_blocked_from_space"] = *req.IsBlockedFromSpace
	}
	if req.Topics != nil {
		updates["topics"] = req.Topics
	}
	if req.StartTime != 0 {
		updates["start_time"] = req.StartTime
	}
	if req.Categories != nil {
		updates["categories"] = []int64(req.Categories) // Convert from dbr.Int64Array to []int64
	}

	logrus.WithFields(logrus.Fields{
		"room_id": req.RoomID,
		"updates": updates,
	}).Info("Prepared updates for room")

	// Update room
	if err := s.Ds.UpdateRoom(req.RoomID, updates); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"room_id": req.RoomID,
			"updates": updates,
		}).Error("Failed to update room")
		return web.ErrUpdateRoomFailed
	}

	logrus.WithField("room_id", req.RoomID).Info("Successfully updated room")
	return nil
}

func (s *coreSrv) GetUserRoom(req *web.GetUserRoomReq) (*web.Room, mir.Error) {
	if req.User == nil || req.User.ID == 0 {
		logrus.Error("GetUserRoom: No user info in request")
		return nil, xerror.UnauthorizedAuthNotExist
	}

	userID := req.User.ID
	logrus.WithField("user_id", userID).Info("Getting room for user")

	// Get room where user is the host
	room, err := s.Ds.GetRoomByHostID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get user room")
		return nil, web.ErrRoomNotFound
	}

	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"room_id": room.ID,
	}).Info("Successfully found user room")

	// Convert to web.Room format and enrich with additional data
	return s.enrichRoomData(room)
}

func (s *coreSrv) GetRoomByID(req *web.GetRoomByIDReq) (*web.Room, mir.Error) {
	logrus.WithField("room_id", req.RoomID).Info("Getting room by ID")
	
	// Get room from database
	room, err := s.Ds.GetRoomByID(req.RoomID)
	if err != nil {
		logrus.WithError(err).WithField("room_id", req.RoomID).Error("Failed to get room")
		return nil, web.ErrRoomNotFound
	}
	
	logrus.WithFields(logrus.Fields{
		"room_id": room.ID,
		"host_id": room.HostID,
	}).Info("Successfully retrieved room")
	
	return s.enrichRoomData(room)
}

func (s *coreSrv) GetRoomByHostID(req *web.GetRoomByHostIDReq) (*web.Room, mir.Error) {
	logrus.WithField("host_id", req.HostID).Info("Getting room by host ID")
	
	room, err := s.Ds.GetRoomByHostID(req.HostID)
	if err != nil {
		logrus.WithError(err).WithField("host_id", req.HostID).Error("Failed to get room by host ID")
		return nil, web.ErrRoomNotFound
	}
	
	logrus.WithFields(logrus.Fields{
		"room_id": room.ID,
		"host_id": req.HostID,
	}).Info("Successfully retrieved room by host ID")
	
	return s.enrichRoomData(room)
}

// Helper function to enrich room data with user information
// For individual room methods: Set online status to false since we don't need real-time online status
func (s *coreSrv) enrichRoomData(room *ms.Room) (*web.Room, mir.Error) {
	webRoom := &web.Room{
		ID:                room.ID,
		HostID:            room.HostID,
		HMSRoomID:         room.HMSRoomID,
		SpeakerIDs:        room.SpeakerIDs,
		StartTime:         room.StartTime,
		CreatedAt:         room.Model.CreatedOn,
		UpdatedAt:         room.Model.ModifiedOn,
		Queue:             web.Queue(*room.Queue),
		IsBlockedFromSpace: room.IsBlockedFromSpace,
		Topics:            room.Topics,
		Categories:        room.Categories, // Include categories in response
	}

	// Get host information
	host, err := s.Ds.GetUserByID(room.HostID)
	if err == nil {
		webRoom.Host = host.Nickname
		webRoom.HostUsername = host.Username
		webRoom.HostImageURL = host.Avatar
		webRoom.IsHostOnline = s.Ds.IsUserOnline(host.ID) // Use real-time online status
	}

	// Get speaker information - use real-time online status for individual room methods
	speakers := make([]web.SpaceParticipant, 0, len(room.SpeakerIDs))
	for _, speakerID := range room.SpeakerIDs {
		if user, err := s.Ds.GetUserByID(speakerID); err == nil {
			speakers = append(speakers, web.SpaceParticipant{
				UserID:   user.ID,
				Username: user.Username,
				Avatar:   user.Avatar,
				IsOnline: s.Ds.IsUserOnline(speakerID), // Use real-time online status
			})
		}
	}
	webRoom.Speakers = speakers

	return webRoom, nil
}

// Helper function to enrich room data with user information including following status
// Optimized for ListRooms: Since Redis filtering only returns online users, we can set all online status to true
func (s *coreSrv) enrichRoomDataWithFollowing(room *ms.Room, userID int64) (*web.Room, mir.Error) {
	webRoom := &web.Room{
		ID:                room.ID,
		HostID:            room.HostID,
		HMSRoomID:         room.HMSRoomID,
		SpeakerIDs:        room.SpeakerIDs,
		StartTime:         room.StartTime,
		CreatedAt:         room.Model.CreatedOn,
		UpdatedAt:         room.Model.ModifiedOn,
		Queue:             web.Queue(*room.Queue),
		IsBlockedFromSpace: room.IsBlockedFromSpace,
		Topics:            room.Topics,
		Categories:        room.Categories, // Include categories in response
	}

	// Get host information
	host, err := s.Ds.GetUserByID(room.HostID)
	if err == nil {
		webRoom.Host = host.Nickname
		webRoom.HostUsername = host.Username
		webRoom.HostImageURL = host.Avatar
		webRoom.IsHostOnline = true // Always true since Redis filtering only returns online users
		
		// Check if authenticated user is following the host
		if userID > 0 {
			isFollowing := s.Ds.IsFollow(userID, host.ID)
			webRoom.IsFollowing = &isFollowing
		}
	}

	// Get speaker information - all speakers are online since we only return rooms with online hosts
	speakers := make([]web.SpaceParticipant, 0, len(room.SpeakerIDs))
	for _, speakerID := range room.SpeakerIDs {
		if user, err := s.Ds.GetUserByID(speakerID); err == nil {
			speakers = append(speakers, web.SpaceParticipant{
				UserID:   user.ID,
				Username: user.Username,
				Avatar:   user.Avatar,
				IsOnline: true, // Always true since Redis filtering only returns online users
			})
		}
	}
	webRoom.Speakers = speakers

	return webRoom, nil
}

// Category methods
func (s *coreSrv) GetAllCategories() web.CategoryListResp {
	logrus.Info("Getting all categories")
	
	categories, err := s.Ds.GetAllCategories()
	if err != nil {
		logrus.WithError(err).Error("Failed to get all categories")
		return web.CategoryListResp{Categories: []*web.Category{}}
	}
	
	// Convert to web.Category format (now using type alias)
	webCategories := make([]*web.Category, 0, len(categories))
	for _, category := range categories {
		webCategories = append(webCategories, &web.Category{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			Icon:        category.Icon,
			Color:       category.Color,
		})
	}
	
	logrus.WithField("category_count", len(webCategories)).Info("Successfully retrieved categories")
	
	return web.CategoryListResp{Categories: webCategories}
}

// SetUserCategories sets categories for a user
func (s *coreSrv) SetUserCategories(req *web.SetUserCategoriesReq) (*web.SetUserCategoriesResp, mir.Error) {
	logrus.WithFields(logrus.Fields{
		"user_id": req.User.ID,
		"category_ids": req.CategoryIDs,
	}).Info("Setting user categories")
	
	// Validate category IDs
	if len(req.CategoryIDs) == 0 {
		return nil, xerror.InvalidParams
	}

	// Update user categories
	if err := s.Ds.SetUserCategories(req.User.ID, []int64(req.CategoryIDs)); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": req.User.ID,
			"category_ids": req.CategoryIDs,
		}).Error("Failed to set user categories")
		return nil, xerror.ServerError
	}

	// Also update room categories for rooms where this user is the host
	// This ensures room categories stay in sync with host preferences
	if err := s.Ds.UpdateCategoriesByHostID(req.User.ID, []int64(req.CategoryIDs)); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": req.User.ID,
			"category_ids": req.CategoryIDs,
		}).Error("Failed to update room categories for host")
		// Don't fail the entire operation if room update fails
		// Just log the error and continue
	} else {
		logrus.WithFields(logrus.Fields{
			"user_id": req.User.ID,
			"category_ids": req.CategoryIDs,
		}).Info("Successfully updated room categories for host")
	}

	logrus.WithFields(logrus.Fields{
		"user_id": req.User.ID,
		"category_ids": req.CategoryIDs,
	}).Info("Successfully set user categories")

	return &web.SetUserCategoriesResp{
		Success: true,
	}, nil
}

// User Reaction Methods - These belong in core service, not private service

// CreateUserReaction creates a user-to-user reaction
func (s *coreSrv) CreateUserReaction(req *web.CreateUserReactionReq) (*web.CreateUserReactionResp, mir.Error) {
	// Check if user is trying to react to themselves
	if req.Uid == req.TargetUserID {
		return nil, xerror.InvalidParams
	}
	
	// Create or update reaction in a single efficient operation
	_, err := s.Ds.CreateUserReaction(req.Uid, req.TargetUserID, req.ReactionTypeID)
	if err != nil {
		return nil, xerror.ServerError
	}
	
	// Get reaction details for response
	reactionName, reactionIcon := s.getReactionDetails(req.ReactionTypeID)
	
	// Send notification for reactions (not for basic 'like')

		// Get the reactor user info for the notification
		reactorUser, err := s.Ds.GetUserByID(req.Uid)
		if err == nil && reactorUser != nil {
					// Create notification message
		onCreateMessageEvent(&ms.Message{
			SenderUserID:   req.Uid,
			ReceiverUserID: req.TargetUserID,
			Type:           ms.MsgTypeUserReaction, // Using dedicated user reaction message type
			Brief:          fmt.Sprintf("%s 给你发送了 %s 反应", reactorUser.Nickname, reactionName),
			Content:        fmt.Sprintf("用户 %s 给你发送了 %s 反应", reactorUser.Nickname, reactionName),
		})
		}
	
	
	return &web.CreateUserReactionResp{
		Status:         true,
		ReactionTypeID: req.ReactionTypeID,
		ReactionName:   reactionName,
		ReactionIcon:   reactionIcon,
	}, nil
}



// GetUserReactions gets reaction counts for a user (reactions received)
func (s *coreSrv) GetUserReactionsCounts(req *web.GetUserReactionsReq) (*web.GetUserReactionsResp, mir.Error) {
	counts, err := s.Ds.GetUserReactionCounts(req.UserID)
	if err != nil {
		return nil, xerror.ServerError
	}
	
	return &web.GetUserReactionsResp{
		ReactionCounts: counts,
	}, nil
}

// GetUserReactionUsers gets users who reacted to a user with a specific reaction type
func (s *coreSrv) GetUserReactionUsers(req *web.GetUserReactionUsersReq) (*web.GetUserReactionUsersResp, mir.Error) {
	users, total, err := s.Ds.GetUserReactionUsers(req.UserID, req.ReactionTypeID, req.Limit, req.Offset)
	if err != nil {
		return nil, xerror.ServerError
	}
	
	// If we have users and an authenticated user, get follow status
	if len(users) > 0 && req.User != nil && req.User.Model != nil {
		// Collect user IDs
		var userIds []int64
		for _, user := range users {
			userIds = append(userIds, user.ID)
		}
		
		// Get follow status from relation service
		followStatus, err := s.Ds.IsMyFollow(req.User.Model.ID, userIds...)
		if err != nil {
			logrus.Errorf("Ds.IsMyFollow err: %s", err)
			// Continue without follow status
		} else {
			// Set follow status for each user
			for _, user := range users {
				if followStatus != nil {
					user.IsFollowing = followStatus[user.ID]
				}
			}
		}
	}
	
	return &web.GetUserReactionUsersResp{
		Users: users,
		Total: total,
	}, nil
}

// GetUserGivenReactions gets reaction counts for a user (reactions given)
func (s *coreSrv) GetUserGivenReactionsCounts(req *web.GetUserGivenReactionsReq) (*web.GetUserGivenReactionsResp, mir.Error) {
	counts, err := s.Ds.GetUserGivenReactionCounts(req.UserID)
	if err != nil {
		return nil, xerror.ServerError
	}
	
	return &web.GetUserGivenReactionsResp{
		ReactionCounts: counts,
	}, nil
}

// GetUserGivenReactionUsers gets users that a user has reacted to with a specific reaction type
func (s *coreSrv) GetUserGivenReactionUsers(req *web.GetUserGivenReactionUsersReq) (*web.GetUserGivenReactionUsersResp, mir.Error) {
	users, total, err := s.Ds.GetUserGivenReactionUsers(req.UserID, req.ReactionTypeID, req.Limit, req.Offset)
	if err != nil {
		return nil, xerror.ServerError
	}
	
	// If we have users and an authenticated user, get follow status
	if len(users) > 0 && req.User != nil && req.User.Model != nil {
		// Collect user IDs
		var userIds []int64
		for _, user := range users {
			userIds = append(userIds, user.ID)
		}
		
		// Get follow status from relation service
		followStatus, err := s.Ds.IsMyFollow(req.User.Model.ID, userIds...)
		if err != nil {
			logrus.Errorf("Ds.IsMyFollow err: %s", err)
			// Continue without follow status
		} else {
			// Set follow status for each user
			for _, user := range users {
				if followStatus != nil {
					user.IsFollowing = followStatus[user.ID]
				}
			}
		}
	}
	
	return &web.GetUserGivenReactionUsersResp{
		Users: users,
		Total: total,
	}, nil
}

// GetReactionsToTwoUsers gets reactions TO two users with pagination
func (s *coreSrv) GetReactionsToTwoUsers(req *web.GetReactionsToTwoUsersReq) (*web.GetReactionsToTwoUsersResp, mir.Error) {
	// Calculate offset from page and page size
	offset := (req.Page - 1) * req.PageSize
	
	// Get reactions TO both users with complete data in one efficient query
	reactions, total, err := s.Ds.GetReactionsToTwoUsers(req.User1ID, req.User2ID, req.PageSize, offset)
	if err != nil {
		logrus.Errorf("Failed to get reactions to two users: %v", err)
		return nil, xerror.ServerError
	}
	
	// Create pagination response using the same pattern as other endpoints
	pager := base.Pager{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalRows: total,
	}
	
	return &web.GetReactionsToTwoUsersResp{
		List:  reactions,
		Pager: pager,
	}, nil
}

// GetGlobalReactionTimeline gets all reactions given by all users in chronological order
func (s *coreSrv) GetGlobalReactionTimeline(req *web.GetGlobalReactionTimelineReq) (*web.GetGlobalReactionTimelineResp, mir.Error) {
	// Calculate offset from page and page size
	offset := (req.Page - 1) * req.PageSize
	
	// Get global reaction timeline
	reactions, total, err := s.Ds.GetGlobalReactionTimeline(req.PageSize, offset)
	if err != nil {
		logrus.Errorf("Failed to get global reaction timeline: %v", err)
		return nil, xerror.ServerError
	}
	
	// Create pagination response
	pager := base.Pager{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalRows: total,
	}
	
	return &web.GetGlobalReactionTimelineResp{
		List:  reactions,
		Pager: pager,
	}, nil
}

// GetUserReactionTimeline gets all reactions given by a specific user in chronological order
func (s *coreSrv) GetUserReactionTimeline(req *web.GetUserReactionTimelineReq) (*web.GetUserReactionTimelineResp, mir.Error) {
	// Calculate offset from page and page size
	offset := (req.Page - 1) * req.PageSize
	
	// Get user reaction timeline
	reactions, total, err := s.Ds.GetUserReactionTimeline(req.UserID, req.PageSize, offset)
	if err != nil {
		logrus.Errorf("Failed to get user reaction timeline: %v", err)
		return nil, xerror.ServerError
	}
	
	// Create pagination response
	pager := base.Pager{
		Page:      req.Page,
		PageSize:  req.PageSize,
		TotalRows: total,
	}
	
	return &web.GetUserReactionTimelineResp{
		List:  reactions,
		Pager: pager,
	}, nil
}

// GetUserOnlineStatus returns only the online status of a specific user
func (s *coreSrv) GetUserOnlineStatus(req *web.UserOnlineStatusReq) (*web.UserOnlineStatusResp, mir.Error) {
	// Get online status from Redis cache
	isOnline := s.Ds.IsUserOnline(req.UserID)
	
	return &web.UserOnlineStatusResp{
		UserID:   req.UserID,
		IsOnline: isOnline,
	}, nil
}

// GetCentrifugoToken generates a JWT token for Centrifugo connection
func (s *coreSrv) GetCentrifugoToken(req *web.CentrifugoTokenReq) (*web.CentrifugoTokenResp, mir.Error) {
	// Get user ID from authenticated user
	userID := req.User.ID
	
	// Generate JWT token for Centrifugo (no expiration)
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", userID),
		// No "exp" claim = token never expires
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("paopao-centrifugo-secret-key-2024"))
	if err != nil {
		return nil, xerror.ServerError
	}
	
	return &web.CentrifugoTokenResp{
		Token: tokenString,
	}, nil
}

// Helper functions for user reactions

// getReactionDetails returns reaction name and icon for a given reaction type ID
func (s *coreSrv) getReactionDetails(reactionTypeID int64) (name, icon string) {
	return cs.GetReactionName(reactionTypeID), cs.GetReactionIcon(reactionTypeID)
}

// UpdateUserLocation implements LocationService interface - directly updates Redis
func (s *coreSrv) UpdateUserLocation(userID int64, locationData *web.LocationData) error {
	if locationData == nil {
		return fmt.Errorf("location data cannot be nil")
	}
	
	// Format location data as string (country|city or just country)
	locationStr := locationData.Country
	if locationData.City != "" {
		locationStr = locationData.Country + "|" + locationData.City
	}
	
	// Set location in Redis with 24 hour expiration (same as online status)
	expire := int64(24 * 60 * 60) // 24 hours in seconds
	err := s.wc.SetUserLocation(userID, locationStr, expire)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user location in Redis")
		return err
	}
	
	logrus.WithFields(logrus.Fields{
		"user_id": userID,
		"location": locationStr,
		"country": locationData.Country,
		"city": locationData.City,
	}).Info("Successfully updated user location in Redis")
	
	return nil
}

// UpdateUserLocationAPI implements the API endpoint for updating user location
func (s *coreSrv) UpdateUserLocationAPI(req *web.UpdateUserLocationReq) (*web.UpdateUserLocationResp, mir.Error) {
	if req.User == nil {
		return nil, web.ErrGetUserFailed
	}
	
	userID := req.User.ID
	if userID <= 0 {
		return nil, web.ErrGetUserFailed
	}
	
	// Update location in Redis using LocationService
	err := s.UpdateUserLocation(userID, req.LocationData)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user location")
		return nil, web.ErrUpdateUserLocationFailed
	}
	
	return &web.UpdateUserLocationResp{
		Success: true,
		Message: "Location updated successfully",
	}, nil
}

func newCoreSrv(s *base.DaoServant, oss core.ObjectStorageService, wc core.WebCache) api.Core {
	cs := conf.CacheSetting
	return &coreSrv{
		DaoServant:     s,
		oss:            oss,
		wc:             wc,
		messagesExpire: cs.MessagesExpire,
		prefixMessages: conf.PrefixMessages,
	}
}

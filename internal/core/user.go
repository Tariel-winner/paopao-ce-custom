// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
)

// UserManageService 用户管理服务
type UserManageService interface {
	GetUserByID(id int64) (*ms.User, error)
	GetUserByUsername(username string) (*ms.User, error)
	GetUserByPhone(phone string) (*ms.User, error)
	GetUsersByIDs(ids []int64) ([]*ms.User, error)
	GetUsersByKeyword(keyword string) ([]*cs.UserProfile, error)
	UserProfileByName(username string) (*cs.UserProfile, error)
	CreateUser(user *ms.User) (*ms.User, error)
	UpdateUser(user *ms.User) error
	SetUserCategories(userID int64, categoryIDs []int64) error
	GetRegisterUserCount() (int64, error)
	
	// User Reaction Methods - User-to-user reactions belong in user service
	CreateUserReaction(reactorUserID, targetUserID, reactionTypeID int64) (*ms.UserReaction, error)
	GetUserReactionCounts(targetUserID int64) (map[int64]int64, error)
	GetUserGivenReactionCounts(reactorUserID int64) (map[int64]int64, error)
	GetUserReactionUsers(targetUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserFormated, int64, error)
	GetUserGivenReactionUsers(reactorUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserFormated, int64, error)
	GetReactionsToTwoUsers(user1ID, user2ID int64, limit, offset int) ([]*ms.UserReactionWithUser, int64, error)
	
	// Search and discovery methods
	SearchUserReactions(reactorUserID, targetUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserReactionWithUser, int64, error)
	
	// Global and user-specific reaction timelines
	GetGlobalReactionTimeline(limit, offset int) ([]*cs.UserReactionWithBothUsers, int64, error)
	GetUserReactionTimeline(userID int64, limit, offset int) ([]*cs.UserReactionWithBothUsers, int64, error)
}

// ContactManageService 联系人管理服务
type ContactManageService interface {
	RequestingFriend(userId int64, friendId int64, greetings string) error
	AddFriend(userId int64, friendId int64) error
	RejectFriend(userId int64, friendId int64) error
	DeleteFriend(userId int64, friendId int64) error
	GetContacts(userId int64, offset int, limit int) (*ms.ContactList, error)
	IsFriend(userID int64, friendID int64) bool
	
	// New methods for iPhone contact integration
	UploadPhoneContacts(userID int64, contacts []cs.PhoneContact) (int64, int64, error)
	MatchPhoneContacts(userID int64) (int64, error)
}

// DeviceManageService 设备管理服务
type DeviceManageService interface {
	RegisterDevice(userID int64, deviceToken, platform, deviceID, deviceName string) error
	UpdateDeviceToken(deviceID, deviceToken string) error
	GetUserDevices(userID int64) ([]*cs.UserDevice, error)
}

// FollowingManageService 关注管理服务
type FollowingManageService interface {
	FollowUser(userId int64, followId int64) error
	UnfollowUser(userId int64, followId int64) error
	ListFollows(userId int64, limit, offset int) (*ms.ContactList, error)
	ListFollowings(userId int64, limit, offset int) (*ms.ContactList, error)
	GetFollowCount(userId int64) (int64, int64, error)
	IsFollow(userId int64, followId int64) bool
}

// UserRelationService 用户关系服务
type UserRelationService interface {
	MyFriendIds(userId int64) ([]int64, error)
	MyFollowIds(userId int64) ([]int64, error)
	IsMyFriend(userId int64, friendIds ...int64) (map[int64]bool, error)
	IsMyFollow(userId int64, followIds ...int64) (map[int64]bool, error)
}

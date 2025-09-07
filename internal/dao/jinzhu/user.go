// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"gorm.io/gorm"
)

// ReactionWithUserData represents the result of the JOIN query
type ReactionWithUserData struct {
	ReactionID     int64  `json:"reaction_id"`
	TargetUserID   int64  `json:"target_user_id"`
	ReactionTypeID int64  `json:"reaction_type_id"`
	CreatedOn      int64  `json:"created_on"`
	UserID         int64  `json:"user_id"`
	Nickname       string `json:"nickname"`
	Username       string `json:"username"`
	Avatar         string `json:"avatar"`
	IsAdmin        bool   `json:"is_admin"`
	ReactionName   string `json:"reaction_name"`
	ReactionIcon   string `json:"reaction_icon"`
}

var (
	_ core.UserManageService = (*userManageSrv)(nil)
)

type userManageSrv struct {
	db  *gorm.DB
	ums core.UserMetricServantA

	_userProfileJoins    string
	_userProfileWhere    string
	_userProfileColoumns []string
}

type userRelationSrv struct {
	db *gorm.DB
}

func newUserManageService(db *gorm.DB, ums core.UserMetricServantA) core.UserManageService {
	return &userManageSrv{
		db:                db,
		ums:               ums,
		_userProfileJoins: fmt.Sprintf("LEFT JOIN %s m ON %s.id=m.user_id", _userMetric_, _user_),
		_userProfileWhere: fmt.Sprintf("%s.username=? AND %s.is_del=0", _user_, _user_),
		_userProfileColoumns: []string{
			fmt.Sprintf("%s.id", _user_),
			fmt.Sprintf("%s.username", _user_),
			fmt.Sprintf("%s.nickname", _user_),
			fmt.Sprintf("%s.phone", _user_),
			fmt.Sprintf("%s.status", _user_),
			fmt.Sprintf("%s.avatar", _user_),
			fmt.Sprintf("%s.balance", _user_),
			fmt.Sprintf("%s.is_admin", _user_),
			fmt.Sprintf("%s.created_on", _user_),
			fmt.Sprintf("%s.categories", _user_),
			"m.tweets_count",
		},
	}
}

func newUserRelationService(db *gorm.DB) core.UserRelationService {
	return &userRelationSrv{
		db: db,
	}
}

func (s *userManageSrv) GetUserByID(id int64) (*ms.User, error) {
	user := &dbr.User{
		Model: &dbr.Model{
			ID: id,
		},
	}
	return user.Get(s.db)
}

func (s *userManageSrv) GetUserByUsername(username string) (*ms.User, error) {
	user := &dbr.User{
		Username: username,
	}
	return user.Get(s.db)
}

func (s *userManageSrv) UserProfileByName(username string) (res *cs.UserProfile, err error) {
	err = s.db.Table(_user_).Joins(s._userProfileJoins).
		Where(s._userProfileWhere, username).
		Select(s._userProfileColoumns).
		First(&res).Error
	
		if err != nil {
		return nil, err
	}
	
	// Get reaction counts for this user
	reactionCounts, err := s.GetUserReactionCounts(res.ID)
	if err != nil {
		// Don't fail the entire request if reaction counts fail
		// Just log the error and continue with empty reaction counts
		logrus.Warnf("Failed to get reaction counts for user %s: %v", username, err)
		res.ReactionCounts = make(map[int64]int64)
	} else {
		res.ReactionCounts = reactionCounts
	}
	
	// Ensure ReactionCounts is never nil for consistent JSON response
	if res.ReactionCounts == nil {
		res.ReactionCounts = make(map[int64]int64)
	}
	
	return res, nil
}

func (s *userManageSrv) GetUserByPhone(phone string) (*ms.User, error) {
	user := &dbr.User{
		Phone: phone,
	}
	return user.Get(s.db)
}

func (s *userManageSrv) GetUsersByIDs(ids []int64) ([]*ms.User, error) {
	user := &dbr.User{}
	return user.List(s.db, &dbr.ConditionsT{
		"id IN ?": ids,
	}, 0, 0)
}

func (s *userManageSrv) GetUsersByKeyword(keyword string) ([]*cs.UserProfile, error) {
    logrus.Debugf("GetUsersByKeyword called with keyword: %s", keyword)
    logrus.Debugf("Using joins: %s", s._userProfileJoins)
    logrus.Debugf("Using columns: %v", s._userProfileColoumns)
    
    if keyword == "" {
        return nil, nil
    }
    
    var profiles []*cs.UserProfile
    query := s.db.Table(_user_).
        Select(s._userProfileColoumns).
        Joins(s._userProfileJoins).
        Where(fmt.Sprintf("%s.username LIKE ? AND %s.is_del=0", _user_, _user_), keyword+"%").
        Order(fmt.Sprintf("%s.id ASC", _user_)).
        Limit(6)
    
    // Debug the SQL query
    sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
        return tx.Find(&profiles)
    })
    logrus.Debugf("SQL Query: %s", sql)
    
    err := query.Find(&profiles).Error
    if err != nil {
        logrus.Errorf("GetUsersByKeyword query error: %v", err)
        return nil, err
    }
    
    logrus.Debugf("GetUsersByKeyword found %d users", len(profiles))
    for i, p := range profiles {
        logrus.Debugf("User %d: %+v", i, p)
    }
    
    return profiles, nil
}

func (s *userManageSrv) CreateUser(user *dbr.User) (res *ms.User, err error) {
	if res, err = user.Create(s.db); err == nil {
		// 宽松处理错误
		s.ums.AddUserMetric(res.ID)
	}
	return
}

func (s *userManageSrv) UpdateUser(user *ms.User) error {
	return user.Update(s.db)
}

func (s *userManageSrv) SetUserCategories(userID int64, categoryIDs []int64) error {
	// Call DBR layer to update user categories
	user := &dbr.User{}
	return user.SetCategories(s.db, userID, categoryIDs)
}

func (s *userManageSrv) GetRegisterUserCount() (res int64, err error) {
	err = s.db.Model(&dbr.User{}).Count(&res).Error
	return
}

// User Reaction Methods - These belong in user service, not tweet service
func (s *userManageSrv) CreateUserReaction(reactorUserID, targetUserID, reactionTypeID int64) (*ms.UserReaction, error) {
	reaction := &dbr.UserReaction{
		ReactorUserID:  reactorUserID,
		TargetUserID:   targetUserID,
		ReactionTypeID: reactionTypeID,
	}
	return reaction.CreateOrUpdateUserReaction(s.db)
}



func (s *userManageSrv) GetUserReactionCounts(targetUserID int64) (map[int64]int64, error) {
	reaction := &dbr.UserReaction{}
	return reaction.GetUserReactionCounts(s.db, targetUserID)
}

func (s *userManageSrv) GetUserGivenReactionCounts(reactorUserID int64) (map[int64]int64, error) {
	reaction := &dbr.UserReaction{}
	return reaction.GetUserGivenReactionCounts(s.db, reactorUserID)
}

func (s *userManageSrv) GetUserReactionUsers(targetUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserFormated, int64, error) {
	reaction := &dbr.UserReaction{}
	return reaction.GetUserReactionUsers(s.db, targetUserID, reactionTypeID, limit, offset)
}

// GetReactionsToTwoUsers gets reactions TO two users with pagination and time ordering
func (s *userManageSrv) GetReactionsToTwoUsers(user1ID, user2ID int64, limit, offset int) ([]*ms.UserReactionWithUser, int64, error) {
	// Use raw query from DBR layer for better performance
	reaction := &dbr.UserReaction{}
	rawReactions, err := reaction.GetReactionsWithUserData(s.db, user1ID, user2ID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	// Get total count using raw query
	total, err := reaction.CountReactionsToUsers(s.db, user1ID, user2ID)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert directly to cs.UserReactionWithUser format using raw data
	reactionList := make([]*ms.UserReactionWithUser, 0, len(rawReactions))
	for _, raw := range rawReactions {
		reactionList = append(reactionList, &ms.UserReactionWithUser{
			ReactionID:     raw.ReactionID,
			ReactorUser: &ms.UserFormated{
				ID:       raw.UserID,
				Nickname: raw.Nickname,
				Username: raw.Username,
				Avatar:   raw.Avatar,
				IsAdmin:  raw.IsAdmin,
			},
			TargetUserID:   raw.TargetUserID,
			ReactionTypeID: raw.ReactionTypeID,
			ReactionName:   raw.ReactionName,
			ReactionIcon:   raw.ReactionIcon,
			CreatedOn:      raw.CreatedOn,
		})
	}
	
	return reactionList, total, nil
}



func (s *userManageSrv) GetUserGivenReactionUsers(reactorUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserFormated, int64, error) {
	reaction := &dbr.UserReaction{}
	return reaction.GetUserGivenReactionUsers(s.db, reactorUserID, reactionTypeID, limit, offset)
}

func (s *userRelationSrv) MyFriendIds(userId int64) (res []int64, err error) {
	err = s.db.Table(_contact_).Where("user_id=? AND status=2 AND is_del=0", userId).Select("friend_id").Find(&res).Error
	return
}

func (s *userRelationSrv) MyFollowIds(userId int64) (res []int64, err error) {
	err = s.db.Table(_following_).Where("user_id=? AND is_del=0", userId).Select("follow_id").Find(&res).Error
	return
}

func (s *userRelationSrv) IsMyFriend(userId int64, friendIds ...int64) (map[int64]bool, error) {
	size := len(friendIds)
	res := make(map[int64]bool, size)
	if size == 0 {
		return res, nil
	}
	myFriendIds, err := s.MyFriendIds(userId)
	if err != nil {
		return nil, err
	}
	for _, friendId := range friendIds {
		res[friendId] = false
		for _, myFriendId := range myFriendIds {
			if friendId == myFriendId {
				res[friendId] = true
				break
			}
		}
	}
	return res, nil
}

func (s *userRelationSrv) IsMyFollow(userId int64, followIds ...int64) (map[int64]bool, error) {
	size := len(followIds)
	res := make(map[int64]bool, size)
	if size == 0 {
		return res, nil
	}
	myFollowIds, err := s.MyFollowIds(userId)
	if err != nil {
		return nil, err
	}
	for _, followId := range followIds {
		res[followId] = false
		for _, myFollowId := range myFollowIds {
			if followId == myFollowId {
				res[followId] = true
				break
			}
		}
	}
	return res, nil
}

func (s *userManageSrv) SearchUserReactions(reactorUserID, targetUserID, reactionTypeID int64, limit, offset int) ([]*ms.UserReactionWithUser, int64, error) {
	reaction := &dbr.UserReaction{}
	rawReactions, total, err := reaction.SearchUserReactions(s.db, reactorUserID, targetUserID, reactionTypeID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert to cs.UserReactionWithUser format
	reactionList := make([]*ms.UserReactionWithUser, 0, len(rawReactions))
	for _, raw := range rawReactions {
		reactionList = append(reactionList, &ms.UserReactionWithUser{
			ReactionID:     raw.ReactionID,
			ReactorUser: &ms.UserFormated{
				ID:       raw.UserID,
				Nickname: raw.Nickname,
				Username: raw.Username,
				Avatar:   raw.Avatar,
				IsAdmin:  raw.IsAdmin,
			},
			TargetUserID:   raw.TargetUserID,
			ReactionTypeID: raw.ReactionTypeID,
			ReactionName:   raw.ReactionName,
			ReactionIcon:   raw.ReactionIcon,
			CreatedOn:      raw.CreatedOn,
		})
	}
	
	return reactionList, total, nil
}

func (s *userManageSrv) GetGlobalReactionTimeline(limit, offset int) ([]*cs.UserReactionWithBothUsers, int64, error) {
	reaction := &dbr.UserReaction{}
	rawReactions, total, err := reaction.GetGlobalReactionTimeline(s.db, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert to cs.UserReactionWithBothUsers format
	reactionList := make([]*cs.UserReactionWithBothUsers, 0, len(rawReactions))
	for _, raw := range rawReactions {
		reactionList = append(reactionList, &cs.UserReactionWithBothUsers{
			ReactionID:     raw.ReactionID,
			ReactorUser: &cs.UserFormated{
				ID:       raw.ReactorUserID,
				Nickname: raw.ReactorNickname,
				Username: raw.ReactorUsername,
				Avatar:   raw.ReactorAvatar,
				IsAdmin:  raw.ReactorIsAdmin,
			},
			TargetUser: &cs.UserFormated{
				ID:       raw.TargetUserID,
				Nickname: raw.TargetNickname,
				Username: raw.TargetUsername,
				Avatar:   raw.TargetAvatar,
				IsAdmin:  raw.TargetIsAdmin,
			},
			ReactionTypeID: raw.ReactionTypeID,
			ReactionName:   raw.ReactionName,
			ReactionIcon:   raw.ReactionIcon,
			CreatedOn:      raw.CreatedOn,
		})
	}
	
	return reactionList, total, nil
}

func (s *userManageSrv) GetUserReactionTimeline(userID int64, limit, offset int) ([]*cs.UserReactionWithBothUsers, int64, error) {
	reaction := &dbr.UserReaction{}
	rawReactions, total, err := reaction.GetUserReactionTimeline(s.db, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	// Convert to cs.UserReactionWithBothUsers format
	reactionList := make([]*cs.UserReactionWithBothUsers, 0, len(rawReactions))
	for _, raw := range rawReactions {
		reactionList = append(reactionList, &cs.UserReactionWithBothUsers{
			ReactionID:     raw.ReactionID,
			ReactorUser: &cs.UserFormated{
				ID:       raw.ReactorUserID,
				Nickname: raw.ReactorNickname,
				Username: raw.ReactorUsername,
				Avatar:   raw.ReactorAvatar,
				IsAdmin:  raw.ReactorIsAdmin,
			},
			TargetUser: &cs.UserFormated{
				ID:       raw.TargetUserID,
				Nickname: raw.TargetNickname,
				Username: raw.TargetUsername,
				Avatar:   raw.TargetAvatar,
				IsAdmin:  raw.TargetIsAdmin,
			},
			ReactionTypeID: raw.ReactionTypeID,
			ReactionName:   raw.ReactionName,
			ReactionIcon:   raw.ReactionIcon,
			CreatedOn:      raw.CreatedOn,
		})
	}
	
	return reactionList, total, nil
}

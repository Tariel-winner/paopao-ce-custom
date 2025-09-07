// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"strings"
	"time"

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

// ReactionWithBothUsersData represents the result of JOIN query with both reactor and target user data
type ReactionWithBothUsersData struct {
	ReactionID     int64  `json:"reaction_id"`
	TargetUserID   int64  `json:"target_user_id"`
	ReactionTypeID int64  `json:"reaction_type_id"`
	CreatedOn      int64  `json:"created_on"`
	
	// Reactor user data (user who gave the reaction)
	ReactorUserID   int64  `json:"reactor_user_id"`
	ReactorNickname string `json:"reactor_nickname"`
	ReactorUsername string `json:"reactor_username"`
	ReactorAvatar   string `json:"reactor_avatar"`
	ReactorIsAdmin  bool   `json:"reactor_is_admin"`
	
	// Target user data (user who received the reaction)
	TargetNickname string `json:"target_nickname"`
	TargetUsername string `json:"target_username"`
	TargetAvatar   string `json:"target_avatar"`
	TargetIsAdmin  bool   `json:"target_is_admin"`
	
	// Reaction data
	ReactionName string `json:"reaction_name"`
	ReactionIcon string `json:"reaction_icon"`
}

type UserReaction struct {
	*Model
	ReactorUserID  int64 `json:"reactor_user_id"`  // User who is reacting
	TargetUserID   int64 `json:"target_user_id"`   // User being reacted to
	ReactionTypeID int64 `json:"reaction_type_id"` // Type of reaction
}

// TableName specifies the table name for UserReaction
func (UserReaction) TableName() string {
	return "p_user_reactions"
}

func (u *UserReaction) Get(db *gorm.DB) (*UserReaction, error) {
	var reaction UserReaction
	
	if u.Model != nil && u.ID > 0 {
		db = db.Where("id = ? AND is_del = ?", u.ID, 0)
	}
	if u.ReactorUserID > 0 {
		db = db.Where("reactor_user_id = ?", u.ReactorUserID)
	}
	if u.TargetUserID > 0 {
		db = db.Where("target_user_id = ?", u.TargetUserID)
	}
	
	if err := db.First(&reaction).Error; err != nil {
		return nil, err
	}
	return &reaction, nil
}

func (u *UserReaction) Create(db *gorm.DB) (*UserReaction, error) {
	err := db.Create(&u).Error
	return u, err
}

// CreateOrUpdateUserReaction creates a new reaction or updates existing one in a single operation
// This is more efficient than checking existence first
func (u *UserReaction) CreateOrUpdateUserReaction(db *gorm.DB) (*UserReaction, error) {
	// Use ON CONFLICT for PostgreSQL
	// This assumes there's a unique constraint on (reactor_user_id, target_user_id)
	
	// For PostgreSQL
	query := `
		INSERT INTO p_user_reactions (reactor_user_id, target_user_id, reaction_type_id, created_on, modified_on, is_del)
		VALUES ($1, $2, $3, $4, $4, 0)
		ON CONFLICT (reactor_user_id, target_user_id) 
		DO UPDATE SET 
			reaction_type_id = EXCLUDED.reaction_type_id,
			modified_on = EXCLUDED.modified_on
	`
	
	now := time.Now().Unix()
	err := db.Exec(query, u.ReactorUserID, u.TargetUserID, u.ReactionTypeID, now).Error
	if err != nil {
		return nil, err
	}
	
	// Get the created/updated record
	return u.Get(db)
}

func (u *UserReaction) Update(db *gorm.DB) error {
	return db.Model(u).Where("id = ? AND is_del = ?", u.ID, 0).Updates(map[string]interface{}{
		"reaction_type_id": u.ReactionTypeID,
		"modified_on":      time.Now().Unix(),
	}).Error
}

func (u *UserReaction) Delete(db *gorm.DB) error {
	return db.Model(&UserReaction{}).Where("id = ? AND is_del = ?", u.Model.ID, 0).Updates(map[string]any{
		"deleted_on": time.Now().Unix(),
		"is_del":     1,
	}).Error
}

func (u *UserReaction) List(db *gorm.DB, conditions *ConditionsT, offset, limit int) ([]*UserReaction, error) {
	var reactions []*UserReaction
	
	if conditions != nil {
		for key, value := range *conditions {
			db = db.Where(key, value)
		}
	}
	
	if limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	
	err := db.Where("is_del = ?", 0).Find(&reactions).Error
	return reactions, err
}

func (u *UserReaction) Count(db *gorm.DB, conditions *ConditionsT) (int64, error) {
	var count int64
	
	if conditions != nil {
		for key, value := range *conditions {
			db = db.Where(key, value)
		}
	}
	
	err := db.Model(&UserReaction{}).Where("is_del = ?", 0).Count(&count).Error
	return count, err
}

// GetReactionsWithUserData gets reactions with user data in one raw query
func (u *UserReaction) GetReactionsWithUserData(db *gorm.DB, user1ID, user2ID int64, limit, offset int) ([]*ReactionWithUserData, error) {
	var reactions []*ReactionWithUserData
	
	query := `
		SELECT 
			ur.id as reaction_id,
			ur.target_user_id,
			ur.reaction_type_id,
			ur.created_on,
			u.id as user_id,
			u.nickname,
			u.username,
			u.avatar,
			u.is_admin,
			r.name as reaction_name,
			r.icon as reaction_icon
		FROM p_user_reactions ur
		JOIN p_user u ON ur.reactor_user_id = u.id
		JOIN p_reactions r ON ur.reaction_type_id = r.id
		WHERE (ur.target_user_id = ? OR ur.target_user_id = ?) 
		  AND ur.is_del = 0
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	err := db.Raw(query, user1ID, user2ID, limit, offset).Scan(&reactions).Error
	return reactions, err
}

// CountReactionsToUsers counts reactions to two users
func (u *UserReaction) CountReactionsToUsers(db *gorm.DB, user1ID, user2ID int64) (int64, error) {
	var count int64
	
	query := `
		SELECT COUNT(*) 
		FROM p_user_reactions ur
		WHERE (ur.target_user_id = ? OR ur.target_user_id = ?) 
		  AND ur.is_del = 0
	`
	
	err := db.Raw(query, user1ID, user2ID).Scan(&count).Error
	return count, err
}

// GetUserReactionCounts returns reaction counts received by a user grouped by reaction type
func (u *UserReaction) GetUserReactionCounts(db *gorm.DB, targetUserID int64) (map[int64]int64, error) {
	var results []struct {
		ReactionTypeID int64 `json:"reaction_type_id"`
		Count          int64 `json:"count"`
	}
	
	query := `
		SELECT reaction_type_id, COUNT(*) as count
		FROM p_user_reactions 
		WHERE target_user_id = ? AND is_del = 0
		GROUP BY reaction_type_id
	`
	
	err := db.Raw(query, targetUserID).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	counts := make(map[int64]int64)
	for _, result := range results {
		counts[result.ReactionTypeID] = result.Count
	}
	
	return counts, nil
}

// GetUserGivenReactionCounts returns reaction counts given by a user grouped by reaction type
func (u *UserReaction) GetUserGivenReactionCounts(db *gorm.DB, reactorUserID int64) (map[int64]int64, error) {
	var results []struct {
		ReactionTypeID int64 `json:"reaction_type_id"`
		Count          int64 `json:"count"`
	}
	
	query := `
		SELECT reaction_type_id, COUNT(*) as count
		FROM p_user_reactions 
		WHERE reactor_user_id = ? AND is_del = 0
		GROUP BY reaction_type_id
	`
	
	err := db.Raw(query, reactorUserID).Scan(&results).Error
	if err != nil {
		return nil, err
	}
	
	counts := make(map[int64]int64)
	for _, result := range results {
		counts[result.ReactionTypeID] = result.Count
	}
	
	return counts, nil
}

// GetUserReactionUsers returns users who reacted to a specific user with a specific reaction type
func (u *UserReaction) GetUserReactionUsers(db *gorm.DB, targetUserID, reactionTypeID int64, limit, offset int) ([]*UserFormated, int64, error) {
	var users []*UserFormated
	var total int64
	
	// Get total count first
	countQuery := `
		SELECT COUNT(DISTINCT ur.reactor_user_id)
		FROM p_user_reactions ur
		WHERE ur.target_user_id = ? AND ur.reaction_type_id = ? AND ur.is_del = 0
	`
	err := db.Raw(countQuery, targetUserID, reactionTypeID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get users with JOIN in single query
	usersQuery := `
		SELECT DISTINCT
			u.id, u.nickname, u.username, u.avatar, u.status, u.is_admin, u.created_on, ur.created_on as reaction_created_on
		FROM p_user_reactions ur
		JOIN p_user u ON ur.reactor_user_id = u.id
		WHERE ur.target_user_id = ? 
		  AND ur.reaction_type_id = ? 
		  AND ur.is_del = 0 
		  AND u.is_del = 0
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	err = db.Raw(usersQuery, targetUserID, reactionTypeID, limit, offset).Scan(&users).Error
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
}

// GetUserGivenReactionUsers returns users that a specific user has reacted to with a specific reaction type
func (u *UserReaction) GetUserGivenReactionUsers(db *gorm.DB, reactorUserID, reactionTypeID int64, limit, offset int) ([]*UserFormated, int64, error) {
	var users []*UserFormated
	var total int64
	
	// Get total count first
	countQuery := `
		SELECT COUNT(DISTINCT ur.target_user_id)
		FROM p_user_reactions ur
		WHERE ur.reactor_user_id = ? AND ur.reaction_type_id = ? AND ur.is_del = 0
	`
	err := db.Raw(countQuery, reactorUserID, reactionTypeID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get users with JOIN in single query
	usersQuery := `
		SELECT DISTINCT
			u.id, u.nickname, u.username, u.avatar, u.status, u.is_admin, u.created_on, ur.created_on as reaction_created_on
		FROM p_user_reactions ur
		JOIN p_user u ON ur.target_user_id = u.id
		WHERE ur.reactor_user_id = ? 
		  AND ur.reaction_type_id = ? 
		  AND ur.is_del = 0 
		  AND u.is_del = 0
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	err = db.Raw(usersQuery, reactorUserID, reactionTypeID, limit, offset).Scan(&users).Error
	if err != nil {
		return nil, 0, err
	}
	
	return users, total, nil
} 

// SearchUserReactions searches for reactions with flexible parameters
func (u *UserReaction) SearchUserReactions(db *gorm.DB, reactorUserID, targetUserID, reactionTypeID int64, limit, offset int) ([]*ReactionWithUserData, int64, error) {
	var reactions []*ReactionWithUserData
	var total int64
	
	// Build dynamic WHERE conditions
	whereConditions := []string{"ur.is_del = 0"}
	var args []interface{}
	
	if reactorUserID > 0 {
		whereConditions = append(whereConditions, "ur.reactor_user_id = ?")
		args = append(args, reactorUserID)
	}
	
	if targetUserID > 0 {
		whereConditions = append(whereConditions, "ur.target_user_id = ?")
		args = append(args, targetUserID)
	}
	
	if reactionTypeID > 0 {
		whereConditions = append(whereConditions, "ur.reaction_type_id = ?")
		args = append(args, reactionTypeID)
	}
	
	whereClause := strings.Join(whereConditions, " AND ")
	
	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM p_user_reactions ur
		WHERE ` + whereClause
	
	err := db.Raw(countQuery, args...).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get reactions with user data
	reactionsQuery := `
		SELECT 
			ur.id as reaction_id,
			ur.target_user_id,
			ur.reaction_type_id,
			ur.created_on,
			u.id as user_id,
			u.nickname,
			u.username,
			u.avatar,
			u.is_admin,
			r.name as reaction_name,
			r.icon as reaction_icon
		FROM p_user_reactions ur
		JOIN p_user u ON ur.reactor_user_id = u.id
		JOIN p_reactions r ON ur.reaction_type_id = r.id
		WHERE ` + whereClause + `
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	// Add limit and offset to args
	args = append(args, limit, offset)
	
	err = db.Raw(reactionsQuery, args...).Scan(&reactions).Error
	if err != nil {
		return nil, 0, err
	}
	
	return reactions, total, nil
}

// GetGlobalReactionTimeline gets all reactions given by all users in chronological order
func (u *UserReaction) GetGlobalReactionTimeline(db *gorm.DB, limit, offset int) ([]*ReactionWithBothUsersData, int64, error) {
	var reactions []*ReactionWithBothUsersData
	var total int64
	
	// Get total count
	countQuery := `
		SELECT COUNT(*) 
		FROM p_user_reactions ur
		WHERE ur.is_del = 0
	`
	
	err := db.Raw(countQuery).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get reactions with both reactor and target user data in chronological order
	reactionsQuery := `
		SELECT 
			ur.id as reaction_id,
			ur.target_user_id,
			ur.reaction_type_id,
			ur.created_on,
			
			-- Reactor user data (user who gave the reaction)
			ur.reactor_user_id,
			reactor.nickname as reactor_nickname,
			reactor.username as reactor_username,
			reactor.avatar as reactor_avatar,
			reactor.is_admin as reactor_is_admin,
			
			-- Target user data (user who received the reaction)
			target.nickname as target_nickname,
			target.username as target_username,
			target.avatar as target_avatar,
			target.is_admin as target_is_admin,
			
			-- Reaction data
			r.name as reaction_name,
			r.icon as reaction_icon
		FROM p_user_reactions ur
		JOIN p_user reactor ON ur.reactor_user_id = reactor.id
		JOIN p_user target ON ur.target_user_id = target.id
		JOIN p_reactions r ON ur.reaction_type_id = r.id
		WHERE ur.is_del = 0
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	err = db.Raw(reactionsQuery, limit, offset).Scan(&reactions).Error
	if err != nil {
		return nil, 0, err
	}
	
	return reactions, total, nil
}

// GetUserReactionTimeline gets all reactions given by a specific user in chronological order
func (u *UserReaction) GetUserReactionTimeline(db *gorm.DB, userID int64, limit, offset int) ([]*ReactionWithBothUsersData, int64, error) {
	var reactions []*ReactionWithBothUsersData
	var total int64
	
	// Get total count for this user
	countQuery := `
		SELECT COUNT(*) 
		FROM p_user_reactions ur
		WHERE ur.reactor_user_id = ? AND ur.is_del = 0
	`
	
	err := db.Raw(countQuery, userID).Scan(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get reactions with both reactor and target user data for this specific user
	reactionsQuery := `
		SELECT 
			ur.id as reaction_id,
			ur.target_user_id,
			ur.reaction_type_id,
			ur.created_on,
			
			-- Reactor user data (user who gave the reaction)
			ur.reactor_user_id,
			reactor.nickname as reactor_nickname,
			reactor.username as reactor_username,
			reactor.avatar as reactor_avatar,
			reactor.is_admin as reactor_is_admin,
			
			-- Target user data (user who received the reaction)
			target.nickname as target_nickname,
			target.username as target_username,
			target.avatar as target_avatar,
			target.is_admin as target_is_admin,
			
			-- Reaction data
			r.name as reaction_name,
			r.icon as reaction_icon
		FROM p_user_reactions ur
		JOIN p_user reactor ON ur.reactor_user_id = reactor.id
		JOIN p_user target ON ur.target_user_id = target.id
		JOIN p_reactions r ON ur.reaction_type_id = r.id
		WHERE ur.reactor_user_id = ? AND ur.is_del = 0
		ORDER BY ur.created_on DESC
		LIMIT ? OFFSET ?
	`
	
	err = db.Raw(reactionsQuery, userID, limit, offset).Scan(&reactions).Error
	if err != nil {
		return nil, 0, err
	}
	
	return reactions, total, nil
} 
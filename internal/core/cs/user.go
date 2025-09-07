// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cs

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
)

const (
	RelationUnknown RelationTyp = iota
	RelationSelf
	RelationFriend
	RelationFollower
	RelationFollowing
	RelationAdmin
	RelationGuest
)

// Reaction type constants
const (
	ReactionTypeLike    int64 = 1  // 👍
	ReactionTypeLove    int64 = 2  // ❤️
	ReactionTypeHot     int64 = 3  // 🔥
	ReactionTypeSmart   int64 = 4  // 🧠
	ReactionTypeFunny   int64 = 5  // 😂
	ReactionTypeKind    int64 = 6  // 🤗
	ReactionTypeBrave   int64 = 7  // 💪
	ReactionTypeCool    int64 = 8  // 😎
	ReactionTypeSweet   int64 = 9  // 🍯
	ReactionTypeStrong  int64 = 10 // 💪
	ReactionTypeFriendly int64 = 11 // 😊
	ReactionTypeHonest  int64 = 12 // 🤝
	ReactionTypeGenerous int64 = 13 // 🎁
	ReactionTypeFit     int64 = 14 // 🏃
	ReactionTypeCreative int64 = 15 // 🎨
	ReactionTypeStupid  int64 = 16 // 🤦
	ReactionTypeMean    int64 = 17 // 😠
	ReactionTypeFake    int64 = 18 // 🎭
	ReactionTypeLazy    int64 = 19 // 😴
)

// Int64Array - custom type for PostgreSQL arrays with proper GORM mapping
type Int64Array []int64

func (a Int64Array) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	return fmt.Sprintf("{%s}", strings.Trim(strings.Replace(fmt.Sprint(a), " ", ",", -1), "[]")), nil
}

func (a *Int64Array) Scan(value interface{}) error {
	if value == nil {
		*a = Int64Array{}
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("cannot scan %T into Int64Array", value)
	}
	if str == "{}" {
		*a = Int64Array{}
		return nil
	}
	str = strings.Trim(str, "{}")
	parts := strings.Split(str, ",")
	result := make(Int64Array, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			var num int64
			if _, err := fmt.Sscanf(part, "%d", &num); err == nil {
				result = append(result, num)
			}
		}
	}
	*a = result
	return nil
}

// UnmarshalJSON implements json.Unmarshaler for JSON binding
func (a *Int64Array) UnmarshalJSON(data []byte) error {
	// Handle null/empty case
	if string(data) == "null" || string(data) == "[]" {
		*a = Int64Array{}
		return nil
	}
	
	// Try to unmarshal as []int64 first
	var int64Slice []int64
	if err := json.Unmarshal(data, &int64Slice); err == nil {
		*a = Int64Array(int64Slice)
		return nil
	}
	
	// Try to unmarshal as []interface{} (from JSON parsing)
	var interfaceSlice []interface{}
	if err := json.Unmarshal(data, &interfaceSlice); err == nil {
		result := make(Int64Array, 0, len(interfaceSlice))
		for _, v := range interfaceSlice {
			switch val := v.(type) {
			case float64:
				result = append(result, int64(val))
			case int64:
				result = append(result, val)
			case int:
				result = append(result, int64(val))
			case string:
				// Try to parse string as int64
				if num, err := fmt.Sscanf(val, "%d", new(int64)); err == nil {
					result = append(result, int64(num))
				}
			}
		}
		*a = result
		return nil
	}
	
	return fmt.Errorf("cannot unmarshal %s into Int64Array", string(data))
}


type (
	// UserInfoList 用户信息列表
	UserInfoList []*UserInfo

	//
	RelationTyp uint8

	VistUser struct {
		Username string
		UserId   int64
		RelTyp   RelationTyp
	}
)

// UserInfo 用户基本信息
type UserInfo struct {
	ID        int64  `json:"id"`
	Nickname  string `json:"nickname"`
	Username  string `json:"username"`
	Status    int    `json:"status"`
	Avatar    string `json:"avatar"`
	IsAdmin   bool   `json:"is_admin"`
	CreatedOn int64  `json:"created_on"`
	IsOnline  bool   `json:"is_online,omitempty" gorm:"-"` // User's online status (optional, not in DB)
}

type UserProfile struct {
	ID          int64            `json:"id" db:"id"`
	Nickname    string           `json:"nickname"`
	Username    string           `json:"username"`
	Phone       string           `json:"phone"`
	Status      int              `json:"status"`
	Avatar      string           `json:"avatar"`
	Balance     int64            `json:"balance"`
	IsAdmin     bool             `json:"is_admin"`
	CreatedOn   int64            `json:"created_on"`
	TweetsCount int              `json:"tweets_count"`
	Categories  Int64Array       `json:"categories" db:"categories"`
	ReactionCounts map[int64]int64 `json:"reaction_counts"` // reaction_type_id -> count (reactions received)
}

// UserReactionInfo represents user reaction information for API responses
type UserReactionInfo struct {
	ID             int64  `json:"id"`
	ReactorUserID  int64  `json:"reactor_user_id"`
	TargetUserID   int64  `json:"target_user_id"`
	ReactionTypeID int64  `json:"reaction_type_id"`
	ReactionName   string `json:"reaction_name"`
	ReactionIcon   string `json:"reaction_icon"`
	CreatedOn      int64  `json:"created_on"`
	ModifiedOn     int64  `json:"modified_on"`
}

// UserReactionCounts represents reaction counts for a user
type UserReactionCounts struct {
	UserID          int64            `json:"user_id"`
	ReactionCounts  map[int64]int64  `json:"reaction_counts"`  // reaction_type_id -> count (reactions received)
	GivenCounts     map[int64]int64  `json:"given_counts"`     // reaction_type_id -> count (reactions given)
}

// UserProfileWithFollow adds follow status to UserProfile
type UserProfileWithFollow struct {
	ID          int64  `json:"id"`
	Nickname    string `json:"nickname"`
	Username    string `json:"username"`
	Phone       string `json:"phone"`
	Status      int    `json:"status"`
	Avatar      string `json:"avatar"`
	Balance     int64  `json:"balance"`
	IsAdmin     bool   `json:"is_admin"`
	CreatedOn   int64  `json:"created_on"`
	TweetsCount int    `json:"tweets_count"`
	IsFollowing bool   `json:"is_following"`
	Categories  Int64Array `json:"categories,omitempty"`
	IsOnline    bool   `json:"is_online,omitempty" gorm:"-"` // User's online status (optional, not in DB)
}

// UserReactionList represents a list of user reactions with pagination
type UserReactionList struct {
	Reactions []UserReactionInfo `json:"reactions"`
	Total     int64              `json:"total"`
	Limit     int                `json:"limit"`
	Offset    int                `json:"offset"`
}

// UserReactionWithUser represents a reaction with user data for API responses
type UserReactionWithUser struct {
	ReactionID     int64         `json:"reaction_id"`
	ReactorUser    *UserFormated `json:"reactor_user"`    // User who reacted
	TargetUserID   int64         `json:"target_user_id"`  // User who was reacted to
	ReactionTypeID int64         `json:"reaction_type_id"` // Type of reaction
	ReactionName   string        `json:"reaction_name"`   // Human readable name
	ReactionIcon   string        `json:"reaction_icon"`   // Icon for reaction
	CreatedOn      int64         `json:"created_on"`      // When reaction was created
}

// UserReactionWithBothUsers represents a reaction with both reactor and target user data for API responses
type UserReactionWithBothUsers struct {
	ReactionID     int64         `json:"reaction_id"`
	ReactorUser    *UserFormated `json:"reactor_user"`    // User who reacted
	TargetUser     *UserFormated `json:"target_user"`     // User who was reacted to
	ReactionTypeID int64         `json:"reaction_type_id"` // Type of reaction
	ReactionName   string        `json:"reaction_name"`   // Human readable name
	ReactionIcon   string        `json:"reaction_icon"`   // Icon for reaction
	CreatedOn      int64         `json:"created_on"`      // When reaction was created
}

type UserFormated struct {
	ID          int64   `db:"id" json:"id"`
	Nickname    string  `json:"nickname"`
	Username    string  `json:"username"`
	Status      int     `json:"status"`
	Avatar      string  `json:"avatar"`
	IsAdmin     bool    `json:"is_admin"`
	IsFriend    bool    `json:"is_friend"`
	IsFollowing bool    `json:"is_following"`
	Categories  Int64Array `json:"categories" db:"categories"`
	IsOnline    bool    `json:"is_online,omitempty" gorm:"-"` // User's online status (optional, not in DB)
}

func (t RelationTyp) String() string {
	switch t {
	case RelationSelf:
		return "self"
	case RelationFriend:
		return "friend"
	case RelationFollower:
		return "follower"
	case RelationFollowing:
		return "following"
	case RelationAdmin:
		return "admin"
	case RelationUnknown:
		fallthrough
	default:
		return "unknown"
	}
}

// GetReactionName returns the human-readable name for a reaction type
func GetReactionName(reactionTypeID int64) string {
	switch reactionTypeID {
	case ReactionTypeLike:
		return "like"
	case ReactionTypeLove:
		return "love"
	case ReactionTypeHot:
		return "hot"
	case ReactionTypeSmart:
		return "smart"
	case ReactionTypeFunny:
		return "funny"
	case ReactionTypeKind:
		return "kind"
	case ReactionTypeBrave:
		return "brave"
	case ReactionTypeCool:
		return "cool"
	case ReactionTypeSweet:
		return "sweet"
	case ReactionTypeStrong:
		return "strong"
	case ReactionTypeFriendly:
		return "friendly"
	case ReactionTypeHonest:
		return "honest"
	case ReactionTypeGenerous:
		return "generous"
	case ReactionTypeFit:
		return "fit"
	case ReactionTypeCreative:
		return "creative"
	case ReactionTypeStupid:
		return "stupid"
	case ReactionTypeMean:
		return "mean"
	case ReactionTypeFake:
		return "fake"
	case ReactionTypeLazy:
		return "lazy"
	default:
		return "like"
	}
}

// GetReactionIcon returns the emoji/icon for a reaction type
func GetReactionIcon(reactionTypeID int64) string {
	switch reactionTypeID {
	case ReactionTypeLike:
		return "👍"
	case ReactionTypeLove:
		return "❤️"
	case ReactionTypeHot:
		return "🔥"
	case ReactionTypeSmart:
		return "🧠"
	case ReactionTypeFunny:
		return "😂"
	case ReactionTypeKind:
		return "🤗"
	case ReactionTypeBrave:
		return "💪"
	case ReactionTypeCool:
		return "😎"
	case ReactionTypeSweet:
		return "🍯"
	case ReactionTypeStrong:
		return "💪"
	case ReactionTypeFriendly:
		return "😊"
	case ReactionTypeHonest:
		return "🤝"
	case ReactionTypeGenerous:
		return "🎁"
	case ReactionTypeFit:
		return "🏃"
	case ReactionTypeCreative:
		return "🎨"
	case ReactionTypeStupid:
		return "🤦"
	case ReactionTypeMean:
		return "😠"
	case ReactionTypeFake:
		return "🎭"
	case ReactionTypeLazy:
		return "😴"
	default:
		return "👍"
	}
}

// IsPositiveReaction returns true if the reaction type is positive
func IsPositiveReaction(reactionTypeID int64) bool {
	return reactionTypeID >= 1 && reactionTypeID <= 15
}

// IsNegativeReaction returns true if the reaction type is negative
func IsNegativeReaction(reactionTypeID int64) bool {
	return reactionTypeID >= 16 && reactionTypeID <= 19
}

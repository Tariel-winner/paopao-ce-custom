// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rocboss/paopao-ce/internal/core/cs"
	"gorm.io/gorm"
)

const (
	UserStatusNormal int = iota + 1
	UserStatusClosed
)
// Int64Array - minimal custom type for PostgreSQL arrays
type Int64Array []int64


type User struct {
	*Model
	Nickname   string        `json:"nickname"`
	Username   string        `json:"username"`
	Phone      string        `json:"phone"`
	Password   string        `json:"password"`
	Salt       string        `json:"salt"`
	Status     int           `json:"status"`
	Avatar     string        `json:"avatar"`
	Balance    int64         `json:"balance"`
	IsAdmin    bool          `json:"is_admin"`
	Categories Int64Array    `json:"categories" gorm:"type:integer[];default:'{}'"`
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
	Categories  []int64 `json:"categories"`
}

type UserReactionWithUser struct {
	ReactionID     int64         `json:"reaction_id"`
	ReactorUser    *UserFormated `json:"reactor_user"`    // User who reacted
	TargetUserID   int64         `json:"target_user_id"`  // User who was reacted to
	ReactionTypeID int64         `json:"reaction_type_id"` // Type of reaction
	ReactionName   string        `json:"reaction_name"`   // Human readable name
	ReactionIcon   string        `json:"reaction_icon"`   // Icon for reaction
	CreatedOn      int64         `json:"created_on"`      // When reaction was created
}

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


func (u *User) Format() *UserFormated {
	if u.Model != nil {
		return &UserFormated{
			ID:         u.ID,
			Nickname:   u.Nickname,
			Username:   u.Username,
			Status:     u.Status,
			Avatar:     u.Avatar,
			IsAdmin:    u.IsAdmin,
			Categories: []int64(u.Categories),
		}
	}

	return nil
}

func (u *User) Get(db *gorm.DB) (*User, error) {
	var user User
	if u.Model != nil && u.Model.ID > 0 {
		db = db.Where("id= ? AND is_del = ?", u.Model.ID, 0)
	} else if u.Phone != "" {
		db = db.Where("phone = ? AND is_del = ?", u.Phone, 0)
	} else {
		db = db.Where("username = ? AND is_del = ?", u.Username, 0)
	}

	err := db.First(&user).Error
	if err != nil {
		return &user, err
	}

	return &user, nil
}

func (u *User) List(db *gorm.DB, conditions *ConditionsT, offset, limit int) ([]*User, error) {
	var users []*User
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	for k, v := range *conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

func (u *User) ListUserInfoById(db *gorm.DB, ids []int64) (res cs.UserInfoList, err error) {
	err = db.Model(u).Where("id IN ?", ids).Find(&res).Error
	return
}

func (u *User) Create(db *gorm.DB) (*User, error) {
	err := db.Create(&u).Error
	return u, err
}

func (u *User) Update(db *gorm.DB) error {
	return db.Model(&User{}).Where("id = ? AND is_del = ?", u.Model.ID, 0).Save(u).Error
}

func (u *User) SetCategories(db *gorm.DB, userID int64, categoryIDs []int64) error {
	return db.Model(&User{}).
		Where("id = ? AND is_del = ?", userID, 0).
		Update("categories", Int64Array(categoryIDs)).Error
}

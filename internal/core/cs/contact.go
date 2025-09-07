// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cs

const (
	ContactStatusRequesting int8 = iota + 1
	ContactStatusAgree
	ContactStatusReject
	ContactStatusDeleted
)

type Contact struct {
	ID           int64  `db:"id" json:"id"`
	UserId       int64  `db:"user_id" json:"user_id"`
	FriendId     int64  `db:"friend_id" json:"friend_id"`
	GroupId      int64  `json:"group_id"`
	Remark       string `json:"remark"`
	Status       int8   `json:"status"` // 1请求好友, 2已同意好友, 3已拒绝好友, 4已删除好友
	IsTop        int8   `json:"is_top"`
	IsBlack      int8   `json:"is_black"`
	NoticeEnable int8   `json:"notice_enable"`
	IsDel        int8   `json:"-"`
	DeletedOn    int64  `db:"-" json:"-"`
}

// PhoneContact represents a contact from iPhone address book
type PhoneContact struct {
	Name  string `json:"name"`  // Contact name (e.g., "John Doe")
	Phone string `json:"phone"` // Phone number (e.g., "+1234567890")
	Email string `json:"email"` // Email address (optional)
}

// UserDevice represents a user's device for push notifications
type UserDevice struct {
	ID           int64  `db:"id" json:"id"`
	UserID       int64  `db:"user_id" json:"user_id"`
	DeviceToken  string `db:"device_token" json:"device_token"`
	Platform     string `db:"platform" json:"platform"`      // 'ios' or 'android'
	DeviceID     string `db:"device_id" json:"device_id"`
	DeviceName   string `db:"device_name" json:"device_name"`
	IsActive     bool   `db:"is_active" json:"is_active"`
	LastUsedOn   int64  `db:"last_used_on" json:"last_used_on"`
	CreatedOn    int64  `db:"created_on" json:"created_on"`
	ModifiedOn   int64  `db:"modified_on" json:"modified_on"`
	DeletedOn    int64  `db:"deleted_on" json:"deleted_on"`
	IsDel        int8   `db:"is_del" json:"-"`
}

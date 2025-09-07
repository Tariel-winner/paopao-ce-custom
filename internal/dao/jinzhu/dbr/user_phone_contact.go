// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"gorm.io/gorm"
)

// UserPhoneContact represents iPhone contacts uploaded by users
type UserPhoneContact struct {
	*Model
	UserId         int64  `json:"user_id" gorm:"column:user_id"`
	ContactName    string `json:"contact_name" gorm:"column:contact_name"`
	ContactPhone   string `json:"contact_phone" gorm:"column:contact_phone"`
	ContactEmail   string `json:"contact_email" gorm:"column:contact_email"`
	IsMatched      bool   `json:"is_matched" gorm:"column:is_matched"`
	MatchedUserID  *int64 `json:"matched_user_id" gorm:"column:matched_user_id"`
	CreatedOn      int64  `json:"created_on" gorm:"column:created_on"`
	ModifiedOn     int64  `json:"modified_on" gorm:"column:modified_on"`
	DeletedOn      int64  `json:"deleted_on" gorm:"column:deleted_on"`
	IsDel          int8   `json:"-" gorm:"column:is_del"`
}

// TableName specifies the table name for UserPhoneContact
func (UserPhoneContact) TableName() string {
	return "p_user_phone_contacts"
}

// Create saves a new UserPhoneContact to the database
func (c *UserPhoneContact) Create(db *gorm.DB) error {
	return db.Create(c).Error
}

// Update updates an existing UserPhoneContact in the database
func (c *UserPhoneContact) Update(db *gorm.DB) error {
	return db.Model(&UserPhoneContact{}).Where("id = ?", c.Model.ID).Save(c).Error
}

// GetByUserID gets all phone contacts for a specific user
func (c *UserPhoneContact) GetByUserID(db *gorm.DB, userID int64) ([]*UserPhoneContact, error) {
	var contacts []*UserPhoneContact
	err := db.Where("user_id = ? AND is_del = ?", userID, 0).Find(&contacts).Error
	return contacts, err
}

// GetUnmatchedByUserID gets all unmatched phone contacts for a specific user
func (c *UserPhoneContact) GetUnmatchedByUserID(db *gorm.DB, userID int64) ([]*UserPhoneContact, error) {
	var contacts []*UserPhoneContact
	err := db.Where("user_id = ? AND is_matched = ? AND is_del = ?", userID, false, 0).Find(&contacts).Error
	return contacts, err
}

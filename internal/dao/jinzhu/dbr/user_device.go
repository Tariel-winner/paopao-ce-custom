// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"gorm.io/gorm"
)

// UserDevice represents a user's device for push notifications
type UserDevice struct {
	*Model
	UserID       int64  `json:"user_id" gorm:"column:user_id"`
	DeviceToken  string `json:"device_token" gorm:"column:device_token"`
	Platform     string `json:"platform" gorm:"column:platform"`      // 'ios' or 'android'
	DeviceID     string `json:"device_id" gorm:"column:device_id"`
	DeviceName   string `json:"device_name" gorm:"column:device_name"`
	IsActive     bool   `json:"is_active" gorm:"column:is_active"`
	LastUsedOn   int64  `json:"last_used_on" gorm:"column:last_used_on"`
	CreatedOn    int64  `json:"created_on" gorm:"column:created_on"`
	ModifiedOn   int64  `json:"modified_on" gorm:"column:modified_on"`
	DeletedOn    int64  `json:"deleted_on" gorm:"column:deleted_on"`
	IsDel        int8   `json:"-" gorm:"column:is_del"`
}

// TableName specifies the table name for UserDevice
func (UserDevice) TableName() string {
	return "p_user_device_tokens"
}

// Create saves a new UserDevice to the database
func (d *UserDevice) Create(db *gorm.DB) error {
	return db.Create(d).Error
}

// Update updates an existing UserDevice in the database
func (d *UserDevice) Update(db *gorm.DB) error {
	return db.Model(&UserDevice{}).Where("id = ?", d.Model.ID).Save(d).Error
}

// GetByUserID gets all devices for a specific user
func (d *UserDevice) GetByUserID(db *gorm.DB, userID int64) ([]*UserDevice, error) {
	var devices []*UserDevice
	err := db.Where("user_id = ? AND is_del = ?", userID, 0).Find(&devices).Error
	return devices, err
}

// GetActiveByUserID gets all active devices for a specific user
func (d *UserDevice) GetActiveByUserID(db *gorm.DB, userID int64) ([]*UserDevice, error) {
	var devices []*UserDevice
	err := db.Where("user_id = ? AND is_active = ? AND is_del = ?", userID, true, 0).Find(&devices).Error
	return devices, err
}

// GetByDeviceID gets a device by its unique device ID
func (d *UserDevice) GetByDeviceID(db *gorm.DB, deviceID string) (*UserDevice, error) {
	var device UserDevice
	err := db.Where("device_id = ? AND is_del = ?", deviceID, 0).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// UpdateDeviceToken updates the device token for a specific device
func (d *UserDevice) UpdateDeviceToken(db *gorm.DB, deviceID, deviceToken string) error {
	return db.Model(&UserDevice{}).Where("device_id = ? AND is_del = ?", deviceID, 0).Update("device_token", deviceToken).Error
}

// DeactivateDevice deactivates a device (sets is_active to false)
func (d *UserDevice) DeactivateDevice(db *gorm.DB, deviceID string) error {
	return db.Model(&UserDevice{}).Where("device_id = ? AND is_del = ?", deviceID, 0).Update("is_active", false).Error
}

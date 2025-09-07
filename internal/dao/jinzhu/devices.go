// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
	"time"

	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	_ core.DeviceManageService = (*deviceManageSrv)(nil)
)

type deviceManageSrv struct {
	db *gorm.DB
}

func newDeviceManageService(db *gorm.DB) core.DeviceManageService {
	return &deviceManageSrv{
		db: db,
	}
}

// RegisterDevice registers a new device for push notifications
func (s *deviceManageSrv) RegisterDevice(userID int64, deviceToken, platform, deviceID, deviceName string) error {
	db := s.db.Begin()
	defer func() {
		if err := recover(); err != nil {
			db.Rollback()
		}
	}()

	now := time.Now().Unix()

	// Check if device already exists
	var existingDevice dbr.UserDevice
	if err := db.Where("device_id = ? AND is_del = ?", deviceID, 0).First(&existingDevice).Error; err == nil {
		// Device exists, update it
		existingDevice.DeviceToken = deviceToken
		existingDevice.Platform = platform
		existingDevice.DeviceName = deviceName
		existingDevice.IsActive = true
		existingDevice.LastUsedOn = now
		existingDevice.ModifiedOn = now

		if err := existingDevice.Update(db); err != nil {
			db.Rollback()
			return err
		}
	} else if err == gorm.ErrRecordNotFound {
		// Device doesn't exist, create new one
		device := &dbr.UserDevice{
			UserID:      userID,
			DeviceToken: deviceToken,
			Platform:    platform,
			DeviceID:    deviceID,
			DeviceName:  deviceName,
			IsActive:    true,
			LastUsedOn:  now,
			CreatedOn:   now,
			ModifiedOn:  now,
		}

		if err := device.Create(db); err != nil {
			db.Rollback()
			return err
		}
	} else {
		// Database error
		db.Rollback()
		return err
	}

	if err := db.Commit().Error; err != nil {
		return err
	}

	logrus.Infof("Device registered successfully: userID=%d, deviceID=%s, platform=%s", userID, deviceID, platform)
	return nil
}

// UpdateDeviceToken updates the device token for an existing device
func (s *deviceManageSrv) UpdateDeviceToken(deviceID, deviceToken string) error {
	device := &dbr.UserDevice{}
	
	if err := device.UpdateDeviceToken(s.db, deviceID, deviceToken); err != nil {
		logrus.Errorf("Failed to update device token: deviceID=%s, error=%v", deviceID, err)
		return err
	}

	logrus.Infof("Device token updated successfully: deviceID=%s", deviceID)
	return nil
}

// GetUserDevices gets all devices for a specific user
func (s *deviceManageSrv) GetUserDevices(userID int64) ([]*cs.UserDevice, error) {
	device := &dbr.UserDevice{}
	devices, err := device.GetByUserID(s.db, userID)
	if err != nil {
		return nil, err
	}

	// Convert to core.UserDevice
	var result []*cs.UserDevice
	for _, d := range devices {
		result = append(result, &cs.UserDevice{
			ID:           d.ID,
			UserID:       d.UserID,
			DeviceToken:  d.DeviceToken,
			Platform:     d.Platform,
			DeviceID:     d.DeviceID,
			DeviceName:   d.DeviceName,
			IsActive:     d.IsActive,
			LastUsedOn:   d.LastUsedOn,
			CreatedOn:    d.CreatedOn,
			ModifiedOn:   d.ModifiedOn,
			DeletedOn:    d.DeletedOn,
			IsDel:        d.IsDel,
		})
	}

	return result, nil
}

// GetActiveUserDevices gets all active devices for a specific user (for sending push notifications)
func (s *deviceManageSrv) GetActiveUserDevices(userID int64) ([]*cs.UserDevice, error) {
	device := &dbr.UserDevice{}
	devices, err := device.GetActiveByUserID(s.db, userID)
	if err != nil {
		return nil, err
	}

	// Convert to core.UserDevice
	var result []*cs.UserDevice
	for _, d := range devices {
		result = append(result, &cs.UserDevice{
			ID:           d.ID,
			UserID:       d.UserID,
			DeviceToken:  d.DeviceToken,
			Platform:     d.Platform,
			DeviceID:     d.DeviceID,
			DeviceName:   d.DeviceName,
			IsActive:     d.IsActive,
			LastUsedOn:   d.LastUsedOn,
			CreatedOn:    d.CreatedOn,
			ModifiedOn:   d.ModifiedOn,
			DeletedOn:    d.DeletedOn,
			IsDel:        d.IsDel,
		})
	}

	return result, nil
}

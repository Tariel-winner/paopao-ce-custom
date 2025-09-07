// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	
)

type GetCaptchaResp struct {
	Id      string `json:"id"`
	Content string `json:"b64s"`
}

type SendCaptchaReq struct {
	Phone        string `json:"phone" form:"phone" binding:"required"`
	ImgCaptcha   string `json:"img_captcha" form:"img_captcha" binding:"required"`
	ImgCaptchaID string `json:"img_captcha_id" form:"img_captcha_id" binding:"required"`
}

type LoginReq struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type LoginResp struct {
	Token string `json:"token"`
}

// Contact item for iPhone address book integration
type ContactItem struct {
	Name  string `json:"name" binding:"required"`           // Contact name (e.g., "John Doe")
	Phone string `json:"phone" binding:"required"`          // Phone number (e.g., "+1234567890")
	Email string `json:"email"`                             // Email address (optional)
}

// Device info for push notifications
type DeviceInfo struct {
	DeviceToken string `json:"device_token" binding:"required"` // Push notification token from device
	Platform    string `json:"platform" binding:"required,oneof=ios android"` // Device platform
	DeviceID    string `json:"device_id" binding:"required"`   // Unique device identifier
	DeviceName  string `json:"device_name"`                    // User-friendly device name (e.g., "iPhone 15")
}

type RegisterReq struct {
	Username   string   `json:"username" form:"username" binding:"required"`
	Password   string   `json:"password" form:"password" binding:"required"`
	Categories dbr.Int64Array `json:"categories" form:"categories"` // Optional: user's preferred categories
	
	// iPhone contacts for address book integration
	Contacts []ContactItem `json:"contacts"` // Array of contacts from iPhone (optional)
	
	// Device info for push notifications
	Device *DeviceInfo `json:"device"` // Device registration info (optional)
}

type RegisterResp struct {
	UserId   int64  `json:"id"`
	Username string `json:"username"`
	
	// Response info about what was processed
	ContactsUploaded int64 `json:"contacts_uploaded,omitempty"` // How many contacts were uploaded
	ContactsMatched  int64 `json:"contacts_matched,omitempty"`  // How many contacts match app users
	DeviceRegistered bool  `json:"device_registered,omitempty"` // Whether device was registered
}



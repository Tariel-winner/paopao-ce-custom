// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PushNotificationService handles sending push notifications via Gorush
type PushNotificationService struct {
	db           *gorm.DB
	gorushURL    string
	httpClient   *http.Client
	notificationCache *NotificationCache
	cache        core.AppCache
	
	// Cache for device tokens to avoid repeated DB queries every 30 seconds
	deviceTokensCache map[int64][]*dbr.UserDevice // Cache device tokens by user ID
	usersWithDevicesCache []int64                 // Cache list of users with devices
	lastDeviceCacheUpdate int64                   // When device cache was last updated
	deviceCacheExpiry     int64                   // Device cache expiry time (5 minutes)
}

// GorushNotification represents the notification payload for Gorush
type GorushNotification struct {
	Tokens   []string               `json:"tokens"`
	Platform int                    `json:"platform"` // 1 for iOS, 2 for Android
	Message  string                 `json:"message"`
	Title    string                 `json:"title"`
	Priority string                 `json:"priority"`
	Sound    string                 `json:"sound"`
	Badge    int                    `json:"badge"`
	Data     map[string]interface{} `json:"data"`
}

// GorushResponse represents the response from Gorush
type GorushResponse struct {
	Success string `json:"success"`
	Counts  int    `json:"counts"`
	Logs    []struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"logs"`
}

// NewPushNotificationService creates a new push notification service
func NewPushNotificationService(db *gorm.DB, gorushURL string, cache core.AppCache) *PushNotificationService {
	return &PushNotificationService{
		db:        db,
		gorushURL: gorushURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		notificationCache: NewNotificationCache(cache),
		cache:            cache,
		deviceTokensCache: make(map[int64][]*dbr.UserDevice),
		deviceCacheExpiry: 300, // Cache expires after 5 minutes
	}
}

// SendContactOnlineNotification sends notification when someone comes online to ALL users in the app
func (s *PushNotificationService) SendContactOnlineNotification(onlineUserID int64, onlineUsername string) error {
	if onlineUserID == 0 {
		logrus.Info("Sending generic online notification to all users - someone is online")
	} else {
		logrus.Infof("Sending online notification for user %s (ID: %d) to all users", onlineUsername, onlineUserID)
	}

	// Process asynchronously to avoid blocking the main thread
	// Use a separate goroutine with proper cleanup
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Panic in sendOnlineNotificationsAsync: %v", r)
			}
		}()
		s.sendOnlineNotificationsAsync(onlineUserID, onlineUsername)
	}()
	
	return nil
}

// sendOnlineNotificationsAsync processes online notifications asynchronously
func (s *PushNotificationService) sendOnlineNotificationsAsync(onlineUserID int64, onlineUsername string) {
	// Process users in batches to avoid memory issues
	batchSize := 1000 // Process 1000 users at a time
	offset := 0
	notificationsSent := 0

	for {
		// Get batch of users with device tokens
		userBatch, err := s.getUsersWithDevicesBatch(offset, batchSize)
		if err != nil {
			logrus.Errorf("Failed to get users batch (offset: %d, limit: %d): %v", offset, batchSize, err)
			return
		}

		// If no more users, break
		if len(userBatch) == 0 {
			break
		}

		logrus.Debugf("Processing batch of %d users (offset: %d)", len(userBatch), offset)

		// Process this batch
		batchSent := s.processUserBatch(userBatch, onlineUserID, onlineUsername)
		notificationsSent += batchSent

		// Move to next batch
		offset += batchSize

		// If we got fewer users than batch size, we're done
		if len(userBatch) < batchSize {
			break
		}

		// Small delay between batches to avoid overwhelming the system
		time.Sleep(100 * time.Millisecond)
	}

	logrus.Infof("Sent online notifications to %d users for %s", notificationsSent, onlineUsername)
}

// SendContactMatchedNotification sends notification when a contact is matched
func (s *PushNotificationService) SendContactMatchedNotification(contactOwnerID int64, matchedUsername string) error {
	logrus.Infof("Sending contact matched notification to user %d for %s", contactOwnerID, matchedUsername)

	// Get device tokens for the contact owner
	iosTokens, androidTokens := s.getDeviceTokensByPlatform([]int64{contactOwnerID})

	// Send iOS notifications with spam prevention
	if len(iosTokens) > 0 {
		if err := s.sendGorushNotificationWithCache(iosTokens, 1,
			fmt.Sprintf("%s is on the app!", matchedUsername),
			"Contact Found",
			NotificationTypeContactMatched,
			contactOwnerID,
			map[string]interface{}{
				"type": "contact_matched",
				"username": matchedUsername,
			}); err != nil {
			logrus.Errorf("Failed to send iOS contact matched notification: %v", err)
		}
	}

	// Send Android notifications with spam prevention
	if len(androidTokens) > 0 {
		if err := s.sendGorushNotificationWithCache(androidTokens, 2,
			fmt.Sprintf("%s is on the app!", matchedUsername),
			"Contact Found",
			NotificationTypeContactMatched,
			contactOwnerID,
			map[string]interface{}{
				"type": "contact_matched",
				"username": matchedUsername,
			}); err != nil {
			logrus.Errorf("Failed to send Android contact matched notification: %v", err)
		}
	}

	return nil
}

// getUsersWithMatchedContact gets all users who have a specific user in their contacts
func (s *PushNotificationService) getUsersWithMatchedContact(targetUserID int64) ([]int64, error) {
	var userIDs []int64
	err := s.db.Model(&dbr.UserPhoneContact{}).
		Where("matched_user_id = ? AND is_matched = ? AND is_del = ?", targetUserID, true, 0).
		Pluck("user_id", &userIDs).Error
	
	return userIDs, err
}

// getAllUsersWithDevices gets all users who have registered device tokens
func (s *PushNotificationService) getAllUsersWithDevices() ([]int64, error) {
	var userIDs []int64
	err := s.db.Model(&dbr.UserDevice{}).
		Where("is_active = ? AND is_del = ?", true, 0).
		Distinct("user_id").
		Pluck("user_id", &userIDs).Error
	
	return userIDs, err
}

// getUsersWithDevicesBatch gets a batch of users who have registered device tokens
func (s *PushNotificationService) getUsersWithDevicesBatch(offset, limit int) ([]int64, error) {
	// Use cache to avoid DB query every 30 seconds
	usersWithDevices, err := s.getUsersWithDevicesFromCache()
	if err != nil {
		return nil, err
	}
	
	// Apply pagination to cached results
	start := offset
	end := offset + limit
	
	if start >= len(usersWithDevices) {
		return []int64{}, nil // No more users
	}
	
	if end > len(usersWithDevices) {
		end = len(usersWithDevices)
	}
	
	return usersWithDevices[start:end], nil
}

// getUsersWithDevicesFromCache gets users with devices from cache, updates cache if needed
func (s *PushNotificationService) getUsersWithDevicesFromCache() ([]int64, error) {
	now := time.Now().Unix()
	
	// Check if cache needs refresh
	if now-s.lastDeviceCacheUpdate > s.deviceCacheExpiry {
		if err := s.refreshDeviceCache(now); err != nil {
			logrus.Errorf("Failed to refresh device cache: %v", err)
			// Fall back to direct DB query
			return s.getAllUsersWithDevices()
		}
	}
	
	// Return from cache
	return s.usersWithDevicesCache, nil
}

// refreshDeviceCache loads all device tokens into cache
func (s *PushNotificationService) refreshDeviceCache(now int64) error {
	logrus.Debug("Refreshing device cache...")
	
	// Load all active device tokens
	var devices []dbr.UserDevice
	if err := s.db.Where("is_active = ? AND is_del = ?", true, 0).Find(&devices).Error; err != nil {
		return err
	}
	
	// Group devices by user and get unique user IDs
	userDevicesMap := make(map[int64][]*dbr.UserDevice)
	userIDsSet := make(map[int64]bool)
	
	for i := range devices {
		userID := devices[i].UserID
		userDevicesMap[userID] = append(userDevicesMap[userID], &devices[i])
		userIDsSet[userID] = true
	}
	
	// Convert set to slice
	var userIDs []int64
	for userID := range userIDsSet {
		userIDs = append(userIDs, userID)
	}
	
	// Update cache
	s.deviceTokensCache = userDevicesMap
	s.usersWithDevicesCache = userIDs
	s.lastDeviceCacheUpdate = now
	
	logrus.Debugf("Device cache refreshed with %d users having %d total devices", len(userIDs), len(devices))
	return nil
}

// processUserBatch processes a batch of users for online notifications
func (s *PushNotificationService) processUserBatch(userBatch []int64, onlineUserID int64, onlineUsername string) int {
	notificationsSent := 0

	for _, userID := range userBatch {
		// Skip sending notification to the user who came online (only for specific user notifications)
		if onlineUserID != 0 && userID == onlineUserID {
			continue
		}

		// For generic notifications (onlineUserID == 0), we don't need per-user spam prevention
		// The spam prevention is handled at the monitor level (30-second intervals)

		// Get device tokens for this specific user
		iosTokens, androidTokens := s.getDeviceTokensByPlatform([]int64{userID})

		// Prepare notification message
		var message, title string
		var data map[string]interface{}
		
		if onlineUserID == 0 {
			// Get one online user and create dopamine message
			message, title, data = s.createDopamineMessage()
		} else {
			// Specific user notification
			message = fmt.Sprintf("%s is now online!", onlineUsername)
			title = "User Online"
			data = map[string]interface{}{
				"type": "user_online",
				"user_id": onlineUserID,
				"username": onlineUsername,
			}
		}

		// Send iOS notifications
		if len(iosTokens) > 0 {
			if err := s.sendGorushNotification(iosTokens, 1, message, title, data); err != nil {
				logrus.Errorf("Failed to send iOS notification to user %d: %v", userID, err)
			} else {
				notificationsSent++
			}
		}

		// Send Android notifications
		if len(androidTokens) > 0 {
			if err := s.sendGorushNotification(androidTokens, 2, message, title, data); err != nil {
				logrus.Errorf("Failed to send Android notification to user %d: %v", userID, err)
			} else {
				notificationsSent++
			}
		}
	}

	return notificationsSent
}

// getDeviceTokensByPlatform gets device tokens grouped by platform for given user IDs
func (s *PushNotificationService) getDeviceTokensByPlatform(userIDs []int64) ([]string, []string) {
	var iosTokens, androidTokens []string

	for _, userID := range userIDs {
		devices, err := s.getActiveUserDevices(userID)
		if err != nil {
			logrus.Errorf("Failed to get devices for user %d: %v", userID, err)
			continue
		}

		for _, device := range devices {
			if device.IsActive {
				switch device.Platform {
				case "ios":
					iosTokens = append(iosTokens, device.DeviceToken)
				case "android":
					androidTokens = append(androidTokens, device.DeviceToken)
				}
			}
		}
	}

	return iosTokens, androidTokens
}

// getActiveUserDevices gets all active devices for a specific user from cache
func (s *PushNotificationService) getActiveUserDevices(userID int64) ([]*cs.UserDevice, error) {
	// Use cache to avoid DB query every 30 seconds
	devices, err := s.getUserDevicesFromCache(userID)
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

// getUserDevicesFromCache gets user devices from cache, updates cache if needed
func (s *PushNotificationService) getUserDevicesFromCache(userID int64) ([]*dbr.UserDevice, error) {
	now := time.Now().Unix()
	
	// Check if cache needs refresh
	if now-s.lastDeviceCacheUpdate > s.deviceCacheExpiry {
		if err := s.refreshDeviceCache(now); err != nil {
			logrus.Errorf("Failed to refresh device cache: %v", err)
			// Fall back to direct DB query
			device := &dbr.UserDevice{}
			return device.GetActiveByUserID(s.db, userID)
		}
	}
	
	// Return from cache
	if devices, exists := s.deviceTokensCache[userID]; exists {
		return devices, nil
	}
	
	// User not in cache, query directly
	device := &dbr.UserDevice{}
	devices, err := device.GetActiveByUserID(s.db, userID)
	if err == nil {
		// Add to cache for next time
		s.deviceTokensCache[userID] = devices
	}
	return devices, err
}

// sendGorushNotification sends a notification via Gorush
func (s *PushNotificationService) sendGorushNotification(tokens []string, platform int, message, title string, data map[string]interface{}) error {
	if len(tokens) == 0 {
		return nil
	}

	notification := GorushNotification{
		Tokens:   tokens,
		Platform: platform,
		Message:  message,
		Title:    title,
		Priority: "high",
		Sound:    "default",
		Badge:    1,
		Data:     data,
	}

	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %v", err)
	}

	url := fmt.Sprintf("%s/api/push", s.gorushURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gorush returned status %d", resp.StatusCode)
	}

	var gorushResp GorushResponse
	if err := json.NewDecoder(resp.Body).Decode(&gorushResp); err != nil {
		return fmt.Errorf("failed to decode response: %v", err)
	}

	logrus.Infof("Gorush response: %s, sent to %d devices", gorushResp.Success, gorushResp.Counts)
	return nil
}

// sendGorushNotificationWithCache sends notification with spam prevention
func (s *PushNotificationService) sendGorushNotificationWithCache(
	tokens []string, 
	platform int, 
	message, title string, 
	notificationType NotificationType,
	targetUserID int64,
	data map[string]interface{},
) error {
	if len(tokens) == 0 {
		return nil
	}

	// Check if we should send this notification (spam prevention)
	shouldSend := s.notificationCache.ShouldSendNotification(0, targetUserID, notificationType)

	if !shouldSend {
		logrus.Debugf("Skipping notification for user %d (type: %s) due to spam prevention", 
			targetUserID, notificationType)
		return nil
	}

	// Send the notification
	return s.sendGorushNotification(tokens, platform, message, title, data)
}

// createDopamineMessage creates a simple dopamine-triggered message
func (s *PushNotificationService) createDopamineMessage() (string, string, map[string]interface{}) {
	// Get one online user
	pattern := conf.PrefixOnlineUser + "*"
	keys, err := s.cache.Keys(pattern)
	if err != nil || len(keys) == 0 {
		return "🔥 People from different places are live! Join the chat!", "Live Now", map[string]interface{}{
			"type": "user_online",
			"user_id": 0,
		}
	}

	// Pick random user
	rand.Seed(time.Now().UnixNano())
	randomKey := keys[rand.Intn(len(keys))]
	userIDStr := strings.TrimPrefix(randomKey, conf.PrefixOnlineUser)
	
	// Get location (format: "Country|City" or "Country")
	locationKey := conf.PrefixUserLocation + userIDStr
	locationData, err := s.cache.Get(locationKey)
	
	var message string
	if err == nil && len(locationData) > 0 {
		locationStr := string(locationData)
		if locationStr != "" {
			// Parse location format: "Country|City" or "Country"
			parts := strings.Split(locationStr, "|")
			if len(parts) >= 2 && parts[1] != "" {
				// Has city: "People from [City] and other places are live!"
				message = fmt.Sprintf("🔥 People from %s and other places are live! Join the chat!", parts[1])
			} else if len(parts) >= 1 && parts[0] != "" {
				// Only country: "People from [Country] and other places are live!"
				message = fmt.Sprintf("🔥 People from %s and other places are live! Join the chat!", parts[0])
			} else {
				message = "🔥 People from different places are live! Join the chat!"
			}
		} else {
			message = "🔥 People from different places are live! Join the chat!"
		}
	} else {
		message = "🔥 People from different places are live! Join the chat!"
	}

	return message, "Live Now", map[string]interface{}{
		"type": "user_online",
		"user_id": 0,
	}
}

// TestGorushConnection tests the connection to Gorush
func (s *PushNotificationService) TestGorushConnection() error {
	url := fmt.Sprintf("%s/api/stat/go", s.gorushURL)
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("failed to connect to Gorush: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Gorush health check failed with status %d", resp.StatusCode)
	}

	logrus.Info("Gorush connection test successful")
	return nil
}

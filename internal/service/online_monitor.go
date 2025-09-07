// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"fmt"
	"time"

	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OnlineMonitorService monitors user online status changes and triggers notifications
type OnlineMonitorService struct {
	db                *gorm.DB
	cache             core.AppCache
	pushNotification  *PushNotificationService
	lastCheckTime     int64
	checkInterval     time.Duration
	ticker            *time.Ticker
	stopChan          chan struct{}
	
	// Cache for database queries to avoid repeated DB calls every 30 seconds
	userCache         map[int64]*dbr.User        // Cache user info by ID
	userDevicesCache  map[int64][]*dbr.UserDevice // Cache user devices by user ID
	lastCacheUpdate   int64                      // When cache was last updated
	cacheExpiry       int64                      // Cache expiry time (5 minutes)
}

// NewOnlineMonitorService creates a new online monitor service
func NewOnlineMonitorService(db *gorm.DB, cache core.AppCache, pushNotification *PushNotificationService) *OnlineMonitorService {
	return &OnlineMonitorService{
		db:               db,
		cache:            cache,
		pushNotification: pushNotification,
		checkInterval:    30 * time.Second, // Check every 30 seconds
		stopChan:         make(chan struct{}),
		userCache:        make(map[int64]*dbr.User),
		userDevicesCache: make(map[int64][]*dbr.UserDevice),
		cacheExpiry:      300, // Cache expires after 5 minutes
	}
}

// CheckOnlineStatusChanges checks if there are any online users and sends notifications
func (s *OnlineMonitorService) CheckOnlineStatusChanges() error {
	now := time.Now().Unix()
	
	// Get all online users
	onlineUsers, err := s.getOnlineUsers()
	if err != nil {
		logrus.Errorf("Failed to get online users: %v", err)
		return err
	}

	logrus.Debugf("Found %d online users", len(onlineUsers))

	// If there are ANY online users, send notifications to ALL users
	if len(onlineUsers) > 0 {
		// Check if we should send notifications (avoid spam)
		if s.shouldSendOnlineNotification(now) {
			// Send ONE notification to ALL users about "someone is online"
			if err := s.sendOnlineNotificationsToAll(); err != nil {
				logrus.Errorf("Failed to send online notifications: %v", err)
			}
		}
	}

	s.lastCheckTime = now
	return nil
}

// getOnlineUsers gets all currently online user IDs
func (s *OnlineMonitorService) getOnlineUsers() ([]int64, error) {
	// Get all online user keys from Redis
	pattern := conf.PrefixOnlineUser + "*"
	keys, err := s.cache.Keys(pattern)
	if err != nil {
		return nil, err
	}

	var userIDs []int64
	for _, key := range keys {
		// Extract user ID from key (format: "paopao:onlineuser:123")
		if len(key) > len(conf.PrefixOnlineUser) {
			userIDStr := key[len(conf.PrefixOnlineUser):]
			if userID, err := parseUserID(userIDStr); err == nil {
				userIDs = append(userIDs, userID)
			}
		}
	}

	return userIDs, nil
}

// shouldSendOnlineNotification checks if we should send online notifications (spam prevention)
func (s *OnlineMonitorService) shouldSendOnlineNotification(now int64) bool {
	// First check - don't send notifications
	if s.lastCheckTime == 0 {
		return false
	}

	// Check if enough time has passed since last notification (spam prevention)
	timeSinceLastCheck := now - s.lastCheckTime
	if timeSinceLastCheck < 30 { // Less than 30 seconds since last check
		return false
	}

	return true
}

// sendOnlineNotificationsToAll sends notifications to ALL users that someone is online
func (s *OnlineMonitorService) sendOnlineNotificationsToAll() error {
	logrus.Info("Sending online notifications to all users - someone is online")

	// Send generic online notification to ALL users
	return s.pushNotification.SendContactOnlineNotification(0, "Someone")
}

// getUserFromCache gets user info from cache, updates cache if needed
func (s *OnlineMonitorService) getUserFromCache(userID int64) (*dbr.User, error) {
	now := time.Now().Unix()
	
	// Check if cache needs refresh
	if now-s.lastCacheUpdate > s.cacheExpiry {
		if err := s.refreshUserCache(); err != nil {
			logrus.Errorf("Failed to refresh user cache: %v", err)
			// Fall back to direct DB query
			var user dbr.User
			err := s.db.Where("id = ? AND is_del = ?", userID, 0).First(&user).Error
			return &user, err
		}
	}
	
	// Return from cache
	if user, exists := s.userCache[userID]; exists {
		return user, nil
	}
	
	// User not in cache, query directly
	var user dbr.User
	err := s.db.Where("id = ? AND is_del = ?", userID, 0).First(&user).Error
	if err == nil {
		// Add to cache for next time
		s.userCache[userID] = &user
	}
	return &user, err
}

// refreshUserCache loads all active users into cache
func (s *OnlineMonitorService) refreshUserCache() error {
	logrus.Debug("Refreshing user cache...")
	
	// Load all active users
	var users []dbr.User
	if err := s.db.Where("is_del = ?", 0).Find(&users).Error; err != nil {
		return err
	}
	
	// Update cache
	s.userCache = make(map[int64]*dbr.User)
	for i := range users {
		s.userCache[users[i].ID] = &users[i]
	}
	
	s.lastCacheUpdate = time.Now().Unix()
	logrus.Debugf("User cache refreshed with %d users", len(users))
	return nil
}

// StartMonitoring starts the online monitoring service
func (s *OnlineMonitorService) StartMonitoring() {
	logrus.Info("Starting online status monitoring...")
	
	// Run initial check
	if err := s.CheckOnlineStatusChanges(); err != nil {
		logrus.Errorf("Initial online status check failed: %v", err)
	}

	// Start periodic monitoring with proper stop mechanism
	s.ticker = time.NewTicker(s.checkInterval)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logrus.Errorf("Panic in online monitoring goroutine: %v", r)
			}
			s.ticker.Stop()
		}()
		
		for {
			select {
			case <-s.stopChan:
				logrus.Info("Online monitoring stopped")
				return
			case <-s.ticker.C:
				if err := s.CheckOnlineStatusChanges(); err != nil {
					logrus.Errorf("Online status check failed: %v", err)
				}
			}
		}
	}()
	
	logrus.Info("Online status monitoring started")
}

// StopMonitoring stops the online monitoring service
func (s *OnlineMonitorService) StopMonitoring() {
	logrus.Info("Stopping online status monitoring...")
	
	// Signal stop to the monitoring goroutine
	close(s.stopChan)
	
	// Stop the ticker
	if s.ticker != nil {
		s.ticker.Stop()
	}
	
	logrus.Info("Online status monitoring stopped")
}

// parseUserID parses user ID from string
func parseUserID(userIDStr string) (int64, error) {
	// Simple parsing - in production you might want more robust parsing
	var userID int64
	_, err := fmt.Sscanf(userIDStr, "%d", &userID)
	return userID, err
}

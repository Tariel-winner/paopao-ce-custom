// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Notification cache system with configurable TTLs
package service

import (
	"fmt"
	"time"

	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/sirupsen/logrus"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	NotificationTypeContactMatched NotificationType = "contact_matched"
	NotificationTypeContactOnline  NotificationType = "contact_online"
)

// NotificationCache tracks sent notifications to prevent spam using Redis
type NotificationCache struct {
	cache core.AppCache
}

// NewNotificationCache creates a new notification cache using existing Redis
func NewNotificationCache(appCache core.AppCache) *NotificationCache {
	return &NotificationCache{
		cache: appCache,
	}
}

// ShouldSendNotification checks if we should send a notification using Redis TTL
func (nc *NotificationCache) ShouldSendNotification(
	userID int64, 
	targetID int64, 
	notificationType NotificationType,
) bool {
	key := nc.generateKey(userID, targetID, notificationType)
	
	// Check if key exists in cache
	if nc.cache.Exist(key) {
		// Key exists, notification was sent recently
		logrus.Debugf("Skipping notification for key %s (type: %s) - sent recently", key, notificationType)
		return false
	}
	
	// Key doesn't exist, set it with TTL
	ttl := nc.getTTLForNotificationType(notificationType)
	err := nc.cache.Set(key, []byte("1"), int64(ttl.Seconds()))
	if err != nil {
		logrus.Errorf("Failed to set cache key %s: %v", key, err)
		return true // If cache fails, allow notification (fail-safe)
	}
	
	logrus.Debugf("Set notification cache key %s with TTL %v", key, ttl)
	return true
}

// generateKey creates a unique Redis key for notification tracking
func (nc *NotificationCache) generateKey(userID, targetID int64, notificationType NotificationType) string {
	return fmt.Sprintf("notif:%d:%d:%s", userID, targetID, notificationType)
}

// getTTLForNotificationType returns the TTL duration for each notification type
func (nc *NotificationCache) getTTLForNotificationType(notificationType NotificationType) time.Duration {
	switch notificationType {
	case NotificationTypeContactMatched:
		// Contact matched: use configured TTL from cache settings
		return time.Duration(conf.CacheSetting.ContactMatchedExpire) * time.Second
	case NotificationTypeContactOnline:
		// Contact online: use configured TTL from cache settings
		return time.Duration(conf.CacheSetting.ContactOnlineExpire) * time.Second
	default:
		// Default: use contact matched TTL as fallback
		return time.Duration(conf.CacheSetting.ContactMatchedExpire) * time.Second
	}
}

// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"github.com/Masterminds/semver/v3"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/dao/cache"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
)

// ContactPushService integrates contact matching, push notifications, and online monitoring
type ContactPushService struct {
	contactMatching    *ContactMatchingService
	pushNotification   *PushNotificationService
	onlineMonitor      *OnlineMonitorService
}

func (s *ContactPushService) Name() string {
	return "ContactPush"
}

func (s *ContactPushService) Version() *semver.Version {
	return semver.MustParse("v1.0.0")
}

func (s *ContactPushService) OnInit() error {
	logrus.Info("Initializing ContactPush service...")
	
	// Get database connection
	db := conf.MustGormDB()
	
	// Initialize cache first
	appCache := cache.NewAppCache()
	
	// Initialize services
	s.contactMatching = NewContactMatchingService(db)
	s.pushNotification = NewPushNotificationService(db, "http://gorush:8088", appCache)
	
	// Initialize online monitor service
	s.onlineMonitor = NewOnlineMonitorService(db, appCache, s.pushNotification)
	
	logrus.Info("ContactPush service initialized successfully")
	return nil
}

func (s *ContactPushService) OnStart() error {
	logrus.Info("Starting ContactPush service...")
	
	// Contact matching DISABLED for now - only online user monitoring runs
	// Register contact matching cron job using existing infrastructure
	// This will run every 5 minutes and handle both matching and notifications
	// schedule, err := cron.ParseStandard(conf.JobManagerSetting.ContactMatchingInterval)
	// if err != nil {
	//	logrus.Errorf("Failed to parse cron schedule: %v", err)
	//	return err
	// }
	// events.OnTask(schedule, func() {
	//	logrus.Info("Running contact matching job...")
	//	
	//	// Run contact matching
	//	matched, err := s.contactMatching.MatchAllContacts()
	//	if err != nil {
	//		logrus.Errorf("Contact matching job failed: %v", err)
	//		return
	//	}
	//	
	//	logrus.Infof("Contact matching job completed, matched %d contacts", matched)
	//	
	//	// Contact matching notifications DISABLED for now
	//	// if matched > 0 {
	//	//	if err := s.sendNotificationsForNewMatches(); err != nil {
	//	//		logrus.Errorf("Failed to send notifications for new matches: %v", err)
	//	//	}
	//	// }
	// })
	
	// Test Gorush connection
	if err := s.pushNotification.TestGorushConnection(); err != nil {
		logrus.Warnf("Gorush connection test failed: %v", err)
	} else {
		logrus.Info("Gorush connection test successful")
	}
	
	// Start online monitoring
	s.onlineMonitor.StartMonitoring()
	
	logrus.Info("ContactPush service started successfully")
	return nil
}

func (s *ContactPushService) OnStop() error {
	logrus.Info("Stopping ContactPush service...")
	
	// Stop online monitoring
	if s.onlineMonitor != nil {
		s.onlineMonitor.StopMonitoring()
	}
	
	logrus.Info("ContactPush service stopped successfully")
	return nil
}

// sendNotificationsForNewMatches sends notifications for newly matched contacts
func (s *ContactPushService) sendNotificationsForNewMatches() error {
	logrus.Info("Processing notifications for newly matched contacts...")
	
	// Get newly matched contacts from the last run
	newlyMatchedContacts, err := s.contactMatching.GetNewlyMatchedContactsForNotification()
	if err != nil {
		return err
	}
	
	if len(newlyMatchedContacts) == 0 {
		logrus.Debug("No newly matched contacts to notify about")
		return nil
	}
	
	logrus.Infof("Found %d newly matched contacts to notify about", len(newlyMatchedContacts))
	
	// Send notifications for each newly matched contact
	for _, contact := range newlyMatchedContacts {
		if contact.MatchedUserID == nil {
			logrus.Warnf("Contact %s has no matched user ID, skipping notification", contact.ContactName)
			continue
		}
		
		// Get the matched user's information
		var matchedUser dbr.User
		if err := s.contactMatching.db.Where("id = ? AND is_del = ?", *contact.MatchedUserID, 0).First(&matchedUser).Error; err != nil {
			logrus.Errorf("Failed to get matched user info for ID %d: %v", *contact.MatchedUserID, err)
			continue
		}
		
		// Send "contact found" notification to the contact owner
		if err := s.pushNotification.SendContactMatchedNotification(contact.UserId, matchedUser.Username); err != nil {
			logrus.Errorf("Failed to send contact matched notification for %s: %v", contact.ContactName, err)
		} else {
			logrus.Infof("Sent contact matched notification for %s to user %d", contact.ContactName, contact.UserId)
		}
	}
	
	logrus.Info("Completed processing notifications for newly matched contacts")
	return nil
}

func newContactPushService() Service {
	return &ContactPushService{}
}

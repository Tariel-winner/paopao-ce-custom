// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"time"

	"github.com/nyaruka/phonenumbers"

	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// ContactMatchingService handles iPhone contact matching with app users
type ContactMatchingService struct {
	db *gorm.DB
}

// NewContactMatchingService creates a new contact matching service
func NewContactMatchingService(db *gorm.DB) *ContactMatchingService {
	return &ContactMatchingService{
		db: db,
	}
}

// MatchAllContacts processes all unmatched contacts and tries to match them with app users
func (s *ContactMatchingService) MatchAllContacts() (int64, error) {
	logrus.Info("Starting contact matching job...")
	
	var matched int64
	now := time.Now().Unix()

	// Get all unmatched contacts
	var phoneContacts []dbr.UserPhoneContact
	if err := s.db.Where("is_matched = ? AND is_del = ?", false, 0).Find(&phoneContacts).Error; err != nil {
		logrus.Errorf("Failed to fetch unmatched contacts: %v", err)
		return 0, err
	}

	logrus.Infof("Found %d unmatched contacts to process", len(phoneContacts))

	for _, phoneContact := range phoneContacts {
		// Try to match this contact
		if isMatched, err := s.matchSingleContact(&phoneContact, now); err != nil {
			logrus.Errorf("Failed to match contact %s (%s): %v", phoneContact.ContactName, phoneContact.ContactPhone, err)
			continue
		} else if isMatched {
			matched++
		}
	}

	logrus.Infof("Contact matching job completed. Matched %d contacts", matched)
	return matched, nil
}

// matchSingleContact attempts to match a single phone contact with an app user
func (s *ContactMatchingService) matchSingleContact(phoneContact *dbr.UserPhoneContact, now int64) (bool, error) {
	// Normalize the phone number using go-libphonenumber
	normalizedPhone, err := s.normalizePhoneNumber(phoneContact.ContactPhone)
	if err != nil {
		logrus.Debugf("Failed to normalize phone %s: %v", phoneContact.ContactPhone, err)
		return false, nil // Don't fail the job for invalid phone numbers
	}

	// Try to find app user with matching phone number
	var appUser dbr.User
	if err := s.db.Where("phone = ? AND is_del = ?", normalizedPhone, 0).First(&appUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// No match found, this is normal
			return false, nil
		}
		// Database error
		return false, err
	}

	// Found matching user! Update the contact
	phoneContact.IsMatched = true
	phoneContact.MatchedUserID = &appUser.ID
	phoneContact.ModifiedOn = now

	if err := phoneContact.Update(s.db); err != nil {
		return false, err
	}

	logrus.Infof("Matched contact %s (%s) with user %s (ID: %d)", 
		phoneContact.ContactName, phoneContact.ContactPhone, appUser.Username, appUser.ID)
	
	return true, nil
}

// normalizePhoneNumber normalizes a phone number using go-libphonenumber
func (s *ContactMatchingService) normalizePhoneNumber(phone string) (string, error) {
	// Try to parse as international number first
	parsedNumber, err := phonenumbers.Parse(phone, "")
	if err != nil {
		// If that fails, try with US as default country
		parsedNumber, err = phonenumbers.Parse(phone, "US")
		if err != nil {
			return phone, err // Return original if we can't parse
		}
	}

	// Format as E.164 (international format)
	normalized := phonenumbers.Format(parsedNumber, phonenumbers.E164)
	
	// Remove the + prefix for database storage consistency
	if len(normalized) > 0 && normalized[0] == '+' {
		normalized = normalized[1:]
	}
	
	return normalized, nil
}

// GetNewlyMatchedContactsForNotification returns contacts that were matched in the last minute
// This is used by the notification system to send "contact found" notifications
func (s *ContactMatchingService) GetNewlyMatchedContactsForNotification() ([]*dbr.UserPhoneContact, error) {
	// Get contacts matched in the last minute (testing mode - was 1 hour)
	oneMinuteAgo := time.Now().Add(-1 * time.Minute).Unix()
	
	var contacts []*dbr.UserPhoneContact
	err := s.db.Where("is_matched = ? AND modified_on > ? AND is_del = ?", true, oneMinuteAgo, 0).Find(&contacts).Error
	
	return contacts, err
}

// NOTE: StartContactMatchingCron is deprecated - use events.OnTask() in ContactPushService instead

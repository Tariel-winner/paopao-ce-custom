// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
	"time"

	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

var (
	_ core.ContactManageService = (*contactManageSrv)(nil)
)

type contactManageSrv struct {
	db *gorm.DB
}

func newContactManageService(db *gorm.DB) core.ContactManageService {
	return &contactManageSrv{
		db: db,
	}
}

func (s *contactManageSrv) fetchOrNewContact(db *gorm.DB, userId int64, friendId int64, status int8) (*dbr.Contact, error) {
	contact := &dbr.Contact{
		UserId:   userId,
		FriendId: friendId,
	}
	contact, err := contact.FetchUser(db)
	if err != nil {
		contact = &dbr.Contact{
			UserId:   userId,
			FriendId: friendId,
			Status:   status,
		}
		if contact, err = contact.Create(db); err != nil {
			logrus.Errorf("contactManageSrv.fetchOrNewContact create new contact err:%s", err)
			return nil, err
		}
	}
	return contact, nil
}

func (s *contactManageSrv) RequestingFriend(userId int64, friendId int64, greetings string) (err error) {
	db := s.db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	contact, e := s.fetchOrNewContact(db, userId, friendId, dbr.ContactStatusRequesting)
	if e != nil {
		err = e
		return
	}

	// 如果已经好友，啥也不干
	if contact.Status == dbr.ContactStatusAgree {
		return nil
	} else if contact.Status == dbr.ContactStatusReject || contact.Status == dbr.ContactStatusDeleted {
		contact.Status = dbr.ContactStatusRequesting
		contact.IsDel = 0 // remove deleted flag if needed
		if err = contact.UpdateInUnscoped(db); err != nil {
			logrus.Errorf("contactManageSrv.RequestingFriend update exsit contact err:%s", err)
			return
		}
	}

	msg := &dbr.Message{
		SenderUserID:   userId,
		ReceiverUserID: friendId,
		Type:           dbr.MsgTypeRequestingFriend,
		Brief:          "请求添加好友，并附言:",
		Content:        greetings,
		ReplyID:        int64(dbr.ContactStatusRequesting),
	}
	if _, err = msg.Create(db); err != nil {
		logrus.Errorf("contactManageSrv.RequestingFriend create message err:%s", err)
		return
	}
	return nil
}

func (s *contactManageSrv) AddFriend(userId int64, friendId int64) (err error) {
	db := s.db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	contact := &dbr.Contact{
		UserId:   friendId,
		FriendId: userId,
	}
	if contact, err = contact.GetByUserFriend(db); err != nil {
		return
	}
	// 如果还不是请求好友，啥也不干
	if contact.Status != dbr.ContactStatusRequesting {
		logrus.Debugf("contactManageSrv.AddFriend not reuesting status now so skip")
		return nil
	}
	contact.Status = dbr.ContactStatusAgree
	if err = contact.Update(db); err != nil {
		return err
	}

	contact, err = s.fetchOrNewContact(db, userId, friendId, dbr.ContactStatusAgree)
	if err != nil {
		return
	}

	// 如果已经好友，啥也不干
	if contact.Status != dbr.ContactStatusAgree {
		contact.Status = dbr.ContactStatusAgree
		contact.IsDel = 0 // remove deleted flag
		if err = contact.UpdateInUnscoped(db); err != nil {
			logrus.Errorf("contactManageSrv.AddFriend update contact err:%s", err)
			return
		}
	}

	args := []any{userId, friendId, friendId, userId, dbr.MsgTypeRequestingFriend, dbr.ContactStatusRequesting}
	msgs, e := (&dbr.Message{}).FetchBy(db, dbr.Predicates{
		"((sender_user_id = ? AND receiver_user_id = ?) OR (sender_user_id = ? AND receiver_user_id = ?)) AND type = ? AND reply_id = ?": args,
	})
	if e != nil {
		err = e
		return
	}
	for _, msg := range msgs {
		msg.ReplyID = int64(dbr.ContactStatusAgree)
		if err = msg.Update(db); err != nil {
			return
		}
	}
	return nil
}

func (s *contactManageSrv) RejectFriend(userId int64, friendId int64) (err error) {
	db := s.db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	contact := &dbr.Contact{
		UserId:   friendId,
		FriendId: userId,
	}
	if contact, err = contact.GetByUserFriend(db); err != nil {
		return
	}
	// 如果还不是请求好友，啥也不干
	if contact.Status != dbr.ContactStatusRequesting {
		return nil
	}
	contact.Status = dbr.ContactStatusReject
	if err = contact.Update(db); err != nil {
		return err
	}

	args := []any{friendId, userId, dbr.MsgTypeRequestingFriend, dbr.ContactStatusRequesting}
	msgs, e := (&dbr.Message{}).FetchBy(db, dbr.Predicates{
		"sender_user_id = ? AND receiver_user_id = ? AND type = ? AND reply_id = ?": args,
	})
	if e != nil {
		err = e
		return
	}
	for _, msg := range msgs {
		msg.ReplyID = int64(dbr.ContactStatusReject)
		if err = msg.Update(db); err != nil {
			return
		}
	}
	return nil
}

func (s *contactManageSrv) DeleteFriend(userId int64, friendId int64) (err error) {
	db := s.db.Begin()
	defer func() {
		if err == nil {
			db.Commit()
		} else {
			db.Rollback()
		}
	}()

	contact := &dbr.Contact{
		UserId:   userId,
		FriendId: friendId,
	}
	contacts, e := contact.FetchByUserFriendAll(db)
	if e != nil {
		return e
	}

	for _, contact := range contacts {
		// 如果还不是好友，啥也不干
		if contact.Status != dbr.ContactStatusAgree {
			continue
		}
		contact.Status = dbr.ContactStatusDeleted
		contact.DeletedOn = time.Now().Unix()
		contact.IsDel = 1
		if err = contact.Update(db); err != nil {
			return
		}
	}
	return nil
}

func (s *contactManageSrv) GetContacts(userId int64, offset int, limit int) (*ms.ContactList, error) {
	contact := &dbr.Contact{}
	condition := dbr.ConditionsT{
		"user_id": userId,
		"status":  dbr.ContactStatusAgree,
	}
	contacts, err := contact.List(s.db, condition, offset, limit)
	if err != nil {
		return nil, err
	}
	total, err := contact.Count(s.db, condition)
	if err != nil {
		return nil, err
	}
	resp := &ms.ContactList{
		Contacts: make([]ms.ContactItem, 0, len(contacts)),
		Total:    total,
	}
	for _, c := range contacts {
		if c.User != nil {
			resp.Contacts = append(resp.Contacts, ms.ContactItem{
				UserId:    c.FriendId,
				Username:  c.User.Username,
				Nickname:  c.User.Nickname,
				Avatar:    c.User.Avatar,
				Phone:     c.User.Phone,
				CreatedOn: c.User.CreatedOn,
			})
		}
	}
	return resp, nil
}

func (s *contactManageSrv) IsFriend(userId int64, friendId int64) bool {
	contact := &dbr.Contact{
		UserId:   friendId,
		FriendId: userId,
	}
	contact, err := contact.GetByUserFriend(s.db)
	if err == nil && contact.Status == dbr.ContactStatusAgree {
		return true
	}
	return false
}

// UploadPhoneContacts uploads iPhone contacts and matches them with existing app users
func (s *contactManageSrv) UploadPhoneContacts(userID int64, contacts []cs.PhoneContact) (int64, int64, error) {
	db := s.db.Begin()
	defer func() {
		if err := recover(); err != nil {
			db.Rollback()
		}
	}()

	var uploaded int64
	var matched int64
	now := time.Now().Unix()

	// Upload contacts to p_user_phone_contacts table
	for _, contact := range contacts {
		phoneContact := &dbr.UserPhoneContact{
			UserId:       userID,
			ContactName:  contact.Name,
			ContactPhone: contact.Phone,
			ContactEmail: contact.Email,
			IsMatched:    false,
			CreatedOn:    now,
			ModifiedOn:   now,
		}

		if err := phoneContact.Create(db); err != nil {
			db.Rollback()
			return 0, 0, err
		}
		uploaded++
	}

	// Try to match contacts with existing app users
	matched, err := s.matchPhoneContacts(db, userID)
	if err != nil {
		db.Rollback()
		return uploaded, 0, err
	}

	if err := db.Commit().Error; err != nil {
		return uploaded, 0, err
	}

	return uploaded, matched, nil
}

// MatchPhoneContacts matches iPhone contacts with existing app users by phone number
func (s *contactManageSrv) MatchPhoneContacts(userID int64) (int64, error) {
	db := s.db.Begin()
	defer func() {
		if err := recover(); err != nil {
			db.Rollback()
		}
	}()

	matched, err := s.matchPhoneContacts(db, userID)
	if err != nil {
		db.Rollback()
		return 0, err
	}

	if err := db.Commit().Error; err != nil {
		return 0, err
	}

	return matched, nil
}

// matchPhoneContacts internal method to match contacts by phone number
func (s *contactManageSrv) matchPhoneContacts(db *gorm.DB, userID int64) (int64, error) {
	var matched int64

	// Get all unmatched contacts for this user
	var phoneContacts []dbr.UserPhoneContact
	if err := db.Where("user_id = ? AND is_matched = ?", userID, false).Find(&phoneContacts).Error; err != nil {
		return 0, err
	}

	for _, phoneContact := range phoneContacts {
		// Try to find app user with matching phone number
		var appUser dbr.User
		if err := db.Where("phone = ? AND is_del = ?", phoneContact.ContactPhone, 0).First(&appUser).Error; err == nil {
			// Found matching user, update the contact
			phoneContact.IsMatched = true
			phoneContact.MatchedUserID = &appUser.ID
			phoneContact.ModifiedOn = time.Now().Unix()

			if err := phoneContact.Update(db); err != nil {
				return matched, err
			}
			matched++
		}
	}

	return matched, nil
}

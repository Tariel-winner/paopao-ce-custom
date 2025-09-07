// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
    "gorm.io/gorm"
    "time"
)

type Queue struct {
    ID            int64    `json:"id"`
    Name          string   `json:"name"`
    Description   string   `json:"description"`
    IsClosed      bool     `json:"is_closed"`
    Participants  []int64  `json:"participants" gorm:"type:jsonb;default:'[]';serializer:json"`
}

type Room struct {
    *Model
    HostID            int64    `json:"host_id"`
    HMSRoomID         string   `json:"hms_room_id"`
    SpeakerIDs        []int64  `json:"speaker_ids" gorm:"type:jsonb;default:'[]';serializer:json"`
    StartTime         int64    `json:"start_time"`
    Queue             *Queue   `json:"queue" gorm:"type:jsonb;serializer:json"`
    IsBlockedFromSpace int16   `json:"is_blocked_from_space"`
    Topics            []string `json:"topics" gorm:"type:jsonb;default:'[]';serializer:json"`
    Categories        Int64Array `json:"categories" gorm:"type:integer[];default:'{}'"`
}

type RoomFormated struct {
    ID                int64     `json:"id"`
    HostID            int64     `json:"host_id"`
    HMSRoomID         string    `json:"hms_room_id"`
    SpeakerIDs        []int64   `json:"speaker_ids"`
    StartTime         int64     `json:"start_time"`
    Queue             *Queue    `json:"queue"`
    IsBlockedFromSpace int16    `json:"is_blocked_from_space"`
    Topics            []string  `json:"topics"`
    Categories        []int64   `json:"categories"`
    CreatedOn         int64     `json:"created_on"`
    ModifiedOn        int64     `json:"modified_on"`
}

func (r *Room) Format() *RoomFormated {
    if r.Model != nil {
        return &RoomFormated{
            ID:                r.ID,
            HostID:            r.HostID,
            HMSRoomID:         r.HMSRoomID,
            SpeakerIDs:        r.SpeakerIDs,
            StartTime:         r.StartTime,
            Queue:             r.Queue,
            IsBlockedFromSpace: r.IsBlockedFromSpace,
            Topics:            r.Topics,
            Categories:        []int64(r.Categories),
            CreatedOn:         r.CreatedOn,
            ModifiedOn:        r.ModifiedOn,
        }
    }
    return nil
}

func (r *Room) Create(db *gorm.DB) (*Room, error) {
    err := db.Create(&r).Error
    if err != nil {
        return r, err
    }
    return r, nil
}

func (r *Room) Delete(db *gorm.DB) error {
    return db.Model(r).Where("id = ?", r.Model.ID).Updates(map[string]any{
        "deleted_on": time.Now().Unix(),
        "is_del":     1,
    }).Error
}

func (r *Room) Get(db *gorm.DB) (*Room, error) {
    var room Room
    if r.Model != nil && r.ID > 0 {
        db = db.Where("id = ? AND is_del = ?", r.ID, 0)
    } else {
        return nil, gorm.ErrRecordNotFound
    }

    err := db.First(&room).Error
    if err != nil {
        return &room, err
    }

    return &room, nil
}

func (r *Room) GetByHostID(db *gorm.DB) (*Room, error) {
    var room Room
    err := db.Where("host_id = ? AND is_del = ?", r.HostID, 0).First(&room).Error
    if err != nil {
        return nil, err
    }
    return &room, nil
}

func (r *Room) List(db *gorm.DB, conditions ConditionsT, offset, limit int) ([]*Room, error) {
    var rooms []*Room
    var err error
    
    // Start with the base query
    db = db.Model(&Room{}).Where("is_del = ?", 0)
    
    // Apply conditions (including ORDER)
    for k, v := range conditions {
        if k == "ORDER" {
            db = db.Order(v)
        } else {
            db = db.Where(k, v)
        }
    }
    
    // Apply pagination LAST (after ORDER)
    if offset >= 0 && limit > 0 {
        db = db.Offset(offset).Limit(limit)
    }

    if err = db.Find(&rooms).Error; err != nil {
        return nil, err
    }

    return rooms, nil
}

func (r *Room) Count(db *gorm.DB, conditions ConditionsT) (int64, error) {
    var count int64
    
    // Start with the base query
    db = db.Model(&Room{}).Where("is_del = ?", 0)
    
    // Apply conditions (excluding ORDER for count)
    for k, v := range conditions {
        if k != "ORDER" {
            db = db.Where(k, v)
        }
    }
    
    if err := db.Count(&count).Error; err != nil {
        return 0, err
    }
    return count, nil
}

func (r *Room) Update(db *gorm.DB) error {
	return db.Model(&Room{}).Where("id = ? AND is_del = ?", r.Model.ID, 0).Updates(r).Error
}

// UpdateCategoriesByHostID updates categories for all rooms where the user is the host
func (r *Room) UpdateCategoriesByHostID(db *gorm.DB, hostID int64, categoryIDs []int64) error {
	return db.Model(&Room{}).
		Where("host_id = ? AND is_del = ?", hostID, 0).
		Update("categories", Int64Array(categoryIDs)).Error
} 
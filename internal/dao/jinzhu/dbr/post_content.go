// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"time"

	"gorm.io/gorm"
)

// 类型，1标题，2文字段落，3图片地址，4视频地址，5语音地址，6链接地址，7附件资源
type PostContentT int

const (
	ContentTypeTitle PostContentT = iota + 1
	ContentTypeText
	ContentTypeImage
	ContentTypeVideo
	ContentTypeAudio
	ContentTypeLink
	ContentTypeAttachment
	ContentTypeChargeAttachment
)

var (
	mediaContentType = []PostContentT{
		ContentTypeImage,
		ContentTypeVideo,
		ContentTypeAudio,
		ContentTypeAttachment,
		ContentTypeChargeAttachment,
	}
)

type PostContent struct {
	*Model
	PostID   int64        `json:"post_id"`
	UserID   []int64      `json:"user_id" gorm:"type:jsonb;default:'[0,0]';serializer:json"`
	Content  string       `json:"content"`
	Type     PostContentT `json:"type"`
	Sort     int64        `json:"sort"`
	RoomID   string       `json:"room_id"`
	Duration string       `json:"duration"`
	Size     string       `json:"size"`
}

// SetUserIDs sets the UserID field from a slice of int64
func (p *PostContent) SetUserIDs(ids []int64) {
	p.UserID = ids
}

// GetUserIDs returns the UserID field as a slice of int64
func (p *PostContent) GetUserIDs() []int64 {
	return p.UserID
}

type PostContentFormated struct {
	ID       int64        `db:"id" json:"id"`
	PostID   int64        `json:"post_id"`
	Content  string       `json:"content"`
	Type     PostContentT `json:"type"`
	Sort     int64        `json:"sort"`
	Duration string       `json:"duration,omitempty"`
	Size     string       `json:"size,omitempty"`
}

func (p *PostContent) DeleteByPostId(db *gorm.DB, postId int64) error {
	return db.Model(p).Where("post_id = ?", postId).Updates(map[string]any{
		"deleted_on": time.Now().Unix(),
		"is_del":     1,
	}).Error
}

func (p *PostContent) MediaContentsByPostId(db *gorm.DB, postId int64) (contents []string, err error) {
	err = db.Model(p).Where("post_id = ? AND type IN ?", postId, mediaContentType).Select("content").Find(&contents).Error
	return
}

func (p *PostContent) Create(db *gorm.DB) (*PostContent, error) {
	err := db.Create(&p).Error

	return p, err
}

func (p *PostContent) Format() *PostContentFormated {
	if p.Model == nil {
		return nil
	}
	return &PostContentFormated{
		ID:       p.ID,
		PostID:   p.PostID,
		Content:  p.Content,
		Type:     p.Type,
		Sort:     p.Sort,
		Duration: p.Duration,
		Size:     p.Size,
	}
}

func (p *PostContent) List(db *gorm.DB, conditions *ConditionsT, offset, limit int) ([]*PostContent, error) {
	var contents []*PostContent
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if p.PostID > 0 {
		db = db.Where("id = ?", p.PostID)
	}

	for k, v := range *conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&contents).Error; err != nil {
		return nil, err
	}

	return contents, nil
}

func (p *PostContent) Get(db *gorm.DB) (*PostContent, error) {
	var content PostContent
	if p.Model != nil && p.ID > 0 {
		db = db.Where("id = ? AND is_del = ?", p.ID, 0)
	} else {
		return nil, gorm.ErrRecordNotFound
	}

	err := db.First(&content).Error
	if err != nil {
		return &content, err
	}

	return &content, nil
}

func (p *PostContent) UpdateContentByRoomId(db *gorm.DB, roomId string, content string) error {
	return db.Model(p).Where("room_id = ? AND is_del = ?", roomId, 0).Update("content", content).Error
}



// UpdateAudioContentByRoomId updates all audio-related fields in a single database call
func (p *PostContent) UpdateAudioContentByRoomId(db *gorm.DB, roomId string, content string, duration string, size string) error {
	return db.Model(p).Where("room_id = ? AND is_del = ?", roomId, 0).Updates(map[string]interface{}{
		"content":  content,
		"duration": duration,
		"size":     size,
	}).Error
}

// UpdateAudioContentByPostID updates audio content for a specific post
func (p *PostContent) UpdateAudioContentByPostID(db *gorm.DB, postID int64, content string, duration string, size string) error {
	return db.Model(p).Where("post_id = ? AND type = ? AND is_del = ?", postID, ContentTypeAudio, 0).Updates(map[string]interface{}{
		"content":  content,
		"duration": duration,
		"size":     size,
	}).Error
}

func (p *PostContent) GetPostContentByRoomID(db *gorm.DB, roomId string) (*PostContent, error) {
	var content PostContent
	err := db.Where("room_id = ? AND is_del = ?", roomId, 0).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// GetPostContentBySessionID finds post content where the content field contains the session ID
func (p *PostContent) GetPostContentBySessionID(db *gorm.DB, roomId string, sessionId string) (*PostContent, error) {
	var content PostContent
	err := db.Where("room_id = ? AND content = ? AND is_del = ?", roomId, sessionId, 0).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

// GetAudioContentByPostID finds the audio content for a specific post
func (p *PostContent) GetAudioContentByPostID(db *gorm.DB, postID int64) (*PostContent, error) {
	var content PostContent
	err := db.Where("post_id = ? AND type = ? AND is_del = ?", postID, ContentTypeAudio, 0).First(&content).Error
	if err != nil {
		return nil, err
	}
	return &content, nil
}

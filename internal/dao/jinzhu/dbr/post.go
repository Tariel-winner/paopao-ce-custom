// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"strings"
	"time"

	"gorm.io/gorm"
	"github.com/sirupsen/logrus"
)

// PostVisibleT 可访问类型，可见性: 0私密 10充电可见 20订阅可见 30保留 40保留 50好友可见 60关注可见 70保留 80保留 90公开',
type PostVisibleT uint8

const (
	PostVisitPublic    PostVisibleT = 90
	PostVisitPrivate   PostVisibleT = 0
	PostVisitFriend    PostVisibleT = 50
	PostVisitFollowing PostVisibleT = 60
)

type PostByMedia = Post

type PostByComment = Post

// LocationData represents location information from iOS
type LocationData struct {
	Name    string  `json:"name"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Address string  `json:"address"`
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
}
 
type Post struct {
	*Model
	UserID          []int64        `json:"user_id" gorm:"type:jsonb;default:'[0,0]';serializer:json"`
	CommentCount    int64        `json:"comment_count"`
	CollectionCount int64        `json:"collection_count"`
	ShareCount      int64        `json:"share_count"`
	UpvoteCount     int64        `json:"upvote_count"`
	Visibility      PostVisibleT `json:"visibility"`
	IsTop           int          `json:"is_top"`
	IsEssence       int          `json:"is_essence"`
	IsLock          int          `json:"is_lock"`
	LatestRepliedOn int64        `json:"latest_replied_on"`
	Tags            string       `json:"tags"`
	AttachmentPrice int64        `json:"attachment_price"`
	IP              string       `json:"ip"`
	IPLoc           string       `json:"ip_loc"`
	RoomID          string       `json:"room_id"`
	SessionID       string       `json:"session_id"`
	// Location fields
	LocationName    string  `json:"location_name"`
	LocationLat     float64 `json:"location_lat"`
	LocationLng     float64 `json:"location_lng"`
	LocationAddress string  `json:"location_address"`
	LocationCity    string  `json:"location_city"`
	LocationState   string  `json:"location_state"`
	LocationCountry string  `json:"location_country"`
}

type PostFormated struct {
	ID              int64                  `json:"id"`
	UserID          []int64                `json:"user_id"`
	User            *UserFormated          `json:"user"`
	Visitor         *UserFormated          `json:"visitor"`
	Contents        []*PostContentFormated `json:"contents"`
	CommentCount    int64                  `json:"comment_count"`
	CollectionCount int64                  `json:"collection_count"`
	ShareCount      int64                  `json:"share_count"`
	UpvoteCount     int64                  `json:"upvote_count"`
	Visibility      PostVisibleT           `json:"visibility"`
	IsTop           int                    `json:"is_top"`
	IsEssence       int                    `json:"is_essence"`
	IsLock          int                    `json:"is_lock"`
	LatestRepliedOn int64                  `json:"latest_replied_on"`
	CreatedOn       int64                  `json:"created_on"`
	ModifiedOn      int64                  `json:"modified_on"`
	Tags            map[string]int8        `json:"tags"`
	AttachmentPrice int64                  `json:"attachment_price"`
	IPLoc           string                 `json:"ip_loc"`
	RoomID          string                 `json:"room_id"`
	SessionID       string                 `json:"session_id"`
	// Location fields
	LocationName    string  `json:"location_name"`
	LocationLat     float64 `json:"location_lat"`
	LocationLng     float64 `json:"location_lng"`
	LocationAddress string  `json:"location_address"`
	LocationCity    string  `json:"location_city"`
	LocationState   string  `json:"location_state"`
	LocationCountry string  `json:"location_country"`
}

func (t PostVisibleT) ToOutValue() (res uint8) {
	switch t {
	case PostVisitPublic:
		res = 0
	case PostVisitPrivate:
		res = 1
	case PostVisitFriend:
		res = 2
	case PostVisitFollowing:
		res = 3
	default:
		res = 1
	}
	return
}

func (p *Post) GetHostID() int64 {
    if len(p.UserID) > 0 {
        return p.UserID[0]
    }
    return 0
}

func (p *Post) GetVisitorID() int64 {
    if len(p.UserID) > 1 {
        return p.UserID[1]
    }
    return 0
}

func (p *Post) Format() *PostFormated {
	if p.Model != nil {
		tagsMap := map[string]int8{}
		for _, tag := range strings.Split(p.Tags, ",") {
			tagsMap[tag] = 1
		}
		return &PostFormated{
			ID:              p.ID,
			UserID:          p.UserID,
			User:            &UserFormated{},
			Visitor:         &UserFormated{},
			Contents:        []*PostContentFormated{},
			CommentCount:    p.CommentCount,
			CollectionCount: p.CollectionCount,
			ShareCount:      p.ShareCount,
			UpvoteCount:     p.UpvoteCount,
			Visibility:      p.Visibility,
			IsTop:           p.IsTop,
			IsEssence:       p.IsEssence,
			IsLock:          p.IsLock,
			LatestRepliedOn: p.LatestRepliedOn,
			CreatedOn:       p.CreatedOn,
			ModifiedOn:      p.ModifiedOn,
			AttachmentPrice: p.AttachmentPrice,
			Tags:            tagsMap,
			IPLoc:           p.IPLoc,
			RoomID:          p.RoomID,
			SessionID:       p.SessionID,
			// Location fields
			LocationName:    p.LocationName,
			LocationLat:     p.LocationLat,
			LocationLng:     p.LocationLng,
			LocationAddress: p.LocationAddress,
			LocationCity:    p.LocationCity,
			LocationState:   p.LocationState,
			LocationCountry: p.LocationCountry,
		}
	}

	return nil
}

func (p *Post) Create(db *gorm.DB) (*Post, error) {
	// Debug: Print the exact data being passed
	logrus.Debugf("DEBUG Create: post=%+v", p)
	logrus.Debugf("DEBUG Create: user_ids=%v (type: %T)", p.UserID, p.UserID)
	logrus.Debugf("DEBUG Create: model=%+v", p.Model)
	logrus.Infof("DEBUG Create: Location fields - name=%s, lat=%f, lng=%f, city=%s, country=%s", 
		p.LocationName, p.LocationLat, p.LocationLng, p.LocationCity, p.LocationCountry)

	err := db.Create(&p).Error
	if err != nil {
		logrus.Errorf("DEBUG Create Error: %v", err)
		return p, err
	}

	logrus.Debugf("DEBUG Create Success: created post=%+v", p)
	logrus.Infof("DEBUG Create Success: Location fields after save - name=%s, lat=%f, lng=%f, city=%s, country=%s", 
		p.LocationName, p.LocationLat, p.LocationLng, p.LocationCity, p.LocationCountry)
	return p, nil
}

func (s *Post) Delete(db *gorm.DB) error {
	return db.Model(s).Where("id = ?", s.Model.ID).Updates(map[string]any{
		"deleted_on": time.Now().Unix(),
		"is_del":     1,
	}).Error
}

func (p *Post) Get(db *gorm.DB) (*Post, error) {
	var post Post
	if p.Model != nil && p.ID > 0 {
		db = db.Where("id = ? AND is_del = ?", p.ID, 0)
	} else {
		return nil, gorm.ErrRecordNotFound
	}

	err := db.First(&post).Error
	if err != nil {
		return &post, err
	}

	return &post, nil
}

func (p *Post) List(db *gorm.DB, conditions ConditionsT, offset, limit int) ([]*Post, error) {
	var posts []*Post
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if len(p.UserID) > 0 {
		db = db.Where("user_id[1] = ?", p.GetHostID())
	}
	for k, v := range conditions {
		if k == "ORDER" {
			db = db.Order(v)
		} else {
			db = db.Where(k, v)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *Post) Fetch(db *gorm.DB, predicates Predicates, offset, limit int) ([]*Post, error) {
	var posts []*Post
	var err error
	if offset >= 0 && limit > 0 {
		db = db.Offset(offset).Limit(limit)
	}
	if len(p.UserID) > 0 {
		db = db.Where("user_id[1] = ?", p.GetHostID())
	}
	for query, args := range predicates {
		if query == "ORDER" {
			db = db.Order(args[0])
		} else {
			db = db.Where(query, args...)
		}
	}

	if err = db.Where("is_del = ?", 0).Find(&posts).Error; err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *Post) CountBy(db *gorm.DB, predicates Predicates) (count int64, err error) {
	for query, args := range predicates {
		if query != "ORDER" {
			db = db.Where(query, args...)
		}
	}
	err = db.Model(p).Count(&count).Error
	return
}

func (p *Post) Count(db *gorm.DB, conditions ConditionsT) (int64, error) {
	var count int64
	if len(p.UserID) > 0 {
		db = db.Where("user_id[1] = ?", p.GetHostID())
	}
	for k, v := range conditions {
		if k != "ORDER" {
			db = db.Where(k, v)
		}
	}
	if err := db.Model(p).Count(&count).Error; err != nil {
		return 0, err
	}

	return count, nil
}

func (p *Post) Update(db *gorm.DB) error {
	return db.Model(&Post{}).Where("id = ? AND is_del = ?", p.Model.ID, 0).Save(p).Error
}

// GetPostLocation calculates the page and position of a post in a user's timeline
// This method uses the exact same query as getUserPosts to ensure consistency
func (p *Post) GetPostLocation(db *gorm.DB, postID int64, userID int64, pageSize int) (page int, position int, totalPosts int64, err error) {
	// First, get the target post to get its properties
	var targetPost Post
	if err = db.Where("id = ? AND is_del = ?", postID, 0).First(&targetPost).Error; err != nil {
		return 0, 0, 0, err
	}

	// Check if the post belongs to the specified user
	if len(targetPost.UserID) == 0 || targetPost.GetHostID() != userID {
		return 0, 0, 0, gorm.ErrRecordNotFound
	}

	// Count total posts for this user with visibility filtering (same as getUserPosts)
	var total int64
	if err = db.Model(&Post{}).
		Where("CAST(user_id->0 AS bigint) = ? AND is_del = ? AND visibility >= ?", userID, 0, PostVisitPublic).
		Count(&total).Error; err != nil {
		return 0, 0, 0, err
	}

	// Use a simple approach: get all posts in the same order as getUserPosts and find our post
	var allPosts []Post
	if err = db.Model(&Post{}).
		Where("CAST(user_id->0 AS bigint) = ? AND is_del = ? AND visibility >= ?", userID, 0, PostVisitPublic).
		Order("is_top DESC, latest_replied_on DESC").
		Find(&allPosts).Error; err != nil {
		return 0, 0, 0, err
	}

	// Find the position of our target post in the ordered list
	var positionInTimeline int64 = 0
	for i, post := range allPosts {
		if post.ID == postID {
			positionInTimeline = int64(i + 1) // 1-based position
			break
		}
	}

	if positionInTimeline == 0 {
		return 0, 0, 0, gorm.ErrRecordNotFound
	}

	// Calculate page (1-based)
	page = int((positionInTimeline - 1) / int64(pageSize)) + 1

	// Calculate position within the page (1-based)
	position = int((positionInTimeline-1)%int64(pageSize)) + 1

	return page, position, total, nil
}

// GetPostBySessionID finds a post by session_id
func (p *Post) GetPostBySessionID(db *gorm.DB, sessionID string) (*Post, error) {
	var post Post
	err := db.Where("session_id = ? AND is_del = ?", sessionID, 0).First(&post).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (p PostVisibleT) String() string {
	switch p {
	case PostVisitPublic:
		return "public"
	case PostVisitPrivate:
		return "private"
	case PostVisitFriend:
		return "friend"
	default:
		return "unknow"
	}
}

func (p *PostFormated) GetHostID() int64 {
    if len(p.UserID) > 0 {
        return p.UserID[0]
    }
    return 0
}

func (p *PostFormated) GetVisitorID() int64 {
    if len(p.UserID) > 1 {
        return p.UserID[1]
    }
    return 0
}




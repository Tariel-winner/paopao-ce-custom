// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cs

// RoomInfo represents room information for common use
type RoomInfo struct {
	ID                int64     `json:"id"`
	HostID            int64     `json:"host_id"`
	HMSRoomID         string    `json:"hms_room_id,omitempty"`
	SpeakerIDs        []int64   `json:"speaker_ids"`
	StartTime         int64     `json:"start_time,omitempty"`
	CreatedAt         int64     `json:"created_at"`
	UpdatedAt         int64     `json:"updated_at"`
	Queue             QueueInfo `json:"queue"`
	IsBlockedFromSpace int16    `json:"is_blocked_from_space"`
	Speakers          []SpaceParticipant `json:"speakers"`
	Host              string    `json:"host,omitempty"`
	HostImageURL      string    `json:"host_image_url,omitempty"`
	HostUsername      string    `json:"host_username,omitempty"`
	HostLocation      string    `json:"host_location,omitempty"`
	Topics            []string  `json:"topics,omitempty"`
	Categories        []int64   `json:"categories,omitempty"`
	IsHostOnline      bool      `json:"is_host_online"`
	IsFollowing       *bool     `json:"is_following,omitempty"`
}

// QueueInfo represents queue information for common use
type QueueInfo struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name,omitempty"`
	Description   string   `json:"description,omitempty"`
	IsClosed      bool     `json:"is_closed"`
	Participants  []int64  `json:"participants"`
}

// SpaceParticipant represents a speaker in the room
type SpaceParticipant struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	Avatar      string `json:"avatar"`
	IsOnline    bool   `json:"is_online"`
} 
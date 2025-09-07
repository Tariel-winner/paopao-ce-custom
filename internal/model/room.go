package model

import (
    "encoding/json"
    "time"
)

// Room represents a room in the system
type Room struct {
    ID                int64     `json:"id"`
    HostID            int64     `json:"host_id"`
    HMSRoomID         string    `json:"hms_room_id,omitempty"`
    SpeakerIDs        []string  `json:"speaker_ids"`
    StartTime         time.Time `json:"start_time,omitempty"`
    CreatedAt         time.Time `json:"created_at"`
    UpdatedAt         time.Time `json:"updated_at"`
    Queue             Queue     `json:"queue"`
    IsBlockedFromSpace bool     `json:"is_blocked_from_space"`
    Topics            []string  `json:"topics,omitempty"`
    Categories        []int64   `json:"categories,omitempty"`
}

// Queue represents the room's queue
type Queue struct {
    ID            int64    `json:"id"`
    Name          string   `json:"name,omitempty"`
    Description   string   `json:"description,omitempty"`
    IsClosed      bool     `json:"is_closed"`
    Participants  []string `json:"participants"`
}

// RoomList represents a list of rooms with pagination
type RoomList struct {
    Items []*Room `json:"items"`
    Total int64   `json:"total"`
}

// MarshalJSON implements custom JSON marshaling for Room
func (r *Room) MarshalJSON() ([]byte, error) {
    type Alias Room
    return json.Marshal(&struct {
        *Alias
        StartTime int64 `json:"start_time,omitempty"`
        CreatedAt int64 `json:"created_at"`
        UpdatedAt int64 `json:"updated_at"`
    }{
        Alias:     (*Alias)(r),
        StartTime: r.StartTime.Unix(),
        CreatedAt: r.CreatedAt.Unix(),
        UpdatedAt: r.UpdatedAt.Unix(),
    })
}

// UnmarshalJSON implements custom JSON unmarshaling for Room
func (r *Room) UnmarshalJSON(data []byte) error {
    type Alias Room
    aux := &struct {
        *Alias
        StartTime int64 `json:"start_time,omitempty"`
        CreatedAt int64 `json:"created_at"`
        UpdatedAt int64 `json:"updated_at"`
    }{
        Alias: (*Alias)(r),
    }
    if err := json.Unmarshal(data, &aux); err != nil {
        return err
    }
    r.StartTime = time.Unix(aux.StartTime, 0)
    r.CreatedAt = time.Unix(aux.CreatedAt, 0)
    r.UpdatedAt = time.Unix(aux.UpdatedAt, 0)
    return nil
} 
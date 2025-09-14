// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
    "github.com/rocboss/paopao-ce/internal/core"
    "github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
    "github.com/rocboss/paopao-ce/internal/core/ms"
    "gorm.io/gorm"
    "github.com/sirupsen/logrus"
    "errors"
    "encoding/json"
)

var (
    _ core.RoomService = (*roomSrv)(nil)
)

type roomDao struct {
    db *gorm.DB
}

type roomSrv struct {
    db *gorm.DB
    userManageService core.UserManageService
    categoryService core.CategoryService
}

func newRoomDao(db *gorm.DB) *roomDao {
    return &roomDao{
        db: db,
    }
}

func newRoomService(db *gorm.DB, userManageService core.UserManageService, categoryService core.CategoryService) core.RoomService {
    return &roomSrv{
        db: db,
        userManageService: userManageService,
        categoryService: categoryService,
    }
}

// DAO methods
func (d *roomDao) CreateRoom(room *dbr.Room) error {
    logrus.WithFields(logrus.Fields{
        "room": room,
    }).Info("Creating room in database")
    createdRoom, err := room.Create(d.db)
    if err != nil {
        logrus.WithError(err).WithFields(logrus.Fields{
            "room": room,
        }).Error("Failed to create room in database")
        return err
    }
    // Update the original room object with the generated ID and timestamps
    room.ID = createdRoom.ID
    room.CreatedOn = createdRoom.CreatedOn
    room.ModifiedOn = createdRoom.ModifiedOn
    logrus.WithFields(logrus.Fields{
        "room": room,
    }).Info("Successfully created room in database")
    return nil
}

func (d *roomDao) GetRoomByID(id int64) (*dbr.Room, error) {
    room := &dbr.Room{
        Model: &dbr.Model{
            ID: id,
        },
    }
    return room.Get(d.db)
}

func (d *roomDao) GetRoomByHostID(hostID int64) (*dbr.Room, error) {
    // Debug log: hostID being queried
    logrus.WithField("hostID", hostID).Info("[DEBUG] GetRoomByHostID: querying for room with this hostID")
    room := &dbr.Room{
        HostID: hostID,
    }
    result, err := room.GetByHostID(d.db)
    if err != nil {
        logrus.WithFields(logrus.Fields{
            "hostID": hostID,
            "error":  err,
        }).Error("[DEBUG] GetRoomByHostID: error occurred")
        return nil, err
    }
    logrus.WithFields(logrus.Fields{
        "hostID": hostID,
        "roomID": result.ID,
    }).Info("[DEBUG] GetRoomByHostID: found room")
    return result, nil
}

func (d *roomDao) ListRooms(limit, offset int, userID int64, onlineUserIDs ...[]int64) ([]*dbr.Room, int64, error) {
    room := &dbr.Room{}
    
    // Build conditions for the List method
    conditions := dbr.ConditionsT{}
    
    // If online user IDs are provided, filter by them
    if len(onlineUserIDs) > 0 && len(onlineUserIDs[0]) > 0 {
        // Store the user IDs directly for proper GORM IN clause handling
        conditions["host_id IN ?"] = onlineUserIDs[0]
        logrus.Debugf("Filtering rooms by online host IDs: %v", onlineUserIDs[0])
    }
    
    // Simple ordering by creation time (newest first) for all cases
    // This ensures consistent pagination and works well with Redis filtering
    conditions["ORDER"] = "created_on DESC"
    logrus.Debugf("Using simple ordering by created_on DESC for user %d", userID)
    
    // Get rooms using the DBR List method
    rooms, err := room.List(d.db, conditions, offset, limit)
    if err != nil {
        logrus.Errorf("Failed to fetch rooms: %v", err)
        return nil, 0, err
    }
    
    // Count total using the DBR Count method
    total, err := room.Count(d.db, conditions)
    if err != nil {
        logrus.Errorf("Failed to count rooms: %v", err)
        return nil, 0, err
    }
    
    logrus.Debugf("ListRooms with follow priority: offset=%d, limit=%d, total=%d, returned=%d, userID=%d, onlineFilter=%v",
        offset, limit, total, len(rooms), userID, len(onlineUserIDs) > 0)
    
    return rooms, total, nil
}

func (d *roomDao) UpdateRoom(id int64, updates map[string]interface{}) error {
    logrus.WithFields(logrus.Fields{
        "room_id": id,
        "updates": updates,
    }).Info("Attempting to update room in database")

    // Marshal speaker_ids and topics to JSON if present
    if speakerIDs, ok := updates["speaker_ids"]; ok {
        b, err := json.Marshal(speakerIDs)
        if err != nil {
            logrus.WithError(err).Error("Failed to marshal speaker_ids to JSON")
            return err
        }
        updates["speaker_ids"] = string(b)
    }
    if topics, ok := updates["topics"]; ok {
        b, err := json.Marshal(topics)
        if err != nil {
            logrus.WithError(err).Error("Failed to marshal topics to JSON")
            return err
        }
        updates["topics"] = string(b)
    }

    // First check if room exists
    var count int64
    if err := d.db.Model(&dbr.Room{}).Where("id = ? AND is_del = ?", id, 0).Count(&count).Error; err != nil {
        logrus.WithError(err).WithField("room_id", id).Error("Failed to check if room exists")
        return err
    }
    
    if count == 0 {
        logrus.WithField("room_id", id).Error("Room not found or already deleted")
        return gorm.ErrRecordNotFound
    }

    // Perform the update
    result := d.db.Model(&dbr.Room{}).Where("id = ? AND is_del = ?", id, 0).Updates(updates)
    if result.Error != nil {
        logrus.WithError(result.Error).WithFields(logrus.Fields{
            "room_id": id,
            "updates": updates,
        }).Error("Failed to update room in database")
        return result.Error
    }

    if result.RowsAffected == 0 {
        logrus.WithField("room_id", id).Error("No rows were affected by the update")
        return gorm.ErrRecordNotFound
    }

    logrus.WithFields(logrus.Fields{
        "room_id": id,
        "updates": updates,
        "rows_affected": result.RowsAffected,
    }).Info("Successfully updated room in database")
    return nil
}



// Session mapping methods
func (d *roomDao) SaveSessionMapping(roomID, sessionID, peerID, userID string) error {
    logrus.WithFields(logrus.Fields{
        "room_id": roomID,
        "session_id": sessionID,
        "peer_id": peerID,
        "user_id": userID,
    }).Info("Saving session mapping to database")
    
    query := `INSERT INTO session_mappings (room_id, session_id, peer_id, user_id) 
              VALUES (?, ?, ?, ?) 
              ON CONFLICT (room_id, session_id, peer_id) 
              DO UPDATE SET user_id = ?`
    
    result := d.db.Exec(query, roomID, sessionID, peerID, userID, userID)
    if result.Error != nil {
        logrus.WithError(result.Error).WithFields(logrus.Fields{
            "room_id": roomID,
            "session_id": sessionID,
            "peer_id": peerID,
            "user_id": userID,
        }).Error("Failed to save session mapping")
        return result.Error
    }
    
    logrus.WithFields(logrus.Fields{
        "room_id": roomID,
        "session_id": sessionID,
        "peer_id": peerID,
        "user_id": userID,
    }).Info("Successfully saved session mapping")
    return nil
}

func (d *roomDao) GetUserIDFromSession(roomID, sessionID, peerID string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"room_id": roomID,
		"session_id": sessionID,
		"peer_id": peerID,
	}).Info("Looking up user_id from session mapping")
	
	var userID string
	query := `SELECT user_id FROM session_mappings 
              WHERE room_id = ? AND session_id = ? AND peer_id = ?`
	
	result := d.db.Raw(query, roomID, sessionID, peerID).Scan(&userID)
	if result.Error != nil {
		logrus.WithError(result.Error).WithFields(logrus.Fields{
			"room_id": roomID,
			"session_id": sessionID,
			"peer_id": peerID,
		}).Error("Failed to get user_id from session mapping")
		return "", result.Error
	}
	
	if result.RowsAffected == 0 {
		logrus.WithFields(logrus.Fields{
			"room_id": roomID,
			"session_id": sessionID,
			"peer_id": peerID,
		}).Error("No session mapping found")
		return "", gorm.ErrRecordNotFound
	}
	
	logrus.WithFields(logrus.Fields{
		"room_id": roomID,
		"session_id": sessionID,
		"peer_id": peerID,
		"user_id": userID,
	}).Info("Successfully found user_id from session mapping")
	return userID, nil
}



// Service methods
func (s *roomSrv) CreateRoom(room *ms.Room) error {
    logrus.WithFields(logrus.Fields{
        "room": room,
    }).Info("Creating room in service layer")
    
    // Verify host exists
    if _, err := s.userManageService.GetUserByID(room.HostID); err != nil {
        logrus.WithError(err).WithFields(logrus.Fields{
            "host_id": room.HostID,
        }).Error("Failed to verify host exists")
        return err
    }
    
    dbrRoom := &dbr.Room{
        HostID: room.HostID,
        HMSRoomID: room.HMSRoomID,
        SpeakerIDs: room.SpeakerIDs,
        StartTime: room.StartTime,
        Queue: &dbr.Queue{
            ID: room.Queue.ID,
            Name: room.Queue.Name,
            Description: room.Queue.Description,
            IsClosed: room.Queue.IsClosed,
            Participants: room.Queue.Participants,
        },
        IsBlockedFromSpace: room.IsBlockedFromSpace,
        Topics: room.Topics,
        Categories: dbr.Int64Array(room.Categories), // Convert from []int64 to Int64Array
    }
    
    logrus.WithFields(logrus.Fields{
        "dbr_room": dbrRoom,
    }).Info("Converting to DBR room")
    
    err := newRoomDao(s.db).CreateRoom(dbrRoom)
    if err != nil {
        logrus.WithError(err).WithFields(logrus.Fields{
            "dbr_room": dbrRoom,
        }).Error("Failed to create room in service layer")
        return err
    }
    
    // Update the original room object with the generated ID and timestamps
    room.ID = dbrRoom.ID
    room.CreatedOn = dbrRoom.CreatedOn
    room.ModifiedOn = dbrRoom.ModifiedOn
    
    logrus.WithFields(logrus.Fields{
        "dbr_room": dbrRoom,
        "original_room_id": room.ID,
    }).Info("Successfully created room in service layer")
    return nil
}

func (s *roomSrv) GetRoomByID(id int64) (*ms.Room, error) {
    room, err := newRoomDao(s.db).GetRoomByID(id)
    if err != nil {
        return nil, err
    }
    return &ms.Room{
        Model: &ms.Model{
            ID: room.ID,
        },
        HostID: room.HostID,
        HMSRoomID: room.HMSRoomID,
        SpeakerIDs: room.SpeakerIDs,
        StartTime: room.StartTime,
        Queue: &ms.Queue{
            ID: room.Queue.ID,
            Name: room.Queue.Name,
            Description: room.Queue.Description,
            IsClosed: room.Queue.IsClosed,
            Participants: room.Queue.Participants,
        },
        IsBlockedFromSpace: room.IsBlockedFromSpace,
        Topics: room.Topics,
        Categories: []int64(room.Categories), // Convert from Int64Array to []int64
    }, nil
}

func (s *roomSrv) GetRoomByHostID(hostID int64) (*ms.Room, error) {
    room, err := newRoomDao(s.db).GetRoomByHostID(hostID)
    if err != nil {
        return nil, err
    }
    return &ms.Room{
        Model: &ms.Model{
            ID: room.ID,
        },
        HostID: room.HostID,
        HMSRoomID: room.HMSRoomID,
        SpeakerIDs: room.SpeakerIDs,
        StartTime: room.StartTime,
        Queue: &ms.Queue{
            ID: room.Queue.ID,
            Name: room.Queue.Name,
            Description: room.Queue.Description,
            IsClosed: room.Queue.IsClosed,
            Participants: room.Queue.Participants,
        },
        IsBlockedFromSpace: room.IsBlockedFromSpace,
        Topics: room.Topics,
        Categories: []int64(room.Categories), // Convert from Int64Array to []int64
    }, nil
}

func (s *roomSrv) ListRooms(limit, offset int, userID int64, onlineUserIDs ...[]int64) ([]*ms.Room, int64, error) {
    // Call DAO with follow-based prioritization
    rooms, total, err := newRoomDao(s.db).ListRooms(limit, offset, userID, onlineUserIDs...)
    if err != nil {
        return nil, 0, err
    }
    
    msRooms := make([]*ms.Room, 0, len(rooms))
    for _, room := range rooms {
        msRooms = append(msRooms, &ms.Room{
            Model: &ms.Model{
                ID: room.ID,
            },
            HostID: room.HostID,
            HMSRoomID: room.HMSRoomID,
            SpeakerIDs: room.SpeakerIDs,
            StartTime: room.StartTime,
            Queue: &ms.Queue{
                ID: room.Queue.ID,
                Name: room.Queue.Name,
                Description: room.Queue.Description,
                IsClosed: room.Queue.IsClosed,
                Participants: room.Queue.Participants,
            },
            IsBlockedFromSpace: room.IsBlockedFromSpace,
            Topics: room.Topics,
            Categories: []int64(room.Categories),
        })
    }
    return msRooms, total, nil
}



func (s *roomSrv) UpdateRoom(id int64, updates map[string]interface{}) error {
    // Validate speaker IDs if being updated
    if speakerIDs, ok := updates["speaker_ids"].([]int64); ok {
        for _, speakerID := range speakerIDs {
            if _, err := s.userManageService.GetUserByID(speakerID); err != nil {
                logrus.WithError(err).WithField("speaker_id", speakerID).Error("Invalid speaker ID")
                return err
            }
        }
    }

    // Validate HMS Room ID if being updated
    if hmsRoomID, ok := updates["hms_room_id"].(string); ok {
        if hmsRoomID == "" {
            logrus.Error("HMS Room ID cannot be empty")
            return errors.New("invalid HMS Room ID")
        }
    }

    // Validate Queue if being updated
    if queue, ok := updates["queue"].(*ms.Queue); ok {
        if queue == nil {
            logrus.Error("Queue cannot be nil")
            return errors.New("invalid queue")
        }
        // Validate queue participants if present
        if queue.Participants != nil {
            for _, participantID := range queue.Participants {
                if _, err := s.userManageService.GetUserByID(participantID); err != nil {
                    logrus.WithError(err).WithField("participant_id", participantID).Error("Invalid queue participant ID")
                    return err
                }
            }
        }
    }

    // Validate Topics if being updated
    if topics, ok := updates["topics"].([]string); ok {
        if topics == nil {
            logrus.Error("Topics cannot be nil")
            return errors.New("invalid topics")
        }
        // Filter out empty topics and create new array
        validTopics := make([]string, 0, len(topics))
        for _, topic := range topics {
            if topic != "" {
                validTopics = append(validTopics, topic)
            }
        }
        updates["topics"] = validTopics
        logrus.WithFields(logrus.Fields{
            "room_id": id,
            "new_topics": validTopics,
        }).Info("Replacing topics for room")
    }

    // Validate IsBlockedFromSpace if being updated
    if isBlocked, ok := updates["is_blocked_from_space"].(int16); ok {
        if isBlocked < 0 {
            logrus.Error("IsBlockedFromSpace cannot be negative")
            return errors.New("invalid is_blocked_from_space value")
        }
    }

    // Validate Start Time if being updated
    if startTime, ok := updates["start_time"].(int64); ok {
        if startTime < 0 {
            logrus.Error("Start time cannot be negative")
            return errors.New("invalid start time")
        }
    }

    logrus.WithFields(logrus.Fields{
        "room_id": id,
        "updates": updates,
    }).Info("Validated updates for room")

    return newRoomDao(s.db).UpdateRoom(id, updates)
}



func (s *roomSrv) IsUserOnline(userID int64) bool {
    // TODO: This should be implemented through the DataService interface
    // The cache wrapper will handle the actual implementation
    return false
}

// Session mapping service methods
func (s *roomSrv) SaveSessionMapping(roomID, sessionID, peerID, userID string) error {
    logrus.WithFields(logrus.Fields{
        "room_id": roomID,
        "session_id": sessionID,
        "peer_id": peerID,
        "user_id": userID,
    }).Info("Saving session mapping in service layer")
    
    return newRoomDao(s.db).SaveSessionMapping(roomID, sessionID, peerID, userID)
}

func (s *roomSrv) GetUserIDFromSession(roomID, sessionID, peerID string) (string, error) {
	logrus.WithFields(logrus.Fields{
		"room_id": roomID,
		"session_id": sessionID,
		"peer_id": peerID,
	}).Info("Getting user_id from session mapping in service layer")
	
	return newRoomDao(s.db).GetUserIDFromSession(roomID, sessionID, peerID)
}

func (s *roomSrv) UpdateCategoriesByHostID(hostID int64, categoryIDs []int64) error {
	logrus.WithFields(logrus.Fields{
		"host_id": hostID,
		"category_ids": categoryIDs,
	}).Info("Updating room categories for host")
	
	// Call DBR layer directly (same pattern as user service)
	room := &dbr.Room{}
	return room.UpdateCategoriesByHostID(s.db, hostID, categoryIDs)
} 

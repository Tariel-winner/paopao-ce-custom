// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package core

import (
    "github.com/rocboss/paopao-ce/internal/core/ms"
)

// RoomService 房间服务
type RoomService interface {
    // GetRoomByID retrieves a room by its ID
    GetRoomByID(id int64) (*ms.Room, error)
    
    // GetRoomByHostID retrieves a room by its host ID
    GetRoomByHostID(hostID int64) (*ms.Room, error)
    
    // ListRooms returns a paginated list of rooms (with category priority if available)
    ListRooms(limit, offset int, userID int64, onlineUserIDs ...[]int64) ([]*ms.Room, int64, error)
    
    // CreateRoom creates a new room
    CreateRoom(room *ms.Room) error
    
    // UpdateRoom updates a room's information
    UpdateRoom(id int64, updates map[string]interface{}) error
    
    
    
    // IsUserOnline checks if a user is online
    IsUserOnline(userID int64) bool
    
    	// Session mapping methods
	SaveSessionMapping(roomID, sessionID, peerID, userID string) error
	GetUserIDFromSession(roomID, sessionID, peerID string) (string, error)
	
	// Update room categories for a host
	UpdateCategoriesByHostID(hostID int64, categoryIDs []int64) error
} 
// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"github.com/rocboss/paopao-ce/internal/core/cs"
	"github.com/rocboss/paopao-ce/internal/model/joint"
)

// Room represents a single room with enriched data
type Room = cs.RoomInfo

// Queue represents the room's queue
type Queue = cs.QueueInfo

// SpaceParticipant represents a speaker in the room
type SpaceParticipant = cs.SpaceParticipant

// RoomListResp represents the paginated response for rooms
type RoomListResp struct {
	joint.CachePageResp
} 
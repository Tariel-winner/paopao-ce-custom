// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package cache

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/redis/rueidis"
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/pkg/utils"
)

var (
	_webCache core.WebCache = (*webCache)(nil)
	_appCache core.AppCache = (*appCache)(nil)
)

type appCache struct {
	cscExpire time.Duration
	c         rueidis.Client
}

type webCache struct {
	core.AppCache
	c               rueidis.Client
	unreadMsgExpire int64
}

func (s *appCache) Get(key string) ([]byte, error) {
	res, err := rueidis.MGetCache(s.c, context.Background(), s.cscExpire, []string{key})
	if err != nil {
		return nil, err
	}
	message := res[key]
	return message.AsBytes()
}

func (s *appCache) Set(key string, data []byte, ex int64) error {
	ctx := context.Background()
	cmd := s.c.B().Set().Key(key).Value(utils.String(data))
	if ex > 0 {
		return s.c.Do(ctx, cmd.ExSeconds(ex).Build()).Error()
	}
	return s.c.Do(ctx, cmd.Build()).Error()
}

func (s *appCache) SetNx(key string, data []byte, ex int64) error {
	ctx := context.Background()
	cmd := s.c.B().Set().Key(key).Value(utils.String(data)).Nx()
	if ex > 0 {
		return s.c.Do(ctx, cmd.ExSeconds(ex).Build()).Error()
	}
	return s.c.Do(ctx, cmd.Build()).Error()
}

func (s *appCache) Delete(keys ...string) (err error) {
	if len(keys) != 0 {
		err = s.c.Do(context.Background(), s.c.B().Del().Key(keys...).Build()).Error()
	}
	return
}

func (s *appCache) DelAny(pattern string) (err error) {
	var (
		keys   []string
		cursor uint64
		entry  rueidis.ScanEntry
	)
	ctx := context.Background()
	for {
		cmd := s.c.B().Scan().Cursor(cursor).Match(pattern).Count(50).Build()
		if entry, err = s.c.Do(ctx, cmd).AsScanEntry(); err != nil {
			return
		}
		keys = append(keys, entry.Elements...)
		if entry.Cursor != 0 {
			cursor = entry.Cursor
			continue
		}
		break
	}
	if len(keys) != 0 {
		err = s.c.Do(ctx, s.c.B().Del().Key(keys...).Build()).Error()
	}
	return
}

func (s *appCache) Exist(key string) bool {
	cmd := s.c.B().Exists().Key(key).Build()
	count, _ := s.c.Do(context.Background(), cmd).AsInt64()
	return count > 0
}

func (s *appCache) Keys(pattern string) (res []string, err error) {
	ctx, cursor := context.Background(), uint64(0)
	for {
		cmd := s.c.B().Scan().Cursor(cursor).Match(pattern).Count(50).Build()
		entry, err := s.c.Do(ctx, cmd).AsScanEntry()
		if err != nil {
			return nil, err
		}
		res = append(res, entry.Elements...)
		if entry.Cursor != 0 {
			cursor = entry.Cursor
			continue
		}
		break
	}
	return
}

// BatchCheckOnlineUsers checks multiple user online statuses in a single Redis call
func (s *appCache) BatchCheckOnlineUsers(userIDs []int64) (map[int64]bool, error) {
	if len(userIDs) == 0 {
		return make(map[int64]bool), nil
	}
	
	ctx := context.Background()
	onlineStatus := make(map[int64]bool)
	
	// Build Redis keys for all users
	keys := make([]string, len(userIDs))
	for i, userID := range userIDs {
		keys[i] = conf.KeyOnlineUser.Get(userID)
	}
	
	// Use MGET to check all keys at once
	cmd := s.c.B().Mget().Key(keys...).Build()
	res, err := s.c.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		return nil, err
	}
	
	// Map results back to user IDs
	for i, userID := range userIDs {
		// If the key exists (not nil), user is online
		onlineStatus[userID] = res[i] != ""
	}
	
	return onlineStatus, nil
}


func (s *webCache) GetUnreadMsgCountResp(uid int64) ([]byte, error) {
	key := conf.KeyUnreadMsg.Get(uid)
	return s.Get(key)
}

func (s *webCache) PutUnreadMsgCountResp(uid int64, data []byte) error {
	return s.Set(conf.KeyUnreadMsg.Get(uid), data, s.unreadMsgExpire)
}

func (s *webCache) DelUnreadMsgCountResp(uid int64) error {
	return s.Delete(conf.KeyUnreadMsg.Get(uid))
}

func (s *webCache) ExistUnreadMsgCountResp(uid int64) bool {
	return s.Exist(conf.KeyUnreadMsg.Get(uid))
}

func (s *webCache) PutHistoryMaxOnline(newScore int) (int, error) {
	ctx := context.Background()
	cmd := s.c.B().Zadd().
		Key(conf.KeySiteStatus).
		Gt().ScoreMember().
		ScoreMember(float64(newScore), conf.KeyHistoryMaxOnline).Build()
	if err := s.c.Do(ctx, cmd).Error(); err != nil {
		return 0, err
	}
	cmd = s.c.B().Zscore().Key(conf.KeySiteStatus).Member(conf.KeyHistoryMaxOnline).Build()
	if score, err := s.c.Do(ctx, cmd).ToFloat64(); err == nil {
		return int(score), nil
	} else {
		return 0, err
	}
}

func newAppCache() *appCache {
	return &appCache{
		cscExpire: conf.CacheSetting.CientSideCacheExpire,
		c:         conf.MustRedisClient(),
	}
}

// GetOnlineUsersWithCursor implements cursor-based pagination for online users with their locations
// This prevents duplication even with constantly changing Redis keys
// Returns: userIDs, locations map[userID]location, nextCursor, error
func (s *appCache) GetOnlineUsersWithCursor(cursor uint64, limit int) ([]int64, map[int64]string, uint64, error) {
	var userIDs []int64
	var scannedCount int
	var nextCursor uint64
	
	ctx := context.Background()
	
	for {
		// Scan next batch of keys with pattern "paopao:onlineuser:*"
		cmd := s.c.B().Scan().Cursor(cursor).Match("paopao:onlineuser:*").Count(100).Build()
		entry, err := s.c.Do(ctx, cmd).AsScanEntry()
		if err != nil {
			return nil, nil, 0, err
		}
		
		nextCursor = entry.Cursor // Store the cursor outside the loop
		
		// Convert keys to user IDs
		for _, key := range entry.Elements {
			// Extract user ID from key like "paopao:onlineuser:12345"
			userIDStr := strings.TrimPrefix(key, "paopao:onlineuser:")
			if userIDStr != key { // Make sure prefix was found
				if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
					userIDs = append(userIDs, userID)
					scannedCount++
					
					// Stop if we reached the limit
					if scannedCount >= limit {
						break
					}
				}
			}
		}
		
		// Break if we have enough users or Redis cursor is 0
		if scannedCount >= limit || entry.Cursor == 0 {
			break
		}
		
		cursor = entry.Cursor
	}
	
	// Now batch fetch locations for all found online users
	locations := make(map[int64]string)
	if len(userIDs) > 0 {
		locationKeys := make([]string, len(userIDs))
		for i, userID := range userIDs {
			locationKeys[i] = conf.KeyUserLocation.Get(userID)
		}
		
		// Use MGET to get all location keys at once
		cmd := s.c.B().Mget().Key(locationKeys...).Build()
		res, err := s.c.Do(ctx, cmd).AsStrSlice()
		if err == nil { // Don't fail if locations can't be fetched
			// Map results back to user IDs
			for i, userID := range userIDs {
				if i < len(res) && res[i] != "" {
					locations[userID] = res[i] // Format: "Country|City" or "Country"
				}
			}
		}
	}
	
	return userIDs, locations, nextCursor, nil
}


// GetOnlineUsersCount gets the total count of online users by counting keys
func (s *appCache) GetOnlineUsersCount() (int64, error) {
	ctx := context.Background()
	var totalCount int64
	var cursor uint64
	
	for {
		cmd := s.c.B().Scan().Cursor(cursor).Match("paopao:onlineuser:*").Count(100).Build()
		entry, err := s.c.Do(ctx, cmd).AsScanEntry()
		if err != nil {
			return 0, err
		}
		
		totalCount += int64(len(entry.Elements))
		
		if entry.Cursor == 0 {
			break
		}
		cursor = entry.Cursor
	}
	
	return totalCount, nil
}

// SetUserLocation sets user location data in Redis
func (s *appCache) SetUserLocation(userID int64, locationData string, expire int64) error {
	ctx := context.Background()
	key := conf.KeyUserLocation.Get(userID)
	cmd := s.c.B().Set().Key(key).Value(locationData)
	if expire > 0 {
		return s.c.Do(ctx, cmd.ExSeconds(expire).Build()).Error()
	}
	return s.c.Do(ctx, cmd.Build()).Error()
}

// GetUserLocation gets user location data from Redis
func (s *appCache) GetUserLocation(userID int64) (string, error) {
	ctx := context.Background()
	key := conf.KeyUserLocation.Get(userID)
	cmd := s.c.B().Get().Key(key).Build()
	res, err := s.c.Do(ctx, cmd).AsStrSlice()
	if err != nil {
		return "", err
	}
	if len(res) > 0 {
		return res[0], nil
	}
	return "", nil
}

func newWebCache(ac core.AppCache) *webCache {
	return &webCache{
		AppCache:        ac,
		c:               conf.MustRedisClient(),
		unreadMsgExpire: conf.CacheSetting.UnreadMsgExpire,
	}
}

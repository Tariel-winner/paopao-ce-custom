// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package chain

import (
	"github.com/rocboss/paopao-ce/internal/conf"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/infra/metrics"
	"github.com/rocboss/paopao-ce/pkg/utils/iploc"
)

type OnlineUserMetric struct {
	metrics.BaseMetric
	ac       core.AppCache
	uid      int64
	clientIP string
	expire   int64
}

func OnUserOnlineMetric(ac core.AppCache, uid int64, clientIP string) {
	metrics.OnMeasure(&OnlineUserMetric{
		ac:       ac,
		uid:      uid,
		clientIP: clientIP,
		expire:   conf.CacheSetting.OnlineUserExpire,
	})
}

func (m *OnlineUserMetric) Name() string {
	return "OnlineUserMetric"
}

func (m *OnlineUserMetric) Action() (err error) {
	// 设置用户在线状态
	m.ac.Set(conf.KeyOnlineUser.Get(m.uid), []byte{}, m.expire)
	
	// 设置用户位置信息（如果有IP地址）
	if m.clientIP != "" {
		// Check if we already have location for this user (avoid repeated IP lookups)
		locationKey := conf.KeyUserLocation.Get(m.uid)
		if !m.ac.Exist(locationKey) {
			// Only lookup location if we don't have it cached
			country, city := iploc.Find(m.clientIP)
			if country != "" {
				locationData := country
				if city != "" {
					locationData = country + "|" + city
				}
				// Location cached for 24 hours (86400 seconds)
				// Reasoning: User's location changes less frequently than online status
				locationExpire := int64( 60 * 60) // 24 hours in seconds
				m.ac.Set(locationKey, []byte(locationData), locationExpire)
			}
		}
		// If location already exists in Redis, we skip the expensive IP lookup
	}
	return
}

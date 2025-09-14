// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package core

import "github.com/rocboss/paopao-ce/internal/core/ms"

// LocationService 位置服务接口
type LocationService interface {
	// UpdateUserLocation 更新用户位置信息到Redis
	UpdateUserLocation(userID int64, locationData *ms.LocationData) error
}

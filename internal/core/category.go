// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package core

import (
	"github.com/rocboss/paopao-ce/internal/core/ms"
)

// CategoryService defines the interface for category management
type CategoryService interface {
	// Master category management only (user categories are now part of user service)
	GetAllCategories() ([]*ms.Category, error)
	GetCategoryByID(id int64) (*ms.Category, error)
} 
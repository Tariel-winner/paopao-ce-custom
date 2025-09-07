// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package jinzhu

import (
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
	"gorm.io/gorm"
)

type categorySrv struct {
	db *gorm.DB
}

func newCategoryService(db *gorm.DB) core.CategoryService {
	return &categorySrv{
		db: db,
	}
}





// GetAllCategories gets all available categories
func (s *categorySrv) GetAllCategories() ([]*ms.Category, error) {
	category := &dbr.Category{}
	dbrCategories, err := category.GetAllCategories(s.db)
	if err != nil {
		return nil, err
	}

	// Convert dbr.Category to ms.Category
	msCategories := make([]*ms.Category, len(dbrCategories))
	for i, cat := range dbrCategories {
		msCategories[i] = &ms.Category{
			Model: &ms.Model{
				ID:         cat.ID,
				CreatedOn:  cat.CreatedOn,
				ModifiedOn: cat.ModifiedOn,
				DeletedOn:  cat.DeletedOn,
				IsDel:      cat.IsDel,
			},
			Name:        cat.Name,
			Description: cat.Description,
			Icon:        cat.Icon,
			Color:       cat.Color,
		}
	}

	return msCategories, nil
}

// GetCategoryByID gets a category by ID
func (s *categorySrv) GetCategoryByID(id int64) (*ms.Category, error) {
	category := &dbr.Category{}
	dbrCategory, err := category.GetCategoryByID(s.db, id)
	if err != nil {
		return nil, err
	}

	msCategory := &ms.Category{
		Model: &ms.Model{
			ID:         dbrCategory.ID,
			CreatedOn:  dbrCategory.CreatedOn,
			ModifiedOn: dbrCategory.ModifiedOn,
			DeletedOn:  dbrCategory.DeletedOn,
			IsDel:      dbrCategory.IsDel,
		},
		Name:        dbrCategory.Name,
		Description: dbrCategory.Description,
		Icon:        dbrCategory.Icon,
		Color:       dbrCategory.Color,
	}

	return msCategory, nil
}



 
// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package dbr

import (
	"time"
	"gorm.io/gorm"
)

// Category represents a master category
type Category struct {
	*Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

// UserCategory represents user's category preferences
type UserCategory struct {
	*Model
	UserID     int64 `json:"user_id"`
	CategoryID int64 `json:"category_id"`
}

// GetUserCategories gets all categories for a user
func (uc *UserCategory) GetUserCategories(db *gorm.DB, userID int64) ([]int64, error) {
	var categoryIDs []int64
	err := db.Model(&UserCategory{}).
		Where("user_id = ? AND is_del = ?", userID, 0).
		Pluck("category_id", &categoryIDs).Error
	return categoryIDs, err
}

// CreateUserCategory creates a user category preference
func (uc *UserCategory) CreateUserCategory(db *gorm.DB, userID, categoryID int64) error {
	uc.UserID = userID
	uc.CategoryID = categoryID
	uc.CreatedOn = time.Now().Unix()
	uc.ModifiedOn = time.Now().Unix()
	return db.Create(uc).Error
}

// DeleteUserCategory removes a user category preference
func (uc *UserCategory) DeleteUserCategory(db *gorm.DB, userID, categoryID int64) error {
	return db.Model(&UserCategory{}).
		Where("user_id = ? AND category_id = ? AND is_del = ?", userID, categoryID, 0).
		Updates(map[string]interface{}{
			"deleted_on": time.Now().Unix(),
			"is_del":     1,
		}).Error
}



// GetAllCategories gets all active categories
func (c *Category) GetAllCategories(db *gorm.DB) ([]*Category, error) {
	var categories []*Category
	err := db.Model(&Category{}).
		Where("is_del = ?", 0).
		Order("name ASC").
		Find(&categories).Error
	return categories, err
}

// GetCategoryByID gets a category by ID
func (c *Category) GetCategoryByID(db *gorm.DB, id int64) (*Category, error) {
	var category Category
	err := db.Model(&Category{}).
		Where("id = ? AND is_del = ?", id, 0).
		First(&category).Error
	return &category, err
} 
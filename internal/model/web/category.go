package web

import (
	"github.com/rocboss/paopao-ce/internal/core/cs"
)

// Category represents a category in web responses
type Category = cs.CategoryInfo

// CategoryListResp represents the response for listing categories
type CategoryListResp struct {
	Categories []*Category `json:"categories"`
}





 
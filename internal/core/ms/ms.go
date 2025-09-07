// Copyright 2023 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package ms contain core data service interface type
// model define for gorm adapter
package ms

import (
	"github.com/rocboss/paopao-ce/internal/dao/jinzhu/dbr"
)

const (
	UserStatusNormal = dbr.UserStatusNormal
	UserStatusClosed = dbr.UserStatusClosed
)

type (
	User                = dbr.User
	Post                = dbr.Post
	ConditionsT         = dbr.ConditionsT
	PostFormated        = dbr.PostFormated
	PostContentFormated = dbr.PostContentFormated
	UserFormated        = dbr.UserFormated
	UserReactionWithUser = dbr.UserReactionWithUser
	Model               = dbr.Model
	Room                = dbr.Room
	Queue               = dbr.Queue
	Category            = dbr.Category
	UserCategory        = dbr.UserCategory
	UserReaction        = dbr.UserReaction
)

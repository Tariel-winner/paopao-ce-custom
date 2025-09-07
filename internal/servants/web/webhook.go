// Copyright 2022 ROC. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/alimy/mir/v4"
	"github.com/gin-gonic/gin"
	api "github.com/rocboss/paopao-ce/auto/api/v1"
	"github.com/rocboss/paopao-ce/internal/model/web"
	"github.com/rocboss/paopao-ce/internal/servants/base"
	"github.com/rocboss/paopao-ce/internal/core"
	"github.com/rocboss/paopao-ce/internal/core/ms"
	"github.com/sirupsen/logrus"
)

var (
	_ api.Webhook = (*webhookSrv)(nil)
)

type webhookSrv struct {
	api.UnimplementedWebhookServant
	*base.DaoServant
}

func (s *webhookSrv) Chain() gin.HandlersChain {
	// Add raw request logging middleware for webhook debugging
	return gin.HandlersChain{
		func(c *gin.Context) {
			// Log ALL incoming requests to this router group for debugging
			logrus.Infof("=== WEBHOOK MIDDLEWARE TRIGGERED ===")
			logrus.Infof("Path: %s", c.Request.URL.Path)
			logrus.Infof("Method: %s", c.Request.Method)
			logrus.Infof("Headers: %v", c.Request.Header)
			
			// Read the raw body for ALL requests to this group
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logrus.Errorf("Failed to read raw request body: %v", err)
			} else {
				logrus.Infof("=== RAW WEBHOOK REQUEST ===")
				logrus.Infof("Body: %s", string(body))
				logrus.Infof("=== END RAW WEBHOOK REQUEST ===")
				
				// Put the body back for further processing
				c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
			}
			c.Next()
		},
	}
}

// First, let's define the webhookError type
type webhookError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"msg"`
	Data    map[string]interface{} `json:"data"`
}

// Implement mir.Error interface
func (e *webhookError) Error() string {
	return e.Message
}

// Implement mir.Error interface
func (e *webhookError) StatusCode() int {
	if e.Code == 0 {
		return http.StatusOK
	}
	return http.StatusBadRequest
}

// Implement mir.Error interface
func (e *webhookError) Render(c *gin.Context) {
	c.JSON(e.StatusCode(), e)
}

// Helper function to create success response
func newSuccessResponse(data map[string]interface{}) mir.Error {
	return &webhookError{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

// Helper function to create error response
func newErrorResponse(code int, message string, err error) mir.Error {
	return &webhookError{
		Code:    code,
		Message: message,
		Data: map[string]interface{}{
			"error":   err.Error(),
			"details": "Error processing recording webhook",
		},
	}
}

// Session registration endpoint
func (s *webhookSrv) RegisterSession(req *web.SessionRegistrationReq) mir.Error {
	logrus.Infof("Registering %d sessions", len(req.Sessions))

	// Save all session mappings
	for _, session := range req.Sessions {
		logrus.Infof("Registering session: room_id=%s, session_id=%s, peer_id=%s, user_id=%s",
			session.RoomID, session.SessionID, session.PeerID, session.UserID)

		// Save session mapping
		err := s.Ds.SaveSessionMapping(session.RoomID, session.SessionID, session.PeerID, session.UserID)
		if err != nil {
			logrus.Errorf("Failed to save session mapping: %v", err)
			return newErrorResponse(500, "Failed to save session mapping", err)
		}
	}

	return newSuccessResponse(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Successfully registered %d sessions", len(req.Sessions)),
		"count":   len(req.Sessions),
	})
}

func (s *webhookSrv) AudioWebhook(req *web.AudioWebhookReq) mir.Error {
	logrus.Infof("=== FULL WEBHOOK PAYLOAD RECEIVED === %+v", req)
	logrus.Infof("Received webhook payload type: %s", req.Type)
	logrus.Infof("Received webhook data: recording_id=%s, room_id=%s, peer_id=%s, session_id=%s, track_type=%s",
		req.Data.RecordingID,
		req.Data.RoomID,
		req.Data.PeerID,
		req.Data.SessionID,
		req.Data.TrackType,
	)
	
	// Process both track.recording.success and stream.recording.success events
	if req.Type != "track.recording.success" && req.Type != "stream.recording.success" {
		logrus.Infof("Ignoring non-recording webhook event: %s", req.Type)
		return newSuccessResponse(map[string]interface{}{
			"success": true,
			"ignored": true,
		})
	}

	// Validate required fields
	if req.Data.RecordingID == "" || req.Data.RoomID == "" || 
	req.Data.RecordingURL == "" || req.Data.PeerID == "" {
		logrus.Errorf("Missing required fields in webhook payload: recording_id=%s, room_id=%s, peer_id=%s, recording_presigned_url=%s",
			req.Data.RecordingID,
			req.Data.RoomID,
			req.Data.PeerID,
			req.Data.RecordingURL,
		)
		return newErrorResponse(400, "Missing required fields", fmt.Errorf("missing required fields in webhook payload"))
	}

	logrus.Infof("Looking up user mapping with: room_id=%s, session_id=%s, peer_id=%s",
		req.Data.RoomID, req.Data.SessionID, req.Data.PeerID)

	// Get user_id from session mapping using session_id from webhook payload
	userID, err := s.Ds.GetUserIDFromSession(req.Data.RoomID, req.Data.SessionID, req.Data.PeerID)
	if err != nil {
		logrus.Errorf("Failed to get user_id for peer_id %s with session_id %s: %v", req.Data.PeerID, req.Data.SessionID, err)
		return newErrorResponse(400, "No session mapping found", err)
	}

	// Get the existing post for this session using session_id
	post, err := s.Ds.GetPostBySessionID(req.Data.SessionID)
	if err != nil {
		logrus.Errorf("Failed to find post with session_id %s: %v", req.Data.SessionID, err)
		return newErrorResponse(400, "Failed to find post for this session", err)
	}

	// Get the audio content directly for this post (more efficient)
	existingContent, err := s.Ds.GetAudioContentByPostID(post.ID)
	if err != nil {
		logrus.Errorf("Failed to get audio content for post_id %d: %v", post.ID, err)
		return newErrorResponse(400, "No audio content found for this post", err)
	}

	// Format content with user ID mapping
	var newContent string
	if existingContent.Content == "" {
		// First recording
		newContent = fmt.Sprintf("%s:%s", userID, req.Data.RecordingURL)
		logrus.Infof("Creating first recording content: %s", newContent)
	} else {
		// Split content and ensure all parts have user_id prefix
		parts := strings.Split(existingContent.Content, "|")
		var validParts []string
		
		// Filter out any parts without user_id prefix
		for _, part := range parts {
			if strings.Contains(part, ":") {
				validParts = append(validParts, part)
			}
		}
		
		// Check if user already has a recording
		userExists := false
		for i, part := range validParts {
			if strings.HasPrefix(part, userID+":") {
				// Update existing user's recording
				validParts[i] = fmt.Sprintf("%s:%s", userID, req.Data.RecordingURL)
				userExists = true
				break
			}
		}
		
		if !userExists {
			// Add new user's recording
			validParts = append(validParts, fmt.Sprintf("%s:%s", userID, req.Data.RecordingURL))
		}
		
		newContent = strings.Join(validParts, "|")
		logrus.Infof("Updating recordings content. Previous: %s, New: %s", existingContent.Content, newContent)
	}

	// Update content with the formatted string
	// Use duration and size directly from webhook data (now numbers)
	duration := req.Data.Duration
	size := req.Data.Size
	
	err = s.Ds.UpdateContentByPostID(post.ID, newContent, duration, size)
	if err != nil {
		logrus.Errorf("Failed to update post content: %v", err)
		return newErrorResponse(400, "Failed to update post content", err)
	}

	// We already have the post from earlier, no need to fetch it again

	// Update search index
	s.PushPostToSearch(post)
	
	// 私密推文不创建标签与用户提醒
	if post.Visibility != core.PostVisitPrivate {
		// 创建用户消息提醒 - 通知访客用户
		if post.GetVisitorID() > 0 && post.GetVisitorID() != post.GetHostID() {
			// Only create message when both users have completed their recordings
			// Check if we went from 1 user to 2 users (both recordings completed)
			beforeUserCount := s.countUsersInContent(existingContent.Content)
			afterUserCount := s.countUsersInContent(newContent)
			
			logrus.Infof("Content change: before=%d users (%s), after=%d users (%s)", 
				beforeUserCount, existingContent.Content, 
				afterUserCount, newContent)
			
			// Only create message when we transition from 1 user to 2 users
			if beforeUserCount == 1 && afterUserCount == 2 {
				// 创建消息提醒
				onCreateMessageEvent(&ms.Message{
					SenderUserID:   post.GetHostID(),
					ReceiverUserID: post.GetVisitorID(),
					Type:           ms.MsgTypePost,
					Brief:          "在新发布的泡泡动态中@了你",
					PostID:         post.ID,
				})
				logrus.Infof("Both users completed recordings, created message notification for visitor user %d", post.GetVisitorID())
			} else {
				logrus.Infof("Skipping message creation: before=%d users, after=%d users", beforeUserCount, afterUserCount)
			}
		}
	}
	
	logrus.Infof("Successfully updated audio content for recording_id: %s, room_id: %s, peer_id: %s, user_id: %s",
		req.Data.RecordingID,
		req.Data.RoomID,
		req.Data.PeerID,
		userID,
	)

	return newSuccessResponse(map[string]interface{}{
		"success":      true,
		"recording_id": req.Data.RecordingID,
		"room_id":      req.Data.RoomID,
		"peer_id":      req.Data.PeerID,
		"user_id":      userID,
		"track_type":   req.Data.TrackType,
	})
}




// Helper method to count unique users in audio content
func (s *webhookSrv) countUsersInContent(content string) int {
	if content == "" {
		return 0
	}
	
	// Split by | to get individual recordings
	parts := strings.Split(content, "|")
	userIDs := make(map[string]bool)
	
	// Count unique user IDs
	for _, part := range parts {
		if strings.Contains(part, ":") {
			// Extract user ID from "userID:url" format
			userID := strings.Split(part, ":")[0]
			userIDs[userID] = true
		}
	}
	
	return len(userIDs)
}

func newWebhookSrv(s *base.DaoServant) api.Webhook {
	return &webhookSrv{
		DaoServant: s,
	}
} 

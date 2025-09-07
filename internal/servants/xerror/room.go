package xerror

import "errors"

var (
	ErrRoomNotFound = errors.New("room not found")
	ErrRoomDeleted  = errors.New("room has been deleted")
	ErrNotRoomHost  = errors.New("user is not the room host")
	ErrInvalidSpeaker = errors.New("invalid speaker ID")
) 
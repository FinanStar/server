package session

import (
	"context"
	"time"
)

const (
	SESSION_TTL       = time.Duration(14*24) * time.Hour
	SESSION_ID_LENGTH = 16
)

const (
	SESSION_NOT_FOUND_ERROR    = "There is no session with provided sId"
	SESSION_DATA_INVALID_ERROR = "Associated data with sId is invalid"
)

type SessionManager interface {
	CreateSession(ctx context.Context, sData *SessionData) (string, error)
	DeleteSession(ctx context.Context, sId string) error
	RenewalSession(ctx context.Context, sId string) error
	GetSessionData(ctx context.Context, sId string) (*SessionData, error)
	ResetSessions(ctx context.Context, userId uint32) error
}

type SessionData struct {
	UserId uint32
}

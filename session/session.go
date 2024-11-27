package session

import "context"

type SessionManager interface {
	CreateSession(ctx context.Context, sData *SessionData) (string, error)
	DeleteSession(ctx context.Context, sId string) error
	RenewalSession(ctx context.Context, sId string) error
	GetSessionData(ctx context.Context, sId string) (*SessionData, error)
}

type SessionData struct {
	UserId uint32
}
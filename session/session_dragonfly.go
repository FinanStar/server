package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Implementation of session manager using Dragonfly DB

const (
	KNOWN_SESSIONS_SET_KEY_PREFIX = "known-sessions-set"
	SESSION_KEY_PREFIX            = "session"
	SESSION_TTL                   = time.Duration(14*24) * time.Hour
)

func generateSecureId() (string, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", fmt.Errorf("Generate secure id failed with error: %v", err.Error())
	}

	return hex.EncodeToString(randomBytes), nil
}

type CreateDragonflyClientOptions struct {
	Host       string
	Port       uint32
	DatabaseId int
	Username   string
	Password   string
}

func CreateDragonflySessionManager(options *CreateDragonflyClientOptions) *DragonflySessionManager {
	client := redis.NewClient(&redis.Options{
		Username: options.Username,
		Password: options.Password,
		DB:       options.DatabaseId,
		Addr:     fmt.Sprintf("%s:%d", options.Host, options.Port),
	})

	if client == nil {
		panic("Failed to create Redis client, check that options is valid")
	}

	return &DragonflySessionManager{client: client}
}

type DragonflySessionManager struct {
	client *redis.Client
}

func (dsm *DragonflySessionManager) CreateSession(ctx context.Context, sData *SessionData) (string, error) {
	sId, err := generateSecureId()

	if err != nil {
		return "", err
	}

	// Ensuring that sId will saved only if it is unique
	for {
		tx := dsm.client.TxPipeline()

		setCmd := tx.SetNX(
			ctx,
			fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
			sData.UserId,
			SESSION_TTL,
		)
		tx.SAdd(
			ctx,
			fmt.Sprintf("%s:%d", KNOWN_SESSIONS_SET_KEY_PREFIX, sData.UserId),
			sId,
		)

		_, err := tx.Exec(ctx)

		if err != nil {
			return "", err
		}

		if setCmd.Err() != redis.Nil {
			break
		}

		sId, err = generateSecureId()

		if err != nil {
			return "", err
		}
	}

	return sId, nil
}

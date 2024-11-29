package session

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
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
		return "", fmt.Errorf(
			"Generate secure id failed with error: %v",
			err.Error(),
		)
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

func CreateDragonflySessionManager(
	options *CreateDragonflyClientOptions,
) *DragonflySessionManager {
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

func (dsm *DragonflySessionManager) CreateSession(
	ctx context.Context,
	sData *SessionData,
) (string, error) {
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

func (dsm *DragonflySessionManager) DeleteSession(
	ctx context.Context,
	sId string,
) error {
	sessionKey := fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId)
	userId, err := dsm.client.Get(
		ctx,
		sessionKey,
	).Result()

	if err == redis.Nil {
		return errors.New("There is no session with provided sId")
	}

	if err != nil {
		return err
	}

	pipe := dsm.client.TxPipeline()

	pipe.Del(ctx, sessionKey)
	pipe.SRem(
		ctx,
		fmt.Sprintf("%s:%s", KNOWN_SESSIONS_SET_KEY_PREFIX, userId),
		sId,
	)

	_, err = pipe.Exec(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (dsm *DragonflySessionManager) RenewalSession(
	ctx context.Context,
	sId string,
) error {
	success, err := dsm.client.Expire(
		ctx,
		fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
		SESSION_TTL,
	).Result()

	if err != nil {
		return err
	}

	if success {
		return errors.New("Provided session is not existing")
	}

	return nil
}

func (dsm *DragonflySessionManager) GetSessionData(
	ctx context.Context,
	sId string,
) (*SessionData, error) {
	userId, err := dsm.client.Get(
		ctx,
		fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
	).Result()

	if err == redis.Nil {
		return nil, errors.New("There is no session with provided sId")
	}

	if err != nil {
		return nil, err
	}

	userIdConverted, err := strconv.ParseInt(userId, 10, 32)

	if err != nil {
		return nil, errors.New("Associated data with sId is invalid")
	}

	return &SessionData{UserId: uint32(userIdConverted)}, nil
}

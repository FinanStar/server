package session

import (
	"context"
	"errors"
	"finanstar/server/crypto"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

// Implementation of session manager using Dragonfly DB

const (
	KNOWN_SESSIONS_SET_KEY_PREFIX = "known-sessions-set"
	SESSION_KEY_PREFIX            = "session"
)

func NewDragonflySessionManager(client *redis.Client) *DragonflySessionManager {
	return &DragonflySessionManager{client}
}

type CreateDragonflyClientOptions struct {
	Host       string
	Port       uint32
	DatabaseId int
	Username   string
	Password   string
}

func CreateDragonflyClient(
	options *CreateDragonflyClientOptions,
) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Username: options.Username,
		Password: options.Password,
		DB:       options.DatabaseId,
		Addr:     fmt.Sprintf("%s:%d", options.Host, options.Port),
	})

	if client == nil {
		panic("Failed to create Redis client, check options validity")
	}

	return client
}

type DragonflySessionManager struct {
	client *redis.Client
}

func (dsm *DragonflySessionManager) CreateSession(
	ctx context.Context,
	sData *SessionData,
) (string, error) {
	// Ensuring that sId will saved only if it is unique
	for {
		sId, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)

		if err != nil {
			return "", err
		}

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

		_, err = tx.Exec(ctx)

		if err != nil {
			return "", err
		}

		setCmdSuccessful, err := setCmd.Result()

		if setCmdSuccessful {
			return sId, nil
		}
	}
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
		return errors.New(SESSION_NOT_FOUND_ERROR)
	}

	if err != nil {
		return err
	}

	knownSessionsSetKey := fmt.Sprintf(
		"%s:%s",
		KNOWN_SESSIONS_SET_KEY_PREFIX,
		userId,
	)
	pipe := dsm.client.TxPipeline()

	pipe.Del(ctx, sessionKey)
	pipe.SRem(ctx, knownSessionsSetKey, sId)

	_, err = pipe.Exec(ctx)

	if err != nil {
		return err
	}

	// #region Prevent memory leak in storage
	sIds, err := dsm.client.SMembers(ctx, knownSessionsSetKey).Result()

	if err != nil {
		return err
	}

	if len(sIds) == 0 {
		err = dsm.client.Del(ctx, knownSessionsSetKey).Err()

		if err != nil {
			return err
		}
	}
	// #endregion

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

	if !success {
		return errors.New(SESSION_NOT_FOUND_ERROR)
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
		return nil, errors.New(SESSION_NOT_FOUND_ERROR)
	}

	if err != nil {
		return nil, err
	}

	userIdConverted, err := strconv.ParseInt(userId, 10, 32)

	if err != nil {
		return nil, errors.New(SESSION_DATA_INVALID_ERROR)
	}

	return &SessionData{UserId: uint32(userIdConverted)}, nil
}

func (dsm *DragonflySessionManager) ResetSessions(
	ctx context.Context,
	userId uint32,
) error {
	knownSessionsSetKey := fmt.Sprintf(
		"%s:%d",
		KNOWN_SESSIONS_SET_KEY_PREFIX,
		userId,
	)
	sIds, err := dsm.client.SMembers(
		ctx,
		knownSessionsSetKey,
	).Result()

	if err != nil {
		return err
	}

	pipe := dsm.client.TxPipeline()
	delSessionsCmd := pipe.Del(ctx, sIds...)
	delKnownSessionsSetCmd := pipe.Del(ctx, knownSessionsSetKey)

	_, execErr := pipe.Exec(ctx)

	if execErr != nil {
		return execErr
	}

	if err = delSessionsCmd.Err(); err != nil {
		return err
	}

	if err = delKnownSessionsSetCmd.Err(); err != nil {
		return err
	}

	return nil
}

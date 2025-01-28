package session

import (
	"context"
	"errors"
	"finanstar/server/crypto"
	kvs "finanstar/server/key_value_store"
	"fmt"
	"strconv"
)

// Implementation of session manager using in-memory key-value store

const (
	KNOWN_SESSIONS_SET_KEY_PREFIX = "known-sessions-set"
	SESSION_KEY_PREFIX            = "session"
)

type InMemorySMBuilder struct {
	store  kvs.KeyValueStore
	crypto crypto.Crypto
}

func (self *InMemorySMBuilder) Store(
	store kvs.KeyValueStore,
) *InMemorySMBuilder {
	self.store = store

	return self
}

func (self *InMemorySMBuilder) Crypto(
	crypto crypto.Crypto,
) *InMemorySMBuilder {
	self.crypto = crypto

	return self
}

func (self InMemorySMBuilder) Build() SessionManager {
	return InMemorySessionManager{store: self.store, crypto: self.crypto}
}

func CreateInMemorySessionManager(
	store kvs.KeyValueStore,
	crypto crypto.Crypto,
) InMemorySessionManager {
	return InMemorySessionManager{store, crypto}
}

type InMemorySessionManager struct {
	store  kvs.KeyValueStore
	crypto crypto.Crypto
}

func (self InMemorySessionManager) CreateSession(
	ctx context.Context,
	sData *SessionData,
) (string, error) {
	// Ensuring that sId will saved only if it is unique
	for {
		sId, err := self.crypto.Generator().SecureId(SESSION_ID_LENGTH)

		if err != nil {
			return "", err
		}

		tx := self.store.StartTransaction()

		setCmd := tx.SetWithExpiration(
			ctx,
			fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
			sData.UserId,
			SESSION_TTL,
		)
		tx.AddToVector(
			ctx,
			fmt.Sprintf("%s:%d", KNOWN_SESSIONS_SET_KEY_PREFIX, sData.UserId),
			sId,
		)

		err = tx.Exec(ctx)

		if err != nil {
			return "", err
		}

		setCmdSuccessful, err := setCmd.Result()

		if setCmdSuccessful {
			return sId, nil
		}
	}
}

func (self InMemorySessionManager) DeleteSession(
	ctx context.Context,
	sId string,
) error {
	sessionKey := fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId)
	userId, err := self.store.Get(
		ctx,
		sessionKey,
	).Result()

	if err == kvs.NoItem {
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
	tx := self.store.StartTransaction()

	tx.Delete(ctx, sessionKey)
	tx.DeleteFromVector(ctx, knownSessionsSetKey, sId)

	err = tx.Exec(ctx)

	if err != nil {
		return err
	}

	// #region Prevent memory leak in storage
	sIds, err := self.store.GetVector(ctx, knownSessionsSetKey).Result()

	if err != nil {
		return err
	}

	if len(sIds) == 0 {
		_, err = self.store.Delete(ctx, knownSessionsSetKey).Result()

		if err != nil {
			return err
		}
	}
	// #endregion

	return nil
}

func (self InMemorySessionManager) RenewalSession(
	ctx context.Context,
	sId string,
) error {
	success, err := self.store.AssignExpiration(
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

func (self InMemorySessionManager) GetSessionData(
	ctx context.Context,
	sId string,
) (*SessionData, error) {
	userId, err := self.store.Get(
		ctx,
		fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
	).Result()

	if err == kvs.NoItem {
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

func (self InMemorySessionManager) ResetSessions(
	ctx context.Context,
	userId uint32,
) error {
	knownSessionsSetKey := fmt.Sprintf(
		"%s:%d",
		KNOWN_SESSIONS_SET_KEY_PREFIX,
		userId,
	)
	sIds, err := self.store.GetVector(
		ctx,
		knownSessionsSetKey,
	).Result()

	if err != nil {
		return err
	}

	tx := self.store.StartTransaction()
	delSessionsCmds := []kvs.CommandResult[int64]{}

	for _, id := range sIds {
		cmd := tx.Delete(ctx, id)

		delSessionsCmds = append(delSessionsCmds, cmd)
	}

	delKnownSessionsSetCmd := tx.Delete(ctx, knownSessionsSetKey)

	execErr := tx.Exec(ctx)

	if execErr != nil {
		return execErr
	}

	for _, cmd := range delSessionsCmds {
		_, err := cmd.Result()

		if err != nil {
			return err
		}
	}

	if _, err = delKnownSessionsSetCmd.Result(); err != nil {
		return err
	}

	return nil
}

package session

import (
	"context"
	"errors"
	"finanstar/server/crypto"
	kvs "finanstar/server/key_value_store"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateSession(t *testing.T) {
	t.Parallel()
	cryptoMock := crypto.CreateMock()
	keyValueStoreMock := kvs.KeyValueStoreMock{}
	imsm := CreateInMemorySessionManager(&keyValueStoreMock, &cryptoMock)
	require := require.New(t)

	userId := uint32(1337)
	ctx := context.Background()
	secureId := "very-super-secure-id"

	keyValueStoreTxMock := kvs.KeyValueStoreTransactionMock{}

	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId, nil)

	setWithExpirationResultMock := kvs.CreateResultMock[bool]()
	setWithExpirationResultMock.On("Result").Return(true, nil)
	keyValueStoreTxMock.
		On(
			"SetWithExpiration",
			ctx,
			fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, secureId),
			userId,
			SESSION_TTL,
		).
		Return(&setWithExpirationResultMock)

	addToVectorResultMock := kvs.CreateResultMock[int64]()
	keyValueStoreTxMock.
		On(
			"AddToVector",
			ctx,
			fmt.Sprintf("%s:%d", KNOWN_SESSIONS_SET_KEY_PREFIX, userId),
			secureId,
		).
		Return(&addToVectorResultMock)
	keyValueStoreTxMock.On("Exec", ctx).Return(nil)

	keyValueStoreMock.On("StartTransaction").Return(&keyValueStoreTxMock)

	sId, err := imsm.CreateSession(
		context.Background(),
		&SessionData{UserId: userId},
	)

	require.Nilf(err, "Expected sId, but got error")
	require.Equalf(
		len(secureId),
		len(sId),
		"Expected sId to be %d length, but got %d",
		len(secureId),
		len(sId),
	)
}

func TestDeleteSession(t *testing.T) {
	t.Parallel()
	cryptoMock := crypto.CreateMock()
	require := require.New(t)

	userId := uint32(1337)
	secureId := "very-super-secure-id"

	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId, nil)

	testVariants := []struct {
		title          string
		getSuccessful  bool
		delResult      int
		sRemResult     int
		sMembersResult int
	}{
		{"Defined second session", true, 1, 1, 1},
		{"Undefined session", false, 0, 0, 0},
		{"Defined last session", true, 1, 1, 0},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, (func(t *testing.T) {
			keyValueStoreMock := kvs.KeyValueStoreMock{}
			imsm := CreateInMemorySessionManager(&keyValueStoreMock, &cryptoMock)
			ctx := context.Background()

			getResultMock := kvs.CreateResultMock[string]()

			keyValueStoreMock.
				On("Get", ctx, fmt.Sprintf(`%s:%s`, SESSION_KEY_PREFIX, secureId)).
				Return(&getResultMock)

			if tt.getSuccessful {
				getResultMock.On("Result").Return(strconv.Itoa(int(userId)), nil)
				knownSessionSet := fmt.Sprintf(
					`%s:%d`,
					KNOWN_SESSIONS_SET_KEY_PREFIX,
					userId,
				)

				keyValueStoreTxMock := kvs.KeyValueStoreTransactionMock{}
				keyValueStoreMock.On("StartTransaction").Return(&keyValueStoreTxMock)

				deleteResultMock := kvs.CreateResultMock[int64]()
				deleteResultMock.On("Result").Return(int64(tt.delResult), nil)
				keyValueStoreTxMock.
					On("Delete", ctx, fmt.Sprintf(`%s:%s`, SESSION_KEY_PREFIX, secureId)).
					Return(&deleteResultMock)

				deleteFromVectorMock := kvs.CreateResultMock[int64]()
				deleteFromVectorMock.On("Result").Return(int64(tt.sRemResult), nil)
				keyValueStoreTxMock.
					On("DeleteFromVector", ctx, knownSessionSet, secureId).
					Return(&deleteFromVectorMock)

				keyValueStoreTxMock.On("Exec", ctx).Return(nil)

				sIds := make([]string, tt.sMembersResult)

				for index := range sIds {
					sIds[index] = secureId
				}

				getVectorResultMock := kvs.CreateResultMock[[]string]()
				getVectorResultMock.On("Result").Return(sIds, nil)
				keyValueStoreMock.
					On("GetVector", ctx, knownSessionSet).
					Return(&getVectorResultMock)

				if tt.sMembersResult == 0 {
					deleteResultMock := kvs.CreateResultMock[int64]()
					deleteResultMock.On("Result").Return(int64(1), nil)
					keyValueStoreMock.
						On("Delete", ctx, knownSessionSet).
						Return(&deleteResultMock)
				}
			} else {
				getResultMock.On("Result").Return("", errors.New(SESSION_NOT_FOUND_ERROR))
			}

			err := imsm.DeleteSession(context.Background(), secureId)

			if tt.getSuccessful {
				require.Nilf(
					err,
					"Expected run without error, but got %v",
					err,
				)
			} else {
				require.EqualError(err, SESSION_NOT_FOUND_ERROR)
			}
		}))
	}
}

func TestRenewalSession(t *testing.T) {
	t.Parallel()
	cryptoMock := crypto.CreateMock()
	require := require.New(t)

	secureId := "very-super-secure-id"

	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId, nil)

	testVariants := []struct {
		title  string
		result bool
	}{
		{"Defined session", true},
		{"Undefined session", false},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			keyValueStoreMock := kvs.KeyValueStoreMock{}
			imsm := CreateInMemorySessionManager(&keyValueStoreMock, &cryptoMock)
			ctx := context.Background()

			assignExpirationResultMock := kvs.CreateResultMock[bool]()
			assignExpirationResultMock.
				On("Result").Return(tt.result, nil)
			keyValueStoreMock.
				On(
					"AssignExpiration",
					ctx,
					fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, secureId),
					SESSION_TTL,
				).
				Return(&assignExpirationResultMock)

			err := imsm.RenewalSession(context.Background(), secureId)

			if tt.result {
				require.Nil(err)
			} else {
				require.EqualError(err, SESSION_NOT_FOUND_ERROR)
			}
		})
	}
}

func TestGetSessionData(t *testing.T) {
	t.Parallel()
	cryptoMock := crypto.CreateMock()
	require := require.New(t)

	secureId := "very-super-secure-id"
	userId := 1337

	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId, nil)

	testVariants := []struct {
		title        string
		getCmdResult interface{}
		resultError  error
	}{
		{"Defined valid session data", strconv.Itoa(userId), nil},
		{"Undefined session", nil, kvs.NoItem},
		{
			"Defined invalid session data",
			"l33t",
			errors.New(SESSION_DATA_INVALID_ERROR),
		},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			keyValueStoreMock := kvs.KeyValueStoreMock{}
			imsm := CreateInMemorySessionManager(&keyValueStoreMock, &cryptoMock)
			ctx := context.Background()

			getResultMock := kvs.CreateResultMock[string]()
			keyValueStoreMock.
				On("Get", ctx, fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, secureId)).
				Return(&getResultMock)

			if tt.resultError != nil {
				getResultMock.On("Result").Return("", tt.resultError)
			} else {
				getResultMock.On("Result").Return(tt.getCmdResult, nil)
			}

			sData, err := imsm.GetSessionData(context.Background(), secureId)

			if tt.resultError != nil {
				if tt.resultError == kvs.NoItem {
					require.EqualError(err, SESSION_NOT_FOUND_ERROR)
				}

				if tt.resultError.Error() == SESSION_DATA_INVALID_ERROR {
					require.EqualError(err, tt.resultError.Error())
				}
			} else {
				require.Equalf(
					uint32(userId),
					sData.UserId,
					"Expected valid session data %d, but got %d",
					userId,
					sData.UserId,
				)
			}
		})
	}
}

func TestResetSession(t *testing.T) {
	t.Parallel()
	cryptoMock := crypto.CreateMock()
	require := require.New(t)

	secureId1 := "first-very-super-secure-id"
	secureId2 := "second-very-super-secure-id"

	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId1, nil).
		Repeatability = 1
	cryptoMock.
		GeneratorMock.
		On("SecureId", SESSION_ID_LENGTH).
		Return(secureId2, nil).
		Repeatability = 1

	testVariants := []struct {
		title  string
		sIds   []string
		userId uint32
	}{
		{"Defined session", []string{secureId1, secureId2}, 1337},
		{"Undefined session", []string{}, 1336},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			keyValueStoreMock := kvs.KeyValueStoreMock{}
			imsm := CreateInMemorySessionManager(&keyValueStoreMock, &cryptoMock)
			ctx := context.Background()

			knownSessionsSet := fmt.Sprintf(
				"%s:%d",
				KNOWN_SESSIONS_SET_KEY_PREFIX,
				tt.userId,
			)

			getVectorResultMock := kvs.CreateResultMock[[]string]()
			getVectorResultMock.On("Result").Return(tt.sIds, nil)
			keyValueStoreMock.
				On("GetVector", ctx, knownSessionsSet).
				Return(&getVectorResultMock)

			keyValueStoreTxMock := kvs.KeyValueStoreTransactionMock{}
			keyValueStoreMock.On("StartTransaction").Return(&keyValueStoreTxMock)

			for _, id := range tt.sIds {
				deleteIdsResultMock := kvs.CreateResultMock[int64]()
				deleteIdsResultMock.On("Result").Return(int64(1), nil)
				keyValueStoreTxMock.On("Delete", ctx, id).Return(&deleteIdsResultMock)
			}

			deleteSessionSetResultMock := kvs.CreateResultMock[int64]()
			deleteSessionSetResultMock.On("Result").Return(int64(1), nil)
			keyValueStoreTxMock.
				On("Delete", ctx, knownSessionsSet).
				Return(&deleteSessionSetResultMock)

			keyValueStoreTxMock.On("Exec", ctx).Return(nil)

			err := imsm.ResetSessions(context.Background(), tt.userId)

			require.Nil(err)
		})
	}
}

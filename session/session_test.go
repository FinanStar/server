package session

import (
	"context"
	"errors"
	"finanstar/server/crypto"
	"fmt"
	"strconv"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func expectError(t *testing.T, gotErr error, expectedErr error) {
	if gotErr == nil {
		t.Errorf(`Expected error "%s", but got nothing`, expectedErr)
	}

	if expectedErr == nil {
		t.Error("Expected error is not provided")
	}

	if gotErr.Error() != expectedErr.Error() {
		t.Errorf(
			`Expected error "%s", but got "%s"`,
			expectedErr,
			gotErr.Error(),
		)
	}
}

func expectNoError(t *testing.T, err error) {
	if err != nil {
		t.Errorf(`Expected no error, but got "%s"`, err.Error())
	}
}

func checkMockExpectationsWereMet(t *testing.T, mock redismock.ClientMock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met, error = %s", err.Error())
	}
}

func TestCreateSession(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	dsm := NewDragonflySessionManager(client)

	userId := uint32(1337)

	mock.ExpectTxPipeline()
	mock.
		Regexp().
		ExpectSetNX(
			`[a-z]+`,
			userId,
			SESSION_TTL,
		).
		SetVal(true)
	mock.
		Regexp().
		ExpectSAdd(
			fmt.Sprintf("%s:%d", KNOWN_SESSIONS_SET_KEY_PREFIX, userId),
			`[a-z]+`,
		).
		SetVal(1)
	mock.ExpectTxPipelineExec().SetVal(make([]interface{}, 0))

	sId, err := dsm.CreateSession(
		context.Background(),
		&SessionData{UserId: userId},
	)

	if err != nil || len(sId) == 0 {
		t.Errorf("Expected sId, but got error or nil (%s)", err.Error())
	}

	checkMockExpectationsWereMet(t, mock)
}

func TestDeleteSession(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	dsm := NewDragonflySessionManager(client)

	userId := uint32(1337)
	sId, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)

	expectNoError(t, err)

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
			getExpect := mock.ExpectGet(
				fmt.Sprintf(`%s:%s`, SESSION_KEY_PREFIX, sId),
			)

			if tt.getSuccessful {
				getExpect.SetVal(strconv.Itoa(int(userId)))
				knownSessionSet := fmt.Sprintf(
					`%s:%d`,
					KNOWN_SESSIONS_SET_KEY_PREFIX,
					userId,
				)

				mock.ExpectTxPipeline()
				mock.
					ExpectDel(fmt.Sprintf(`%s:%s`, SESSION_KEY_PREFIX, sId)).
					SetVal(int64(tt.delResult))
				mock.
					ExpectSRem(knownSessionSet, sId).
					SetVal(int64(tt.sRemResult))
				mock.ExpectTxPipelineExec().SetVal(make([]interface{}, 0))

				sIds := make([]string, tt.sMembersResult)

				for index := range sIds {
					id, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)

					expectNoError(t, err)
					sIds[index] = id
				}

				mock.ExpectSMembers(knownSessionSet).SetVal(sIds)

				if tt.sMembersResult == 0 {
					mock.ExpectDel(knownSessionSet).SetVal(1)
				}
			} else {
				getExpect.RedisNil()
			}

			err = dsm.DeleteSession(context.Background(), sId)

			if tt.getSuccessful {
				if err != nil {
					t.Errorf("Expected run without error, but got error = %v", err)
				}
			} else {
				expectError(t, err, errors.New(SESSION_NOT_FOUND_ERROR))
			}

			checkMockExpectationsWereMet(t, mock)
		}))
	}
}

func TestRenewalSession(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	dsm := NewDragonflySessionManager(client)

	sId, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)

	expectNoError(t, err)

	testVariants := []struct {
		title  string
		result bool
	}{
		{"Defined session", true},
		{"Undefined session", false},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			mock.
				ExpectExpire(
					fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
					SESSION_TTL,
				).
				SetVal(tt.result)

			err = dsm.RenewalSession(context.Background(), sId)

			if tt.result {
				expectNoError(t, err)
			} else {
				expectError(t, err, errors.New(SESSION_NOT_FOUND_ERROR))
			}

			checkMockExpectationsWereMet(t, mock)
		})
	}
}

func TestGetSessionData(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	dsm := NewDragonflySessionManager(client)

	userId := 1337
	sId, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)

	expectNoError(t, err)

	testVariants := []struct {
		title        string
		getCmdResult interface{}
		resultError  error
	}{
		{"Defined valid session data", strconv.Itoa(userId), nil},
		{"Undefined session", nil, redis.Nil},
		{
			"Defined invalid session data",
			"l33t",
			errors.New(SESSION_DATA_INVALID_ERROR),
		},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			expectGetMock := mock.ExpectGet(
				fmt.Sprintf("%s:%s", SESSION_KEY_PREFIX, sId),
			)

			if tt.resultError == redis.Nil {
				expectGetMock.RedisNil()
			} else {
				expectGetMock.SetVal(tt.getCmdResult.(string))
			}

			sData, err := dsm.GetSessionData(context.Background(), sId)

			if tt.resultError != nil {
				if tt.resultError == redis.Nil {
					expectError(t, err, errors.New(SESSION_NOT_FOUND_ERROR))
				}

				if tt.resultError.Error() == SESSION_DATA_INVALID_ERROR {
					expectError(t, err, tt.resultError)
				}
			} else {
				if sData.UserId != uint32(userId) {
					t.Errorf(
						"Expected valid session data %d, but got %d",
						userId,
						sData.UserId,
					)
				}
			}

			checkMockExpectationsWereMet(t, mock)
		})
	}
}

func TestResetSession(t *testing.T) {
	t.Parallel()
	client, mock := redismock.NewClientMock()
	dsm := NewDragonflySessionManager(client)

	sId1, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)
	expectNoError(t, err)
	sId2, err := crypto.GenerateSecureId(SESSION_ID_LENGTH)
	expectNoError(t, err)

	testVariants := []struct {
		title  string
		sIds   []string
		userId uint32
	}{
		{"Defined session", []string{sId1, sId2}, 1337},
		{"Undefined session", []string{}, 1336},
	}

	for _, tt := range testVariants {
		t.Run(tt.title, func(t *testing.T) {
			knownSessionsSet := fmt.Sprintf(
				"%s:%d",
				KNOWN_SESSIONS_SET_KEY_PREFIX,
				tt.userId,
			)

			mock.ExpectSMembers(knownSessionsSet).SetVal(tt.sIds)
			mock.ExpectTxPipeline()
			mock.ExpectDel(tt.sIds...).SetVal(int64(len(tt.sIds)))
			mock.ExpectDel(knownSessionsSet).SetVal(1)
			mock.ExpectTxPipelineExec()

			err := dsm.ResetSessions(context.Background(), tt.userId)

			expectNoError(t, err)
			checkMockExpectationsWereMet(t, mock)
		})
	}
}

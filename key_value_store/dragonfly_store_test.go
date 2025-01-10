package key_value_store

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func checkMockExpectationsWereMet(t *testing.T, mock redismock.ClientMock) {
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Mock expectations were not met, error = %s", err.Error())
	}
}

// TODO:
func TestSingleOperations(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subtests := []struct {
		name      string
		operation string
		args      []interface{}
		result    interface{}
	}{
		{
			name:      "Set",
			operation: "Set",
			args:      []interface{}{"key", "value"},
			result:    "Ok",
		},
		{
			name:      "SetWithExpiration",
			operation: "SetWithExpiration",
			args:      []interface{}{"key", "value", time.Hour},
			result:    true,
		},
		{
			name:      "AddToVector",
			operation: "AddToVector",
			args:      []interface{}{"key", "value1", "value2"},
			result:    int64(2),
		},
		{
			name:      "Get",
			operation: "Get",
			args:      []interface{}{"key"},
			result:    "value",
		},
		{
			name:      "Delete",
			operation: "Delete",
			args:      []interface{}{"key"},
			result:    int64(1),
		},
		{
			name:      "DeleteFromVector",
			operation: "DeleteFromVector",
			args:      []interface{}{"key", "value1", "value2"},
			result:    int64(2),
		},
		{
			name:      "GetVector",
			operation: "GetVector",
			args:      []interface{}{"key"},
			result:    []string{"value1", "value2"},
		},
		{
			name:      "AssignExpiration",
			operation: "AssignExpiration",
			args:      []interface{}{"key", time.Hour},
			result:    true,
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			var (
				result interface{}
				err    error
			)

			client, mock := redismock.NewClientMock()
			dc := dragonFlyClient{dragonflyClientCommands{client}}
			ctx := context.Background()
			key := test.args[0].(string)

			switch test.operation {
			case "Set":
				{
					mock.
						ExpectSet(key, test.args[1], 0).
						SetVal(test.result.(string))
					result, err = dc.Set(ctx, key, test.args[1]).Result()

					break
				}

			case "SetWithExpiration":
				{
					mock.
						ExpectSetNX(key, test.args[1], test.args[2].(time.Duration)).
						SetVal(true)
					result, err = dc.
						SetWithExpiration(
							ctx,
							key,
							test.args[1],
							test.args[2].(time.Duration),
						).
						Result()

					break
				}

			case "AddToVector":
				{
					values := test.args[1:]
					mock.
						ExpectSAdd(key, values...).
						SetVal(test.result.(int64))
					result, err = dc.
						AddToVector(ctx, key, values...).
						Result()

					break
				}

			case "Get":
				{
					mock.ExpectGet(key).SetVal(test.result.(string))
					result, err = dc.Get(ctx, key).Result()

					break
				}

			case "Delete":
				{
					mock.ExpectDel(key).SetVal(test.result.(int64))
					result, err = dc.Delete(ctx, key).Result()

					break
				}

			case "DeleteFromVector":
				{
					values := test.args[1:]
					mock.ExpectSRem(key, values...).SetVal(test.result.(int64))
					result, err = dc.DeleteFromVector(ctx, key, values...).Result()

					break
				}

			case "GetVector":
				{
					mock.
						ExpectSMembers(key).
						SetVal([]string{test.result.([]string)[0], test.result.([]string)[1]})
					result, err = dc.GetVector(ctx, key).Result()

					break
				}

			case "AssignExpiration":
				{
					mock.
						ExpectExpire(key, test.args[1].(time.Duration)).
						SetVal(test.result.(bool))
					result, err = dc.
						AssignExpiration(ctx, key, test.args[1].(time.Duration)).
						Result()

					break
				}
			}

			checkMockExpectationsWereMet(t, mock)
			require.Nil(err)
			resultType := reflect.TypeOf(test.result).Kind()

			if resultType == reflect.Array {
				require.ElementsMatch(test.result, result)
			} else {
				require.Equal(test.result, result)
			}
		})
	}
}

func TestNoItemError(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	subtests := []struct {
		name      string
		operation string
		args      []interface{}
	}{
		{
			name:      "Get",
			operation: "Get",
			args:      []interface{}{"key"},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			var (
				result interface{}
				err    error
			)

			client, mock := redismock.NewClientMock()
			dc := dragonFlyClient{dragonflyClientCommands{client}}
			ctx := context.Background()
			key := test.args[0].(string)

			switch test.operation {
			case "Get":
				{
					mock.ExpectGet(key).SetErr(redis.Nil)
					result, err = dc.Get(ctx, key).Result()

					break
				}
			}

			checkMockExpectationsWereMet(t, mock)
			resultType := reflect.TypeOf(result).Kind()

			if resultType == reflect.String {
				require.Equal(0, len(result.(string)))
			} else {
				require.Nil(result)
			}

			require.EqualError(err, NO_ITEM_ERROR)
		})
	}
}

func TestTransaction(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	client, mock := redismock.NewClientMock()
	dc := dragonFlyClient{dragonflyClientCommands{client}}
	ctx := context.Background()

	mock.ExpectTxPipeline()
	tx := dc.StartTransaction()
	require.NotNil(tx)

	mock.ExpectSet("key", "value", 0).SetVal("OK")
	setCmd := tx.Set(ctx, "key", "value")

	mock.ExpectTxPipelineExec().SetVal([]interface{}{})
	err := tx.Exec(ctx)
	require.Nil(err)

	checkMockExpectationsWereMet(t, mock)

	setCmdResult, err := setCmd.Result()
	require.Nil(err)
	require.Equal(setCmdResult, "OK")
}

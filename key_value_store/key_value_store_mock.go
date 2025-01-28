package key_value_store

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

type commandResultMock[T interface{}] struct {
	mock.Mock
}

func CreateResultMock[T interface{}]() commandResultMock[T] {
	return commandResultMock[T]{}
}

func (self *commandResultMock[T]) Result() (T, error) {
	args := self.Called()

	return args.Get(0).(T), args.Error(1)
}

type keyValueStoreCommandsMock struct {
	mock.Mock
}

func (self *keyValueStoreCommandsMock) Set(
	ctx context.Context,
	key string,
	value interface{},
) CommandResult[string] {
	args := self.Called(ctx, key, value)

	return args.Get(0).(CommandResult[string])
}

func (self *keyValueStoreCommandsMock) SetWithExpiration(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) CommandResult[bool] {
	args := self.Called(ctx, key, value, expiration)

	return args.Get(0).(CommandResult[bool])
}

func (self *keyValueStoreCommandsMock) AddToVector(
	ctx context.Context,
	key string,
	values ...interface{},
) CommandResult[int64] {
	calledArguments := []interface{}{ctx, key}

	calledArguments = append(calledArguments, values...)

	args := self.Called(calledArguments...)

	return args.Get(0).(CommandResult[int64])
}

func (self *keyValueStoreCommandsMock) Get(
	ctx context.Context,
	key string,
) CommandResult[string] {
	args := self.Called(ctx, key)

	return args.Get(0).(CommandResult[string])
}

func (self *keyValueStoreCommandsMock) Delete(
	ctx context.Context,
	key string,
) CommandResult[int64] {
	args := self.Called(ctx, key)

	return args.Get(0).(CommandResult[int64])
}

func (self *keyValueStoreCommandsMock) DeleteFromVector(
	ctx context.Context,
	key string,
	values ...interface{},
) CommandResult[int64] {
	calledArguments := []interface{}{ctx, key}

	calledArguments = append(calledArguments, values...)

	args := self.Called(calledArguments...)

	return args.Get(0).(CommandResult[int64])
}

func (self *keyValueStoreCommandsMock) GetVector(
	ctx context.Context,
	key string,
) CommandResult[[]string] {
	args := self.Called(ctx, key)

	return args.Get(0).(CommandResult[[]string])
}

func (self *keyValueStoreCommandsMock) AssignExpiration(
	ctx context.Context,
	key string,
	expiration time.Duration,
) CommandResult[bool] {
	args := self.Called(ctx, key, expiration)

	return args.Get(0).(CommandResult[bool])
}

type KeyValueStoreTransactionMock struct {
	keyValueStoreCommandsMock
}

func (self *KeyValueStoreTransactionMock) Exec(ctx context.Context) error {
	args := self.Called(ctx)

	return args.Error(0)
}

type KeyValueStoreMock struct {
	keyValueStoreCommandsMock
}

func (self *KeyValueStoreMock) StartTransaction() KeyValueStoreTransaction {
	args := self.Called()

	return args.Get(0).(KeyValueStoreTransaction)
}

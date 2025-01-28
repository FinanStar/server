package key_value_store

import (
	"context"
	"time"
)

type Error string

func (self Error) Error() string { return string(self) }

const (
	NO_ITEM_ERROR = "There is no items"

	NoItem = Error(NO_ITEM_ERROR)
)

type CommandResult[T interface{}] interface {
	Result() (T, error)
}

type KeyValueStoreCommands interface {
	Set(ctx context.Context, key string, value interface{}) CommandResult[string]
	SetWithExpiration(
		ctx context.Context,
		key string,
		value interface{},
		expiration time.Duration,
	) CommandResult[bool]
	AddToVector(
		ctx context.Context,
		key string,
		values ...interface{},
	) CommandResult[int64]
	Get(ctx context.Context, key string) CommandResult[string]
	Delete(ctx context.Context, key string) CommandResult[int64]
	DeleteFromVector(
		ctx context.Context,
		key string,
		values ...interface{},
	) CommandResult[int64]
	GetVector(ctx context.Context, key string) CommandResult[[]string]
	AssignExpiration(
		ctx context.Context,
		key string,
		expiration time.Duration,
	) CommandResult[bool]
}

type KeyValueStoreTransaction interface {
	KeyValueStoreCommands
	Exec(ctx context.Context) error
}

type KeyValueStore interface {
	KeyValueStoreCommands
	StartTransaction() KeyValueStoreTransaction
}

type kvsBuilder interface {
	Build() KeyValueStore
}

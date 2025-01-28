package key_value_store

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type DragonflyBuilder struct {
	client redis.Cmdable
}

func (self *DragonflyBuilder) Client(client redis.Cmdable) *DragonflyBuilder {
	self.client = client

	return self
}

func (self DragonflyBuilder) Build() KeyValueStore {
	return dragonFlyClient{dragonflyClientCommands{client: self.client}}
}

type redisCmd[T interface{}] interface {
	Result() (T, error)
}

type dragonflyClientCommands struct {
	client redis.Cmdable
}

type commandResult[T interface{}] struct {
	cmd redisCmd[T]
}

func (self commandResult[T]) Result() (T, error) {
	value, error := self.cmd.Result()

	if error == redis.Nil {
		var placeholder T

		return placeholder, NoItem
	}

	return value, nil
}

func (self dragonflyClientCommands) Set(
	ctx context.Context,
	key string,
	value interface{},
) CommandResult[string] {
	cmd := self.client.Set(ctx, key, value, 0)

	return commandResult[string]{cmd}
}

func (self dragonflyClientCommands) SetWithExpiration(
	ctx context.Context,
	key string,
	value interface{},
	expiration time.Duration,
) CommandResult[bool] {
	cmd := self.client.SetNX(ctx, key, value, expiration)

	return commandResult[bool]{cmd}
}

func (self dragonflyClientCommands) AddToVector(
	ctx context.Context,
	key string,
	values ...interface{},
) CommandResult[int64] {
	cmd := self.client.SAdd(ctx, key, values...)

	return commandResult[int64]{cmd}
}

func (self dragonflyClientCommands) Get(
	ctx context.Context,
	key string,
) CommandResult[string] {
	cmd := self.client.Get(ctx, key)

	return commandResult[string]{cmd}
}

func (self dragonflyClientCommands) Delete(
	ctx context.Context,
	key string,
) CommandResult[int64] {
	cmd := self.client.Del(ctx, key)

	return commandResult[int64]{cmd}
}

func (self dragonflyClientCommands) DeleteFromVector(
	ctx context.Context,
	key string,
	values ...interface{},
) CommandResult[int64] {
	cmd := self.client.SRem(ctx, key, values...)

	return commandResult[int64]{cmd}
}

func (self dragonflyClientCommands) GetVector(
	ctx context.Context,
	key string,
) CommandResult[[]string] {
	cmd := self.client.SMembers(ctx, key)

	return commandResult[[]string]{cmd}
}

func (self dragonflyClientCommands) AssignExpiration(
	ctx context.Context,
	key string,
	expiration time.Duration,
) CommandResult[bool] {
	cmd := self.client.Expire(ctx, key, expiration)

	return commandResult[bool]{cmd}
}

type dragonFlyTransaction struct {
	dragonflyClientCommands
}

func (self dragonFlyTransaction) Exec(ctx context.Context) error {
	_, err := self.client.(redis.Pipeliner).Exec(ctx)

	return err
}

type dragonFlyClient struct {
	dragonflyClientCommands
}

func (self dragonFlyClient) StartTransaction() KeyValueStoreTransaction {
	tx := self.client.TxPipeline()

	return dragonFlyTransaction{
		dragonflyClientCommands{
			client: tx,
		},
	}
}

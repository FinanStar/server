package user

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type expectTuple struct {
	User  *userEntity
	Error error
}

type userRepositoryMock struct {
	mock.Mock
}

func NewRepositoryMock() *userRepositoryMock {
	return new(userRepositoryMock)
}

func (self *userRepositoryMock) Create(
	ctx context.Context,
	dto createUserRepositoryDto,
) (*userEntity, error) {
	args := self.Called(ctx, dto)

	return args.Get(0).(*userEntity), args.Error(1)
}

func (self *userRepositoryMock) Update(
	ctx context.Context,
	id uint32,
	dto updateUserRepositoryDto,
) (*userEntity, error) {
	args := self.Called(ctx, id, dto)

	return args.Get(0).(*userEntity), args.Error(1)
}

func (self *userRepositoryMock) GetByLogin(
	ctx context.Context,
	login string,
) (*userEntity, error) {
	args := self.Called(ctx, login)

	return args.Get(0).(*userEntity), args.Error(1)
}

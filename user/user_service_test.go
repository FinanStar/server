package user

import (
	"context"
	"errors"
	"finanstar/server/crypto"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServiceCreate(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	hashedPassword, err := crypto.HashPassword(`secure-password`)

	require.Nil(err)

	subtests := []struct {
		name   string
		dto    CreateUserDto
		result expectTuple
	}{
		{
			name: `CreateNewUser`,
			dto: CreateUserDto{
				Login:    `test@example.com`,
				Password: `secure-password`,
			},
			result: expectTuple{
				User: &userEntity{
					Id:       1,
					Login:    `test@example.com`,
					Password: hashedPassword,
				},
			},
		},
		{
			name: `CreateDuplicateUser`,
			dto: CreateUserDto{
				Login:    `test@example.com`,
				Password: `secure-password`,
			},
			result: expectTuple{
				User:  nil,
				Error: errors.New(USER_ALREADY_EXISTS_ERROR),
			},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			userRepository := NewTestUserRepository()
			userService := NewUserService(&userRepository)

			userRepository.CreateExpectResult(test.result.User, test.result.Error)
			createdUser, err := userService.Create(
				context.Background(),
				test.dto,
			)

			if test.result.Error != nil {
				require.EqualError(err, test.result.Error.Error())
			}

			if test.result.User != nil {
				expectedUser := *test.result.User

				require.Equal(createdUser.Id, expectedUser.Id)
				require.Equal(createdUser.Login, expectedUser.Login)
				require.Equal(createdUser.Password, expectedUser.Password)
			}
		})
	}
}

func TestServiceUpdate(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	testUser := UserDto{
		Id:       1,
		Login:    `test@example.com`,
		Password: `secure-password`,
	}
	hashedPassword, err := crypto.HashPassword(testUser.Password)

	require.Nil(err)

	subtests := []struct {
		name   string
		userId uint32
		dto    UpdateUserDto
		result expectTuple
	}{
		{
			name:   `UpdateUserLoginAndPassword`,
			userId: testUser.Id,
			dto: UpdateUserDto{
				Login:    &testUser.Login,
				Password: &testUser.Password,
			},
			result: expectTuple{
				User: &userEntity{
					Id:       testUser.Id,
					Login:    `test@example.com`,
					Password: hashedPassword,
				},
			},
		},
		{
			name:   `UpdateUnknownUser`,
			userId: testUser.Id,
			dto: UpdateUserDto{
				Login:    &testUser.Login,
				Password: &testUser.Password,
			},
			result: expectTuple{
				Error: errors.New(USER_NOT_FOUND_ERROR),
			},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			userRepository := NewTestUserRepository()
			userService := NewUserService(&userRepository)

			userRepository.UpdateExpectResult(test.result.User, test.result.Error)
			updatedUser, err := userService.Update(
				context.Background(),
				test.userId,
				test.dto,
			)

			if test.result.User != nil {
				expectedUser := *test.result.User

				require.Nil(err)
				require.Equal(updatedUser.Id, expectedUser.Id)
				require.Equal(updatedUser.Login, expectedUser.Login)
				require.Equal(updatedUser.Password, expectedUser.Password)
			}

			if test.result.Error != nil {
				require.EqualError(err, test.result.Error.Error())
			}
		})
	}
}

func TestServiceGetByLogin(t *testing.T) {
	t.Parallel()
	require := require.New(t)
	hashedPassword, err := crypto.HashPassword(`secure-password`)

	require.Nil(err)

	testUser := UserDto{
		Id:       1,
		Login:    `test@example.com`,
		Password: hashedPassword,
	}

	subtests := []struct {
		name   string
		userId uint32
		result expectTuple
	}{
		{
			name:   `GetExistedUser`,
			userId: testUser.Id,
			result: expectTuple{
				User: &userEntity{
					Id:       testUser.Id,
					Login:    testUser.Login,
					Password: testUser.Password,
				},
			},
		},
		{
			name:   `GetNonExistedUser`,
			userId: testUser.Id,
			result: expectTuple{
				Error: errors.New(USER_NOT_FOUND_ERROR),
			},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			userRepository := NewTestUserRepository()
			userService := NewUserService(&userRepository)

			userRepository.GetByLoginExpectResult(test.result.User, test.result.Error)
			user, err := userService.GetByLogin(context.Background(), testUser.Login)

			if test.result.User != nil {
				require.Nil(err)
				require.Equal(user.Id, testUser.Id)
				require.Equal(user.Login, testUser.Login)
				require.Equal(user.Password, testUser.Password)
			}

			if test.result.Error != nil {
				require.EqualError(err, test.result.Error.Error())
			}
		})
	}
}

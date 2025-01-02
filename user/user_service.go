package user

import (
	"context"
	"finanstar/server/crypto"
)

type UserService struct {
	repository UserRepository
}

type UserDto struct {
	Id       uint32
	Login    string
	Password string
}

type UpdateUserDto struct {
	Login    *string
	Password *string
}

type CreateUserDto struct {
	Login    string
	Password string
}

func NewUserService(repository UserRepository) UserService {
	return UserService{repository}
}

func makeUserDto(user *userEntity) *UserDto {
	return &UserDto{
		Id:       user.Id,
		Login:    user.Login,
		Password: user.Password,
	}
}

func (self *UserService) GetByLogin(
	ctx context.Context,
	login string,
) (*UserDto, error) {
	userEntity, err := self.repository.GetByLogin(ctx, login)

	if err != nil {
		return nil, err
	}

	return makeUserDto(userEntity), nil
}

func (self *UserService) Update(
	ctx context.Context,
	id uint32,
	dto UpdateUserDto,
) (*UserDto, error) {
	var hashedPassword *string

	if dto.Password != nil {
		password, err := crypto.HashPassword(*dto.Password)

		if err != nil {
			return nil, err
		}

		hashedPassword = &password
	}

	userEntity, err := self.repository.Update(
		ctx,
		id,
		updateUserRepositoryDto{Login: dto.Login, Password: hashedPassword},
	)

	if err != nil {
		return nil, err
	}

	return makeUserDto(userEntity), nil
}

func (self *UserService) Create(
	ctx context.Context,
	dto CreateUserDto,
) (*UserDto, error) {
	hashedPassword, err := crypto.HashPassword(dto.Password)

	if err != nil {
		return nil, err
	}

	userEntity, err := self.repository.Create(
		ctx,
		createUserRepositoryDto{Login: dto.Login, Password: hashedPassword},
	)

	if err != nil {
		return nil, err
	}

	return makeUserDto(userEntity), nil
}

package user

import "context"

const (
	USER_NOT_FOUND_ERROR            = "User not found"
	THERE_IS_NO_UPDATE_PARAMS_ERROR = "There is not update params"
	USER_ALREADY_EXISTS_ERROR       = "User already exists"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*userEntity, error)
	Update(ctx context.Context, id uint32, dto updateUserRepositoryDto) (*userEntity, error)
	Create(ctx context.Context, dto createUserRepositoryDto) (*userEntity, error)
}

type userEntity struct {
	Id       uint32
	Login    string
	Password string
}

type updateUserRepositoryDto struct {
	Login    *string
	Password *string
}

type createUserRepositoryDto struct {
	Login    string
	Password string
}

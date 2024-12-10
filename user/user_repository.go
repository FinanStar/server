package user

import "context"

const (
	USER_NOT_FOUND_ERROR            = "User not found"
	THERE_IS_NO_UPDATE_PARAMS_ERROR = "There is not update params"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*UserEntity, error)
	Update(ctx context.Context, id uint32, dto UpdateUserDto) (*UserEntity, error)
}

type UserEntity struct {
	Id       uint32
	Login    string
	Password string
}

type UpdateUserDto struct {
	Login    *string
	Password *string
}

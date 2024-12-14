package user

import "context"

const (
	USER_NOT_FOUND_ERROR            = "User not found"
	THERE_IS_NO_UPDATE_PARAMS_ERROR = "There is not update params"
)

type UserRepository interface {
	GetByLogin(ctx context.Context, login string) (*userEntity, error)
	Update(ctx context.Context, id uint32, dto updateUserRepositoryDto) (*userEntity, error)
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

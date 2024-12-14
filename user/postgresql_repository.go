package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"finanstar/server/utils"
)

func NewPostgresqlUserRepository(db utils.PgxPoolIface) postgresqlUserRepository {
	return postgresqlUserRepository{db}
}

type postgresqlUserRepository struct {
	db utils.PgxPoolIface
}

func (self *postgresqlUserRepository) GetByLogin(
	ctx context.Context,
	login string,
) (*userEntity, error) {
	user := userEntity{
		Id:       0,
		Login:    ``,
		Password: ``,
	}

	err := self.db.
		QueryRow(
			ctx,
			`SELECT id, login, password FROM users WHERE login = $1;`,
			login,
		).
		Scan(&user.Id, &user.Login, &user.Password)

	if err == pgx.ErrNoRows {
		return nil, errors.New(USER_NOT_FOUND_ERROR)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (self *postgresqlUserRepository) Update(
	ctx context.Context,
	id uint32,
	dto updateUserRepositoryDto,
) (*userEntity, error) {
	user := userEntity{
		Id:       id,
		Login:    ``,
		Password: ``,
	}
	queryArgs := make([]interface{}, 1)
	updateParams := make([]string, 0)

	queryArgs[0] = id

	if dto.Login != nil {
		updateParams = append(
			updateParams,
			fmt.Sprintf(`login = $%d`, len(queryArgs)+1),
		)
		queryArgs = append(queryArgs, *dto.Login)
	}

	if dto.Password != nil {
		updateParams = append(
			updateParams,
			fmt.Sprintf(`password = $%d`, len(queryArgs)+1),
		)
		queryArgs = append(queryArgs, *dto.Password)
	}

	if len(updateParams) == 0 {
		return nil, errors.New(THERE_IS_NO_UPDATE_PARAMS_ERROR)
	}

	err := self.db.
		QueryRow(
			ctx,
			fmt.Sprintf(
				`UPDATE users WHERE id = $1 SET %s RETURNING login, password;`,
				strings.Join(updateParams, `,`),
			),
			queryArgs...,
		).
		Scan(&user.Login, &user.Password)

	if err == pgx.ErrNoRows {
		return nil, errors.New(USER_NOT_FOUND_ERROR)
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (self *postgresqlUserRepository) Create(
	ctx context.Context,
	dto createUserRepositoryDto,
) (*userEntity, error) {
	user := userEntity{
		Id:       0,
		Login:    ``,
		Password: ``,
	}

	err := self.db.
		QueryRow(
			ctx,
			`
				INSERT INTO users (login, password)
				VALUES ($1, $2)
				RETURNING id, login, password;
			`,
			dto.Login,
			dto.Password,
		).
		Scan(&user.Id, &user.Login, &user.Password)

	if err != nil {
		if strings.Contains(err.Error(), utils.DUPLICATE_VALUE_ERROR) {
			return nil, errors.New(USER_ALREADY_EXISTS_ERROR)
		}

		return nil, err
	}

	return &user, nil
}

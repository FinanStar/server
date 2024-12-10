package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"finanstar/server/utils"
)

func NewPostgresqlUserRepository(db utils.PgxPoolIface) PostgresqlUserRepository {
	return PostgresqlUserRepository{db}
}

type PostgresqlUserRepository struct {
	db utils.PgxPoolIface
}

func (pur *PostgresqlUserRepository) GetByLogin(
	ctx context.Context,
	login string,
) (*UserEntity, error) {
	user := UserEntity{
		Id:       0,
		Login:    ``,
		Password: ``,
	}

	err := pur.db.
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

func (pur *PostgresqlUserRepository) Update(
	ctx context.Context,
	id uint32,
	dto UpdateUserDto,
) (*UserEntity, error) {
	user := UserEntity{
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
		queryArgs = append(queryArgs, dto.Login)
	}

	if dto.Password != nil {
		updateParams = append(
			updateParams,
			fmt.Sprintf(`password = $%d`, len(queryArgs)+1),
		)
		queryArgs = append(queryArgs, dto.Password)
	}

	if len(updateParams) == 0 {
		return nil, errors.New(THERE_IS_NO_UPDATE_PARAMS_ERROR)
	}

	err := pur.db.
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

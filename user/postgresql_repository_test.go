package user

import (
	"context"
	"errors"
	utils_pgx "finanstar/server/utils"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

type result struct {
	user    *userEntity
	dbError error
	error   error
}

func TestRepositoryGetByLogin(t *testing.T) {
	t.Parallel()

	expectedSql := `SELECT id, login, password FROM users WHERE login = \$1;`
	testLogin := `test@example.com`

	subtests := []struct {
		name   string
		result result
	}{
		{
			name: `ReturnsUser`,
			result: result{
				user: &userEntity{
					Id:       1,
					Login:    testLogin,
					Password: `hashed_password`,
				},
			},
		},
		{
			name: `ReturnsUserNotFoundError`,
			result: result{
				error: errors.New(USER_NOT_FOUND_ERROR),
			},
		},
		{
			name: `ReturnsUnknownError`,
			result: result{
				dbError: errors.New(`UnknownError`),
			},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			require := require.New(t)
			db, err := pgxmock.NewPool()

			require.Nil(err)
			pur := postgresqlUserRepository{db: db}

			rows := db.NewRows([]string{`id`, `login`, `password`})
			resultUser := test.result.user

			if resultUser != nil {
				rows.AddRow(resultUser.Id, resultUser.Login, resultUser.Password)
			}

			query := db.ExpectQuery(expectedSql).WithArgs(testLogin)

			if test.result.dbError != nil {
				query.WillReturnError(test.result.dbError)
			} else {
				query.WillReturnRows(rows)
			}

			user, err := pur.GetByLogin(context.Background(), testLogin)

			if resultUser != nil {
				require.Nil(err)

				require.NotNil(user)
				require.Equal(resultUser.Id, user.Id)
				require.Equal(resultUser.Login, user.Login)
				require.Equal(resultUser.Password, user.Password)
			} else {
				require.Nil(user)

				require.NotNil(err)

				if test.result.error != nil {
					require.EqualError(err, test.result.error.Error())
				} else {
					require.EqualError(err, test.result.dbError.Error())
				}
			}
		})
	}
}

func TestRepositoryUpdate(t *testing.T) {
	t.Parallel()

	type updateDto struct {
		login    string
		password string
	}

	subtests := []struct {
		name        string
		userId      uint32
		expectedSql string
		updateDto   updateDto
		result      result
	}{
		{
			name:   `UpdateAllFields`,
			userId: 1,
			expectedSql: `
				UPDATE users
				WHERE id = \$1
				SET login = \$2,password = \$3
				RETURNING login, password;
			`,
			updateDto: updateDto{
				login:    `test@example.com`,
				password: `hashed_password`,
			},
			result: result{
				user: &userEntity{
					Id:       1,
					Login:    `test@example.com`,
					Password: `hashed_password`,
				},
			},
		},
		{
			name:        `ReturnsNoUpdateParamsError`,
			userId:      1,
			expectedSql: ``,
			updateDto: updateDto{
				login:    ``,
				password: ``,
			},
			result: result{
				error: errors.New(THERE_IS_NO_UPDATE_PARAMS_ERROR),
			},
		},
		{
			name:   `UpdateOnlyLoginField`,
			userId: 1,
			expectedSql: `
				UPDATE users
				WHERE id = \$1
				SET login = \$2
				RETURNING login, password;
			`,
			updateDto: updateDto{
				login:    `test@example.com`,
				password: ``,
			},
			result: result{
				user: &userEntity{
					Id:       1,
					Login:    `test@example.com`,
					Password: `hashed_password`,
				},
			},
		},
		{
			name:   `UpdateOnlyPasswordField`,
			userId: 1,
			expectedSql: `
				UPDATE users
				WHERE id = \$1
				SET password = \$2
				RETURNING login, password;
			`,
			updateDto: updateDto{
				login:    ``,
				password: `hashed_password`,
			},
			result: result{
				user: &userEntity{
					Id:       1,
					Login:    `test@example.com`,
					Password: `hashed_password`,
				},
			},
		},
		{
			name:   `UpdateNonExistingUser`,
			userId: 1,
			expectedSql: `
				UPDATE users
				WHERE id = \$1
				SET login = \$2,password = \$3
				RETURNING login, password;
			`,
			updateDto: updateDto{
				login:    `test@example.com`,
				password: `hashed_password`,
			},
			result: result{
				error: errors.New(USER_NOT_FOUND_ERROR),
			},
		},
	}

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			require := require.New(t)
			db, err := pgxmock.NewPool()
			require.Nil(err)
			pur := postgresqlUserRepository{db: db}

			rows := db.NewRows([]string{`login`, `password`})
			resultUser := test.result.user

			if resultUser != nil {
				rows.AddRow(resultUser.Login, resultUser.Password)
			}

			args := []any{test.userId}
			dto := updateUserRepositoryDto{}

			if len(test.updateDto.login) != 0 {
				args = append(args, test.updateDto.login)
				dto.Login = &test.updateDto.login
			}

			if len(test.updateDto.password) != 0 {
				args = append(args, test.updateDto.password)
				dto.Password = &test.updateDto.password
			}

			query := db.ExpectQuery(test.expectedSql).WithArgs(args...)

			if test.result.dbError != nil {
				query.WillReturnError(test.result.dbError)
			} else {
				query.WillReturnRows(rows)
			}

			user, err := pur.Update(
				context.Background(),
				test.userId,
				dto,
			)

			if resultUser != nil {
				require.Nil(err)

				require.NotNil(user)
				require.Equal(resultUser.Id, user.Id)
				require.Equal(resultUser.Login, user.Login)
				require.Equal(resultUser.Password, user.Password)
			} else {
				require.Nil(user)

				require.NotNil(err)

				if test.result.error != nil {
					require.EqualError(err, test.result.error.Error())
				} else {
					require.EqualError(err, test.result.dbError.Error())
				}
			}
		})
	}
}

func TestRepositoryCreate(t *testing.T) {
	t.Parallel()

	type createDto struct {
		login    string
		password string
	}

	subtests := []struct {
		name      string
		createDto createDto
		result    result
	}{
		{
			name: `CreateNewUser`,
			createDto: createDto{
				login:    `test@example.com`,
				password: `hashed_password`,
			},
			result: result{
				user: &userEntity{
					Id:       1,
					Login:    `test@example.com`,
					Password: `hashed_password`,
				},
			},
		},
		{
			name: `CreateDuplicateUser`,
			createDto: createDto{
				login:    `test@example.com`,
				password: `hashed_password`,
			},
			result: result{
				dbError: errors.New(utils_pgx.DUPLICATE_VALUE_ERROR),
				error:   errors.New(USER_ALREADY_EXISTS_ERROR),
			},
		},
		{
			name: `ReturnsUnknownError`,
			createDto: createDto{
				login:    `test@example.com`,
				password: `hashed_password`,
			},
			result: result{
				dbError: errors.New(`Unknown Error`),
			},
		},
	}

	expectedSql := `
		INSERT INTO users \(login, password\)
		VALUES \(\$1, \$2\)
		RETURNING id, login, password;
	`

	for _, test := range subtests {
		t.Run(test.name, func(tt *testing.T) {
			db, err := pgxmock.NewPool()
			require := require.New(t)

			require.Nil(err)
			pur := postgresqlUserRepository{db: db}
			resultUser := test.result.user
			rows := db.NewRows([]string{})

			if resultUser != nil {
				rows = db.NewRows([]string{`id`, `login`, `password`})
				rows.AddRow(resultUser.Id, resultUser.Login, resultUser.Password)
			}

			query := db.
				ExpectQuery(expectedSql).
				WithArgs(test.createDto.login, test.createDto.password)

			if test.result.dbError != nil {
				query.WillReturnError(test.result.dbError)
			} else {
				query.WillReturnRows(rows)
			}

			user, err := pur.Create(
				context.Background(),
				createUserRepositoryDto{
					Login:    test.createDto.login,
					Password: test.createDto.password,
				},
			)

			if resultUser != nil {
				require.Nil(err)

				require.NotNil(user)
				require.Equal(resultUser.Id, user.Id)
				require.Equal(resultUser.Login, user.Login)
				require.Equal(resultUser.Password, user.Password)
			} else {
				require.Nil(user)

				require.NotNil(err)

				if test.result.error != nil {
					require.EqualError(err, test.result.error.Error())
				} else {
					require.EqualError(err, test.result.dbError.Error())
				}
			}
		})
	}
}

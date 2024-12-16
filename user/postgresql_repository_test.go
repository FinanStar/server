package user

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestGetByLogin(t *testing.T) {
	t.Parallel()
	db, err := pgxmock.NewPool()
	require := require.New(t)

	require.Nil(err)
	pur := postgresqlUserRepository{db: db}
	expectedSql := `SELECT id, login, password FROM users WHERE login = \$1;`

	t.Run(`ReturnsUser`, func(tt *testing.T) {
		testLogin := `test@example.com`
		rows := db.NewRows([]string{`id`, `login`, `password`})

		rows.AddRow(uint32(1), testLogin, `hashed_password`)

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnRows(rows)

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(err)
		require.NotNilf(user, "Expected get user, but got nil")
		require.Equalf(
			user.Id,
			uint32(1),
			"Expected get id = 1, but got %d",
			user.Id,
		)
		require.Equalf(
			user.Login,
			testLogin,
			"Expected get login = '%s', but got '%s'",
			testLogin,
			user.Login,
		)
		require.Equalf(
			user.Password,
			`hashed_password`,
			"Expected get password = 'hashed_password', but got '%s'",
			user.Password,
		)
	})

	t.Run(`ReturnsUserNotFoundError`, func(tt *testing.T) {
		testLogin := `test@example.com`
		rows := db.NewRows([]string{})

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnRows(rows)

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(user)
		require.EqualError(err, USER_NOT_FOUND_ERROR)
	})

	t.Run(`ReturnsUnknownError`, func(tt *testing.T) {
		testLogin := `test@example.com`
		errorMessage := `UnknownError`

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnError(errors.New(errorMessage))

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(user)
		require.EqualError(err, errorMessage)
	})
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	db, err := pgxmock.NewPool()
	require := require.New(t)

	require.Nil(err)
	pur := postgresqlUserRepository{db: db}

	t.Run(`UpdateAllFields`, func(tt *testing.T) {
		expectedUser := userEntity{
			Id:       1,
			Login:    `test@example.com`,
			Password: `hashed_password`,
		}

		rows := db.NewRows([]string{"login", "password"})
		expectedSql := `UPDATE users WHERE id = \$1 SET login = \$2,password = \$3 RETURNING login, password;`

		rows.AddRow(expectedUser.Login, expectedUser.Password)

		db.
			ExpectQuery(expectedSql).
			WithArgs(expectedUser.Id, expectedUser.Login, expectedUser.Password).
			WillReturnRows(rows)
		user, err := pur.Update(
			context.Background(),
			expectedUser.Id,
			updateUserRepositoryDto{
				Login:    &expectedUser.Login,
				Password: &expectedUser.Password,
			},
		)

		require.Nil(err)
		require.NotNilf(user, "Expected get user, but got nil")
		require.Equal(expectedUser.Id, user.Id)
		require.Equal(expectedUser.Login, user.Login)
		require.Equal(expectedUser.Password, user.Password)
	})

	t.Run(`ReturnsNoUpdateParamsError`, func(tt *testing.T) {
		user, err := pur.Update(context.Background(), uint32(1), updateUserRepositoryDto{})

		require.Nil(user)
		require.EqualError(err, THERE_IS_NO_UPDATE_PARAMS_ERROR)
	})

	t.Run(`UpdateOnlyLoginField`, func(tt *testing.T) {
		expectedUser := userEntity{
			Id:       1,
			Login:    `test@example.com`,
			Password: `hashed_password`,
		}

		rows := db.NewRows([]string{"login", "password"})
		expectedSql := `UPDATE users WHERE id = \$1 SET login = \$2 RETURNING login, password;`

		rows.AddRow(expectedUser.Login, expectedUser.Password)

		db.
			ExpectQuery(expectedSql).
			WithArgs(expectedUser.Id, expectedUser.Login).
			WillReturnRows(rows)
		user, err := pur.Update(
			context.Background(),
			expectedUser.Id,
			updateUserRepositoryDto{
				Login: &expectedUser.Login,
			},
		)

		require.Nil(err)
		require.NotNilf(user, "Expected get user, but got nil")
		require.Equal(expectedUser.Id, user.Id)
		require.Equal(expectedUser.Login, user.Login)
		require.Equal(expectedUser.Password, user.Password)
	})

	t.Run(`UpdateOnlyPasswordField`, func(tt *testing.T) {
		expectedUser := userEntity{
			Id:       1,
			Login:    `test@example.com`,
			Password: `hashed_password`,
		}

		rows := db.NewRows([]string{"login", "password"})
		expectedSql := `UPDATE users WHERE id = \$1 SET password = \$2 RETURNING login, password;`

		rows.AddRow(expectedUser.Login, expectedUser.Password)

		db.
			ExpectQuery(expectedSql).
			WithArgs(expectedUser.Id, expectedUser.Password).
			WillReturnRows(rows)
		user, err := pur.Update(
			context.Background(),
			expectedUser.Id,
			updateUserRepositoryDto{
				Password: &expectedUser.Password,
			},
		)

		require.Nil(err)
		require.NotNilf(user, "Expected get user, but got nil")
		require.Equal(expectedUser.Id, user.Id)
		require.Equal(expectedUser.Login, user.Login)
		require.Equal(expectedUser.Password, user.Password)
	})

	t.Run(`UpdateNonExistingUser`, func(tt *testing.T) {
		expectedUser := userEntity{
			Id:       1,
			Login:    `test@example.com`,
			Password: `hashed_password`,
		}

		rows := db.NewRows([]string{"login", "password"})
		expectedSql := `
			UPDATE users
			WHERE id = \$1
			SET login = \$2,password = \$3
			RETURNING login, password;
		`

		db.
			ExpectQuery(expectedSql).
			WithArgs(expectedUser.Id, expectedUser.Login, expectedUser.Password).
			WillReturnRows(rows)
		user, err := pur.Update(
			context.Background(),
			expectedUser.Id,
			updateUserRepositoryDto{
				Login:    &expectedUser.Login,
				Password: &expectedUser.Password,
			},
		)

		require.Nil(user)
		require.EqualError(err, USER_NOT_FOUND_ERROR)
	})
}

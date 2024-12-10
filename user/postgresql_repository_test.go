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
	pur := PostgresqlUserRepository{db: db}
	expectedSql := `SELECT id, login, password FROM users WHERE login = \$1;`

	t.Run(`DefinedUser`, func(tt *testing.T) {
		testLogin := `test@example.com`
		rows := db.NewRows([]string{`id`, `login`, `password`})

		rows.AddRow(uint32(1), testLogin, `hashed_password`)

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnRows(rows)

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(err)
		require.NotNilf(user, "Expected get user, but got nil")
		require.Equalf(user.Id, uint32(1), "Expected get id = 1, but got %d", user.Id)
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

	t.Run(`UndefinedUser`, func(tt *testing.T) {
		testLogin := `test@example.com`
		rows := db.NewRows([]string{})

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnRows(rows)

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(user)
		require.EqualError(err, USER_NOT_FOUND_ERROR)
	})

	t.Run(`ReturnsUnkownError`, func(tt *testing.T) {
		testLogin := `test@example.com`
		errorMessage := `UnknownError`

		db.ExpectQuery(expectedSql).WithArgs(testLogin).WillReturnError(errors.New(errorMessage))

		user, err := pur.GetByLogin(context.Background(), testLogin)

		require.Nil(user)
		require.EqualError(err, errorMessage)
	})
}

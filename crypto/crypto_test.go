package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSecureId(t *testing.T) {
	require := require.New(t)
	crypto := Create()

	id, err := crypto.Generator().SecureId(16)

	require.Nil(err)
	require.Equal(len(id), 32)
}

func TestHashPassword(t *testing.T) {
	require := require.New(t)
	crypto := Create()

	hashedPassword, err := crypto.PasswordManager().Hash(`secure-password`)

	require.Nil(err)
	require.Contains(hashedPassword, `$argon2id$v=19$m=19456,t=2,p=1$`)
}

func TestComparePasswords(t *testing.T) {
	require := require.New(t)
	crypto := Create()

	originalPassword := `secure-password`
	hashedPassword, err := crypto.PasswordManager().Hash(originalPassword)

	require.Nil(err)

	match, err := crypto.
		PasswordManager().
		Compare(originalPassword, hashedPassword)

	require.Nil(err)
	require.Equal(match, true)
}

func TestGenerateRandomString(t *testing.T) {
	require := require.New(t)
	crypto := Create()

	randomString := crypto.Generator().RandomString(32)

	require.Equal(32, len(randomString))
}

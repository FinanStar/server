package crypto

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSecureId(t *testing.T) {
	require := require.New(t)

	id, err := GenerateSecureId(16)

	require.Nil(err)
	require.Equal(len(id), 32)
}

func TestHashPassword(t *testing.T) {
	require := require.New(t)

	hashedPassword, err := HashPassword(`secure-password`)

	require.Nil(err)
	require.Contains(hashedPassword, `$argon2id$v=19$m=19456,t=2,p=1$`)
}

func TestComparePasswords(t *testing.T) {
	require := require.New(t)

	originalPassword := `secure-password`
	hashedPassword, err := HashPassword(originalPassword)

	require.Nil(err)

	match, err := ComparePasswords(originalPassword, hashedPassword)

	require.Nil(err)
	require.Equal(match, true)
}

func TestGenerateRandomString(t *testing.T) {
	require := require.New(t)

	randomString := GenerateRandomString(32)

	require.Equal(32, len(randomString))
}

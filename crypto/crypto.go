package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/alexedwards/argon2id"
)

const (
	ARGON2ID_COST        = 2
	ARGON2ID_MEMORY      = 19 * 1024
	ARGON2ID_PARALLELISM = 1
	ARGON2ID_KEY_LENGTH  = 32
	ARGON2ID_SALT_LENGTH = 16
)

func GenerateSecureId(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)

	if err != nil {
		return "", fmt.Errorf(
			"Generate secure id failed with error: %v",
			err.Error(),
		)
	}

	return hex.EncodeToString(randomBytes), nil
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(
		password,
		&argon2id.Params{
			Memory:      ARGON2ID_MEMORY,
			Iterations:  ARGON2ID_COST,
			Parallelism: ARGON2ID_PARALLELISM,
			SaltLength:  ARGON2ID_SALT_LENGTH,
			KeyLength:   ARGON2ID_KEY_LENGTH,
		},
	)

	if err != nil {
		return ``, err
	}

	return hashedPassword, err
}

func ComparePasswords(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

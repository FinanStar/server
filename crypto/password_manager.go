package crypto

import "github.com/alexedwards/argon2id"

const (
	argon2IdCost        = 2
	argon2IdMemory      = 19 * 1024
	argon2IdParallelism = 1
	argon2IdKeyLength   = 32
	argon2IdSaltLength  = 16
)

type passwordManagerImpl struct{}

func (self passwordManagerImpl) Hash(password string) (string, error) {
	hashedPassword, err := argon2id.CreateHash(
		password,
		&argon2id.Params{
			Memory:      argon2IdMemory,
			Iterations:  argon2IdCost,
			Parallelism: argon2IdParallelism,
			SaltLength:  argon2IdSaltLength,
			KeyLength:   argon2IdKeyLength,
		},
	)

	if err != nil {
		return ``, err
	}

	return hashedPassword, err
}

func (self passwordManagerImpl) Compare(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

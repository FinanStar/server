package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathRand "math/rand"
	"time"
	"unsafe"

	"github.com/alexedwards/argon2id"
)

const (
	argon2IdCost        = 2
	argon2IdMemory      = 19 * 1024
	argon2IdParallelism = 1
	argon2IdKeyLength   = 32
	argon2IdSaltLength  = 16
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

func ComparePasswords(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = mathRand.NewSource(time.Now().UnixNano())

func GenerateRandomString(length int) string {
	bytes := make([]byte, length)

	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := length-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}

		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			bytes[i] = letterBytes[idx]
			i--
		}

		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&bytes))
}

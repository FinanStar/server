package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	mathRand "math/rand"
	"time"
	"unsafe"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var src = mathRand.NewSource(time.Now().UnixNano())

type generatorImpl struct{}

func (self generatorImpl) SecureId(length int) (string, error) {
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

func (self generatorImpl) RandomString(length int) string {
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

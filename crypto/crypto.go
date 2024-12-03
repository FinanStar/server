package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

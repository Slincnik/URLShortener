package utils

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

func GenerateShortKey() string {
	uuid := uuid.New()
	hash := sha256.Sum256(uuid[:])
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}

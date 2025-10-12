package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateAPIKey generates a secure random API key
func GenerateAPIKey() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to time-based key if random fails
		hash := sha256.Sum256([]byte(time.Now().String()))
		return fmt.Sprintf("rr_live_%s", hex.EncodeToString(hash[:]))
	}
	return fmt.Sprintf("rr_live_%s", hex.EncodeToString(bytes))
}

// HashAPIKey hashes an API key using HMAC-SHA256
func HashAPIKey(key, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(key))
	return hex.EncodeToString(h.Sum(nil))
}
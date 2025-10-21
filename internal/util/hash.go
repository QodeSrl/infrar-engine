package util

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashString returns a SHA-256 hash of the input string
func HashString(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// HashBytes returns a SHA-256 hash of the input bytes
func HashBytes(input []byte) string {
	hash := sha256.Sum256(input)
	return hex.EncodeToString(hash[:])
}

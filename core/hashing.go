package core

import (
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

// GetSHA3256Hash returns the SHA3-256 hash of a given string.
func GetSHA3256Hash(str []byte) ([]byte, error) {
	// Create a new sha object.
	h := sha3.New256()

	// Add our string to the hash.
	if _, err := h.Write([]byte(str)); err != nil {
		return nil, err
	}

	// Return the SHA3-256 digest.
	return h.Sum(nil), nil
}

// ByteArrayToHex converts a set of bytes to a hex encoded string.
func ByteArrayToHex(payload []byte) string {
	return hex.EncodeToString(payload)
}

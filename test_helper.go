package main

import (
	"crypto/sha256"
	"encoding/hex"
)

func computeHash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

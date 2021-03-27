package session

import (
	"crypto/rand"
	"encoding/hex"
	"io"
)

func randomString(length int) string {
	uuid := make([]byte, length/2)
	_, _ = io.ReadFull(rand.Reader, uuid[:])
	dst := make([]byte, length)
	hex.Encode(dst[:], uuid[:])
	return string(dst[:])
}

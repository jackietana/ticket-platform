package hash

import (
	"crypto/sha1"
	"fmt"
)

type SHA1Hasher struct {
	salt string
}

func NewSHA1Hasher(salt string) *SHA1Hasher {
	return &SHA1Hasher{salt}
}

func (h *SHA1Hasher) Hash(input string) (string, error) {
	hasher := sha1.New()

	if _, err := hasher.Write([]byte(h.salt + input)); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

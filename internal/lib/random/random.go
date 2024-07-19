package random

import (
	"crypto/rand"
	"math/big"
)

func NewRandomString(size int) string {
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, size)
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return ""
		}

		b[i] = chars[num.Int64()]
	}

	return string(b)
}

package filemon

import (
	"crypto/rand"
	"fmt"
	gen "math/rand"
)

const (
	source string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateString() (string, error) {

	length := 10 + gen.Intn(10)
	b := make([]byte, length)

	// Write random byte values into the bytearray
	_, err := rand.Read(b)
	if err != nil {

		return "", fmt.Errorf("GenerateString failed to do so: %w", err)
	}
	for i := range b {
		b[i] = source[int(b[i])%len(source)]
	}
	return string(b), nil
}

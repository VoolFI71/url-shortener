package shortener

import (
	"crypto/rand"
	"fmt"
)

const (
	CodeLength = 10
	Alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	byteLimit  = 256 - 256%len(Alphabet)
)

type Generator interface {
	Generate() (string, error)
}

type RandomGenerator struct{}

func (RandomGenerator) Generate() (string, error) {
	result := make([]byte, CodeLength)
	buffer := make([]byte, CodeLength*2)
	written := 0

	for written < CodeLength {
		if _, err := rand.Read(buffer); err != nil {
			return "", fmt.Errorf("read random bytes: %w", err)
		}
		for _, value := range buffer {
			// Discard the remainder to avoid modulo bias.
			if int(value) >= byteLimit {
				continue
			}
			result[written] = Alphabet[int(value)%len(Alphabet)]
			written++
			if written == CodeLength {
				break
			}
		}
	}

	return string(result), nil
}

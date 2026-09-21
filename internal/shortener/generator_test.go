package shortener_test

import (
	"strings"
	"testing"

	"github.com/VoolFI71/url-shortener/internal/shortener"
)

func TestRandomGeneratorProducesValidCodes(t *testing.T) {
	generator := shortener.RandomGenerator{}

	for i := 0; i < 1000; i++ {
		code, err := generator.Generate()
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if len(code) != shortener.CodeLength {
			t.Fatalf("code length = %d, want %d", len(code), shortener.CodeLength)
		}
		for _, character := range code {
			if !strings.ContainsRune(shortener.Alphabet, character) {
				t.Fatalf("code %q contains an invalid character %q", code, character)
			}
		}
	}
}

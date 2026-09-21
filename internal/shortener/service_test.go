package shortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/VoolFI71/url-shortener/internal/shortener"
	"github.com/VoolFI71/url-shortener/internal/storage/memory"
)

type sequenceGenerator struct {
	codes []string
	next  int
}

func (g *sequenceGenerator) Generate() (string, error) {
	code := g.codes[g.next]
	g.next++
	return code, nil
}

func TestShortenReturnsSameCodeForSameURL(t *testing.T) {
	generator := &sequenceGenerator{codes: []string{"AAAAAAAAAA", "BBBBBBBBBB"}}
	service := shortener.New(memory.New(), generator)

	first, created, err := service.Shorten(context.Background(), "https://example.com/article")
	if err != nil || !created {
		t.Fatalf("first Shorten() error = %v, created = %v", err, created)
	}
	second, created, err := service.Shorten(context.Background(), "https://example.com/article")
	if err != nil || created {
		t.Fatalf("second Shorten() error = %v, created = %v", err, created)
	}
	if second.Code != first.Code {
		t.Fatalf("second code = %q, want %q", second.Code, first.Code)
	}
}

func TestShortenRetriesAfterCodeCollision(t *testing.T) {
	generator := &sequenceGenerator{codes: []string{"AAAAAAAAAA", "AAAAAAAAAA", "BBBBBBBBBB"}}
	service := shortener.New(memory.New(), generator)

	first, _, err := service.Shorten(context.Background(), "https://first.example")
	if err != nil {
		t.Fatalf("first Shorten() error = %v", err)
	}
	second, _, err := service.Shorten(context.Background(), "https://second.example")
	if err != nil {
		t.Fatalf("second Shorten() error = %v", err)
	}
	if first.Code != "AAAAAAAAAA" || second.Code != "BBBBBBBBBB" {
		t.Fatalf("codes = (%q, %q), want (AAAAAAAAAA, BBBBBBBBBB)", first.Code, second.Code)
	}
}

func TestShortenRejectsInvalidURL(t *testing.T) {
	service := shortener.New(memory.New(), &sequenceGenerator{codes: []string{"AAAAAAAAAA"}})
	_, _, err := service.Shorten(context.Background(), "not a URL")
	if !errors.Is(err, shortener.ErrInvalidURL) {
		t.Fatalf("Shorten() error = %v, want ErrInvalidURL", err)
	}
}

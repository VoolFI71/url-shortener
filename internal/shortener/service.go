package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var (
	ErrInvalidURL              = errors.New("invalid URL")
	ErrNotFound                = errors.New("link not found")
	ErrCodeTaken               = errors.New("short code is already taken")
	ErrCodeGenerationExhausted = errors.New("could not generate a free short code")
)

const maxGenerateAttempts = 10

type Link struct {
	Code        string
	OriginalURL string
}

// Store keeps both the code and the original URL unique.
type Store interface {
	Create(ctx context.Context, link Link) (stored Link, created bool, err error)
	FindByCode(ctx context.Context, code string) (Link, error)
}

type Service struct {
	store     Store
	generator Generator
}

func New(store Store, generator Generator) *Service {
	return &Service{store: store, generator: generator}
}

func (s *Service) Shorten(ctx context.Context, rawURL string) (Link, bool, error) {
	originalURL, err := validateURL(rawURL)
	if err != nil {
		return Link{}, false, err
	}

	for attempt := 0; attempt < maxGenerateAttempts; attempt++ {
		code, err := s.generator.Generate()
		if err != nil {
			return Link{}, false, fmt.Errorf("generate short code: %w", err)
		}

		link, created, err := s.store.Create(ctx, Link{
			Code:        code,
			OriginalURL: originalURL,
		})
		if errors.Is(err, ErrCodeTaken) {
			continue
		}
		if err != nil {
			return Link{}, false, fmt.Errorf("save link: %w", err)
		}
		return link, created, nil
	}

	return Link{}, false, ErrCodeGenerationExhausted
}

func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	if !validCode(code) {
		return "", ErrNotFound
	}

	link, err := s.store.FindByCode(ctx, code)
	if err != nil {
		return "", err
	}
	return link.OriginalURL, nil
}

func validateURL(rawURL string) (string, error) {
	rawURL = strings.TrimSpace(rawURL)
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", ErrInvalidURL
	}
	return parsed.String(), nil
}

func validCode(code string) bool {
	if len(code) != CodeLength {
		return false
	}
	for _, char := range code {
		if !strings.ContainsRune(Alphabet, char) {
			return false
		}
	}
	return true
}

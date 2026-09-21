package memory

import (
	"context"
	"sync"

	"github.com/VoolFI71/url-shortener/internal/shortener"
)

type Store struct {
	mu     sync.RWMutex
	byCode map[string]shortener.Link
	byURL  map[string]shortener.Link
}

func New() *Store {
	return &Store{
		byCode: make(map[string]shortener.Link),
		byURL:  make(map[string]shortener.Link),
	}
}

func (s *Store) Create(_ context.Context, link shortener.Link) (shortener.Link, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.byURL[link.OriginalURL]; ok {
		return existing, false, nil
	}
	if _, ok := s.byCode[link.Code]; ok {
		return shortener.Link{}, false, shortener.ErrCodeTaken
	}

	s.byCode[link.Code] = link
	s.byURL[link.OriginalURL] = link
	return link, true, nil
}

func (s *Store) FindByCode(_ context.Context, code string) (shortener.Link, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	link, ok := s.byCode[code]
	if !ok {
		return shortener.Link{}, shortener.ErrNotFound
	}
	return link, nil
}

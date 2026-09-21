package memory_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"url-shortener/internal/shortener"
	"url-shortener/internal/storage/memory"
)

func TestSaveIsIdempotentForOriginalURL(t *testing.T) {
	store := memory.New()
	first := shortener.Link{Code: "AAAAAAAAAA", OriginalURL: "https://example.com"}
	second := shortener.Link{Code: "BBBBBBBBBB", OriginalURL: "https://example.com"}

	stored, created, err := store.Create(context.Background(), first)
	if err != nil || !created || stored != first {
		t.Fatalf("first Create() = (%+v, %v, %v), want (%+v, true, nil)", stored, created, err, first)
	}

	stored, created, err = store.Create(context.Background(), second)
	if err != nil || created || stored != first {
		t.Fatalf("second Create() = (%+v, %v, %v), want (%+v, false, nil)", stored, created, err, first)
	}
}

func TestSaveRejectsCodeCollision(t *testing.T) {
	store := memory.New()
	_, _, _ = store.Create(context.Background(), shortener.Link{
		Code: "AAAAAAAAAA", OriginalURL: "https://first.example",
	})

	_, _, err := store.Create(context.Background(), shortener.Link{
		Code: "AAAAAAAAAA", OriginalURL: "https://second.example",
	})
	if !errors.Is(err, shortener.ErrCodeTaken) {
		t.Fatalf("Create() error = %v, want ErrCodeTaken", err)
	}
}

func TestCreateConcurrentForSameURL(t *testing.T) {
	const requests = 100

	type result struct {
		link    shortener.Link
		created bool
		err     error
	}

	store := memory.New()
	results := make(chan result, requests)
	var wg sync.WaitGroup

	for i := 0; i < requests; i++ {
		wg.Add(1)
		go func(code string) {
			defer wg.Done()
			link, created, err := store.Create(context.Background(), shortener.Link{
				Code:        code,
				OriginalURL: "https://example.com",
			})
			results <- result{link: link, created: created, err: err}
		}(fmt.Sprintf("%010d", i))
	}

	wg.Wait()
	close(results)

	var stored shortener.Link
	createdCount := 0
	for result := range results {
		if result.err != nil {
			t.Fatalf("Create() error = %v", result.err)
		}
		if stored == (shortener.Link{}) {
			stored = result.link
		}
		if result.link != stored {
			t.Fatalf("Create() returned %+v, want %+v", result.link, stored)
		}
		if result.created {
			createdCount++
		}
	}

	if createdCount != 1 {
		t.Fatalf("created count = %d, want 1", createdCount)
	}
}

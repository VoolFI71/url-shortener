package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/VoolFI71/url-shortener/internal/shortener"
)

const uniqueViolation = "23505"

type database interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Store struct {
	db   database
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return &Store{db: pool, pool: pool}, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Create(ctx context.Context, link shortener.Link) (shortener.Link, bool, error) {
	_, err := s.db.Exec(ctx,
		`INSERT INTO links (code, original_url) VALUES ($1, $2)`,
		link.Code,
		link.OriginalURL,
	)
	if err == nil {
		return link, true, nil
	}
	if !isUniqueViolation(err) {
		return shortener.Link{}, false, fmt.Errorf("insert link: %w", err)
	}

	existing, err := s.findByURL(ctx, link.OriginalURL)
	if err == nil {
		return existing, false, nil
	}
	if errors.Is(err, shortener.ErrNotFound) {
		return shortener.Link{}, false, shortener.ErrCodeTaken
	}
	return shortener.Link{}, false, fmt.Errorf("find link after conflict: %w", err)
}

func (s *Store) FindByCode(ctx context.Context, code string) (shortener.Link, error) {
	return scanLink(s.db.QueryRow(ctx,
		`SELECT code, original_url FROM links WHERE code = $1`,
		code,
	))
}

func (s *Store) findByURL(ctx context.Context, originalURL string) (shortener.Link, error) {
	return scanLink(s.db.QueryRow(ctx,
		`SELECT code, original_url FROM links WHERE original_url = $1`,
		originalURL,
	))
}

func scanLink(row pgx.Row) (shortener.Link, error) {
	var link shortener.Link
	if err := row.Scan(&link.Code, &link.OriginalURL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shortener.Link{}, shortener.ErrNotFound
		}
		return shortener.Link{}, fmt.Errorf("scan link: %w", err)
	}
	return link, nil
}

func isUniqueViolation(err error) bool {
	var postgresError *pgconn.PgError
	return errors.As(err, &postgresError) && postgresError.Code == uniqueViolation
}

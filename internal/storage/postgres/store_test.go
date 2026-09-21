package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/VoolFI71/url-shortener/internal/shortener"
)

type fakeDatabase struct {
	execErr error
	row     pgx.Row
}

func (db *fakeDatabase) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, db.execErr
}

func (db *fakeDatabase) QueryRow(context.Context, string, ...any) pgx.Row {
	return db.row
}

type fakeRow struct {
	link shortener.Link
	err  error
}

func (r fakeRow) Scan(destinations ...any) error {
	if r.err != nil {
		return r.err
	}
	*destinations[0].(*string) = r.link.Code
	*destinations[1].(*string) = r.link.OriginalURL
	return nil
}

func TestCreateReturnsExistingLinkOnURLConflict(t *testing.T) {
	existing := shortener.Link{Code: "AAAAAAAAAA", OriginalURL: "https://example.com"}
	store := &Store{db: &fakeDatabase{
		execErr: &pgconn.PgError{Code: uniqueViolation},
		row:     fakeRow{link: existing},
	}}

	link, created, err := store.Create(context.Background(), shortener.Link{
		Code:        "BBBBBBBBBB",
		OriginalURL: existing.OriginalURL,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created {
		t.Fatal("Create() created = true, want false")
	}
	if link != existing {
		t.Fatalf("Create() link = %+v, want %+v", link, existing)
	}
}

func TestCreateReportsCodeConflict(t *testing.T) {
	store := &Store{db: &fakeDatabase{
		execErr: &pgconn.PgError{Code: uniqueViolation},
		row:     fakeRow{err: pgx.ErrNoRows},
	}}

	_, _, err := store.Create(context.Background(), shortener.Link{
		Code:        "AAAAAAAAAA",
		OriginalURL: "https://example.com",
	})
	if !errors.Is(err, shortener.ErrCodeTaken) {
		t.Fatalf("Create() error = %v, want ErrCodeTaken", err)
	}
}

func TestFindByCodeReportsMissingLink(t *testing.T) {
	store := &Store{db: &fakeDatabase{row: fakeRow{err: pgx.ErrNoRows}}}

	_, err := store.FindByCode(context.Background(), "AAAAAAAAAA")
	if !errors.Is(err, shortener.ErrNotFound) {
		t.Fatalf("FindByCode() error = %v, want ErrNotFound", err)
	}
}

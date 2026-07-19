package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/beavrest/linkshort/internal/service"
	"github.com/beavrest/linkshort/migrations"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Save(shortID, originalURL string) (string, error) {
	const q = `
		INSERT INTO short_urls (short_url, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_url`
	var resultID string
	if err := s.db.QueryRow(q, shortID, originalURL).Scan(&resultID); err != nil {
		return "", err
	}
	if resultID != shortID {
		return resultID, service.ErrURLExists
	}
	return resultID, nil
}

func (s *PostgresStore) SaveBatch(items map[string]string) error {
	if len(items) == 0 {
		return nil
	}

	placeholders := make([]string, 0, len(items))
	args := make([]any, 0, len(items)*2)
	i := 1
	for shortID, originalURL := range items {
		placeholders = append(placeholders, fmt.Sprintf("($%d,$%d)", i, i+1))
		args = append(args, shortID, originalURL)
		i += 2
	}

	q := fmt.Sprintf(
		`INSERT INTO short_urls (short_url, original_url) VALUES %s
		 ON CONFLICT (short_url) DO UPDATE SET original_url = EXCLUDED.original_url`,
		strings.Join(placeholders, ","),
	)
	_, err := s.db.Exec(q, args...)
	return err
}

func (s *PostgresStore) Get(shortID string) (string, bool) {
	const q = `SELECT original_url FROM short_urls WHERE short_url = $1`
	var originalURL string
	if err := s.db.QueryRow(q, shortID).Scan(&originalURL); err != nil {
		return "", false
	}
	return originalURL, true
}

func Migrate(db *sql.DB) error {
	sourceDriver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("migrate: init source driver: %w", err)
	}

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: init postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("migrate: init migrate instance: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: apply migrations: %w", err)
	}
	return nil
}

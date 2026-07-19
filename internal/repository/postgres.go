package repository

import (
	"database/sql"
	"errors"
	"fmt"

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

func (s *PostgresStore) Save(shortID, originalURL string) error {
	const q = `
		INSERT INTO short_urls (short_url, original_url)
		VALUES ($1, $2)
		ON CONFLICT (short_url) DO UPDATE SET original_url = EXCLUDED.original_url`
	_, err := s.db.Exec(q, shortID, originalURL)
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

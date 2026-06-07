package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Store struct {
	db *sql.DB
}

func NewStore(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping mysql: %w", err)
	}

	store := &Store{db: db}
	if err := store.Migrate(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS user_languages (
	user_id BIGINT NOT NULL PRIMARY KEY,
	language_code VARCHAR(8) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`)
	if err != nil {
		return fmt.Errorf("migrate user_languages: %w", err)
	}
	return nil
}

func (s *Store) GetUserLanguage(ctx context.Context, userID int64) (string, bool, error) {
	var lang string
	err := s.db.QueryRowContext(ctx,
		`SELECT language_code FROM user_languages WHERE user_id = ?`,
		userID,
	).Scan(&lang)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("get user language: %w", err)
	}
	return lang, true, nil
}

func (s *Store) SetUserLanguage(ctx context.Context, userID int64, lang string) error {
	_, err := s.db.ExecContext(ctx, `
INSERT INTO user_languages (user_id, language_code)
VALUES (?, ?)
ON DUPLICATE KEY UPDATE language_code = VALUES(language_code), updated_at = CURRENT_TIMESTAMP`,
		userID, lang,
	)
	if err != nil {
		return fmt.Errorf("set user language: %w", err)
	}
	return nil
}

package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS urls (
		    id INTEGER PRIMARY KEY AUTOINCREMENT,
		    alias TEXT NOT NULL UNIQUE,
		    url TEXT NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_alias on urls (alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"

	stmt, err := s.db.Prepare("INSERT INTO urls(url, alias) VALUES(?, ?)")

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: Failed to get last insert id: %w", op, err)
	}

	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM urls WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	var resURL string
	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) (bool, error) {
	const op = "storage.sqlite.DeleteURL"

	stmt, err := s.db.Prepare("DELETE FROM urls WHERE alias = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(alias)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrURLNotDeleted)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	rowsDeleted, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	var isDeleted = false
	if rowsDeleted != 0 {
		isDeleted = true
	}

	return isDeleted, nil
}

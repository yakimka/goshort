package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/mattn/go-sqlite3"
)

type URLStorage interface {
	Get(id string) (string, error)
	Set(id, url string) error
	Init() error
}

type SQLiteURLStorage struct {
	db *sql.DB
}

func (s *SQLiteURLStorage) Get(id string) (string, error) {
	var url string
	row := s.db.QueryRow("SELECT url FROM urls WHERE id = ?", id)
	switch err := row.Scan(&url); err {
	case sql.ErrNoRows:
		return "", errors.New("can't find URL")
	case nil:
		return url, nil
	default:
		return "", errors.Join(errors.New("db error"), err)
	}
}

var ErrUniqueConstraint = errors.New("can't insert URL: already exists")

func (s *SQLiteURLStorage) Set(id, url string) error {
	_, err := s.db.Exec("INSERT INTO urls (id, url) VALUES (?, ?)", id, url)
	if err != nil {
		if isSQLUniqueConstraintError(err) {
			return ErrUniqueConstraint
		}
		return fmt.Errorf("can't insert URL: %w", err)
	}
	return nil
}

func (s *SQLiteURLStorage) Init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS urls (
			id TEXT PRIMARY KEY,
			url TEXT NOT NULL
		);
	`)
	if err != nil {
		return errors.Join(errors.New("can't create table"), err)
	}
	return nil
}

func NewSQLiteURLStorage(path string) (*SQLiteURLStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, errors.Join(errors.New("can't open DB"), err)
	}
	return &SQLiteURLStorage{db: db}, nil
}

func isSQLUniqueConstraintError(original error) bool {
	var sqliteErr sqlite3.Error
	if errors.As(original, &sqliteErr) {
		return errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) ||
			errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintPrimaryKey)
	}

	return false
}

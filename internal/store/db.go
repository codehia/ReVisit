package store

import (
	"database/sql"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	goose.SetDialect("sqlite3")
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) DB() *sql.DB { return s.db }

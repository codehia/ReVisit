package store

import (
	"database/sql"
	"os"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
	_ "modernc.org/sqlite"
)

var (
	logger = zap.Must(zap.NewDevelopment())
	sugar  = logger.Sugar()
)

type Store struct {
	db *sql.DB
}

func New(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	goose.SetDialect("sqlite3") //nolint:errcheck
	if err := goose.Up(db, "migrations"); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) DB() *sql.DB { return s.db }

func Open() (*sql.DB, error) {
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		sugar.Warnw("DATABASE_NAME is not set. Setting it to goflash.db")
		databaseName = "goflash.db"
	}
	s, err := New(databaseName)
	if err != nil {
		sugar.Fatalf("failed to open store: %v", err)
		return nil, err
	}
	return s.DB(), nil
}

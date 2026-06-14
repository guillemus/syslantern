package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Conn struct {
	*sql.DB
	*Queries
}

func Connect(dbPath string) (*Conn, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open DB: %w", err)
	}
	if _, err := db.Exec(`
		PRAGMA journal_mode = WAL;
		PRAGMA busy_timeout = 5000;
		PRAGMA synchronous = NORMAL;
		PRAGMA cache_size = -262144;
		PRAGMA foreign_keys = ON;
		PRAGMA temp_store = MEMORY;
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("configure DB: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping DB: %w", err)
	}

	return &Conn{DB: db, Queries: New(db)}, nil
}

func (c *Conn) SessionDB() *sql.DB {
	return c.DB
}

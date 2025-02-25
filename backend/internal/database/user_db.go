// internal/database/user_db.go
package database

import (
	"context"
	"database/sql"
	"fmt"

	"backend/internal/config"

	_ "github.com/lib/pq"
)

type UserDB struct {
	db *sql.DB
}

func NewUserDB(cfg config.DatabaseConfig) (*UserDB, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("connecting to user database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging user database: %w", err)
	}

	return &UserDB{db: db}, nil
}

// DB returns the underlying sql.DB instance
func (u *UserDB) DB() *sql.DB {
	return u.db
}

func (u *UserDB) Close() error {
	return u.db.Close()
}

func (u *UserDB) InitSchema(ctx context.Context) error {
	queries := []string{
		`DO $$ BEGIN
            CREATE TYPE user_type AS ENUM ('regular', 'guest');
        EXCEPTION
            WHEN duplicate_object THEN null;
        END $$;`,
		`CREATE TABLE IF NOT EXISTS users (
            id UUID PRIMARY KEY,
            email TEXT UNIQUE,
            password_hash TEXT,
            name TEXT,
            user_type user_type NOT NULL,
            created_at TIMESTAMP WITH TIME ZONE NOT NULL,
            updated_at TIMESTAMP WITH TIME ZONE NOT NULL
        )`,
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)`,
	}

	for _, query := range queries {
		if _, err := u.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("executing schema query: %w", err)
		}
	}

	return nil
}

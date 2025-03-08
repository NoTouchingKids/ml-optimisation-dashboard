package store

import (
	"backend/internal/database"
	"backend/internal/database/models"
	"context"
	"errors"
)

type UserStore struct {
	db *database.UserDB
}

func NewUserStore(db *database.UserDB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (id, email, password_hash, name, user_type, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `
	_, err := s.db.DB().ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.Password,
		user.Name,
		user.Type,
		user.CreatedAt,
		user.UpdatedAt,
	)
	return err
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
        SELECT id, email, password_hash, name, user_type, created_at, updated_at
        FROM users
        WHERE email = $1
    `
	user := &models.User{}
	err := s.db.DB().QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Password,
		&user.Name,
		&user.Type,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

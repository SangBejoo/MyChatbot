package repository

import (
	"context"
	"project_masAde/internal/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *entities.User) error {
	_, err := r.db.Exec(context.Background(),
		"INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)",
		user.Username, user.PasswordHash, user.Role)
	return err
}

func (r *UserRepository) GetByUsername(username string) (*entities.User, error) {
	var user entities.User
	err := r.db.QueryRow(context.Background(),
		"SELECT id, username, password_hash, role FROM users WHERE username = $1",
		username).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Role)
	
	if err == pgx.ErrNoRows {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

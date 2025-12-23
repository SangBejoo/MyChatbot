package usecases

import (
	"errors"
	"fmt"
	"project_masAde/internal/entities"
	"project_masAde/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthUsecase(repo *repository.UserRepository, secret string) *AuthUsecase {
	return &AuthUsecase{
		userRepo:  repo,
		jwtSecret: []byte(secret),
	}
}

func (uc *AuthUsecase) Register(username, password string) error {
	existing, err := uc.userRepo.GetByUsername(username)
	if err != nil {
		return err
	}
	if existing != nil {
		return errors.New("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entities.User{
		Username:     username,
		PasswordHash: string(hashed),
		Role:         "user", // Default
	}

	return uc.userRepo.Create(user)
}

func (uc *AuthUsecase) Login(username, password string) (string, error) {
	user, err := uc.userRepo.GetByUsername(username)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(uc.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %v", err)
	}

	return tokenString, nil
}

// EnsureAdmin creates a root user if none exists (called on startup)
func (uc *AuthUsecase) EnsureAdmin(username, password string) error {
	user, err := uc.userRepo.GetByUsername(username)
	if err != nil {
		return err
	}
	if user == nil {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		admin := &entities.User{
			Username:     username,
			PasswordHash: string(hashed),
			Role:         "admin",
		}
		return uc.userRepo.Create(admin)
	}
	return nil
}

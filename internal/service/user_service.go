package service

import (
	"errors"
	"go-auth-api/internal/auth"
	"go-auth-api/internal/models"
	"go-auth-api/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Register(req models.RegisterRequest) error {
	hashedPassword, err := models.HashPassword(req.Password)
	if err != nil {
		return err
	}

	user := &models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}
	return s.repo.Create(user)
}

func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.repo.FindByUsername(username)
	if err != nil {
		return "", errors.New("user not found")
	}

	if !models.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT with RBAC role
	return auth.GenerateToken(user.ID, user.IsAdmin)
}



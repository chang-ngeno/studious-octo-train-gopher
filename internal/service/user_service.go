package service

import (
	"errors"
	"go-auth-api/internal/auth"
	"go-auth-api/internal/models"
	"go-auth-api/internal/repository"	
	"github.com/google/uuid"
	"gorm.io/gorm"
	"go-auth-api/internal/database"
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

// func (s *UserService) Login(username, password string) (string, error) {
// 	user, err := s.repo.FindByUsername(username)
// 	if err != nil {
// 		return "", errors.New("user not found")
// 	}

// 	if !models.CheckPasswordHash(password, user.Password) {
// 		return "", errors.New("invalid credentials")
// 	}

// 	// Generate JWT with RBAC role
// 	return auth.GenerateToken(user.ID, user.IsAdmin)
// }
func (s *UserService) Login(username, password string) (string, error) {
	var user models.User
	// Preload the Role and its Permissions
	err := s.repo.DB.Preload("Role.Permissions").Where("username = ?", username).First(&user).Error
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !models.CheckPasswordHash(password, user.Password) {
		return "", errors.New("invalid credentials")
	}

	// Extract permission slugs
	var permSlugs []string
	for _, p := range user.Role.Permissions {
		permSlugs = append(permSlugs, p.Slug)
	}

	return auth.GenerateToken(user.ID, user.Role.Name, permSlugs)
}

// func (s *UserService) UpdateUserRole(targetUserID uuid.UUID, newRoleID uint) error {
// 	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
// 		// 1. Update the Role
// 		if err := tx.Model(&models.User{}).Where("id = ?", targetUserID).
// 			Update("role_id", newRoleID).Error; err != nil {
// 			return err
// 		}

// 		// 2. Rotate Security Version (Revokes all current tokens)
// 		newVersion := uuid.New()
// 		return tx.Model(&models.User{}).Where("id = ?", targetUserID).
// 			Update("security_version", newVersion).Error
// 	})
// }
func (s *UserService) UpdateUserRole(targetUserID uuid.UUID, newRoleID uint) error {
	return s.repo.DB.Transaction(func(tx *gorm.DB) error {
		newVersion := uuid.New()
		
		// 1. Update DB
		if err := tx.Model(&models.User{}).Where("id = ?", targetUserID).
			Updates(map[string]interface{}{
				"role_id":          newRoleID,
				"security_version": newVersion,
			}).Error; err != nil {
			return err
		}

		// 2. Invalidate Cache
		// On the next request, the middleware won't find the key and will hit the DB
		cacheKey := "user_ver:" + targetUserID.String()
		database.RDB.Del(database.Ctx, cacheKey)

		return nil
	})
}
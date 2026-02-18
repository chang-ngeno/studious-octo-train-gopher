package handlers

import (
	"net/http"
	"go-auth-api/internal/models"
	"go-auth-api/internal/service"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svc *service.UserService
}

func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Register(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create user"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
}

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

// Dummy stats for the Admin RBAC example
func (h *UserHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "Welcome Admin", "active_users": 42})
}



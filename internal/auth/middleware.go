package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"go-auth-api/internal/database"
	"go-auth-api/internal/models"
	"net/http"
	"os"
	"strings"
	"time"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_role", claims.Role)
		c.Set("user_permissions", claims.Permissions)
		c.Next()
	}
}

func AuthorizeRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, _ := c.Get("user_role")
		if userRole != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
		c.Next()
	}
}

func HasPermission(requiredPerm string, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get UserID from JWT context
		uid, _ := c.Get("user_id")

		// 2. Check DB for user's permissions (In production, use Redis/Cache here!)
		var user models.User
		err := db.Preload("Role.Permissions").Where("id = ?", uid).First(&user).Error

		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User role not found"})
			return
		}

		// 3. Verify permission exists in the role
		hasPerm := false
		for _, p := range user.Role.Permissions {
			if p.Slug == requiredPerm {
				hasPerm = true
				break
			}
		}

		if !hasPerm {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Missing required permission: " + requiredPerm})
			return
		}

		c.Next()
	}
}

func RequirePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get permissions slice from context (set by AuthMiddleware)
		perms, exists := c.Get("user_permissions")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permissions found"})
			return
		}

		slice := perms.([]string)
		authorized := false
		for _, p := range slice {
			if p == required {
				authorized = true
				break
			}
		}

		if !authorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Missing permission: " + required})
			return
		}

		c.Next()
	}
}

// func (a *AuthMiddleware) ValidateSecurityVersion(db *gorm.DB) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		claims := c.MustGet("claims").(*Claims)

// 		// Optimization: Check Redis first!
// 		// If not in Redis, check DB:
// 		var currentVersion uuid.UUID
// 		err := db.Model(&models.User{}).
// 			Where("id = ?", claims.UserID).
// 			Select("security_version").
// 			Row().Scan(&currentVersion)

// 		if err != nil || currentVersion != claims.SecurityVersion {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
// 				"error": "Session invalidated. Please log in again.",
// 			})
// 			return
// 		}

// 		c.Next()
// 	}
// }

func ValidateSecurityVersion(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("claims").(*Claims)
		cacheKey := "user_ver:" + claims.UserID.String()

		// 1. Try Redis First
		ver, err := database.RDB.Get(database.Ctx, cacheKey).Result()

		if err == nil {
			// Cache Hit: Compare versions
			if ver != claims.SecurityVersion.String() {
				c.AbortWithStatusJSON(401, gin.H{"error": "Session revoked"})
				return
			}
			c.Next()
			return
		}

		// 2. Cache Miss: Fallback to Postgres
		var currentVersion uuid.UUID
		err = db.Model(&models.User{}).Where("id = ?", claims.UserID).
			Select("security_version").Row().Scan(&currentVersion)

		if err != nil || currentVersion != claims.SecurityVersion {
			c.AbortWithStatusJSON(401, gin.H{"error": "Invalid session"})
			return
		}

		// 3. Repopulate Cache for the next request
		database.RDB.Set(database.Ctx, cacheKey, currentVersion.String(), 24*time.Hour)

		c.Next()
	}
}

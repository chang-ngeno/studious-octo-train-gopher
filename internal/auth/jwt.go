package auth

import (
	"os"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Permissions []string  `json:"permissions"`
	SecurityVersion uuid.UUID `json:"version"`
	jwt.RegisteredClaims
}

func GenerateToken(uid uuid.UUID, role string, permissions []string) (string, error) {	
	claims := &Claims{
		UserID: uid,
		Role:   role,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

package middleware

import (
	"fmt"
	"os"


	"github.com/golang-jwt/jwt"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ValidateToken validates the JWT token and returns the user ID
func ValidateToken(tokenString string) (primitive.ObjectID, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return primitive.NilObjectID, fmt.Errorf("JWT_SECRET not set")
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid token: %v", err)
	}

	// Check if token is valid
	if !token.Valid {
		return primitive.NilObjectID, fmt.Errorf("invalid token")
	}

	// Get claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("invalid token claims")
	}

	// Get user ID from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return primitive.NilObjectID, fmt.Errorf("user_id not found in token")
	}

	// Convert string to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid user_id format")
	}

	return userID, nil
}

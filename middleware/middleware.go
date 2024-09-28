package middleware

import (
	"net/http"
	"strings"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// Define a secret key to sign JWTs (this should be stored securely)
var jwtSecretKey = []byte("your-secret-key")

// JWTAuthMiddleware validates the JWT token
func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header (it should have the format "Bearer <token>")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization token required"})
			c.Abort()
			return
		}

		// Check if the Authorization header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		// Extract the token part (the string after "Bearer ")
		tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token missing in Authorization header"})
			c.Abort()
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			// Return the secret key
			return jwtSecretKey, nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// If the token is valid, extract claims and set them in the context
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Extract user_id from claims (ensure user_id was stored in the token when generated)
			userID := claims["user_id"].(string)
			c.Set("user_id", userID)  // Attach the user_id to the request context
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Proceed to the next middleware or handler
		c.Next()
	}
}

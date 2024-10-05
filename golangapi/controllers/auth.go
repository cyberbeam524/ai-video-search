package controllers

import (
	"context"
	"fmt"
	// "encoding/json"
	"net/http"
	"time"
	"video_search_project/config"
	"video_search_project/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/api/idtoken"
)

// JWT secret key
var jwtSecret = []byte("your-secret-key")

// Function to generate a JWT token
func GenerateJWT(userID string) (string, error) {
	// Define token claims
	claims := jwt.MapClaims{
		"user_id": userID,           // Store the user ID in the token
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Set token expiration (24 hours)
	}

	// Create a new token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}


var client = &http.Client{}

func GoogleLogin(c *gin.Context) {
	token := c.PostForm("id_token")

	// Verify Google ID token
	payload, err := idtoken.Validate(context.Background(), token, config.GoogleClientID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	email := payload.Claims["email"].(string)
	var user models.User

	// Check if user exists in MongoDB
	err = config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// If user does not exist, sign them up (create a new user)
		newUser := models.User{
			Email:       email,
			GoogleID:    payload.Subject,
			Subscription: "freemium",
			Uploads:      0,
			IsPremium:    false,
		}
		_, err = config.MongoDB.Collection("users").InsertOne(context.Background(), newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}
		user = newUser
	}

	// Generate JWT token for the user
	jwtToken, err := GenerateJWT(user.ID.Hex()) // user.ID.Hex() to convert ObjectID to string
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// Return the JWT token in the response
	c.JSON(http.StatusOK, gin.H{
		"status": "logged in",
		"email":  user.Email,
		"token":  jwtToken, // Include JWT token in the response
	})
}

func EmailLogin(c *gin.Context) {
	email := c.PostForm("email")
	fmt.Println(email)
	var user models.User

	err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		// Create a new user
		newUser := models.User{Email: email, Subscription: "freemium", Uploads: 0, IsPremium: false}
		_, err = config.MongoDB.Collection("users").InsertOne(context.Background(), newUser)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create user"})
			return
		}
		// Update the user variable to hold the newly created user data
		user = newUser
	}

	// Assuming you've successfully authenticated the user
	jwtToken, err := GenerateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	// Return the token in the response
	c.JSON(http.StatusOK, gin.H{
		"status": "logged in",
		"email": user.Email,
		"token": jwtToken,  // Send the JWT token in the response
	})
}

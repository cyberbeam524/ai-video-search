package main

import (
	"log"
	"video_search_project/controllers"
	"video_search_project/config"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"video_search_project/middleware"
	// "fmt"
	"net/http"
)


// enableCORS Middleware function for handling CORS
func enableCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins (adjust for production)
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight OPTIONS requests
	if c.Request.Method == http.MethodOptions {
		c.AbortWithStatus(http.StatusOK)
		return
	}

	// Pass through to the next handler
	c.Next()
}



func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Setup MongoDB, Stripe, and other configs
	config.Setup()

	r := gin.Default()
	// Apply CORS middleware
	r.Use(enableCORS)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB (change this value to whatever you need)

	// Serve the videos directory as static files
	r.Static("/videos", "./videos") // This will make the videos folder accessible


	// Authentication routes
	r.POST("/auth/google", controllers.GoogleLogin)
	r.POST("/auth/email", controllers.EmailLogin)
	// Protected routes (JWT required)
	protected := r.Group("/")
	protected.Use(middleware.JWTAuthMiddleware())  // Apply JWT middleware
	{
		protected.POST("/video/upload2", controllers.RunFeatureExtractionHandler)  // Protect the upload endpoint
		protected.POST("/payment/checkout", controllers.CreateCheckoutSession)
		protected.GET("/payment/success", controllers.HandlePaymentSuccess)
		protected.POST("/payment/cancel", controllers.CancelSubscription)
		protected.POST("/search", controllers.CallFlaskAPI)
	}

	r.Run() // Start the server on default port 8080
}

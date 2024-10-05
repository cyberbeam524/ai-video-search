package controllers

import (
	"context"
	"net/http"
	"video_search_project/config"
	"video_search_project/models"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/sub"
	"go.mongodb.org/mongo-driver/bson"
)

// Create Stripe session for subscription
func CreateCheckoutSession(c *gin.Context) {
	email := c.PostForm("email")

	// Set the Stripe secret key
	stripe.Key = config.StripeSecretKey

	// Create a checkout session with recurring price (subscription mode)
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(string(stripe.CurrencyUSD)),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("Premium Subscription"),
					},
					UnitAmount: stripe.Int64(1000), // $10 per month
					Recurring: &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
						Interval: stripe.String("month"),  // Set recurring interval to "month"
					},
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode:           stripe.String(string(stripe.CheckoutSessionModeSubscription)),  // Subscription mode
		SuccessURL:     stripe.String(config.StripeSuccessURL),
		CancelURL:      stripe.String(config.StripeCancelURL),
		CustomerEmail:  stripe.String(email),
	}

	// Create a new Stripe checkout session
	s, err := session.New(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the session ID to the client
	c.JSON(http.StatusOK, gin.H{"sessionId": s.ID})
}

// Handle successful payment
func HandlePaymentSuccess(c *gin.Context) {
	email := c.Query("email")
	stripeSubscriptionID := c.Query("stripeSubscriptionID")

	var user models.User
	err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found"})
		return
	}

	// Update user to premium status after successful payment
	update := bson.M{"$set": bson.M{"subscription": "premium", "is_premium": true, "stripe_id": stripeSubscriptionID}}
	_, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update user subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "payment successful, user upgraded to premium"})
}

// CancelSubscription handles canceling a user's Stripe subscription
func CancelSubscription(c *gin.Context) {
	email := c.PostForm("email")
	// Fetch the user's data from MongoDB
	var user models.User
	err := config.MongoDB.Collection("users").FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Check if the user has a Stripe subscription ID
	if user.StripeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no active subscription"})
		return
	}

	// Set your Stripe API key
	stripe.Key = config.StripeSecretKey

	
	// Cancel the subscription using the Stripe API
	params := &stripe.SubscriptionCancelParams{}
	_, err = sub.Cancel(user.StripeID, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cancel subscription"})
		return
	}

	// Update the user's subscription status in the database
	update := bson.M{
		"$set": bson.M{
			"subscription": "none",
			"is_premium":   false,
			"stripe_id": "",
		},
	}
	_, err = config.MongoDB.Collection("users").UpdateOne(context.Background(), bson.M{"email": email}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user status"})
		return
	}

	// Return a success response
	c.JSON(http.StatusOK, gin.H{"status": "subscription canceled"})
}

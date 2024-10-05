package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User model
type User struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	Email         string             `bson:"email"`
	GoogleID      string             `bson:"google_id,omitempty"`
	StripeID      string             `bson:"stripe_id,omitempty"`
	Subscription  string             `bson:"subscription"`
	Uploads       int                `bson:"uploads"`
	IsPremium     bool               `bson:"is_premium"`
}

// Video model
type Video struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserID      primitive.ObjectID `bson:"user_id"`
	VideoID     string             `bson:"video_id"`
	Title       string             `bson:"title"`
	Description string             `bson:"description"`
	CreatedAt   int64              `bson:"created_at"`
}

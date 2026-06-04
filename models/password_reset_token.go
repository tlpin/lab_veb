package models

import "time"

type PasswordResetToken struct {
	ID        string     `bson:"_id" json:"id"`
	UserID    string     `bson:"user_id" json:"user_id"`
	User      User       `bson:"-" json:"-"`
	TokenHash string     `bson:"token_hash" json:"-"`
	ExpiresAt time.Time  `bson:"expires_at" json:"expires_at"`
	Used      bool       `bson:"used" json:"used"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"-"`
}

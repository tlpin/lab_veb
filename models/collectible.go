package models

import "time"

type Collectible struct {
	ID        string     `bson:"_id" json:"id"`
	UserID    string     `bson:"user_id" json:"user_id"`
	User      User       `bson:"-" json:"-"`
	Name      string     `bson:"name" json:"name"`
	Year      int        `bson:"year" json:"year"`
	Country   string     `bson:"country" json:"country"`
	Price     float64    `bson:"price" json:"price"`
	Condition string     `bson:"condition" json:"condition"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"deleted_at,omitempty"`
}

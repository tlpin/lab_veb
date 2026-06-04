package models

import "time"

type File struct {
	ID           string     `bson:"_id" json:"id"`
	UserID       string     `bson:"user_id" json:"user_id"`
	OriginalName string     `bson:"original_name" json:"original_name"`
	ObjectKey    string     `bson:"object_key" json:"-"`
	Bucket       string     `bson:"bucket" json:"-"`
	Size         int64      `bson:"size" json:"size"`
	MimeType     string     `bson:"mime_type" json:"mime_type"`
	CreatedAt    time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `bson:"deleted_at,omitempty" json:"-"`
}

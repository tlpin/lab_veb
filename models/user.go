package models

import "time"

type User struct {
	ID           string  `bson:"_id" json:"id"`
	Email        string  `bson:"email" json:"email"`
	PasswordHash string  `bson:"password_hash" json:"-"`
	DisplayName  string  `bson:"display_name,omitempty" json:"display_name,omitempty"`
	Bio          string  `bson:"bio,omitempty" json:"bio,omitempty"`
	AvatarFileID *string `bson:"avatar_file_id,omitempty" json:"avatar_file_id,omitempty"`

	YandexID *string `bson:"yandex_id,omitempty" json:"-"`
	VKID     *string `bson:"vk_id,omitempty" json:"-"`

	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
	DeletedAt *time.Time `bson:"deleted_at,omitempty" json:"-"`
}

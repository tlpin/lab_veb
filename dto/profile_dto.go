package dto

// ProfileResponseDTO представляет профиль пользователя для API ответов
type ProfileResponseDTO struct {
	ID           string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email        string  `json:"email" example:"user@example.com"`
	DisplayName  string  `json:"display_name,omitempty" example:"John Doe"`
	Bio          string  `json:"bio,omitempty" example:"Collector from NYC"`
	AvatarFileID *string `json:"avatar_file_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	CreatedAt    string  `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// UpdateProfileDTO представляет данные для обновления профиля
type UpdateProfileDTO struct {
	DisplayName  *string `json:"display_name,omitempty" example:"John Updated"`
	Bio          *string `json:"bio,omitempty" example:"Updated bio"`
	AvatarFileID *string `json:"avatar_file_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
}

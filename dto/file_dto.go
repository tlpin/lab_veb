package dto

type FileResponseDTO struct {
	ID           string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       string `json:"user_id" example:"661e8400-e29b-41d4-a716-446655440001"`
	OriginalName string `json:"original_name" example:"photo.jpg"`
	Size         int64  `json:"size" example:"204800"`
	MimeType     string `json:"mime_type" example:"image/jpeg"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

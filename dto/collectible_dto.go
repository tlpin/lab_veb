package dto

import "time"

// Используется при POST
type CreateCollectibleDTO struct {
	Name      string  `json:"name" binding:"required" example:"Монета 1 рубль 1961 года"`
	Year      int     `json:"year" binding:"required" example:"1961"`
	Country   string  `json:"country" binding:"required" example:"СССР"`
	Price     float64 `json:"price" binding:"required" example:"5000"`
	Condition string  `json:"condition" example:"excellent"`
}

// Используется при PUT
type UpdateCollectibleDTO struct {
	Name      string  `json:"name" example:"Монета 1 рубль 1961 года"`
	Year      int     `json:"year" example:"1961"`
	Country   string  `json:"country" example:"СССР"`
	Price     float64 `json:"price" example:"5000"`
	Condition string  `json:"condition" example:"excellent"`
}

// Используется при PATCH
type PatchCollectibleDTO struct {
	Name      *string  `json:"name,omitempty" example:"Монета 1 рубль 1961 года"`
	Year      *int     `json:"year,omitempty" example:"1961"`
	Country   *string  `json:"country,omitempty" example:"СССР"`
	Price     *float64 `json:"price,omitempty" example:"5000"`
	Condition *string  `json:"condition,omitempty" example:"excellent"`
}

type CollectibleResponseDTO struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string    `json:"user_id" example:"661e8400-e29b-41d4-a716-446655440001"`
	Name      string    `json:"name" example:"Монета 1 рубль 1961 года"`
	Year      int       `json:"year" example:"1961"`
	Country   string    `json:"country" example:"СССР"`
	Price     float64   `json:"price" example:"5000"`
	Condition string    `json:"condition" example:"excellent"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2024-01-02T00:00:00Z"`
}

type PaginationMetaDTO struct {
	Total      int64 `json:"total" example:"100"`
	Page       int   `json:"page" example:"1"`
	Limit      int   `json:"limit" example:"10"`
	TotalPages int   `json:"totalPages" example:"10"`
}

type CollectibleListResponseDTO struct {
	Data []CollectibleResponseDTO `json:"data"`
	Meta PaginationMetaDTO        `json:"meta"`
}

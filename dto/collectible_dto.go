package dto

// Используется при создании (POST /items)
type CreateCollectibleDTO struct {
	Name      string  `json:"name" binding:"required"`
	Year      int     `json:"year" binding:"required"`
	Country   string  `json:"country" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
	Condition string  `json:"condition"`
}

// Используется при полном обновлении (PUT /items/:id)
type UpdateCollectibleDTO struct {
	Name      string  `json:"name"`
	Year      int     `json:"year"`
	Country   string  `json:"country"`
	Price     float64 `json:"price"`
	Condition string  `json:"condition"`
}

// Используется при частичном обновлении (PATCH /items/:id)
type PatchCollectibleDTO struct {
	Name      *string  `json:"name,omitempty"`
	Year      *int     `json:"year,omitempty"`
	Country   *string  `json:"country,omitempty"`
	Price     *float64 `json:"price,omitempty"`
	Condition *string  `json:"condition,omitempty"`
}

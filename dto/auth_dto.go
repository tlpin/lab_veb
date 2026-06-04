package dto

type RegisterDTO struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

type LoginDTO struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type ForgotPasswordDTO struct {
	Email string `json:"email" binding:"required,email" example:"user@example.com"`
}

type ResetPasswordDTO struct {
	Token       string `json:"token" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	NewPassword string `json:"new_password" binding:"required,min=6" example:"newpassword123"`
}

type UserResponseDTO struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string `json:"email" example:"user@example.com"`
	CreatedAt string `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

type MessageDTO struct {
	Message string `json:"message" example:"Operation successful"`
}

type ErrorDTO struct {
	Error string `json:"error" example:"Something went wrong"`
}

type RegisterResponseDTO struct {
	Message string          `json:"message" example:"User registered successfully"`
	User    UserResponseDTO `json:"user"`
}

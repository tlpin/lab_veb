package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Collectible struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Year      int            `gorm:"not null" json:"year"`
	Country   string         `gorm:"not null" json:"country"`
	Price     float64        `gorm:"not null" json:"price"`
	Condition string         `gorm:"default:'excellent'" json:"condition"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

func (c *Collectible) BeforeCreate(tx *gorm.DB) error {
	c.ID = uuid.New().String()
	return nil
}

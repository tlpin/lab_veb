package repositories

import (
	"newyear-api/models"

	"gorm.io/gorm"
)

type CollectibleRepository struct {
	db *gorm.DB
}

func NewCollectibleRepository(db *gorm.DB) *CollectibleRepository {
	return &CollectibleRepository{db: db}
}

func (r *CollectibleRepository) Create(collectible *models.Collectible) error {
	return r.db.Create(collectible).Error
}

func (r *CollectibleRepository) GetAll(limit, offset int) ([]models.Collectible, int64, error) {
	var collectibles []models.Collectible
	var total int64

	if err := r.db.Model(&models.Collectible{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Limit(limit).Offset(offset).Find(&collectibles).Error; err != nil {
		return nil, 0, err
	}

	return collectibles, total, nil
}

func (r *CollectibleRepository) GetByID(id string) (*models.Collectible, error) {
	var collectible models.Collectible
	result := r.db.Where("id = ?", id).First(&collectible)
	return &collectible, result.Error
}

func (r *CollectibleRepository) Delete(id string) error {
	return r.db.Where("id = ?", id).Delete(&models.Collectible{}).Error
}

func (r *CollectibleRepository) Update(collectible *models.Collectible) error {
	return r.db.Save(collectible).Error
}

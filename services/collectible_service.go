package services

import (
	"newyear-api/dto"
	"newyear-api/models"
	"newyear-api/repositories"
)

type CollectibleService struct {
	repo *repositories.CollectibleRepository
}

func NewCollectibleService(repo *repositories.CollectibleRepository) *CollectibleService {
	return &CollectibleService{repo: repo}
}

func (s *CollectibleService) Create(input dto.CreateCollectibleDTO) error {
	collectible := models.Collectible{
		Name:      input.Name,
		Year:      input.Year,
		Country:   input.Country,
		Price:     input.Price,
		Condition: input.Condition,
	}
	return s.repo.Create(&collectible)
}

func (s *CollectibleService) GetAll(page, limit int) ([]models.Collectible, int64, error) {
	offset := (page - 1) * limit
	return s.repo.GetAll(limit, offset)
}

func (s *CollectibleService) GetByID(id string) (*models.Collectible, error) {
	return s.repo.GetByID(id)
}

func (s *CollectibleService) Update(id string, input dto.UpdateCollectibleDTO) (*models.Collectible, error) {
	collectible, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	collectible.Name = input.Name
	collectible.Year = input.Year
	collectible.Country = input.Country
	collectible.Price = input.Price
	collectible.Condition = input.Condition

	return collectible, s.repo.Update(collectible)
}

func (s *CollectibleService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ====================== PATCH ======================
func (s *CollectibleService) Patch(id string, input dto.PatchCollectibleDTO) (*models.Collectible, error) {
	collectible, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Обновляем ТОЛЬКО те поля, которые переданы (не nil)
	if input.Name != nil {
		collectible.Name = *input.Name
	}
	if input.Year != nil {
		collectible.Year = *input.Year
	}
	if input.Country != nil {
		collectible.Country = *input.Country
	}
	if input.Price != nil {
		collectible.Price = *input.Price
	}
	if input.Condition != nil {
		collectible.Condition = *input.Condition
	}

	return collectible, s.repo.Update(collectible)
}

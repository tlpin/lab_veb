package services

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"newyear-api/cache"
	"newyear-api/dto"
	"newyear-api/models"
	"newyear-api/repositories"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrCollectibleNotFound  = errors.New("collectible not found")
	ErrCollectibleForbidden = errors.New("forbidden")
)

const (
	cacheKeyItemsList    = "wp:items:user:%s:page:%d:limit:%d"
	cacheKeyItemsSingle  = "wp:items:single:%s"
	cacheKeyItemsPattern = "wp:items:user:%s:*"
)

type CollectibleService struct {
	repo  *repositories.CollectibleRepository
	cache *cache.RedisCache
}

func NewCollectibleService(repo *repositories.CollectibleRepository, redisCache *cache.RedisCache) *CollectibleService {
	return &CollectibleService{
		repo:  repo,
		cache: redisCache,
	}
}

func getDefaultCacheTTL() time.Duration {
	raw := os.Getenv("CACHE_TTL_DEFAULT")
	if raw == "" {
		return 5 * time.Minute
	}

	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 5 * time.Minute
	}

	return time.Duration(seconds) * time.Second
}

func (s *CollectibleService) Create(userID string, input dto.CreateCollectibleDTO) error {
	collectible := models.Collectible{
		UserID:    userID,
		Name:      input.Name,
		Year:      input.Year,
		Country:   input.Country,
		Price:     input.Price,
		Condition: input.Condition,
	}

	if err := s.repo.Create(&collectible); err != nil {
		return err
	}

	_ = s.cache.DeleteByPattern(fmt.Sprintf(cacheKeyItemsPattern, userID))

	return nil
}

func (s *CollectibleService) GetAll(userID string, page, limit int) ([]models.Collectible, int64, error) {
	cacheKey := fmt.Sprintf(cacheKeyItemsList, userID, page, limit)

	var cached struct {
		Collectibles []models.Collectible `json:"collectibles"`
		Total        int64                `json:"total"`
	}

	if err := s.cache.Get(cacheKey, &cached); err == nil {
		log.Printf("Cache HIT: %s", cacheKey)
		return cached.Collectibles, cached.Total, nil
	}

	log.Printf("Cache MISS: %s", cacheKey)

	offset := (page - 1) * limit
	collectibles, total, err := s.repo.GetAllByUser(userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	cached = struct {
		Collectibles []models.Collectible `json:"collectibles"`
		Total        int64                `json:"total"`
	}{
		Collectibles: collectibles,
		Total:        total,
	}

	_ = s.cache.Set(cacheKey, cached, getDefaultCacheTTL())

	return collectibles, total, nil
}

func (s *CollectibleService) GetByID(id, userID string) (*models.Collectible, error) {
	cacheKey := fmt.Sprintf(cacheKeyItemsSingle, id)

	var cached models.Collectible

	if err := s.cache.Get(cacheKey, &cached); err == nil {
		log.Printf("Cache HIT: %s", cacheKey)

		if cached.UserID != userID {
			return nil, ErrCollectibleForbidden
		}
		return &cached, nil
	}

	log.Printf("Cache MISS: %s", cacheKey)

	collectible, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCollectibleNotFound
		}
		return nil, err
	}

	if collectible.UserID != userID {
		return nil, ErrCollectibleForbidden
	}

	_ = s.cache.Set(cacheKey, collectible, getDefaultCacheTTL())

	return collectible, nil
}

func (s *CollectibleService) Update(id, userID string, input dto.UpdateCollectibleDTO) (*models.Collectible, error) {
	collectible, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCollectibleNotFound
		}
		return nil, err
	}

	if collectible.UserID != userID {
		return nil, ErrCollectibleForbidden
	}

	collectible.Name = input.Name
	collectible.Year = input.Year
	collectible.Country = input.Country
	collectible.Price = input.Price
	collectible.Condition = input.Condition

	if err := s.repo.Update(collectible); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(fmt.Sprintf(cacheKeyItemsSingle, id))
	_ = s.cache.DeleteByPattern(fmt.Sprintf(cacheKeyItemsPattern, userID))

	return collectible, nil
}

func (s *CollectibleService) Delete(id, userID string) error {
	collectible, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrCollectibleNotFound
		}
		return err
	}

	if collectible.UserID != userID {
		return ErrCollectibleForbidden
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	_ = s.cache.Delete(fmt.Sprintf(cacheKeyItemsSingle, id))
	_ = s.cache.DeleteByPattern(fmt.Sprintf(cacheKeyItemsPattern, userID))

	return nil
}

func (s *CollectibleService) Patch(id, userID string, input dto.PatchCollectibleDTO) (*models.Collectible, error) {
	collectible, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrCollectibleNotFound
		}
		return nil, err
	}

	if collectible.UserID != userID {
		return nil, ErrCollectibleForbidden
	}

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

	if err := s.repo.Update(collectible); err != nil {
		return nil, err
	}

	_ = s.cache.Delete(fmt.Sprintf(cacheKeyItemsSingle, id))
	_ = s.cache.DeleteByPattern(fmt.Sprintf(cacheKeyItemsPattern, userID))

	return collectible, nil
}

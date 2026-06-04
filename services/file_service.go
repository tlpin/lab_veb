package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"newyear-api/cache"
	"newyear-api/models"
	"newyear-api/repositories"
	"newyear-api/storage"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var (
	ErrFileNotFound  = errors.New("file not found")
	ErrFileForbidden = errors.New("forbidden")
	ErrFileTooBig    = errors.New("file too large")
	ErrFileType      = errors.New("invalid file type")
)

var allowedMimeTypes = map[string]bool{
	"image/png":  true,
	"image/jpeg": true,
	"image/jpg":  true,
}

const (
	cacheKeyFileMeta = "wp:files:%s:meta"
)

type FileService struct {
	repo    *repositories.FileRepository
	storage *storage.MinIOStorage
	cache   *cache.RedisCache
}

func NewFileService(repo *repositories.FileRepository, storage *storage.MinIOStorage, cache *cache.RedisCache) *FileService {
	return &FileService{
		repo:    repo,
		storage: storage,
		cache:   cache,
	}
}

func (s *FileService) getMaxFileSize() int64 {
	raw := os.Getenv("MAX_FILE_SIZE")
	if raw == "" {
		return 10 * 1024 * 1024
	}
	val, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || val <= 0 {
		return 10 * 1024 * 1024
	}
	return val
}

func (s *FileService) Upload(userID, originalName, mimeType string, size int64, reader io.Reader) (*models.File, error) {
	if size > s.getMaxFileSize() {
		return nil, ErrFileTooBig
	}

	normalizedMime := strings.ToLower(strings.TrimSpace(mimeType))
	if !allowedMimeTypes[normalizedMime] {
		return nil, ErrFileType
	}

	objectKey := fmt.Sprintf("users/%s/%s/%s", userID, uuid.New().String(), originalName)

	// Увеличиваем таймаут для загрузки
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := s.storage.Upload(ctx, objectKey, reader, size, normalizedMime); err != nil {
		return nil, fmt.Errorf("failed to upload to minio: %w", err)
	}

	file := &models.File{
		UserID:       userID,
		OriginalName: originalName,
		ObjectKey:    objectKey,
		Bucket:       s.storage.GetBucket(),
		Size:         size,
		MimeType:     normalizedMime,
	}

	if err := s.repo.Create(file); err != nil {
		return nil, err
	}

	return file, nil
}

func (s *FileService) GetFileByID(fileID, userID string) (*models.File, error) {
	cacheKey := fmt.Sprintf(cacheKeyFileMeta, fileID)

	var cached models.File
	if err := s.cache.Get(cacheKey, &cached); err == nil {
		if cached.UserID != userID {
			return nil, ErrFileForbidden
		}
		return &cached, nil
	}

	file, err := s.repo.FindByID(fileID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	if file.UserID != userID {
		return nil, ErrFileForbidden
	}

	_ = s.cache.Set(cacheKey, file, 5*time.Minute)

	return file, nil
}

// ИСПРАВЛЕННЫЙ МЕТОД DOWNLOAD - без контекста с таймаутом
func (s *FileService) Download(fileID, userID string) (io.ReadCloser, *models.File, error) {
	log.Printf("[DEBUG] Download: fileID=%s, userID=%s", fileID, userID)

	file, err := s.GetFileByID(fileID, userID)
	if err != nil {
		log.Printf("[ERROR] GetFileByID failed: %v", err)
		return nil, nil, err
	}

	log.Printf("[DEBUG] File found: objectKey=%s", file.ObjectKey)

	// НЕ ИСПОЛЬЗУЙ context.WithTimeout здесь!
	// Используй обычный background context
	reader, err := s.storage.Download(context.Background(), file.ObjectKey)
	if err != nil {
		log.Printf("[ERROR] MinIO download failed: %v", err)
		return nil, nil, fmt.Errorf("failed to download from minio: %w", err)
	}

	log.Printf("[DEBUG] Download successful")
	return reader, file, nil
}

func (s *FileService) DeleteFile(fileID, userID string) error {
	file, err := s.GetFileByID(fileID, userID)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = s.storage.Delete(ctx, file.ObjectKey)

	if err := s.repo.SoftDelete(fileID); err != nil {
		return err
	}

	_ = s.cache.Delete(fmt.Sprintf(cacheKeyFileMeta, fileID))

	return nil
}

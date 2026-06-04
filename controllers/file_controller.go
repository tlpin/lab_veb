package controllers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"newyear-api/services"

	"github.com/gin-gonic/gin"
)

type FileController struct {
	service *services.FileService
}

func NewFileController(service *services.FileService) *FileController {
	return &FileController{service: service}
}

// Upload godoc
// @Summary      Загрузить файл
// @Description  Загружает файл в MinIO. Поддерживает PNG и JPEG до 10MB
// @Tags         Files
// @Accept       multipart/form-data
// @Produce      json
// @Security     CookieAuth
// @Param        file formData file true "Файл для загрузки"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /files [post]
func (fc *FileController) Upload(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	contentType := fileHeader.Header.Get("Content-Type")

	file, err := fc.service.Upload(userID.(string), fileHeader.Filename, contentType, fileHeader.Size, src)
	if err != nil {
		if errors.Is(err, services.ErrFileTooBig) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large (max 10MB)"})
			return
		}
		if errors.Is(err, services.ErrFileType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: png, jpeg"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":            file.ID,
		"original_name": file.OriginalName,
		"size":          file.Size,
		"mime_type":     file.MimeType,
		"created_at":    file.CreatedAt,
	})
}

// Download godoc
// @Summary      Скачать файл
// @Description  Скачивает файл из MinIO по ID
// @Tags         Files
// @Produce      application/octet-stream
// @Security     CookieAuth
// @Param        fileId path string true "ID файла"
// @Success      200  {string}  string  "Файл"
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /files/{fileId} [get]
func (fc *FileController) Download(c *gin.Context) {
	log.Println("[DEBUG] Download endpoint called")

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID := c.Param("fileId")
	log.Printf("[DEBUG] Downloading fileID=%s for userID=%v", fileID, userID)

	reader, file, err := fc.service.Download(fileID, userID.(string))
	if err != nil {
		log.Printf("[ERROR] Download failed: %v", err)
		if errors.Is(err, services.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if errors.Is(err, services.ErrFileForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}
	defer reader.Close()

	log.Printf("[DEBUG] Sending file: name=%s, size=%d", file.OriginalName, file.Size)

	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))

	// Отправляем файл
	c.DataFromReader(http.StatusOK, file.Size, file.MimeType, reader, nil)
}

// Delete godoc
// @Summary      Удалить файл
// @Description  Удаляет файл из MinIO и БД
// @Tags         Files
// @Produce      json
// @Security     CookieAuth
// @Param        fileId path string true "ID файла"
// @Success      204  {string}  string  "No Content"
// @Failure      401  {object}  map[string]interface{}
// @Failure      403  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Router       /files/{fileId} [delete]
func (fc *FileController) Delete(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	fileID := c.Param("fileId")

	if err := fc.service.DeleteFile(fileID, userID.(string)); err != nil {
		if errors.Is(err, services.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		if errors.Is(err, services.ErrFileForbidden) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	c.Status(http.StatusNoContent)
}

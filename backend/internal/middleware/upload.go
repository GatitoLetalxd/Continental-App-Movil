package middleware

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"continental-backend/internal/config"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

func EnsureUploadDirs() error {
	baseDir := "uploads"
	if config.AppConfig != nil && config.AppConfig.UploadDir != "" {
		baseDir = config.AppConfig.UploadDir
	}
	evidenciaDir := filepath.Join(baseDir, "evidencia")
	chatDir := filepath.Join(baseDir, "chat")

	if err := os.MkdirAll(evidenciaDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(chatDir, 0755); err != nil {
		return err
	}
	return nil
}

type UploadedFileResult struct {
	RelativePath string
	FullPath     string
	MimeType     string
	SizeBytes    int64
	Filename     string
}

func SaveUpload(c *gin.Context, fieldName string, subDir string, maxSizeBytes int64, allowVideo bool) (*UploadedFileResult, error) {
	file, err := c.FormFile(fieldName)
	if err != nil {
		return nil, utils.NewValidationError("Debe seleccionar un archivo válido")
	}

	if file.Size > maxSizeBytes {
		maxMB := maxSizeBytes / (1024 * 1024)
		return nil, utils.NewValidationError(fmt.Sprintf("El archivo excede el tamaño máximo permitido de %dMB", maxMB))
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	isImage := ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
	isVideo := ext == ".mp4" || ext == ".mov" || ext == ".quicktime"

	if !isImage && (!allowVideo || !isVideo) {
		if allowVideo {
			return nil, utils.NewValidationError("Solo se permiten imágenes (jpg, png, webp) y videos (mp4, mov)")
		}
		return nil, utils.NewValidationError("Solo se permiten imágenes (jpg, png, webp)")
	}

	if err := EnsureUploadDirs(); err != nil {
		return nil, utils.NewAppError("Error al preparar directorios de almacenamiento", 500)
	}

	uniqueSuffix := fmt.Sprintf("%d-%d", time.Now().UnixMilli(), rand.Int63n(1e9))
	filename := fmt.Sprintf("%s-%s%s", subDir, uniqueSuffix, ext)

	baseDir := "uploads"
	if config.AppConfig != nil && config.AppConfig.UploadDir != "" {
		baseDir = config.AppConfig.UploadDir
	}
	fullPath := filepath.Join(baseDir, subDir, filename)

	if err := c.SaveUploadedFile(file, fullPath); err != nil {
		return nil, utils.NewAppError(fmt.Sprintf("Error al guardar archivo en disco: %v", err), 500)
	}

	mimeType := file.Header.Get("Content-Type")
	if mimeType == "" {
		if isImage {
			mimeType = "image/" + strings.TrimPrefix(ext, ".")
		} else if isVideo {
			mimeType = "video/" + strings.TrimPrefix(ext, ".")
		}
	}

	// El path relativo guardado en la BD es uploads/subDir/filename
	relPath := fmt.Sprintf("uploads/%s/%s", subDir, filename)

	return &UploadedFileResult{
		RelativePath: relPath,
		FullPath:     fullPath,
		MimeType:     mimeType,
		SizeBytes:    file.Size,
		Filename:     filename,
	}, nil
}

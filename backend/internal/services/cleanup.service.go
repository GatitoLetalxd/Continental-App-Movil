package services

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"continental-backend/internal/config"
	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"

	"gorm.io/gorm"
)

type CleanupService struct {
	db *gorm.DB
}

func NewCleanupService() *CleanupService {
	return &CleanupService{db: database.DB}
}

func (s *CleanupService) LimpiarFotosEvidencia() (int, error) {
	fechaLimite := time.Now().AddDate(0, 0, -constants.DiasExpiracionFotosEvidencia)
	log.Printf("[Cleanup] Buscando fotos de evidencia anteriores a %s", fechaLimite.Format(time.RFC3339))

	var citasInicio []models.Cita
	s.db.Where("foto_inicio IS NOT NULL AND foto_inicio_expirada = false AND created_at < ?", fechaLimite).
		Find(&citasInicio)

	totalEliminadas := 0
	baseDir := config.AppConfig.UploadDir

	for _, c := range citasInicio {
		if c.FotoInicio != nil && *c.FotoInicio != "" {
			// Los paths en BD son relativos tipo "uploads/evidencia/foto.jpg"
			// baseDir ya apunta a la carpeta uploads, así que subimos un nivel
			fullPath := filepath.Join(baseDir, "..", *c.FotoInicio)
			fullPath = filepath.Clean(fullPath)
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				log.Printf("[Cleanup] Eliminada foto inicio: %s", fullPath)
			}
			// Actualizar usando el ID para evitar el bug de "value with no primary key"
			s.db.Model(&models.Cita{}).Where("id = ?", c.ID).Update("foto_inicio_expirada", true)
			totalEliminadas++
		}
	}

	var citasFin []models.Cita
	s.db.Where("foto_fin IS NOT NULL AND foto_fin_expirada = false AND created_at < ?", fechaLimite).
		Find(&citasFin)

	for _, c := range citasFin {
		if c.FotoFin != nil && *c.FotoFin != "" {
			fullPath := filepath.Join(baseDir, "..", *c.FotoFin)
			fullPath = filepath.Clean(fullPath)
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				log.Printf("[Cleanup] Eliminada foto fin: %s", fullPath)
			}
			s.db.Model(&models.Cita{}).Where("id = ?", c.ID).Update("foto_fin_expirada", true)
			totalEliminadas++
		}
	}

	log.Printf("[Cleanup] Fotos de evidencia procesadas: %d", totalEliminadas)
	return totalEliminadas, nil
}

func (s *CleanupService) LimpiarMultimediaChat() (int, error) {
	ahora := time.Now()
	log.Println("[Cleanup] Buscando multimedia de chat expirada")

	var mensajes []models.Mensaje
	s.db.Where("tipo IN ('IMAGEN', 'VIDEO') AND archivo_url IS NOT NULL AND expirado = false AND expira_at < ?", ahora).
		Find(&mensajes)

	baseDir := config.AppConfig.UploadDir
	totalEliminados := 0

	for _, m := range mensajes {
		if m.ArchivoURL != nil && *m.ArchivoURL != "" {
			fullPath := filepath.Join(baseDir, "..", *m.ArchivoURL)
			fullPath = filepath.Clean(fullPath)
			if _, err := os.Stat(fullPath); err == nil {
				_ = os.Remove(fullPath)
				log.Printf("[Cleanup] Eliminado archivo chat: %s", fullPath)
			}
			// Actualizar usando ID para evitar el bug de "value with no primary key"
			s.db.Model(&models.Mensaje{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
				"expirado":    true,
				"archivo_url": nil,
			})
			totalEliminados++
		}
	}

	log.Printf("[Cleanup] Multimedia de chat procesada: %d", totalEliminados)
	return totalEliminados, nil
}

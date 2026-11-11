package jobs

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"continental-backend/internal/database"
	"continental-backend/internal/services"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

type JobScheduler struct {
	cron           *cron.Cron
	cleanupService *services.CleanupService
	db             *gorm.DB
}

func NewJobScheduler() *JobScheduler {
	return &JobScheduler{
		cron:           cron.New(),
		cleanupService: services.NewCleanupService(),
		db:             database.DB,
	}
}

func (s *JobScheduler) Start() {
	// 1. Backup diario a las 2:00 AM
	_, err := s.cron.AddFunc("0 2 * * *", func() {
		log.Println("[Cron] Ejecutando backup de base de datos SQLite...")
		if err := s.EjecutarBackupSQLite(); err != nil {
			log.Printf("[Cron] Error en backup SQLite: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Cron] Error registrando job de backup: %v", err)
	}

	// 2. Limpieza de fotos de evidencia a las 3:00 AM
	_, err = s.cron.AddFunc("0 3 * * *", func() {
		log.Println("[Cron] Ejecutando limpieza de fotos de evidencia...")
		if _, err := s.cleanupService.LimpiarFotosEvidencia(); err != nil {
			log.Printf("[Cron] Error en limpieza de fotos: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Cron] Error registrando job de limpieza de fotos: %v", err)
	}

	// 3. Limpieza de chat multimedia a las 3:30 AM
	_, err = s.cron.AddFunc("30 3 * * *", func() {
		log.Println("[Cron] Ejecutando limpieza de multimedia de chat...")
		if _, err := s.cleanupService.LimpiarMultimediaChat(); err != nil {
			log.Printf("[Cron] Error en limpieza de chat: %v", err)
		}
	})
	if err != nil {
		log.Printf("[Cron] Error registrando job de limpieza de chat: %v", err)
	}

	s.cron.Start()
	log.Println("⏰ [Cron] Tareas programadas iniciadas correctamente (Backup 2:00 AM, Limpieza fotos 3:00 AM, Limpieza chat 3:30 AM)")
}

func (s *JobScheduler) Stop() {
	if s.cron != nil {
		s.cron.Stop()
	}
}

func (s *JobScheduler) EjecutarBackupSQLite() error {
	backupDir, err := filepath.Abs("backups")
	if err != nil {
		backupDir = "backups"
	}
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("error creando directorio de backups: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02-15-04-05")
	filename := fmt.Sprintf("backup-%s.db", timestamp)
	filepathDst := filepath.Join(backupDir, filename)

	// Usar comando VACUUM INTO nativo de SQLite para un backup atómico y limpio
	query := fmt.Sprintf("VACUUM INTO '%s'", filepathDst)
	if err := s.db.Exec(query).Error; err != nil {
		return fmt.Errorf("error ejecutando VACUUM INTO: %w", err)
	}

	log.Printf("[Backup] Backup SQLite completado: %s", filepathDst)

	// Rotación: mantener solo los últimos 7 backups
	entries, err := os.ReadDir(backupDir)
	if err == nil {
		var backupFiles []string
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "backup-") && strings.HasSuffix(e.Name(), ".db") {
				backupFiles = append(backupFiles, e.Name())
			}
		}

		sort.Sort(sort.Reverse(sort.StringSlice(backupFiles)))

		if len(backupFiles) > 7 {
			for _, old := range backupFiles[7:] {
				oldPath := filepath.Join(backupDir, old)
				if err := os.Remove(oldPath); err == nil {
					log.Printf("[Backup] Eliminado backup antiguo: %s", old)
				}
			}
		}
	}

	return nil
}

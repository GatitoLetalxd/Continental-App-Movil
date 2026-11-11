package database

import (
	"log"
	"time"

	"continental-backend/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(sqlite.Open(dbPath), gormConfig)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// SQLite funciona mejor con 1 conexión para escrituras secuenciales sin bloqueos
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Activar modo WAL y configuraciones de rendimiento / integridad
	pragmas := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, p := range pragmas {
		if err := db.Exec(p).Error; err != nil {
			log.Printf("[DB] Advertencia al ejecutar pragma '%s': %v", p, err)
		}
	}

	// AutoMigrate de todos los modelos
	err = db.AutoMigrate(
		&models.Admin{},
		&models.Cliente{},
		&models.Vehiculo{},
		&models.Tecnico{},
		&models.Servicio{},
		&models.Accesorio{},
		&models.Cita{},
		&models.CitaAccesorio{},
		&models.CitaFaseHistorial{},
		&models.Mensaje{},
		&models.Evidencia{},
	)
	if err != nil {
		return nil, err
	}

	DB = db
	log.Println("✅ [DB] Base de datos SQLite inicializada y modelos migrados correctamente")
	return db, nil
}

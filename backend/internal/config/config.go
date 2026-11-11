package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DBPath                string
	PGDatabaseURL         string
	JWTSecret             string
	JWTExpiresIn          string
	Port                  string
	GinMode               string
	NodeEnv               string
	UploadDir             string
	Domain                string
	MinAnticipacionRemota int
}

var AppConfig *Config

func LoadConfig() *Config {
	// Intentar cargar .env si existe, ignorar error si no existe
	if err := godotenv.Load(); err != nil {
		// En producción pueden estar en variables de entorno directamente
		log.Println("[Config] Archivo .env no encontrado o no cargado, leyendo variables de entorno")
	}

	minAnticipacion, err := strconv.Atoi(getEnv("MIN_ANTICIPACION_REMOTA", "120"))
	if err != nil {
		minAnticipacion = 120
	}

	AppConfig = &Config{
		DBPath:                getEnv("DB_PATH", "./continental.db"),
		PGDatabaseURL:         getEnv("PG_DATABASE_URL", "postgresql://postgres:1234@localhost:5432/continental_taller?sslmode=disable"),
		JWTSecret:             getEnv("JWT_SECRET", "continental_jwt_secret_dev_2026"),
		JWTExpiresIn:          getEnv("JWT_EXPIRES_IN", "7d"),
		Port:                  getEnv("PORT", "3001"),
		GinMode:               getEnv("GIN_MODE", "debug"),
		NodeEnv:               getEnv("NODE_ENV", "development"),
		UploadDir:             getEnv("UPLOAD_DIR", "/root/Proyectos/ContinentalApp/backend/uploads"),
		Domain:                getEnv("DOMAIN", "continental.moondev.online"),
		MinAnticipacionRemota: minAnticipacion,
	}

	return AppConfig
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

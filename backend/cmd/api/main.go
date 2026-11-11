package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"continental-backend/internal/config"
	"continental-backend/internal/database"
	"continental-backend/internal/handlers"
	"continental-backend/internal/jobs"
	"continental-backend/internal/middleware"
	"continental-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

var startTime time.Time

func main() {
	startTime = time.Now()

	fmt.Println(`
╔══════════════════════════════════════════════════════╗
║   🔧 Continental Taller - Backend API (Go + SQLite)  ║
╚══════════════════════════════════════════════════════╝`)

	// 1. Cargar configuración
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	// 2. Inicializar SQLite
	_, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Error crítico inicializando base de datos SQLite: %v", err)
	}

	// 3. Inicializar Socket.io Hub
	hub, err := websocket.InitSocketHub()
	if err != nil {
		log.Fatalf("❌ Error inicializando Socket.io Hub: %v", err)
	}
	go func() {
		if err := hub.Server.Serve(); err != nil {
			log.Fatalf("❌ Error en Socket.io server: %v", err)
		}
	}()
	defer hub.Server.Close()

	// 4. Iniciar Tareas Programadas (Cron)
	scheduler := jobs.NewJobScheduler()
	scheduler.Start()
	defer scheduler.Stop()

	// 5. Configurar Gin Router (SetMode ya fue llamado antes de gin.Default)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.SetupCORS())

	// Servir archivos estáticos de uploads
	r.Static("/uploads", cfg.UploadDir)

	// Rutas de Socket.io (soporta tanto /socket.io como /socket.io/ y subrutas)
	r.GET("/socket.io", gin.WrapH(hub.Server))
	r.POST("/socket.io", gin.WrapH(hub.Server))
	r.GET("/socket.io/*any", gin.WrapH(hub.Server))
	r.POST("/socket.io/*any", gin.WrapH(hub.Server))

	// Grupo /api
	api := r.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "ok",
				"engine":    "go",
				"database":  "sqlite",
				"timestamp": time.Now().Format(time.RFC3339),
				"uptime":    time.Since(startTime).Seconds(),
			})
		})

		// Registrar handlers
		handlers.NewAuthHandler().RegisterRoutes(api)
		handlers.NewClienteHandler().RegisterRoutes(api)
		handlers.NewVehiculoHandler().RegisterRoutes(api)
		handlers.NewTecnicoHandler().RegisterRoutes(api)
		handlers.NewServicioHandler().RegisterRoutes(api)
		handlers.NewAccesorioHandler().RegisterRoutes(api)
		handlers.NewCitaHandler().RegisterRoutes(api)
		handlers.NewChatHandler().RegisterRoutes(api)
		handlers.NewAdminHandler().RegisterRoutes(api)
	}

	// Iniciar servidor HTTP
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("📡 Servidor Go listo escuchando en http://localhost:%s\n", cfg.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("❌ Error al arrancar el servidor: %v", err)
	}
}

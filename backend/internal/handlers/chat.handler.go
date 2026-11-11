package handlers

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"continental-backend/internal/config"
	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"
	"continental-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatService *services.ChatService
}

func NewChatHandler() *ChatHandler {
	return &ChatHandler{
		chatService: services.NewChatService(),
	}
}

func (h *ChatHandler) RegisterRoutes(rg *gin.RouterGroup) {
	chat := rg.Group("/chat")
	chat.Use(middleware.VerifyToken())
	chat.Use(middleware.RequireRole(constants.RoleCliente, constants.RoleTecnico))
	{
		chat.GET("/:citaId/mensajes", h.ListarMensajes)
		chat.POST("/:citaId/mensaje", h.EnviarMensaje)
		chat.POST("/:citaId/multimedia", h.EnviarMultimedia)
		chat.GET("/:citaId/archivo/*filePath", h.DescargarArchivo)
	}
}

func (h *ChatHandler) ListarMensajes(c *gin.Context) {
	citaId, err := strconv.Atoi(c.Param("citaId"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("citaId inválido"))
		return
	}

	autorTipo := middleware.GetUserRol(c)
	autorId := middleware.GetUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	res, err := h.chatService.ListarMensajes(citaId, autorTipo, autorId, page, limit)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

type EnviarMensajeRequest struct {
	Contenido string `json:"contenido" binding:"required"`
}

func (h *ChatHandler) EnviarMensaje(c *gin.Context) {
	citaId, err := strconv.Atoi(c.Param("citaId"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("citaId inválido"))
		return
	}

	var req EnviarMensajeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("El contenido es obligatorio"))
		return
	}

	autorTipo := middleware.GetUserRol(c)
	autorId := middleware.GetUserID(c)

	data, err := h.chatService.EnviarMensajeTexto(citaId, autorTipo, autorId, req.Contenido)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirNuevoMensajeChat(citaId, data)
	utils.Created(c, data)
}

func (h *ChatHandler) EnviarMultimedia(c *gin.Context) {
	citaId, err := strconv.Atoi(c.Param("citaId"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("citaId inválido"))
		return
	}

	uploadRes, err := middleware.SaveUpload(c, "archivo", "chat", 250*1024*1024, true)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	autorTipo := middleware.GetUserRol(c)
	autorId := middleware.GetUserID(c)

	tipo := "IMAGEN"
	if strings.HasPrefix(uploadRes.MimeType, "video/") {
		tipo = "VIDEO"
	}

	data, err := h.chatService.EnviarMensajeMultimedia(
		citaId,
		autorTipo,
		autorId,
		tipo,
		uploadRes.RelativePath,
		uploadRes.SizeBytes,
	)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirNuevoMensajeChat(citaId, data)
	utils.Created(c, data)
}

func (h *ChatHandler) DescargarArchivo(c *gin.Context) {
	citaId, err := strconv.Atoi(c.Param("citaId"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("citaId inválido"))
		return
	}

	autorTipo := middleware.GetUserRol(c)
	autorId := middleware.GetUserID(c)

	if err := h.chatService.VerificarAccesoChat(citaId, autorTipo, autorId); err != nil {
		utils.RespondError(c, err)
		return
	}

	rawPath := strings.TrimPrefix(c.Param("filePath"), "/")
	if rawPath == "" {
		rawPath = c.Query("path")
	}
	if rawPath == "" {
		utils.RespondError(c, utils.NewValidationError("Ruta de archivo no proporcionada"))
		return
	}

	baseDir, _ := filepath.Abs(config.AppConfig.UploadDir)
	baseDir = filepath.Clean(baseDir)
	baseDirWithSlash := baseDir + string(filepath.Separator)

	var targetPath string
	if filepath.IsAbs(rawPath) {
		targetPath = filepath.Clean(rawPath)
	} else {
		// En la BD se guarda como "uploads/chat/..."
		// por lo que si viene con prefijo "uploads/", lo normalizamos
		cleaned := strings.TrimPrefix(rawPath, "uploads/")
		targetPath = filepath.Clean(filepath.Join(baseDir, cleaned))
	}

	// Seguridad: Validar que no haya path traversal fuera de baseDir
	if targetPath != baseDir && !strings.HasPrefix(targetPath, baseDirWithSlash) {
		utils.RespondError(c, utils.NewForbiddenError("Acceso no permitido fuera del directorio de archivos"))
		return
	}

	fi, err := os.Stat(targetPath)
	if err != nil || fi.IsDir() {
		utils.RespondError(c, utils.NewNotFoundError("Archivo"))
		return
	}

	c.File(targetPath)
}

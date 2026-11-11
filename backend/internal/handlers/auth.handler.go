package handlers

import (
	"encoding/json"

	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		authService: services.NewAuthService(),
	}
}

func (h *AuthHandler) RegisterRoutes(rg *gin.RouterGroup) {
	auth := rg.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/register", h.Register)
		auth.GET("/me", middleware.VerifyToken(), h.Me)
	}
}

type LoginRequest struct {
	Usuario  string `json:"usuario" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("Usuario y contraseña son obligatorios"))
		return
	}

	res, err := h.authService.Login(req.Usuario, req.Password)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AuthHandler) Register(c *gin.Context) {
	var dto services.RegistrarClienteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.authService.RegistrarCliente(dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.Created(c, res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	userId := middleware.GetUserID(c)
	userRol := middleware.GetUserRol(c)

	data, err := h.authService.ObtenerPerfil(userId, userRol)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	// Inyectar rol explícito para compatibilidad con Node:
	// res.json({ success: true, data: { ...data, rol: req.user.rol } })
	b, err := json.Marshal(data)
	if err != nil {
		utils.RespondError(c, utils.NewAppError("Error serializando perfil", 500))
		return
	}
	var resMap map[string]interface{}
	if err := json.Unmarshal(b, &resMap); err != nil {
		utils.RespondError(c, utils.NewAppError("Error procesando perfil", 500))
		return
	}
	resMap["rol"] = userRol

	utils.OK(c, resMap)
}

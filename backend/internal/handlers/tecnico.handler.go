package handlers

import (
	"strconv"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type TecnicoHandler struct {
	tecnicoService *services.TecnicoService
}

func NewTecnicoHandler() *TecnicoHandler {
	return &TecnicoHandler{
		tecnicoService: services.NewTecnicoService(),
	}
}

func (h *TecnicoHandler) RegisterRoutes(rg *gin.RouterGroup) {
	tecnicos := rg.Group("/tecnicos")
	tecnicos.Use(middleware.VerifyToken())
	{
		tecnicos.GET("", h.Listar)
		tecnicos.GET("/:id", h.ObtenerPorId)
		tecnicos.GET("/:id/disponibilidad", h.ObtenerDisponibilidad)

		tecnicos.POST("", middleware.RequireRole(constants.RoleAdmin), h.Crear)
		tecnicos.PUT("/:id", middleware.RequireRole(constants.RoleAdmin), h.Actualizar)
		tecnicos.PATCH("/:id/estado", middleware.RequireRole(constants.RoleAdmin), h.CambiarEstado)
	}
}

func (h *TecnicoHandler) Listar(c *gin.Context) {
	estado := c.Query("estado")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.tecnicoService.Listar(estado, page, limit)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *TecnicoHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.tecnicoService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *TecnicoHandler) ObtenerDisponibilidad(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.tecnicoService.ObtenerDisponibilidad(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *TecnicoHandler) Crear(c *gin.Context) {
	var dto services.CrearTecnicoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.tecnicoService.Crear(dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.Created(c, res)
}

func (h *TecnicoHandler) Actualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var dto services.ActualizarTecnicoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.tecnicoService.Actualizar(id, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

type CambiarEstadoTecnicoRequest struct {
	Estado string `json:"estado" binding:"required"`
}

func (h *TecnicoHandler) CambiarEstado(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var req CambiarEstadoTecnicoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("El estado es obligatorio"))
		return
	}

	res, err := h.tecnicoService.CambiarEstado(id, req.Estado)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

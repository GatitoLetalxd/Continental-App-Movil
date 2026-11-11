package handlers

import (
	"strconv"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ServicioHandler struct {
	servicioService *services.ServicioService
}

func NewServicioHandler() *ServicioHandler {
	return &ServicioHandler{
		servicioService: services.NewServicioService(),
	}
}

func (h *ServicioHandler) RegisterRoutes(rg *gin.RouterGroup) {
	servicios := rg.Group("/servicios")
	servicios.Use(middleware.VerifyToken())
	{
		servicios.GET("", h.Listar)
		servicios.GET("/:id", h.ObtenerPorId)

		servicios.POST("", middleware.RequireRole(constants.RoleAdmin), h.Crear)
		servicios.PUT("/:id", middleware.RequireRole(constants.RoleAdmin), h.Actualizar)
		servicios.PATCH("/:id/estado", middleware.RequireRole(constants.RoleAdmin), h.CambiarEstado)
	}
}

func (h *ServicioHandler) Listar(c *gin.Context) {
	soloActivos := c.Query("soloActivos") == "true"
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	res, err := h.servicioService.Listar(soloActivos, page, limit)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *ServicioHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.servicioService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *ServicioHandler) Crear(c *gin.Context) {
	var dto services.CrearServicioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.servicioService.Crear(dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.Created(c, res)
}

func (h *ServicioHandler) Actualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var dto services.ActualizarServicioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.servicioService.Actualizar(id, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

type CambiarEstadoServicioRequest struct {
	Activo *bool `json:"activo"`
}

func (h *ServicioHandler) CambiarEstado(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var req CambiarEstadoServicioRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Activo == nil {
		utils.RespondError(c, utils.NewValidationError("El campo activo es obligatorio"))
		return
	}

	res, err := h.servicioService.CambiarEstado(id, *req.Activo)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

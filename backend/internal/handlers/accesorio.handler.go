package handlers

import (
	"strconv"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AccesorioHandler struct {
	accesorioService *services.AccesorioService
}

func NewAccesorioHandler() *AccesorioHandler {
	return &AccesorioHandler{
		accesorioService: services.NewAccesorioService(),
	}
}

func (h *AccesorioHandler) RegisterRoutes(rg *gin.RouterGroup) {
	accesorios := rg.Group("/accesorios")
	accesorios.Use(middleware.VerifyToken())
	{
		accesorios.GET("", h.Listar)
		accesorios.GET("/alertas", middleware.RequireRole(constants.RoleAdmin), h.AlertasStock)
		accesorios.GET("/:id", h.ObtenerPorId)

		accesorios.POST("", middleware.RequireRole(constants.RoleAdmin), h.Crear)
		accesorios.PUT("/:id", middleware.RequireRole(constants.RoleAdmin), h.Actualizar)
	}
}

func (h *AccesorioHandler) Listar(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	res, err := h.accesorioService.Listar(page, limit)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AccesorioHandler) AlertasStock(c *gin.Context) {
	res, err := h.accesorioService.ObtenerAlertasStock()
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AccesorioHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.accesorioService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AccesorioHandler) Crear(c *gin.Context) {
	var dto services.CrearAccesorioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.accesorioService.Crear(dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.Created(c, res)
}

func (h *AccesorioHandler) Actualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var dto services.ActualizarAccesorioDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.accesorioService.Actualizar(id, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

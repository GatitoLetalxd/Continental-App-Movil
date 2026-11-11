package handlers

import (
	"strconv"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type VehiculoHandler struct {
	vehiculoService *services.VehiculoService
}

func NewVehiculoHandler() *VehiculoHandler {
	return &VehiculoHandler{
		vehiculoService: services.NewVehiculoService(),
	}
}

func (h *VehiculoHandler) RegisterRoutes(rg *gin.RouterGroup) {
	vehiculos := rg.Group("/vehiculos")
	vehiculos.Use(middleware.VerifyToken())
	{
		vehiculos.GET("/cliente/:clienteId", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.ListarPorCliente)
		vehiculos.GET("/mis-vehiculos", middleware.RequireRole(constants.RoleCliente), h.MisVehiculos)
		vehiculos.GET("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico, constants.RoleCliente), h.ObtenerPorId)
		vehiculos.POST("", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico, constants.RoleCliente), h.Crear)
		vehiculos.PUT("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleCliente), h.Actualizar)
		vehiculos.DELETE("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleCliente), h.Eliminar)
	}
}

func (h *VehiculoHandler) ListarPorCliente(c *gin.Context) {
	clienteId, err := strconv.Atoi(c.Param("clienteId"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("clienteId inválido"))
		return
	}

	res, err := h.vehiculoService.ListarPorCliente(clienteId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *VehiculoHandler) MisVehiculos(c *gin.Context) {
	clienteId := middleware.GetUserID(c)
	res, err := h.vehiculoService.ListarPorCliente(clienteId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *VehiculoHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.vehiculoService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *VehiculoHandler) Crear(c *gin.Context) {
	userRol := middleware.GetUserRol(c)
	userId := middleware.GetUserID(c)

	var dto services.CrearVehiculoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	clienteId := dto.ClienteID
	if userRol == constants.RoleCliente {
		clienteId = userId
	}

	if clienteId <= 0 {
		utils.RespondError(c, utils.NewValidationError("clienteId es obligatorio"))
		return
	}

	res, err := h.vehiculoService.Crear(clienteId, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.Created(c, res)
}

func (h *VehiculoHandler) Actualizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var dto services.ActualizarVehiculoDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.vehiculoService.Actualizar(id, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *VehiculoHandler) Eliminar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	if err := h.vehiculoService.Eliminar(id); err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(200, gin.H{"success": true, "message": "Vehículo eliminado"})
}

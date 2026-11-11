package handlers

import (
	"strconv"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type ClienteHandler struct {
	clienteService *services.ClienteService
}

func NewClienteHandler() *ClienteHandler {
	return &ClienteHandler{
		clienteService: services.NewClienteService(),
	}
}

func (h *ClienteHandler) RegisterRoutes(rg *gin.RouterGroup) {
	clientes := rg.Group("/clientes")
	clientes.Use(middleware.VerifyToken())
	{
		clientes.GET("", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.Listar)
		clientes.GET("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleCliente), h.ObtenerPorId)
		clientes.PUT("/:id", middleware.RequireRole(constants.RoleAdmin, constants.RoleCliente), h.Actualizar)
	}
}

func (h *ClienteHandler) Listar(c *gin.Context) {
	busqueda := c.Query("busqueda")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.clienteService.Listar(busqueda, page, limit)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *ClienteHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.clienteService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *ClienteHandler) Actualizar(c *gin.Context) {
	userRol := middleware.GetUserRol(c)
	userId := middleware.GetUserID(c)

	targetId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	// Si es cliente, solo puede actualizar su propio perfil
	if userRol == constants.RoleCliente {
		targetId = userId
	}

	var dto services.ActualizarClienteDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	res, err := h.clienteService.Actualizar(targetId, dto)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

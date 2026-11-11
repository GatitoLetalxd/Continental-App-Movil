package handlers

import (
	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	adminService *services.AdminService
}

func NewAdminHandler() *AdminHandler {
	return &AdminHandler{
		adminService: services.NewAdminService(),
	}
}

func (h *AdminHandler) RegisterRoutes(rg *gin.RouterGroup) {
	admin := rg.Group("/admin")
	admin.Use(middleware.VerifyToken())
	admin.Use(middleware.RequireRole(constants.RoleAdmin))
	{
		admin.GET("/dashboard", h.Dashboard)
		admin.GET("/reportes/productividad", h.ReporteProductividad)
		admin.GET("/reportes/servicios", h.ReporteServicios)
	}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	res, err := h.adminService.Dashboard()
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AdminHandler) ReporteProductividad(c *gin.Context) {
	fechaInicio := c.Query("fechaInicio")
	fechaFin := c.Query("fechaFin")

	res, err := h.adminService.ReporteProductividad(fechaInicio, fechaFin)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *AdminHandler) ReporteServicios(c *gin.Context) {
	fechaInicio := c.Query("fechaInicio")
	fechaFin := c.Query("fechaFin")

	res, err := h.adminService.ReporteServicios(fechaInicio, fechaFin)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

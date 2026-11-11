package handlers

import (
	"strconv"
	"time"

	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"
	"continental-backend/internal/utils"
	"continental-backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

type CitaHandler struct {
	citaService       *services.CitaService
	citaEstadoService *services.CitaEstadoService
	faseService       *services.FaseService
	whatsappService   *services.WhatsAppService
}

func NewCitaHandler() *CitaHandler {
	return &CitaHandler{
		citaService:       services.NewCitaService(),
		citaEstadoService: services.NewCitaEstadoService(),
		faseService:       services.NewFaseService(),
		whatsappService:   services.NewWhatsAppService(),
	}
}

func (h *CitaHandler) RegisterRoutes(rg *gin.RouterGroup) {
	citas := rg.Group("/citas")
	citas.Use(middleware.VerifyToken())
	{
		// CRUD
		citas.POST("", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico, constants.RoleCliente), h.Crear)
		citas.GET("", h.Listar)
		citas.GET("/sin-asignar", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.ListarSinAsignar)
		citas.GET("/pospuestas", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.ListarPospuestas)
		citas.GET("/fases-defecto", h.ObtenerFasesPorDefecto)
		citas.GET("/:id", h.ObtenerPorId)

		// Transiciones de estado
		citas.PATCH("/:id/aceptar", middleware.RequireRole(constants.RoleTecnico), h.Aceptar)
		citas.PATCH("/:id/iniciar", middleware.RequireRole(constants.RoleTecnico), h.Iniciar)
		citas.PATCH("/:id/pausar", middleware.RequireRole(constants.RoleTecnico), h.Pausar)
		citas.PATCH("/:id/reanudar", middleware.RequireRole(constants.RoleTecnico), h.Reanudar)
		citas.PATCH("/:id/finalizar", middleware.RequireRole(constants.RoleTecnico), h.Finalizar)
		citas.PATCH("/:id/duracion", middleware.RequireRole(constants.RoleTecnico, constants.RoleAdmin), h.ActualizarDuracion)
		citas.PATCH("/:id/entregar", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.Entregar)
		citas.PATCH("/:id/posponer", middleware.RequireRole(constants.RoleTecnico), h.Posponer)
		citas.PATCH("/:id/cancelar", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.Cancelar)
		citas.PATCH("/:id/reasignar", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.Reasignar)

		// Fotos y evidencias
		citas.POST("/:id/foto/:tipo", middleware.RequireRole(constants.RoleTecnico), h.SubirFoto)
		citas.POST("/:id/evidencia", middleware.RequireRole(constants.RoleTecnico), h.SubirEvidencia)

		// Fases
		citas.PATCH("/:id/fase", middleware.RequireRole(constants.RoleTecnico), h.CambiarFase)
		citas.PATCH("/:id/fase/avanzar", middleware.RequireRole(constants.RoleTecnico), h.AvanzarFase)
		citas.PATCH("/:id/fase/resetear", middleware.RequireRole(constants.RoleTecnico), h.ResetearFase)
		citas.GET("/:id/fases", h.ObtenerHistorialFases)

		// WhatsApp
		citas.GET("/:id/whatsapp/:tipo", middleware.RequireRole(constants.RoleAdmin, constants.RoleTecnico), h.ObtenerLinkWhatsApp)
	}
}

func (h *CitaHandler) Crear(c *gin.Context) {
	userRol := middleware.GetUserRol(c)
	userId := middleware.GetUserID(c)

	var dto services.CrearCitaDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		utils.RespondError(c, utils.NewValidationError("Datos inválidos: "+err.Error()))
		return
	}

	if userRol == constants.RoleCliente {
		dto.ClienteID = userId
	}
	dto.CreadorID = userId

	data, err := h.citaService.Crear(dto, userRol)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	if userRol == constants.RoleTecnico && dto.AutoAceptar {
		websocket.Hub.EmitirEstadoCambiado(data.ID, "ACEPTADA")
		websocket.Hub.EmitirDisponibilidadCambiada(userId)
	}

	utils.Created(c, data)
}

func (h *CitaHandler) Listar(c *gin.Context) {
	userRol := middleware.GetUserRol(c)
	userId := middleware.GetUserID(c)

	estado := c.Query("estado")
	todas := c.Query("todas")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	var tecnicoId *int
	var clienteId *int

	if userRol == constants.RoleCliente {
		clienteId = &userId
	} else if userRol == constants.RoleTecnico && todas != "true" {
		tecnicoId = &userId
	} else {
		if tidStr := c.Query("tecnicoId"); tidStr != "" {
			if tid, err := strconv.Atoi(tidStr); err == nil {
				tecnicoId = &tid
			}
		}
		if cidStr := c.Query("clienteId"); cidStr != "" {
			if cid, err := strconv.Atoi(cidStr); err == nil {
				clienteId = &cid
			}
		}
	}

	filtros := services.FiltrosCitaDTO{
		Estado:    estado,
		TecnicoID: tecnicoId,
		ClienteID: clienteId,
		Page:      page,
		Limit:     limit,
	}

	res, err := h.citaService.Listar(filtros)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *CitaHandler) ListarSinAsignar(c *gin.Context) {
	res, err := h.citaService.ListarSinAsignar()
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.OK(c, res)
}

func (h *CitaHandler) ListarPospuestas(c *gin.Context) {
	res, err := h.citaService.ListarPospuestas()
	if err != nil {
		utils.RespondError(c, err)
		return
	}
	utils.OK(c, res)
}

func (h *CitaHandler) ObtenerFasesPorDefecto(c *gin.Context) {
	res := h.faseService.ObtenerFasesPorDefecto()
	utils.OK(c, res)
}

func (h *CitaHandler) ObtenerPorId(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.citaService.ObtenerPorId(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *CitaHandler) Aceptar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	res, err := h.citaEstadoService.Aceptar(id, userId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "ACEPTADA")
	websocket.Hub.EmitirDisponibilidadCambiada(userId)

	utils.OK(c, res)
}

func (h *CitaHandler) Iniciar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	res, err := h.citaEstadoService.Iniciar(id, userId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "EN_CURSO")
	utils.OK(c, res)
}

func (h *CitaHandler) Pausar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	res, err := h.citaEstadoService.Pausar(id, userId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "PAUSADA")
	utils.OK(c, res)
}

func (h *CitaHandler) Reanudar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	res, err := h.citaEstadoService.Reanudar(id, userId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "EN_CURSO")
	utils.OK(c, res)
}

func (h *CitaHandler) Finalizar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	res, err := h.citaEstadoService.Finalizar(id, userId)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "FINALIZADA")
	utils.OK(c, res)
}

type ActualizarDuracionRequest struct {
	DuracionEstimadaMinutos int `json:"duracionEstimadaMinutos" binding:"required"`
}

func (h *CitaHandler) ActualizarDuracion(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)
	userRol := middleware.GetUserRol(c)

	var req ActualizarDuracionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("duracionEstimadaMinutos es requerida"))
		return
	}

	res, err := h.citaService.ActualizarDuracion(id, userId, userRol, req.DuracionEstimadaMinutos)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *CitaHandler) Entregar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.citaEstadoService.Entregar(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "ENTREGADA")
	utils.OK(c, res)
}

type PosponerRequest struct {
	Motivo *string `json:"motivo"`
}

func (h *CitaHandler) Posponer(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)

	var req PosponerRequest
	_ = c.ShouldBindJSON(&req)

	res, err := h.citaEstadoService.Posponer(id, userId, req.Motivo)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "POSPUESTA")
	websocket.Hub.EmitirCitaPospuesta(id)

	utils.OK(c, res)
}

func (h *CitaHandler) Cancelar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	userId := middleware.GetUserID(c)
	userRol := middleware.GetUserRol(c)

	res, err := h.citaEstadoService.Cancelar(id, userId, userRol)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "CANCELADA")
	utils.OK(c, res)
}

type ReasignarRequest struct {
	TecnicoID          int     `json:"tecnicoId" binding:"required"`
	NuevaHoraAgendada *string `json:"nuevaHoraAgendada"`
}

func (h *CitaHandler) Reasignar(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var req ReasignarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("tecnicoId es requerido"))
		return
	}

	var nuevaHora *time.Time
	if req.NuevaHoraAgendada != nil && *req.NuevaHoraAgendada != "" {
		if parsed, err := time.Parse(time.RFC3339, *req.NuevaHoraAgendada); err == nil {
			nuevaHora = &parsed
		}
	}

	res, err := h.citaEstadoService.Reasignar(id, req.TecnicoID, nuevaHora)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEstadoCambiado(id, "ACEPTADA")
	utils.OK(c, res)
}

func (h *CitaHandler) SubirFoto(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	tipo := c.Param("tipo") // 'inicio' o 'fin'

	uploadRes, err := middleware.SaveUpload(c, "foto", "evidencia", 50*1024*1024, false)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	data, err := h.citaService.SubirFoto(id, tipo, uploadRes.RelativePath)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success":    true,
		"data":       data,
		"archivoUrl": uploadRes.RelativePath,
	})
}

func (h *CitaHandler) SubirEvidencia(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	tipo := c.PostForm("tipo")
	fase := c.PostForm("fase")

	if tipo != "INICIO" && tipo != "AVANCE" && tipo != "FIN" {
		utils.RespondError(c, utils.NewValidationError("Tipo de evidencia no válido. Use INICIO, AVANCE o FIN"))
		return
	}

	uploadRes, err := middleware.SaveUpload(c, "archivo", "evidencia", 200*1024*1024, true)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	data, err := h.citaService.SubirEvidencia(id, tipo, fase, uploadRes.RelativePath, uploadRes.MimeType)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirEvidenciaSubida(id, data)
	utils.OK(c, data)
}

type CambiarFaseRequest struct {
	NombreFase string `json:"nombreFase" binding:"required"`
}

func (h *CitaHandler) CambiarFase(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	var req CambiarFaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondError(c, utils.NewValidationError("nombreFase es requerido"))
		return
	}

	res, err := h.faseService.CambiarFase(id, req.NombreFase)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	websocket.Hub.EmitirFaseCambiada(id, req.NombreFase)
	utils.OK(c, res)
}

func (h *CitaHandler) AvanzarFase(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.faseService.AvanzarFase(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	faseActual := ""
	if res.FaseActual != nil {
		faseActual = *res.FaseActual
	}
	websocket.Hub.EmitirFaseCambiada(id, faseActual)
	utils.OK(c, res)
}

func (h *CitaHandler) ResetearFase(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.faseService.ResetearFasePorDefecto(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *CitaHandler) ObtenerHistorialFases(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}

	res, err := h.faseService.ObtenerHistorial(id)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

func (h *CitaHandler) ObtenerLinkWhatsApp(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.RespondError(c, utils.NewValidationError("ID inválido"))
		return
	}
	tipo := c.Param("tipo")

	res, err := h.whatsappService.GenerarLinkParaCita(id, tipo)
	if err != nil {
		utils.RespondError(c, err)
		return
	}

	utils.OK(c, res)
}

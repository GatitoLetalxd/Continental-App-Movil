package services

import (
	"fmt"
	"math"
	"strings"
	"time"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type CitaService struct {
	db                *gorm.DB
	accesorioService  *AccesorioService
	faseService       *FaseService
	citaEstadoService *CitaEstadoService
}

func NewCitaService() *CitaService {
	return &CitaService{
		db:                database.DB,
		accesorioService:  NewAccesorioService(),
		faseService:       NewFaseService(),
		citaEstadoService: NewCitaEstadoService(),
	}
}

type ItemAccesorioCitaDTO struct {
	AccesorioID int `json:"accesorioId"`
	Cantidad    int `json:"cantidad"`
}

type CrearCitaDTO struct {
	ClienteID               int                    `json:"clienteId"`
	VehiculoID              int                    `json:"vehiculoId" binding:"required"`
	ServicioID              *int                   `json:"servicioId"`
	HoraAgendada            string                 `json:"horaAgendada"`
	DuracionEstimadaMinutos int                    `json:"duracionEstimadaMinutos"`
	EsInmediata             bool                   `json:"esInmediata"`
	Adicionales             *string                `json:"adicionales"`
	ServiciosDetalle        *string                `json:"serviciosDetalle"`
	EsPersonalizado         bool                   `json:"esPersonalizado"`
	AutoAceptar             bool                   `json:"autoAceptar"`
	CreadorID               int                    `json:"creadorId"`
	Accesorios              []ItemAccesorioCitaDTO `json:"accesorios"`
}

type FiltrosCitaDTO struct {
	Estado    string
	TecnicoID *int
	ClienteID *int
	Page      int
	Limit     int
}

type ListarCitasResponse struct {
	Citas      []models.Cita `json:"citas"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	TotalPages int           `json:"totalPages"`
}

func (s *CitaService) Crear(dto CrearCitaDTO, userRol string) (*models.Cita, error) {
	var cliente models.Cliente
	if err := s.db.First(&cliente, dto.ClienteID).Error; err != nil {
		return nil, utils.NewNotFoundError("Cliente")
	}

	var vehiculo models.Vehiculo
	if err := s.db.Where("id = ? AND cliente_id = ?", dto.VehiculoID, dto.ClienteID).First(&vehiculo).Error; err != nil {
		return nil, utils.NewValidationError("El vehículo no pertenece a este cliente")
	}

	var servicio *models.Servicio
	duracion := dto.DuracionEstimadaMinutos

	if dto.ServicioID != nil {
		var srv models.Servicio
		if err := s.db.First(&srv, *dto.ServicioID).Error; err != nil {
			return nil, utils.NewNotFoundError("Servicio")
		}
		if !srv.Activo {
			return nil, utils.NewValidationError("Este servicio no está disponible actualmente")
		}
		servicio = &srv
		if duracion <= 0 {
			duracion = srv.DuracionEstimadaMinutos
		}
	}

	var horaAgendada time.Time
	if dto.EsInmediata {
		horaAgendada = time.Now()
	} else if dto.HoraAgendada != "" {
		parsed, err := time.Parse(time.RFC3339, dto.HoraAgendada)
		if err != nil {
			// Intentar formato sin zona horaria
			parsed, err = time.Parse("2006-01-02T15:04:05", dto.HoraAgendada)
			if err != nil {
				parsed, err = time.Parse("2006-01-02 15:04:05", dto.HoraAgendada)
				if err != nil {
					horaAgendada = time.Now()
				} else {
					horaAgendada = parsed
				}
			} else {
				horaAgendada = parsed
			}
		} else {
			horaAgendada = parsed
		}
	} else {
		horaAgendada = time.Now()
	}

	estado := models.EstadoCitaAgendada
	var tecnicoId *int
	if userRol == constants.RoleTecnico && dto.AutoAceptar {
		estado = models.EstadoCitaAceptada
		tid := dto.CreadorID
		tecnicoId = &tid
		horaAgendada = s.citaEstadoService.CalcularProximaHoraDisponible(tid)
	}

	faseInicial := constants.FasesPorDefecto[0]
	serviciosDetalle := "Servicio Personalizado"
	if dto.ServiciosDetalle != nil && *dto.ServiciosDetalle != "" {
		serviciosDetalle = *dto.ServiciosDetalle
	} else if servicio != nil {
		serviciosDetalle = servicio.Nombre
	}

	cita := models.Cita{
		ClienteID:               dto.ClienteID,
		VehiculoID:              dto.VehiculoID,
		TecnicoID:               tecnicoId,
		ServicioID:              dto.ServicioID,
		Estado:                  estado,
		HoraAgendada:            horaAgendada,
		DuracionEstimadaMinutos: duracion,
		FaseActual:              &faseInicial,
		EsInmediata:             dto.EsInmediata,
		Adicionales:             dto.Adicionales,
		ServiciosDetalle:        &serviciosDetalle,
		EsPersonalizado:         dto.EsPersonalizado,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&cita).Error; err != nil {
			return utils.NewAppError("Error al crear cita: "+err.Error(), 500)
		}

		// Registrar fase inicial en historial
		historial := models.CitaFaseHistorial{
			CitaID:     cita.ID,
			NombreFase: faseInicial,
		}
		if err := tx.Create(&historial).Error; err != nil {
			return utils.NewAppError("Error registrando fase inicial: "+err.Error(), 500)
		}

		// Reservar accesorios si se proporcionaron
		if len(dto.Accesorios) > 0 {
			for _, acc := range dto.Accesorios {
				var accesorio models.Accesorio
				if err := tx.First(&accesorio, acc.AccesorioID).Error; err != nil {
					return utils.NewNotFoundError(fmt.Sprintf("Accesorio %d", acc.AccesorioID))
				}

				res := tx.Model(&models.Accesorio{}).
					Where("id = ? AND stock >= ?", acc.AccesorioID, acc.Cantidad).
					Update("stock", gorm.Expr("stock - ?", acc.Cantidad))

				if res.Error != nil {
					return utils.NewAppError("Error actualizando stock de accesorio: "+res.Error.Error(), 500)
				}
				if res.RowsAffected == 0 {
					return utils.NewValidationError(fmt.Sprintf("Stock insuficiente para \"%s\". Disponible: %d, solicitado: %d", accesorio.Nombre, accesorio.Stock, acc.Cantidad))
				}

				citaAcc := models.CitaAccesorio{
					CitaID:      cita.ID,
					AccesorioID: acc.AccesorioID,
					Cantidad:    acc.Cantidad,
				}
				if err := tx.Create(&citaAcc).Error; err != nil {
					return utils.NewAppError("Error vinculando accesorio a cita: "+err.Error(), 500)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return s.ObtenerPorId(cita.ID)
}

func (s *CitaService) Listar(filtros FiltrosCitaDTO) (*ListarCitasResponse, error) {
	page := filtros.Page
	if page < 1 {
		page = 1
	}
	limit := filtros.Limit
	if limit < 1 {
		limit = 20
	}

	query := s.db.Model(&models.Cita{})
	if filtros.Estado != "" {
		query = query.Where("estado = ?", filtros.Estado)
	}
	if filtros.TecnicoID != nil {
		query = query.Where("tecnico_id = ?", *filtros.TecnicoID)
	}
	if filtros.ClienteID != nil {
		query = query.Where("cliente_id = ?", *filtros.ClienteID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar citas: "+err.Error(), 500)
	}

	var citas []models.Cita
	offset := (page - 1) * limit
	err := query.
		Preload("Cliente").
		Preload("Vehiculo").
		Preload("Servicio").
		Preload("Tecnico").
		Preload("Accesorios.Accesorio").
		Order("hora_agendada DESC").
		Offset(offset).
		Limit(limit).
		Find(&citas).Error

	if err != nil {
		return nil, utils.NewAppError("Error al listar citas: "+err.Error(), 500)
	}

	for i := range citas {
		citas[i].TiempoTranscurrido = utils.CalcularTiempoTranscurrido(citas[i].HoraInicioReal, citas[i].HoraFinReal)
		citas[i].TiempoRestante = utils.CalcularTiempoRestante(citas[i].HoraInicioReal, citas[i].DuracionEstimadaMinutos)
		entrega := utils.CalcularHoraEstimadaEntrega(citas[i].HoraInicioReal, citas[i].HoraAgendada, citas[i].DuracionEstimadaMinutos)
		citas[i].HoraEstimadaEntrega = &entrega
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarCitasResponse{
		Citas:      citas,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *CitaService) ListarSinAsignar() ([]models.Cita, error) {
	var citas []models.Cita
	err := s.db.Where("estado = 'AGENDADA' AND tecnico_id IS NULL").
		Preload("Cliente").
		Preload("Vehiculo").
		Preload("Servicio").
		Order("hora_agendada ASC").
		Find(&citas).Error

	if err != nil {
		return nil, utils.NewAppError("Error al listar citas sin asignar: "+err.Error(), 500)
	}
	return citas, nil
}

func (s *CitaService) ListarPospuestas() ([]models.Cita, error) {
	var citas []models.Cita
	err := s.db.Where("estado = 'POSPUESTA'").
		Preload("Cliente").
		Preload("Vehiculo").
		Preload("Servicio").
		Preload("Accesorios.Accesorio").
		Order("updated_at DESC").
		Find(&citas).Error

	if err != nil {
		return nil, utils.NewAppError("Error al listar citas pospuestas: "+err.Error(), 500)
	}
	return citas, nil
}

func (s *CitaService) ObtenerPorId(id int) (*models.Cita, error) {
	var cita models.Cita
	err := s.db.
		Preload("Cliente").
		Preload("Vehiculo").
		Preload("Servicio").
		Preload("Tecnico").
		Preload("Accesorios.Accesorio").
		Preload("FaseHistorial", func(db *gorm.DB) *gorm.DB {
			return db.Order("timestamp ASC")
		}).
		Preload("Evidencias", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&cita, id).Error

	if err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	cita.TiempoTranscurrido = utils.CalcularTiempoTranscurrido(cita.HoraInicioReal, cita.HoraFinReal)
	cita.TiempoRestante = utils.CalcularTiempoRestante(cita.HoraInicioReal, cita.DuracionEstimadaMinutos)
	entrega := utils.CalcularHoraEstimadaEntrega(cita.HoraInicioReal, cita.HoraAgendada, cita.DuracionEstimadaMinutos)
	cita.HoraEstimadaEntrega = &entrega

	return &cita, nil
}

func (s *CitaService) ActualizarDuracion(citaId, userId int, userRol string, nuevaDuracion int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if userRol != constants.RoleAdmin && (cita.TecnicoID == nil || *cita.TecnicoID != userId) {
		return nil, utils.NewForbiddenError("Solo el técnico asignado o un administrador pueden ajustar la duración")
	}

	if nuevaDuracion <= 0 {
		return nil, utils.NewValidationError("La duración debe ser mayor a 0 minutos")
	}

	if err := s.db.Model(&cita).Update("duracion_estimada_minutos", nuevaDuracion).Error; err != nil {
		return nil, utils.NewAppError("Error al actualizar duración: "+err.Error(), 500)
	}

	return s.ObtenerPorId(citaId)
}

func (s *CitaService) SubirFoto(citaId int, tipo, filePath string) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if tipo == "inicio" {
		if err := s.db.Model(&cita).Update("foto_inicio", filePath).Error; err != nil {
			return nil, utils.NewAppError("Error al guardar foto de inicio: "+err.Error(), 500)
		}
	} else if tipo == "fin" {
		if err := s.db.Model(&cita).Update("foto_fin", filePath).Error; err != nil {
			return nil, utils.NewAppError("Error al guardar foto de fin: "+err.Error(), 500)
		}
	} else {
		return nil, utils.NewValidationError("Tipo de foto inválido. Use 'inicio' o 'fin'")
	}

	return s.ObtenerPorId(citaId)
}

func (s *CitaService) SubirEvidencia(citaId int, tipo, fase, filePath, mimetype string) (*models.Evidencia, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	esVideo := strings.HasPrefix(mimetype, "video/")
	archivoTipo := models.ArchivoTipoImagen
	if esVideo {
		archivoTipo = models.ArchivoTipoVideo
	}

	var evidenciasExistentes []models.Evidencia
	s.db.Where("cita_id = ? AND tipo = ?", citaId, tipo).Find(&evidenciasExistentes)

	numFotos := 0
	numVideos := 0
	for _, e := range evidenciasExistentes {
		if e.ArchivoTipo == models.ArchivoTipoVideo {
			numVideos++
		} else {
			numFotos++
		}
	}

	if esVideo {
		if numVideos >= 1 {
			return nil, utils.NewValidationError(fmt.Sprintf("Solo se permite subir como máximo 1 video para el estado de tipo %s", tipo))
		}
	} else {
		if numFotos >= 5 {
			return nil, utils.NewValidationError(fmt.Sprintf("Solo se permiten subir como máximo 5 fotos para el estado de tipo %s", tipo))
		}
	}

	var fasePtr *string
	if fase != "" {
		fasePtr = &fase
	}

	evidencia := models.Evidencia{
		CitaID:      citaId,
		Tipo:        tipo,
		Fase:        fasePtr,
		ArchivoURL:  filePath,
		ArchivoTipo: archivoTipo,
	}

	if err := s.db.Create(&evidencia).Error; err != nil {
		return nil, utils.NewAppError("Error al guardar evidencia: "+err.Error(), 500)
	}

	return &evidencia, nil
}

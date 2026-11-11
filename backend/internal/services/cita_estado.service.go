package services

import (
	"fmt"
	"strings"
	"time"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type CitaEstadoService struct {
	db               *gorm.DB
	accesorioService *AccesorioService
}

func NewCitaEstadoService() *CitaEstadoService {
	return &CitaEstadoService{
		db:               database.DB,
		accesorioService: NewAccesorioService(),
	}
}

type AceptarCitaResponse struct {
	Cita        models.Cita `json:"cita"`
	Advertencia *string     `json:"advertencia"`
}

func (s *CitaEstadoService) ValidarTransicion(estadoActual, nuevoEstado string) error {
	transicionesValidas, ok := constants.TransicionesCita[estadoActual]
	if !ok {
		return utils.NewConflictError(fmt.Sprintf("Estado actual \"%s\" no reconocido", estadoActual))
	}

	permitido := false
	for _, t := range transicionesValidas {
		if t == nuevoEstado {
			permitido = true
			break
		}
	}

	if !permitido {
		return utils.NewConflictError(fmt.Sprintf(
			"Transición inválida: %s → %s. Transiciones permitidas desde %s: %s",
			estadoActual, nuevoEstado, estadoActual, strings.Join(transicionesValidas, ", "),
		))
	}

	return nil
}

func (s *CitaEstadoService) popularCitaCompleta(citaId int) (*models.Cita, error) {
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
		First(&cita, citaId).Error
	if err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	cita.TiempoTranscurrido = utils.CalcularTiempoTranscurrido(cita.HoraInicioReal, cita.HoraFinReal)
	cita.TiempoRestante = utils.CalcularTiempoRestante(cita.HoraInicioReal, cita.DuracionEstimadaMinutos)
	entrega := utils.CalcularHoraEstimadaEntrega(cita.HoraInicioReal, cita.HoraAgendada, cita.DuracionEstimadaMinutos)
	cita.HoraEstimadaEntrega = &entrega

	return &cita, nil
}

func (s *CitaEstadoService) Aceptar(citaId, tecnicoId int) (*AceptarCitaResponse, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaAceptada); err != nil {
		return nil, err
	}

	var tecnico models.Tecnico
	if err := s.db.First(&tecnico, tecnicoId).Error; err != nil || tecnico.Estado == models.EstadoTecnicoInactivo {
		return nil, utils.NewForbiddenError("El técnico no está activo")
	}

	var citaPausada models.Cita
	errPausada := s.db.Where("tecnico_id = ? AND estado = 'PAUSADA'", tecnicoId).First(&citaPausada).Error
	var advertencia *string
	if errPausada == nil {
		msg := "Tienes una cita en pausa. Cuida tus tiempos para entregar ambas a tiempo."
		advertencia = &msg
	}

	horaAgendadaInteligente := s.CalcularProximaHoraDisponible(tecnicoId)

	cita.Estado = models.EstadoCitaAceptada
	cita.TecnicoID = &tecnicoId
	cita.HoraAgendada = horaAgendadaInteligente

	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al aceptar cita: "+err.Error(), 500)
	}

	completa, err := s.popularCitaCompleta(citaId)
	if err != nil {
		return nil, err
	}

	return &AceptarCitaResponse{
		Cita:        *completa,
		Advertencia: advertencia,
	}, nil
}

func (s *CitaEstadoService) Iniciar(citaId, tecnicoId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.TecnicoID == nil || *cita.TecnicoID != tecnicoId {
		return nil, utils.NewForbiddenError("Solo el técnico asignado puede iniciar esta cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaEnCurso); err != nil {
		return nil, err
	}

	now := time.Now()
	cita.Estado = models.EstadoCitaEnCurso
	cita.HoraInicioReal = &now

	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al iniciar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Pausar(citaId, tecnicoId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.TecnicoID == nil || *cita.TecnicoID != tecnicoId {
		return nil, utils.NewForbiddenError("Solo el técnico asignado puede pausar esta cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaPausada); err != nil {
		return nil, err
	}

	cita.Estado = models.EstadoCitaPausada
	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al pausar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Reanudar(citaId, tecnicoId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.TecnicoID == nil || *cita.TecnicoID != tecnicoId {
		return nil, utils.NewForbiddenError("Solo el técnico asignado puede reanudar esta cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaEnCurso); err != nil {
		return nil, err
	}

	cita.Estado = models.EstadoCitaEnCurso
	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al reanudar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Finalizar(citaId, tecnicoId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.TecnicoID == nil || *cita.TecnicoID != tecnicoId {
		return nil, utils.NewForbiddenError("Solo el técnico asignado puede finalizar esta cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaFinalizada); err != nil {
		return nil, err
	}

	now := time.Now()
	faseFinalizado := "Finalizado"
	cita.Estado = models.EstadoCitaFinalizada
	cita.HoraFinReal = &now
	cita.FaseActual = &faseFinalizado

	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al finalizar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Entregar(citaId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaEntregada); err != nil {
		return nil, err
	}

	cita.Estado = models.EstadoCitaEntregada
	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al entregar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Posponer(citaId, tecnicoId int, motivo *string) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.TecnicoID == nil || *cita.TecnicoID != tecnicoId {
		return nil, utils.NewForbiddenError("Solo el técnico asignado puede posponer esta cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaPospuesta); err != nil {
		return nil, err
	}

	// Accesorios permanecen reservados; técnico se desasigna
	cita.Estado = models.EstadoCitaPospuesta
	cita.TecnicoID = nil
	cita.MotivoPosposicion = motivo

	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al posponer cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Cancelar(citaId, userId int, userRol string) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.Preload("Accesorios").First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if userRol == constants.RoleTecnico {
		if cita.TecnicoID == nil || *cita.TecnicoID != userId {
			return nil, utils.NewForbiddenError("Solo el técnico asignado puede cancelar esta cita")
		}
	} else if userRol == constants.RoleCliente {
		return nil, utils.NewForbiddenError("El cliente no puede cancelar citas directamente. Contacte al técnico o al taller.")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaCancelada); err != nil {
		return nil, err
	}

	// Liberar accesorios reservados de vuelta al stock
	for _, acc := range cita.Accesorios {
		_ = s.accesorioService.LiberarStock(acc.AccesorioID, acc.Cantidad)
	}

	cita.Estado = models.EstadoCitaCancelada
	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al cancelar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) Reasignar(citaId, tecnicoId int, nuevaHoraAgendada *time.Time) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if err := s.ValidarTransicion(cita.Estado, models.EstadoCitaAceptada); err != nil {
		return nil, err
	}

	cita.Estado = models.EstadoCitaAceptada
	cita.TecnicoID = &tecnicoId
	cita.MotivoPosposicion = nil

	if nuevaHoraAgendada != nil {
		cita.HoraAgendada = *nuevaHoraAgendada
	}

	if err := s.db.Save(&cita).Error; err != nil {
		return nil, utils.NewAppError("Error al reasignar cita: "+err.Error(), 500)
	}

	return s.popularCitaCompleta(citaId)
}

func (s *CitaEstadoService) CalcularProximaHoraDisponible(tecnicoId int) time.Time {
	if tecnicoId == 0 {
		return time.Now()
	}

	var citaEnCurso models.Cita
	err := s.db.Where("tecnico_id = ? AND estado = 'EN_CURSO'", tecnicoId).First(&citaEnCurso).Error
	if err == nil {
		baseTime := time.Now()
		if citaEnCurso.HoraInicioReal != nil {
			baseTime = *citaEnCurso.HoraInicioReal
		} else {
			baseTime = citaEnCurso.HoraAgendada
		}
		finEstimado := baseTime.Add(time.Duration(citaEnCurso.DuracionEstimadaMinutos) * time.Minute)
		if finEstimado.After(time.Now()) {
			return finEstimado
		}
	}

	return time.Now()
}

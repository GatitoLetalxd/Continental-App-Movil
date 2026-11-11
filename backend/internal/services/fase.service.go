package services

import (
	"fmt"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type FaseService struct {
	db *gorm.DB
}

func NewFaseService() *FaseService {
	return &FaseService{db: database.DB}
}

func (s *FaseService) CambiarFase(citaId int, nombreFase string) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if len(nombreFase) > constants.MaxCaracteresFase {
		return nil, utils.NewValidationError(fmt.Sprintf("El nombre de la fase no puede exceder %d caracteres", constants.MaxCaracteresFase))
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&cita).Update("fase_actual", nombreFase).Error; err != nil {
			return err
		}

		historial := models.CitaFaseHistorial{
			CitaID:     citaId,
			NombreFase: nombreFase,
		}
		if err := tx.Create(&historial).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, utils.NewAppError("Error al cambiar fase: "+err.Error(), 500)
	}

	var citaActualizada models.Cita
	if err := s.db.Preload("Cliente").Preload("Vehiculo").Preload("Servicio").Preload("Tecnico").First(&citaActualizada, citaId).Error; err == nil {
		citaActualizada.TiempoTranscurrido = utils.CalcularTiempoTranscurrido(citaActualizada.HoraInicioReal, citaActualizada.HoraFinReal)
		citaActualizada.TiempoRestante = utils.CalcularTiempoRestante(citaActualizada.HoraInicioReal, citaActualizada.DuracionEstimadaMinutos)
		entrega := utils.CalcularHoraEstimadaEntrega(citaActualizada.HoraInicioReal, citaActualizada.HoraAgendada, citaActualizada.DuracionEstimadaMinutos)
		citaActualizada.HoraEstimadaEntrega = &entrega
		return &citaActualizada, nil
	}

	cita.FaseActual = &nombreFase
	return &cita, nil
}

func (s *FaseService) ResetearFasePorDefecto(citaId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	faseActual := ""
	if cita.FaseActual != nil {
		faseActual = *cita.FaseActual
	}

	indexActual := -1
	for i, f := range constants.FasesPorDefecto {
		if f == faseActual {
			indexActual = i
			break
		}
	}

	fasePorDefecto := constants.FasesPorDefecto[0]
	if indexActual >= 0 {
		fasePorDefecto = constants.FasesPorDefecto[indexActual]
	} else {
		// Buscar última fase por defecto en historial
		var ultimoHistorial models.CitaFaseHistorial
		err := s.db.Where("cita_id = ? AND nombre_fase IN ?", citaId, constants.FasesPorDefecto).
			Order("timestamp DESC").
			First(&ultimoHistorial).Error

		if err == nil {
			fasePorDefecto = ultimoHistorial.NombreFase
		}
	}

	return s.CambiarFase(citaId, fasePorDefecto)
}

func (s *FaseService) AvanzarFase(citaId int) (*models.Cita, error) {
	var cita models.Cita
	if err := s.db.First(&cita, citaId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	faseActual := ""
	if cita.FaseActual != nil {
		faseActual = *cita.FaseActual
	}

	indexActual := -1
	for i, f := range constants.FasesPorDefecto {
		if f == faseActual {
			indexActual = i
			break
		}
	}

	var siguienteFase string
	if indexActual < 0 {
		var ultimoHistorial models.CitaFaseHistorial
		err := s.db.Where("cita_id = ? AND nombre_fase IN ?", citaId, constants.FasesPorDefecto).
			Order("timestamp DESC").
			First(&ultimoHistorial).Error

		lastIdx := -1
		if err == nil {
			for i, f := range constants.FasesPorDefecto {
				if f == ultimoHistorial.NombreFase {
					lastIdx = i
					break
				}
			}
		}

		nextIdx := lastIdx + 1
		if nextIdx >= len(constants.FasesPorDefecto) {
			nextIdx = len(constants.FasesPorDefecto) - 1
		}
		siguienteFase = constants.FasesPorDefecto[nextIdx]
	} else if indexActual < len(constants.FasesPorDefecto)-1 {
		siguienteFase = constants.FasesPorDefecto[indexActual+1]
	} else {
		return nil, utils.NewValidationError("Ya está en la última fase")
	}

	return s.CambiarFase(citaId, siguienteFase)
}

func (s *FaseService) RegistrarFase(citaId int, nombreFase string) error {
	historial := models.CitaFaseHistorial{
		CitaID:     citaId,
		NombreFase: nombreFase,
	}
	return s.db.Create(&historial).Error
}

func (s *FaseService) ObtenerHistorial(citaId int) ([]models.CitaFaseHistorial, error) {
	var historial []models.CitaFaseHistorial
	if err := s.db.Where("cita_id = ?", citaId).Order("timestamp ASC").Find(&historial).Error; err != nil {
		return nil, utils.NewAppError("Error al consultar historial de fases: "+err.Error(), 500)
	}
	return historial, nil
}

func (s *FaseService) ObtenerFasesPorDefecto() []string {
	return constants.FasesPorDefecto
}

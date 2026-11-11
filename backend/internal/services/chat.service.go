package services

import (
	"fmt"
	"math"
	"time"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type ChatService struct {
	db *gorm.DB
}

func NewChatService() *ChatService {
	return &ChatService{db: database.DB}
}

type ListarMensajesResponse struct {
	Mensajes   []models.Mensaje `json:"mensajes"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	TotalPages int              `json:"totalPages"`
}

func (s *ChatService) VerificarAccesoChat(citaId int, autorTipo string, autorId int) error {
	var cita models.Cita
	if err := s.db.Select("id, cliente_id, tecnico_id").First(&cita, citaId).Error; err != nil {
		return utils.NewNotFoundError("Cita")
	}

	if autorTipo == constants.RoleCliente && cita.ClienteID != autorId {
		return utils.NewForbiddenError("No tienes acceso al chat de esta cita")
	}

	if autorTipo == constants.RoleTecnico {
		if cita.TecnicoID == nil || *cita.TecnicoID != autorId {
			return utils.NewForbiddenError("No tienes acceso al chat de esta cita")
		}
	}

	return nil
}

func (s *ChatService) EnviarMensajeTexto(citaId int, autorTipo string, autorId int, contenido string) (*models.Mensaje, error) {
	if err := s.VerificarAccesoChat(citaId, autorTipo, autorId); err != nil {
		return nil, err
	}

	expiraAt := time.Now().AddDate(0, 0, constants.DiasExpiracionChatMultimedia)

	mensaje := models.Mensaje{
		CitaID:    citaId,
		AutorTipo: autorTipo,
		AutorID:   autorId,
		Contenido: &contenido,
		Tipo:      models.TipoMensajeTexto,
		ExpiraAt:  expiraAt,
	}

	if err := s.db.Create(&mensaje).Error; err != nil {
		return nil, utils.NewAppError("Error al enviar mensaje: "+err.Error(), 500)
	}

	return &mensaje, nil
}

func (s *ChatService) EnviarMensajeMultimedia(citaId int, autorTipo string, autorId int, tipo, archivoUrl string, tamanoBytes int64) (*models.Mensaje, error) {
	if err := s.VerificarAccesoChat(citaId, autorTipo, autorId); err != nil {
		return nil, err
	}

	tamanoMB := float64(tamanoBytes) / (1024 * 1024)

	if tipo == models.TipoMensajeVideo {
		var videosExistentes int64
		s.db.Model(&models.Mensaje{}).
			Where("cita_id = ? AND tipo = 'VIDEO' AND expirado = false", citaId).
			Count(&videosExistentes)

		if videosExistentes >= constants.MaxVideosPorChat {
			return nil, utils.NewValidationError(fmt.Sprintf(
				"Máximo %d videos por conversación. Ya tienes %d.",
				constants.MaxVideosPorChat, videosExistentes,
			))
		}

		if tamanoMB > constants.MaxTamanoVideoCombinadoMB {
			return nil, utils.NewValidationError(fmt.Sprintf(
				"El tamaño del video excede el límite de %dMB",
				constants.MaxTamanoVideoCombinadoMB,
			))
		}
	}

	if tipo == models.TipoMensajeImagen {
		if tamanoMB > constants.MaxTamanoFotoMB {
			return nil, utils.NewValidationError(fmt.Sprintf(
				"La foto excede el límite de %dMB",
				constants.MaxTamanoFotoMB,
			))
		}
	}

	expiraAt := time.Now().AddDate(0, 0, constants.DiasExpiracionChatMultimedia)

	mensaje := models.Mensaje{
		CitaID:     citaId,
		AutorTipo:  autorTipo,
		AutorID:    autorId,
		Tipo:       tipo,
		ArchivoURL: &archivoUrl,
		ExpiraAt:   expiraAt,
	}

	if err := s.db.Create(&mensaje).Error; err != nil {
		return nil, utils.NewAppError("Error al enviar mensaje multimedia: "+err.Error(), 500)
	}

	return &mensaje, nil
}

func (s *ChatService) ListarMensajes(citaId int, autorTipo string, autorId int, page, limit int) (*ListarMensajesResponse, error) {
	if err := s.VerificarAccesoChat(citaId, autorTipo, autorId); err != nil {
		return nil, err
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	var total int64
	if err := s.db.Model(&models.Mensaje{}).Where("cita_id = ?", citaId).Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar mensajes: "+err.Error(), 500)
	}

	var mensajes []models.Mensaje
	offset := (page - 1) * limit
	err := s.db.Where("cita_id = ?", citaId).
		Order("created_at ASC").
		Offset(offset).
		Limit(limit).
		Find(&mensajes).Error

	if err != nil {
		return nil, utils.NewAppError("Error al listar mensajes: "+err.Error(), 500)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarMensajesResponse{
		Mensajes:   mensajes,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

package services

import (
	"math"

	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type ServicioService struct {
	db *gorm.DB
}

func NewServicioService() *ServicioService {
	return &ServicioService{db: database.DB}
}

type ListarServiciosResponse struct {
	Servicios  []models.Servicio `json:"servicios"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	TotalPages int               `json:"totalPages"`
}

type CrearServicioDTO struct {
	Nombre                  string   `json:"nombre" binding:"required"`
	Descripcion             *string  `json:"descripcion"`
	DuracionEstimadaMinutos int      `json:"duracionEstimadaMinutos" binding:"required"`
	Precio                  float64  `json:"precio" binding:"required"`
}

type ActualizarServicioDTO struct {
	Nombre                  *string  `json:"nombre"`
	Descripcion             *string  `json:"descripcion"`
	DuracionEstimadaMinutos *int     `json:"duracionEstimadaMinutos"`
	Precio                  *float64 `json:"precio"`
}

func (s *ServicioService) Listar(soloActivos bool, page, limit int) (*ListarServiciosResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	query := s.db.Model(&models.Servicio{})
	if soloActivos {
		query = query.Where("activo = ?", true)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar servicios: "+err.Error(), 500)
	}

	var servicios []models.Servicio
	offset := (page - 1) * limit
	if err := query.Order("nombre ASC").Offset(offset).Limit(limit).Find(&servicios).Error; err != nil {
		return nil, utils.NewAppError("Error al listar servicios: "+err.Error(), 500)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarServiciosResponse{
		Servicios:  servicios,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *ServicioService) ObtenerPorId(id int) (*models.Servicio, error) {
	var servicio models.Servicio
	if err := s.db.First(&servicio, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Servicio")
	}
	return &servicio, nil
}

func (s *ServicioService) Crear(dto CrearServicioDTO) (*models.Servicio, error) {
	servicio := models.Servicio{
		Nombre:                  dto.Nombre,
		Descripcion:             dto.Descripcion,
		DuracionEstimadaMinutos: dto.DuracionEstimadaMinutos,
		Precio:                  dto.Precio,
		Activo:                  true,
	}

	if err := s.db.Create(&servicio).Error; err != nil {
		return nil, utils.NewAppError("Error al crear servicio: "+err.Error(), 500)
	}

	return &servicio, nil
}

func (s *ServicioService) Actualizar(id int, dto ActualizarServicioDTO) (*models.Servicio, error) {
	var servicio models.Servicio
	if err := s.db.First(&servicio, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Servicio")
	}

	updates := make(map[string]interface{})
	if dto.Nombre != nil && *dto.Nombre != "" {
		updates["nombre"] = *dto.Nombre
	}
	if dto.Descripcion != nil {
		updates["descripcion"] = *dto.Descripcion
	}
	if dto.DuracionEstimadaMinutos != nil {
		updates["duracion_estimada_minutos"] = *dto.DuracionEstimadaMinutos
	}
	if dto.Precio != nil {
		updates["precio"] = *dto.Precio
	}

	if len(updates) > 0 {
		if err := s.db.Model(&servicio).Updates(updates).Error; err != nil {
			return nil, utils.NewAppError("Error al actualizar servicio: "+err.Error(), 500)
		}
	}

	return s.ObtenerPorId(id)
}

func (s *ServicioService) CambiarEstado(id int, activo bool) (*models.Servicio, error) {
	var servicio models.Servicio
	if err := s.db.First(&servicio, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Servicio")
	}

	if err := s.db.Model(&servicio).Update("activo", activo).Error; err != nil {
		return nil, utils.NewAppError("Error al cambiar estado de servicio: "+err.Error(), 500)
	}

	return s.ObtenerPorId(id)
}

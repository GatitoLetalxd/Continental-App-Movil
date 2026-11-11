package services

import (
	"fmt"
	"math"
	"regexp"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ClienteService struct {
	db *gorm.DB
}

func NewClienteService() *ClienteService {
	return &ClienteService{db: database.DB}
}

type ListarClientesResponse struct {
	Clientes   []models.Cliente `json:"clientes"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	TotalPages int              `json:"totalPages"`
}

type ActualizarClienteDTO struct {
	Nombre    *string  `json:"nombre"`
	Contactos []string `json:"contactos"`
	Pin       *string  `json:"pin"`
}

func (s *ClienteService) Listar(busqueda string, page, limit int) (*ListarClientesResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	query := s.db.Model(&models.Cliente{})
	if busqueda != "" {
		query = query.Where("nombre LIKE ?", "%"+busqueda+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar clientes: "+err.Error(), 500)
	}

	var clientes []models.Cliente
	offset := (page - 1) * limit
	if err := query.Preload("Vehiculos").Order("created_at DESC").Offset(offset).Limit(limit).Find(&clientes).Error; err != nil {
		return nil, utils.NewAppError("Error al listar clientes: "+err.Error(), 500)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarClientesResponse{
		Clientes:   clientes,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *ClienteService) ObtenerPorId(id int) (*models.Cliente, error) {
	var cliente models.Cliente
	if err := s.db.Preload("Vehiculos").First(&cliente, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Cliente")
	}
	return &cliente, nil
}

func (s *ClienteService) Actualizar(id int, dto ActualizarClienteDTO) (*models.Cliente, error) {
	var cliente models.Cliente
	if err := s.db.First(&cliente, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Cliente")
	}

	updates := make(map[string]interface{})

	if dto.Nombre != nil && *dto.Nombre != "" {
		updates["nombre"] = *dto.Nombre
	}

	if dto.Contactos != nil {
		if len(dto.Contactos) > constants.MaxContactosCliente {
			return nil, utils.NewValidationError(fmt.Sprintf("Máximo %d contactos permitidos", constants.MaxContactosCliente))
		}
		contactosVal, err := models.StringSlice(dto.Contactos).Value()
		if err != nil {
			return nil, utils.NewValidationError("Error serializando contactos")
		}
		updates["contactos"] = contactosVal
	}

	if dto.Pin != nil && *dto.Pin != "" {
		match, _ := regexp.MatchString(`^\d{4,8}$`, *dto.Pin)
		if !match {
			return nil, utils.NewValidationError("El PIN debe ser numérico de 4 a 8 dígitos")
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*dto.Pin), 10)
		if err != nil {
			return nil, utils.NewAppError("Error al hashear el PIN", 500)
		}
		updates["pin"] = string(hash)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&cliente).Updates(updates).Error; err != nil {
			return nil, utils.NewAppError("Error al actualizar cliente: "+err.Error(), 500)
		}
	}

	return s.ObtenerPorId(id)
}

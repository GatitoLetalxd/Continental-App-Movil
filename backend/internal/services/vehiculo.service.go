package services

import (
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type VehiculoService struct {
	db *gorm.DB
}

func NewVehiculoService() *VehiculoService {
	return &VehiculoService{db: database.DB}
}

type CrearVehiculoDTO struct {
	Marca     string  `json:"marca" binding:"required"`
	Modelo    string  `json:"modelo" binding:"required"`
	Anio      int     `json:"anio" binding:"required"`
	Placa     string  `json:"placa"`
	Notas     *string `json:"notas"`
	ClienteID int     `json:"clienteId"`
}

type ActualizarVehiculoDTO struct {
	Marca  *string `json:"marca"`
	Modelo *string `json:"modelo"`
	Anio   *int    `json:"anio"`
	Placa  *string `json:"placa"`
	Notas  *string `json:"notas"`
}

func (s *VehiculoService) ListarPorCliente(clienteId int) ([]models.Vehiculo, error) {
	var vehiculos []models.Vehiculo
	if err := s.db.Where("cliente_id = ?", clienteId).Order("id DESC").Find(&vehiculos).Error; err != nil {
		return nil, utils.NewAppError("Error al listar vehículos: "+err.Error(), 500)
	}
	return vehiculos, nil
}

func (s *VehiculoService) ObtenerPorId(id int) (*models.Vehiculo, error) {
	var vehiculo models.Vehiculo
	if err := s.db.Preload("Cliente").First(&vehiculo, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Vehículo")
	}
	return &vehiculo, nil
}

func (s *VehiculoService) Crear(clienteId int, dto CrearVehiculoDTO) (*models.Vehiculo, error) {
	var cliente models.Cliente
	if err := s.db.First(&cliente, clienteId).Error; err != nil {
		return nil, utils.NewNotFoundError("Cliente")
	}

	vehiculo := models.Vehiculo{
		Marca:     dto.Marca,
		Modelo:    dto.Modelo,
		Anio:      dto.Anio,
		Placa:     dto.Placa,
		Notas:     dto.Notas,
		ClienteID: clienteId,
	}

	if err := s.db.Create(&vehiculo).Error; err != nil {
		return nil, utils.NewAppError("Error al crear vehículo: "+err.Error(), 500)
	}

	return s.ObtenerPorId(vehiculo.ID)
}

func (s *VehiculoService) Actualizar(id int, dto ActualizarVehiculoDTO) (*models.Vehiculo, error) {
	var vehiculo models.Vehiculo
	if err := s.db.First(&vehiculo, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Vehículo")
	}

	updates := make(map[string]interface{})
	if dto.Marca != nil {
		updates["marca"] = *dto.Marca
	}
	if dto.Modelo != nil {
		updates["modelo"] = *dto.Modelo
	}
	if dto.Anio != nil {
		updates["anio"] = *dto.Anio
	}
	if dto.Placa != nil {
		updates["placa"] = *dto.Placa
	}
	if dto.Notas != nil {
		updates["notas"] = *dto.Notas
	}

	if len(updates) > 0 {
		if err := s.db.Model(&vehiculo).Updates(updates).Error; err != nil {
			return nil, utils.NewAppError("Error al actualizar vehículo: "+err.Error(), 500)
		}
	}

	return s.ObtenerPorId(id)
}

func (s *VehiculoService) Eliminar(id int) error {
	var vehiculo models.Vehiculo
	if err := s.db.First(&vehiculo, id).Error; err != nil {
		return utils.NewNotFoundError("Vehículo")
	}

	var totalCitas int64
	s.db.Model(&models.Cita{}).Where("vehiculo_id = ?", id).Count(&totalCitas)
	if totalCitas > 0 {
		return utils.NewValidationError("No se puede eliminar un vehículo con citas asociadas")
	}

	if err := s.db.Delete(&vehiculo).Error; err != nil {
		return utils.NewAppError("Error al eliminar vehículo: "+err.Error(), 500)
	}

	return nil
}

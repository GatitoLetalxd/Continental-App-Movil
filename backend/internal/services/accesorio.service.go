package services

import (
	"fmt"
	"math"

	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type AccesorioService struct {
	db *gorm.DB
}

func NewAccesorioService() *AccesorioService {
	return &AccesorioService{db: database.DB}
}

type ListarAccesoriosResponse struct {
	Accesorios []models.Accesorio `json:"accesorios"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	TotalPages int                `json:"totalPages"`
}

type CrearAccesorioDTO struct {
	Nombre            string   `json:"nombre" binding:"required"`
	Descripcion       *string  `json:"descripcion"`
	Precio            float64  `json:"precio" binding:"required"`
	Stock             int      `json:"stock"`
	StockMinimoAlerta int      `json:"stockMinimoAlerta"`
}

type ActualizarAccesorioDTO struct {
	Nombre            *string  `json:"nombre"`
	Descripcion       *string  `json:"descripcion"`
	Precio            *float64 `json:"precio"`
	Stock             *int     `json:"stock"`
	StockMinimoAlerta *int     `json:"stockMinimoAlerta"`
}

func (s *AccesorioService) Listar(page, limit int) (*ListarAccesoriosResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	var total int64
	if err := s.db.Model(&models.Accesorio{}).Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar accesorios: "+err.Error(), 500)
	}

	var accesorios []models.Accesorio
	offset := (page - 1) * limit
	if err := s.db.Order("nombre ASC").Offset(offset).Limit(limit).Find(&accesorios).Error; err != nil {
		return nil, utils.NewAppError("Error al listar accesorios: "+err.Error(), 500)
	}

	for i := range accesorios {
		accesorios[i].StockBajo = accesorios[i].Stock <= accesorios[i].StockMinimoAlerta
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarAccesoriosResponse{
		Accesorios: accesorios,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *AccesorioService) ObtenerPorId(id int) (*models.Accesorio, error) {
	var accesorio models.Accesorio
	if err := s.db.First(&accesorio, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Accesorio")
	}
	accesorio.StockBajo = accesorio.Stock <= accesorio.StockMinimoAlerta
	return &accesorio, nil
}

func (s *AccesorioService) Crear(dto CrearAccesorioDTO) (*models.Accesorio, error) {
	stockMinimo := dto.StockMinimoAlerta
	if stockMinimo <= 0 {
		stockMinimo = 5
	}

	accesorio := models.Accesorio{
		Nombre:            dto.Nombre,
		Descripcion:       dto.Descripcion,
		Precio:            dto.Precio,
		Stock:             dto.Stock,
		StockMinimoAlerta: stockMinimo,
	}

	if err := s.db.Create(&accesorio).Error; err != nil {
		return nil, utils.NewAppError("Error al crear accesorio: "+err.Error(), 500)
	}

	accesorio.StockBajo = accesorio.Stock <= accesorio.StockMinimoAlerta
	return &accesorio, nil
}

func (s *AccesorioService) Actualizar(id int, dto ActualizarAccesorioDTO) (*models.Accesorio, error) {
	var accesorio models.Accesorio
	if err := s.db.First(&accesorio, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Accesorio")
	}

	updates := make(map[string]interface{})
	if dto.Nombre != nil && *dto.Nombre != "" {
		updates["nombre"] = *dto.Nombre
	}
	if dto.Descripcion != nil {
		updates["descripcion"] = *dto.Descripcion
	}
	if dto.Precio != nil {
		updates["precio"] = *dto.Precio
	}
	if dto.Stock != nil {
		updates["stock"] = *dto.Stock
	}
	if dto.StockMinimoAlerta != nil {
		updates["stock_minimo_alerta"] = *dto.StockMinimoAlerta
	}

	if len(updates) > 0 {
		if err := s.db.Model(&accesorio).Updates(updates).Error; err != nil {
			return nil, utils.NewAppError("Error al actualizar accesorio: "+err.Error(), 500)
		}
	}

	return s.ObtenerPorId(id)
}

func (s *AccesorioService) ObtenerAlertasStock() ([]models.Accesorio, error) {
	var accesorios []models.Accesorio
	if err := s.db.Where("stock <= stock_minimo_alerta").Order("stock ASC").Find(&accesorios).Error; err != nil {
		return nil, utils.NewAppError("Error al consultar alertas de stock: "+err.Error(), 500)
	}

	for i := range accesorios {
		accesorios[i].StockBajo = true
		accesorios[i].Deficit = accesorios[i].StockMinimoAlerta - accesorios[i].Stock
	}

	return accesorios, nil
}

func (s *AccesorioService) ReservarStock(accesorioId int, cantidad int) error {
	var accesorio models.Accesorio
	if err := s.db.First(&accesorio, accesorioId).Error; err != nil {
		return utils.NewNotFoundError("Accesorio")
	}

	res := s.db.Model(&models.Accesorio{}).
		Where("id = ? AND stock >= ?", accesorioId, cantidad).
		Update("stock", gorm.Expr("stock - ?", cantidad))

	if res.Error != nil {
		return utils.NewAppError("Error actualizando stock de accesorio: "+res.Error.Error(), 500)
	}
	if res.RowsAffected == 0 {
		return utils.NewValidationError(fmt.Sprintf("Stock insuficiente para \"%s\". Disponible: %d, solicitado: %d", accesorio.Nombre, accesorio.Stock, cantidad))
	}

	return nil
}

func (s *AccesorioService) LiberarStock(accesorioId int, cantidad int) error {
	if err := s.db.Model(&models.Accesorio{}).Where("id = ?", accesorioId).Update("stock", gorm.Expr("stock + ?", cantidad)).Error; err != nil {
		return utils.NewAppError("Error liberando stock de accesorio: "+err.Error(), 500)
	}
	return nil
}

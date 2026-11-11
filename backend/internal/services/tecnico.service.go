package services

import (
	"math"

	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type TecnicoService struct {
	db *gorm.DB
}

func NewTecnicoService() *TecnicoService {
	return &TecnicoService{db: database.DB}
}

type TecnicoItem struct {
	models.Tecnico
	CitasCount int64 `gorm:"->" json:"_count"` // Equivalente a { citas: count } en Prisma
}

type ListarTecnicosResponse struct {
	Tecnicos   []map[string]interface{} `json:"tecnicos"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	TotalPages int                      `json:"totalPages"`
}

type CrearTecnicoDTO struct {
	Nombre   string `json:"nombre" binding:"required"`
	Usuario  string `json:"usuario" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type ActualizarTecnicoDTO struct {
	Nombre   *string `json:"nombre"`
	Usuario  *string `json:"usuario"`
	Password *string `json:"password"`
}

type DisponibilidadTecnicoResponse struct {
	TecnicoID        int    `json:"tecnicoId"`
	Nombre           string `json:"nombre"`
	Estado           string `json:"estado"`
	Disponibilidad   string `json:"disponibilidad"`
	CitasActivas     int    `json:"citasActivas"`
	TieneCitaPausada bool   `json:"tieneCitaPausada"`
}

func (s *TecnicoService) Listar(estado string, page, limit int) (*ListarTecnicosResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	query := s.db.Model(&models.Tecnico{})
	if estado != "" {
		query = query.Where("estado = ?", estado)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, utils.NewAppError("Error al contar técnicos: "+err.Error(), 500)
	}

	var tecnicos []models.Tecnico
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tecnicos).Error; err != nil {
		return nil, utils.NewAppError("Error al listar técnicos: "+err.Error(), 500)
	}

	// Mapear incluyendo el count de citas
	resultado := make([]map[string]interface{}, 0, len(tecnicos))
	for _, t := range tecnicos {
		var citasCount int64
		s.db.Model(&models.Cita{}).Where("tecnico_id = ?", t.ID).Count(&citasCount)

		item := map[string]interface{}{
			"id":        t.ID,
			"nombre":    t.Nombre,
			"usuario":   t.Usuario,
			"estado":    t.Estado,
			"createdAt": t.CreatedAt,
			"_count": map[string]interface{}{
				"citas": citasCount,
			},
		}
		resultado = append(resultado, item)
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return &ListarTecnicosResponse{
		Tecnicos:   resultado,
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
	}, nil
}

func (s *TecnicoService) ObtenerPorId(id int) (*models.Tecnico, error) {
	var tecnico models.Tecnico
	if err := s.db.First(&tecnico, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Técnico")
	}
	return &tecnico, nil
}

func (s *TecnicoService) Crear(dto CrearTecnicoDTO) (*models.Tecnico, error) {
	var existente models.Tecnico
	if err := s.db.Where("usuario = ?", dto.Usuario).First(&existente).Error; err == nil {
		return nil, utils.NewValidationError("Ya existe un técnico con ese usuario")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), 10)
	if err != nil {
		return nil, utils.NewAppError("Error al hashear contraseña", 500)
	}

	tecnico := models.Tecnico{
		Nombre:   dto.Nombre,
		Usuario:  dto.Usuario,
		Password: string(hash),
		Estado:   models.EstadoTecnicoActivo,
	}

	if err := s.db.Create(&tecnico).Error; err != nil {
		return nil, utils.NewAppError("Error al crear técnico: "+err.Error(), 500)
	}

	return s.ObtenerPorId(tecnico.ID)
}

func (s *TecnicoService) Actualizar(id int, dto ActualizarTecnicoDTO) (*models.Tecnico, error) {
	var tecnico models.Tecnico
	if err := s.db.First(&tecnico, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Técnico")
	}

	updates := make(map[string]interface{})
	if dto.Nombre != nil && *dto.Nombre != "" {
		updates["nombre"] = *dto.Nombre
	}
	if dto.Usuario != nil && *dto.Usuario != "" {
		var existente models.Tecnico
		if err := s.db.Where("usuario = ? AND id != ?", *dto.Usuario, id).First(&existente).Error; err == nil {
			return nil, utils.NewValidationError("Ya existe otro técnico con ese usuario")
		}
		updates["usuario"] = *dto.Usuario
	}
	if dto.Password != nil && *dto.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(*dto.Password), 10)
		if err != nil {
			return nil, utils.NewAppError("Error al hashear contraseña", 500)
		}
		updates["password"] = string(hash)
	}

	if len(updates) > 0 {
		if err := s.db.Model(&tecnico).Updates(updates).Error; err != nil {
			return nil, utils.NewAppError("Error al actualizar técnico: "+err.Error(), 500)
		}
	}

	return s.ObtenerPorId(id)
}

func (s *TecnicoService) CambiarEstado(id int, nuevoEstado string) (*models.Tecnico, error) {
	if nuevoEstado != models.EstadoTecnicoActivo && nuevoEstado != models.EstadoTecnicoInactivo {
		return nil, utils.NewValidationError("Estado inválido. Debe ser ACTIVO o INACTIVO")
	}

	var tecnico models.Tecnico
	if err := s.db.First(&tecnico, id).Error; err != nil {
		return nil, utils.NewNotFoundError("Técnico")
	}

	if err := s.db.Model(&tecnico).Update("estado", nuevoEstado).Error; err != nil {
		return nil, utils.NewAppError("Error al cambiar estado del técnico: "+err.Error(), 500)
	}

	return s.ObtenerPorId(id)
}

func (s *TecnicoService) ObtenerDisponibilidad(id int) (*DisponibilidadTecnicoResponse, error) {
	tecnico, err := s.ObtenerPorId(id)
	if err != nil {
		return nil, err
	}

	var citasActivas []models.Cita
	s.db.Where("tecnico_id = ? AND estado IN ('ACEPTADA', 'EN_CURSO', 'PAUSADA')", id).Find(&citasActivas)

	var tieneEnCurso bool
	var tienePausada bool
	for _, c := range citasActivas {
		if c.Estado == models.EstadoCitaEnCurso {
			tieneEnCurso = true
		} else if c.Estado == models.EstadoCitaPausada {
			tienePausada = true
		}
	}

	disponibilidad := "LIBRE"
	if tieneEnCurso {
		disponibilidad = "EN_CURSO"
	} else if tienePausada {
		disponibilidad = "PAUSADO"
	} else if len(citasActivas) > 0 {
		disponibilidad = "OCUPADO"
	}

	return &DisponibilidadTecnicoResponse{
		TecnicoID:        tecnico.ID,
		Nombre:           tecnico.Nombre,
		Estado:           tecnico.Estado,
		Disponibilidad:   disponibilidad,
		CitasActivas:     len(citasActivas),
		TieneCitaPausada: tienePausada,
	}, nil
}

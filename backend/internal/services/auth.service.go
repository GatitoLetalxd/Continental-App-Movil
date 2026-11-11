package services

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"continental-backend/internal/config"
	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/middleware"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService() *AuthService {
	return &AuthService{db: database.DB}
}

type UserPayload struct {
	ID     int    `json:"id"`
	Rol    string `json:"rol"`
	Nombre string `json:"nombre"`
}

type TokenResponse struct {
	Token string      `json:"token"`
	User  UserPayload `json:"user"`
}

type RegistrarClienteDTO struct {
	Nombre    string   `json:"nombre" binding:"required"`
	Pin       string   `json:"pin" binding:"required"`
	Contactos []string `json:"contactos"`
}

func (s *AuthService) Login(usuario, password string) (*TokenResponse, error) {
	usuario = strings.TrimSpace(usuario)

	// 1. Buscar en Admin
	var admin models.Admin
	if err := s.db.Where("usuario = ?", usuario).First(&admin).Error; err == nil {
		if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
			return nil, utils.NewUnauthorizedError("Credenciales incorrectas")
		}
		return s.GenerarToken(admin.ID, constants.RoleAdmin, admin.Nombre)
	}

	// 2. Buscar en Técnico
	var tecnico models.Tecnico
	if err := s.db.Where("usuario = ?", usuario).First(&tecnico).Error; err == nil {
		if tecnico.Estado == models.EstadoTecnicoInactivo {
			return nil, utils.NewUnauthorizedError("Tu cuenta de técnico está inactiva. Contacta al administrador.")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(tecnico.Password), []byte(password)); err != nil {
			return nil, utils.NewUnauthorizedError("Credenciales incorrectas")
		}
		return s.GenerarToken(tecnico.ID, constants.RoleTecnico, tecnico.Nombre)
	}

	// 3. Buscar en Cliente (primero por número de contacto, luego por nombre)
	var clientesEncontrados []models.Cliente

	// En SQLite contactos se guarda como JSON string ["58412...", ...]
	// Escapar % y _ para evitar que el teléfono sea interpretado como patrón SQL LIKE
	usuarioEscapado := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(usuario)
	pattern := fmt.Sprintf("%%\"%s\"%%", usuarioEscapado)
	s.db.Where("contactos LIKE ? ESCAPE '\\'", pattern).Find(&clientesEncontrados)

	if len(clientesEncontrados) == 0 {
		// Fallback: buscar por nombre exacto
		s.db.Where("nombre = ?", usuario).Find(&clientesEncontrados)
		if len(clientesEncontrados) > 1 {
			return nil, utils.NewUnauthorizedError(
				"Hay más de un cliente con ese nombre. Por favor, ingrese su número de teléfono como usuario para acceder de forma segura.",
			)
		}
	}

	for _, cliente := range clientesEncontrados {
		if err := bcrypt.CompareHashAndPassword([]byte(cliente.Pin), []byte(password)); err == nil {
			return s.GenerarToken(cliente.ID, constants.RoleCliente, cliente.Nombre)
		}
	}

	return nil, utils.NewUnauthorizedError("Credenciales incorrectas")
}

func (s *AuthService) RegistrarCliente(dto RegistrarClienteDTO) (*TokenResponse, error) {
	if dto.Nombre == "" || dto.Pin == "" {
		return nil, utils.NewValidationError("Nombre y PIN son obligatorios")
	}

	match, _ := regexp.MatchString(`^\d{4,8}$`, dto.Pin)
	if !match {
		return nil, utils.NewValidationError("El PIN debe ser numérico de 4 a 8 dígitos")
	}

	if len(dto.Contactos) > constants.MaxContactosCliente {
		return nil, utils.NewValidationError(fmt.Sprintf("Máximo %d contactos permitidos", constants.MaxContactosCliente))
	}

	pinHasheado, err := s.HashPassword(dto.Pin)
	if err != nil {
		return nil, utils.NewAppError("Error al hashear el PIN", 500)
	}

	cliente := models.Cliente{
		Nombre:    dto.Nombre,
		Pin:       pinHasheado,
		Contactos: dto.Contactos,
	}

	if err := s.db.Create(&cliente).Error; err != nil {
		return nil, utils.NewAppError("Error al registrar cliente: "+err.Error(), 500)
	}

	return s.GenerarToken(cliente.ID, constants.RoleCliente, cliente.Nombre)
}

func (s *AuthService) ObtenerPerfil(userId int, rol string) (any, error) {
	switch rol {
	case constants.RoleAdmin:
		var admin models.Admin
		if err := s.db.Select("id, usuario, nombre, created_at").First(&admin, userId).Error; err != nil {
			return nil, utils.NewNotFoundError("Admin")
		}
		return admin, nil

	case constants.RoleTecnico:
		var tecnico models.Tecnico
		if err := s.db.Select("id, nombre, usuario, estado, created_at").First(&tecnico, userId).Error; err != nil {
			return nil, utils.NewNotFoundError("Técnico")
		}
		return tecnico, nil

	case constants.RoleCliente:
		var cliente models.Cliente
		if err := s.db.Preload("Vehiculos").First(&cliente, userId).Error; err != nil {
			return nil, utils.NewNotFoundError("Cliente")
		}
		return cliente, nil

	default:
		return nil, utils.NewUnauthorizedError("Rol desconocido")
	}
}

func (s *AuthService) GenerarToken(id int, rol, nombre string) (*TokenResponse, error) {
	duration := 7 * 24 * time.Hour

	claims := middleware.JWTClaims{
		ID:     id,
		Rol:    rol,
		Nombre: nombre,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return nil, utils.NewAppError("Error generando token JWT: "+err.Error(), 500)
	}

	return &TokenResponse{
		Token: tokenString,
		User: UserPayload{
			ID:     id,
			Rol:    rol,
			Nombre: nombre,
		},
	}, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

package utils

import "net/http"

type AppError struct {
	Message       string
	StatusCode    int
	IsOperational bool
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(message string, statusCode int) *AppError {
	return &AppError{
		Message:       message,
		StatusCode:    statusCode,
		IsOperational: true,
	}
}

func NewNotFoundError(entity string) *AppError {
	return &AppError{
		Message:       entity + " no encontrado",
		StatusCode:    http.StatusNotFound,
		IsOperational: true,
	}
}

func NewValidationError(message string) *AppError {
	if message == "" {
		message = "Datos inválidos"
	}
	return &AppError{
		Message:       message,
		StatusCode:    http.StatusBadRequest,
		IsOperational: true,
	}
}

func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "No autorizado"
	}
	return &AppError{
		Message:       message,
		StatusCode:    http.StatusUnauthorized,
		IsOperational: true,
	}
}

func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Acceso denegado"
	}
	return &AppError{
		Message:       message,
		StatusCode:    http.StatusForbidden,
		IsOperational: true,
	}
}

func NewConflictError(message string) *AppError {
	if message == "" {
		message = "Conflicto con el estado actual"
	}
	return &AppError{
		Message:       message,
		StatusCode:    http.StatusConflict,
		IsOperational: true,
	}
}

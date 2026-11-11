package services

import (
	"continental-backend/internal/models"
	"testing"
)

func TestCitaEstadoService_ValidarTransicion(t *testing.T) {
	service := &CitaEstadoService{}

	// Transiciones válidas
	casosValidos := []struct {
		actual string
		nuevo  string
	}{
		{models.EstadoCitaAgendada, models.EstadoCitaAceptada},
		{models.EstadoCitaAgendada, models.EstadoCitaCancelada},
		{models.EstadoCitaAceptada, models.EstadoCitaEnCurso},
		{models.EstadoCitaAceptada, models.EstadoCitaPospuesta},
		{models.EstadoCitaAceptada, models.EstadoCitaCancelada},
		{models.EstadoCitaEnCurso, models.EstadoCitaPausada},
		{models.EstadoCitaEnCurso, models.EstadoCitaFinalizada},
		{models.EstadoCitaPausada, models.EstadoCitaEnCurso},
		{models.EstadoCitaFinalizada, models.EstadoCitaEntregada},
		{models.EstadoCitaPospuesta, models.EstadoCitaAceptada},
	}

	for _, c := range casosValidos {
		if err := service.ValidarTransicion(c.actual, c.nuevo); err != nil {
			t.Errorf("Transición válida rechazada: %s -> %s: %v", c.actual, c.nuevo, err)
		}
	}

	// Transiciones inválidas
	casosInvalidos := []struct {
		actual string
		nuevo  string
	}{
		{models.EstadoCitaAgendada, models.EstadoCitaFinalizada},
		{models.EstadoCitaAgendada, models.EstadoCitaEntregada},
		{models.EstadoCitaEntregada, models.EstadoCitaAceptada},
		{models.EstadoCitaCancelada, models.EstadoCitaEnCurso},
		{models.EstadoCitaFinalizada, models.EstadoCitaEnCurso},
		{models.EstadoCitaPospuesta, models.EstadoCitaFinalizada},
		{"ESTADO_INEXISTENTE", models.EstadoCitaAceptada},
	}

	for _, c := range casosInvalidos {
		if err := service.ValidarTransicion(c.actual, c.nuevo); err == nil {
			t.Errorf("Transición inválida fue permitida: %s -> %s", c.actual, c.nuevo)
		}
	}
}

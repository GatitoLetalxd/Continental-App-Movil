package utils

import (
	"testing"
	"time"
)

func TestCalcularTiempoTranscurrido(t *testing.T) {
	if res := CalcularTiempoTranscurrido(nil, nil); res != 0 {
		t.Fatalf("Esperado 0 cuando horaInicioReal es nil, obtenido %d", res)
	}

	inicio := time.Now().Add(-30 * time.Minute)
	fin := time.Now()
	res := CalcularTiempoTranscurrido(&inicio, &fin)
	if res < 29 || res > 31 {
		t.Fatalf("Esperado ~30 minutos, obtenido %d", res)
	}
}

func TestCalcularTiempoRestante(t *testing.T) {
	if res := CalcularTiempoRestante(nil, 60); res != 60 {
		t.Fatalf("Esperado 60 cuando inicio es nil, obtenido %d", res)
	}

	inicio := time.Now().Add(-20 * time.Minute)
	res := CalcularTiempoRestante(&inicio, 60)
	if res < 39 || res > 41 {
		t.Fatalf("Esperado ~40 minutos, obtenido %d", res)
	}
}

func TestCalcularHoraEstimadaEntrega(t *testing.T) {
	agendada := time.Date(2026, 11, 11, 10, 0, 0, 0, time.UTC)
	entrega := CalcularHoraEstimadaEntrega(nil, agendada, 60)
	esperada := agendada.Add(60 * time.Minute)
	if !entrega.Equal(esperada) {
		t.Fatalf("Esperada hora de entrega %v, obtenida %v", esperada, entrega)
	}

	inicio := agendada.Add(15 * time.Minute)
	entregaConInicio := CalcularHoraEstimadaEntrega(&inicio, agendada, 60)
	esperadaConInicio := inicio.Add(60 * time.Minute)
	if !entregaConInicio.Equal(esperadaConInicio) {
		t.Fatalf("Esperada hora de entrega %v, obtenida %v", esperadaConInicio, entregaConInicio)
	}
}

func TestFormatearHora(t *testing.T) {
	// 2026-11-11 es miércoles
	fecha := time.Date(2026, 11, 11, 14, 30, 0, 0, time.UTC)
	res := FormatearHora(fecha)
	esperado := "miércoles, 11 de noviembre a las 14:30"
	if res != esperado {
		t.Fatalf("Esperado '%s', obtenido '%s'", esperado, res)
	}
}

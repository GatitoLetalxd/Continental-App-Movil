package utils

import (
	"fmt"
	"math"
	"time"
)

// CalcularTiempoTranscurrido calcula el tiempo transcurrido en minutos desde horaInicioReal hasta ahora (o horaFinReal).
func CalcularTiempoTranscurrido(horaInicioReal, horaFinReal *time.Time) int {
	if horaInicioReal == nil {
		return 0
	}
	fin := time.Now()
	if horaFinReal != nil {
		fin = *horaFinReal
	}
	diff := fin.Sub(*horaInicioReal)
	return int(math.Round(diff.Minutes()))
}

// CalcularTiempoRestante calcula el tiempo restante estimado en minutos.
func CalcularTiempoRestante(horaInicioReal *time.Time, duracionEstimadaMinutos int) int {
	if horaInicioReal == nil {
		return duracionEstimadaMinutos
	}
	transcurrido := CalcularTiempoTranscurrido(horaInicioReal, nil)
	return duracionEstimadaMinutos - transcurrido
}

// CalcularHoraEstimadaEntrega calcula la hora estimada de entrega.
func CalcularHoraEstimadaEntrega(horaInicioReal *time.Time, horaAgendada time.Time, duracionEstimadaMinutos int) time.Time {
	base := horaAgendada
	if horaInicioReal != nil {
		base = *horaInicioReal
	}
	return base.Add(time.Duration(duracionEstimadaMinutos) * time.Minute)
}

// FormatearHora formatea una fecha para las plantillas de WhatsApp.
func FormatearHora(fecha time.Time) string {
	dias := []string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}
	meses := []string{"", "enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}

	diaSemana := dias[fecha.Weekday()]
	dia := fecha.Day()
	mes := meses[fecha.Month()]
	hora := fecha.Format("15:04")

	return fmt.Sprintf("%s, %d de %s a las %s", diaSemana, dia, mes, hora)
}

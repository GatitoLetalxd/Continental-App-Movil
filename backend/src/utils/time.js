/**
 * Utilidades de cálculo de tiempo basadas en timestamps.
 * 
 * IMPORTANTE (Sección 5.1 de la spec):
 * NO usar cronómetro en memoria. Todo se calcula por diferencia de timestamps
 * guardados en base de datos. Si el servidor se reinicia, nada se pierde.
 */

/**
 * Calcula el tiempo transcurrido en minutos desde horaInicioReal hasta ahora (o horaFinReal).
 */
function calcularTiempoTranscurrido(horaInicioReal, horaFinReal = null) {
  if (!horaInicioReal) return 0;
  const fin = horaFinReal ? new Date(horaFinReal) : new Date();
  const inicio = new Date(horaInicioReal);
  return Math.round((fin - inicio) / (1000 * 60));
}

/**
 * Calcula el tiempo restante estimado en minutos.
 * Basado en horaInicioReal + duracionEstimadaMinutos - ahora.
 * Retorna negativo si se pasó del estimado.
 */
function calcularTiempoRestante(horaInicioReal, duracionEstimadaMinutos) {
  if (!horaInicioReal) return duracionEstimadaMinutos;
  const transcurrido = calcularTiempoTranscurrido(horaInicioReal);
  return duracionEstimadaMinutos - transcurrido;
}

/**
 * Calcula la hora estimada de entrega.
 * horaInicioReal + duracionEstimadaMinutos, o horaAgendada + duracionEstimadaMinutos si no ha iniciado.
 */
function calcularHoraEstimadaEntrega(horaInicioReal, horaAgendada, duracionEstimadaMinutos) {
  const base = horaInicioReal ? new Date(horaInicioReal) : new Date(horaAgendada);
  return new Date(base.getTime() + duracionEstimadaMinutos * 60 * 1000);
}

/**
 * Formatea una fecha para las plantillas de WhatsApp.
 */
function formatearHora(fecha) {
  return new Date(fecha).toLocaleString('es-ES', {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    hour: '2-digit',
    minute: '2-digit',
  });
}

module.exports = {
  calcularTiempoTranscurrido,
  calcularTiempoRestante,
  calcularHoraEstimadaEntrega,
  formatearHora,
};

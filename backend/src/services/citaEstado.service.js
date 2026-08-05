const prisma = require('../config/database');
const { TRANSICIONES_CITA } = require('../config/constants');
const { ConflictError, ValidationError, ForbiddenError } = require('../utils/errors');
const accesorioService = require('./accesorio.service');

/**
 * Máquina de estados de la Cita.
 * 
 * Estados: AGENDADA → ACEPTADA → EN_CURSO → PAUSADA → FINALIZADA → ENTREGADA
 * Alternativos: POSPUESTA, CANCELADA
 * 
 * Reglas:
 * - Transiciones válidas definidas en TRANSICIONES_CITA (constants.js)
 * - Al cancelar: liberar accesorios reservados
 * - Al posponer: mantener accesorios, desasignar técnico
 * - Al aceptar nueva cita: advertir si tiene cita en pausa
 */

/**
 * Valida si una transición de estado es válida.
 */
function validarTransicion(estadoActual, nuevoEstado) {
  const transicionesValidas = TRANSICIONES_CITA[estadoActual];
  if (!transicionesValidas) {
    throw new ConflictError(`Estado actual "${estadoActual}" no reconocido`);
  }
  if (!transicionesValidas.includes(nuevoEstado)) {
    throw new ConflictError(
      `Transición inválida: ${estadoActual} → ${nuevoEstado}. ` +
      `Transiciones permitidas desde ${estadoActual}: ${transicionesValidas.join(', ') || 'ninguna'}`
    );
  }
}

/**
 * ACEPTAR cita: un técnico la acepta manualmente.
 * AGENDADA → ACEPTADA (o POSPUESTA → ACEPTADA para reasignación)
 */
async function aceptar(citaId, tecnicoId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');

  validarTransicion(cita.estado, 'ACEPTADA');

  // Verificar que el técnico está activo
  const tecnico = await prisma.tecnico.findUnique({ where: { id: tecnicoId } });
  if (!tecnico || tecnico.estado === 'INACTIVO') {
    throw new ForbiddenError('El técnico no está activo');
  }

  // Advertencia si el técnico tiene cita en pausa
  const citaPausada = await prisma.cita.findFirst({
    where: { tecnicoId, estado: 'PAUSADA' },
  });

  const horaAgendadaInteligente = await calcularProximaHoraDisponible(tecnicoId);

  const citaActualizada = await prisma.cita.update({
    where: { id: citaId },
    data: {
      estado: 'ACEPTADA',
      tecnicoId,
      horaAgendada: horaAgendadaInteligente,
    },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });

  return {
    cita: citaActualizada,
    advertencia: citaPausada
      ? 'Tienes una cita en pausa. Cuida tus tiempos para entregar ambas a tiempo.'
      : null,
  };
}

/**
 * INICIAR cita: el técnico presiona "Iniciar".
 * ACEPTADA → EN_CURSO. Guarda horaInicioReal.
 */
async function iniciar(citaId, tecnicoId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');
  if (cita.tecnicoId !== tecnicoId) throw new ForbiddenError('Solo el técnico asignado puede iniciar esta cita');

  validarTransicion(cita.estado, 'EN_CURSO');

  return prisma.cita.update({
    where: { id: citaId },
    data: {
      estado: 'EN_CURSO',
      horaInicioReal: new Date(),
    },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * PAUSAR cita.
 * EN_CURSO → PAUSADA.
 */
async function pausar(citaId, tecnicoId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');
  if (cita.tecnicoId !== tecnicoId) throw new ForbiddenError('Solo el técnico asignado puede pausar esta cita');

  validarTransicion(cita.estado, 'PAUSADA');

  return prisma.cita.update({
    where: { id: citaId },
    data: { estado: 'PAUSADA' },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * REANUDAR cita.
 * PAUSADA → EN_CURSO.
 */
async function reanudar(citaId, tecnicoId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');
  if (cita.tecnicoId !== tecnicoId) throw new ForbiddenError('Solo el técnico asignado puede reanudar esta cita');

  validarTransicion(cita.estado, 'EN_CURSO');

  return prisma.cita.update({
    where: { id: citaId },
    data: { estado: 'EN_CURSO' },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * FINALIZAR cita.
 * EN_CURSO → FINALIZADA. Requiere fotoFin. Guarda horaFinReal.
 */
async function finalizar(citaId, tecnicoId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');
  if (cita.tecnicoId !== tecnicoId) throw new ForbiddenError('Solo el técnico asignado puede finalizar esta cita');

  validarTransicion(cita.estado, 'FINALIZADA');

  return prisma.cita.update({
    where: { id: citaId },
    data: {
      estado: 'FINALIZADA',
      horaFinReal: new Date(),
      faseActual: 'Finalizado',
    },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * ENTREGAR cita.
 * FINALIZADA → ENTREGADA. Marcado manual al retirar el vehículo.
 */
async function entregar(citaId) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');

  validarTransicion(cita.estado, 'ENTREGADA');

  return prisma.cita.update({
    where: { id: citaId },
    data: { estado: 'ENTREGADA' },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * POSPONER cita (solo técnico).
 * [ACEPTADA|EN_CURSO|PAUSADA] → POSPUESTA.
 * Accesorios PERMANECEN reservados. Técnico se desasigna.
 */
async function posponer(citaId, tecnicoId, motivo) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');
  if (cita.tecnicoId !== tecnicoId) throw new ForbiddenError('Solo el técnico asignado puede posponer esta cita');

  validarTransicion(cita.estado, 'POSPUESTA');

  // Accesorios PERMANECEN reservados (no se liberan)
  // Técnico se desasigna para que otro pueda tomarla
  return prisma.cita.update({
    where: { id: citaId },
    data: {
      estado: 'POSPUESTA',
      tecnicoId: null,
      motivoPosposicion: motivo || null,
    },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      accesorios: { include: { accesorio: true } },
    },
  });
}

/**
 * CANCELAR cita (técnico o admin).
 * [ACEPTADA|EN_CURSO|PAUSADA] → CANCELADA.
 * Accesorios SE LIBERAN de vuelta al stock.
 */
async function cancelar(citaId, userId, userRol) {
  const cita = await prisma.cita.findUnique({
    where: { id: citaId },
    include: { accesorios: true },
  });
  if (!cita) throw new ConflictError('Cita no encontrada');

  // Solo técnico asignado o admin pueden cancelar
  if (userRol === 'TECNICO' && cita.tecnicoId !== userId) {
    throw new ForbiddenError('Solo el técnico asignado puede cancelar esta cita');
  }
  if (userRol === 'CLIENTE') {
    throw new ForbiddenError('El cliente no puede cancelar citas directamente. Contacte al técnico o al taller.');
  }

  validarTransicion(cita.estado, 'CANCELADA');

  // Liberar accesorios reservados
  for (const citaAccesorio of cita.accesorios) {
    await accesorioService.liberarStock(citaAccesorio.accesorioId, citaAccesorio.cantidad);
  }

  return prisma.cita.update({
    where: { id: citaId },
    data: { estado: 'CANCELADA' },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * REASIGNAR cita pospuesta a otro técnico/fecha (admin o técnico auto-asignándose).
 * POSPUESTA → ACEPTADA.
 */
async function reasignar(citaId, tecnicoId, nuevaHoraAgendada) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new ConflictError('Cita no encontrada');

  validarTransicion(cita.estado, 'ACEPTADA');

  const updateData = {
    estado: 'ACEPTADA',
    tecnicoId,
    motivoPosposicion: null,
  };

  if (nuevaHoraAgendada) {
    updateData.horaAgendada = new Date(nuevaHoraAgendada);
  }

  return prisma.cita.update({
    where: { id: citaId },
    data: updateData,
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      tecnico: true,
    },
  });
}

/**
 * Calcula la fecha y hora en que el técnico estará disponible.
 * Si tiene una cita EN_CURSO, devuelve la hora de finalización estimada.
 * Si no, devuelve la hora actual.
 */
async function calcularProximaHoraDisponible(tecnicoId) {
  if (!tecnicoId) return new Date();

  // Buscar cita activa del técnico
  const citaEnCurso = await prisma.cita.findFirst({
    where: {
      tecnicoId: parseInt(tecnicoId),
      estado: 'EN_CURSO',
    },
  });

  if (citaEnCurso) {
    const baseTime = citaEnCurso.horaInicioReal || citaEnCurso.horaAgendada || new Date();
    const duracionMinutos = citaEnCurso.duracionEstimadaMinutos || 0;
    const finEstimado = new Date(new Date(baseTime).getTime() + duracionMinutos * 60 * 1000);
    
    // Si la hora de fin estimada es menor que el momento actual, usar el momento actual
    return finEstimado > new Date() ? finEstimado : new Date();
  }

  return new Date();
}

module.exports = {
  validarTransicion,
  aceptar,
  iniciar,
  pausar,
  reanudar,
  finalizar,
  entregar,
  posponer,
  cancelar,
  reasignar,
  calcularProximaHoraDisponible,
};

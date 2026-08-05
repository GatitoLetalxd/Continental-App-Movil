const prisma = require('../config/database');
const { FASES_POR_DEFECTO, MAX_CARACTERES_FASE } = require('../config/constants');
const { NotFoundError, ValidationError } = require('../utils/errors');

/**
 * Cambiar la fase actual de una cita.
 * El técnico puede usar un nombre de la plantilla por defecto o texto libre (máx. 30 chars).
 * Cada cambio se registra en CitaFaseHistorial para trazabilidad.
 */
async function cambiarFase(citaId, nombreFase) {
  const cita = await prisma.cita.findUnique({ where: { id: parseInt(citaId) } });
  if (!cita) throw new NotFoundError('Cita');

  // Validar longitud
  if (nombreFase.length > MAX_CARACTERES_FASE) {
    throw new ValidationError(`El nombre de la fase no puede exceder ${MAX_CARACTERES_FASE} caracteres`);
  }

  // Actualizar fase actual y registrar en historial
  const [citaActualizada] = await prisma.$transaction([
    prisma.cita.update({
      where: { id: parseInt(citaId) },
      data: { faseActual: nombreFase },
    }),
    prisma.citaFaseHistorial.create({
      data: {
        citaId: parseInt(citaId),
        nombreFase,
      },
    }),
  ]);

  return citaActualizada;
}

/**
 * Resetear la fase actual al nombre por defecto de la plantilla.
 * Busca cuál es la fase por defecto que corresponde según el historial.
 */
async function resetearFasePorDefecto(citaId) {
  const cita = await prisma.cita.findUnique({ where: { id: parseInt(citaId) } });
  if (!cita) throw new NotFoundError('Cita');

  // Buscar la siguiente fase por defecto
  const indexActual = FASES_POR_DEFECTO.indexOf(cita.faseActual);
  
  // Si la fase actual no es una por defecto, intentar determinar cuál debería ser
  // basado en la última fase por defecto registrada
  let fasePorDefecto;
  if (indexActual >= 0) {
    fasePorDefecto = FASES_POR_DEFECTO[indexActual];
  } else {
    // Buscar la última fase por defecto en el historial
    const ultimaFasePorDefecto = await prisma.citaFaseHistorial.findMany({
      where: {
        citaId: parseInt(citaId),
        nombreFase: { in: FASES_POR_DEFECTO },
      },
      orderBy: { timestamp: 'desc' },
      take: 1,
    });

    if (ultimaFasePorDefecto.length > 0) {
      fasePorDefecto = ultimaFasePorDefecto[0].nombreFase;
    } else {
      fasePorDefecto = FASES_POR_DEFECTO[0]; // Fallback a primera fase
    }
  }

  return cambiarFase(citaId, fasePorDefecto);
}

/**
 * Avanzar a la siguiente fase por defecto en la plantilla.
 */
async function avanzarFase(citaId) {
  const cita = await prisma.cita.findUnique({ where: { id: parseInt(citaId) } });
  if (!cita) throw new NotFoundError('Cita');

  const indexActual = FASES_POR_DEFECTO.indexOf(cita.faseActual);
  
  let siguienteFase;
  if (indexActual < 0) {
    // Si el texto actual es libre, buscar la última fase por defecto y avanzar desde ahí
    const ultimaFasePorDefecto = await prisma.citaFaseHistorial.findMany({
      where: {
        citaId: parseInt(citaId),
        nombreFase: { in: FASES_POR_DEFECTO },
      },
      orderBy: { timestamp: 'desc' },
      take: 1,
    });

    const lastIdx = ultimaFasePorDefecto.length > 0
      ? FASES_POR_DEFECTO.indexOf(ultimaFasePorDefecto[0].nombreFase)
      : -1;
    
    siguienteFase = FASES_POR_DEFECTO[Math.min(lastIdx + 1, FASES_POR_DEFECTO.length - 1)];
  } else if (indexActual < FASES_POR_DEFECTO.length - 1) {
    siguienteFase = FASES_POR_DEFECTO[indexActual + 1];
  } else {
    throw new ValidationError('Ya está en la última fase');
  }

  return cambiarFase(citaId, siguienteFase);
}

/**
 * Registrar una fase en el historial (usado internamente).
 */
async function registrarFase(citaId, nombreFase) {
  return prisma.citaFaseHistorial.create({
    data: {
      citaId: parseInt(citaId),
      nombreFase,
    },
  });
}

/**
 * Obtener historial de fases de una cita.
 */
async function obtenerHistorial(citaId) {
  return prisma.citaFaseHistorial.findMany({
    where: { citaId: parseInt(citaId) },
    orderBy: { timestamp: 'asc' },
  });
}

/**
 * Obtener las fases por defecto (plantilla).
 */
function obtenerFasesPorDefecto() {
  return FASES_POR_DEFECTO;
}

module.exports = {
  cambiarFase,
  resetearFasePorDefecto,
  avanzarFase,
  registrarFase,
  obtenerHistorial,
  obtenerFasesPorDefecto,
};

const prisma = require('../config/database');
const { NotFoundError, ValidationError, ForbiddenError } = require('../utils/errors');
const { FASES_POR_DEFECTO } = require('../config/constants');
const accesorioService = require('./accesorio.service');
const faseService = require('./fase.service');
const { calcularTiempoTranscurrido, calcularTiempoRestante, calcularHoraEstimadaEntrega } = require('../utils/time');

/**
 * Crear una nueva cita.
 */
async function crear(datos, userRol) {
  // Verificar que el cliente existe
  const cliente = await prisma.cliente.findUnique({ where: { id: datos.clienteId } });
  if (!cliente) throw new NotFoundError('Cliente');

  // Verificar que el vehículo pertenece al cliente
  const vehiculo = await prisma.vehiculo.findFirst({
    where: { id: datos.vehiculoId, clienteId: datos.clienteId },
  });
  if (!vehiculo) throw new ValidationError('El vehículo no pertenece a este cliente');

  // Verificar que el servicio existe y está activo (si se suministra servicioId)
  let servicio = null;
  let duracionEstimadaMinutos = datos.duracionEstimadaMinutos || 0;

  if (datos.servicioId) {
    servicio = await prisma.servicio.findUnique({ where: { id: datos.servicioId } });
    if (!servicio) throw new NotFoundError('Servicio');
    if (!servicio.activo) throw new ValidationError('Este servicio no está disponible actualmente');
    if (!duracionEstimadaMinutos) {
      duracionEstimadaMinutos = servicio.duracionEstimadaMinutos;
    }
  }

  let horaAgendada = new Date(datos.horaAgendada);
  const ahora = new Date();
  const esInmediata = datos.esInmediata || false;

  if (esInmediata) {
    // Cita inmediata: horaAgendada = ahora
    horaAgendada = ahora;
  }

  // Determinar si la cita se auto-acepta inmediatamente
  let estado = 'AGENDADA';
  let tecnicoId = null;
  if (userRol === 'TECNICO' && datos.autoAceptar) {
    estado = 'ACEPTADA';
    tecnicoId = parseInt(datos.creadorId);

    // Calcular hora de inicio inteligente si se auto-acepta
    const { calcularProximaHoraDisponible } = require('./citaEstado.service');
    horaAgendada = await calcularProximaHoraDisponible(tecnicoId);
  }

  // Crear la cita
  const cita = await prisma.cita.create({
    data: {
      clienteId: datos.clienteId,
      vehiculoId: datos.vehiculoId,
      tecnicoId,
      estado,
      servicioId: datos.servicioId || null,
      horaAgendada,
      duracionEstimadaMinutos,
      faseActual: FASES_POR_DEFECTO[0],
      esInmediata,
      adicionales: datos.adicionales || null,
      serviciosDetalle: datos.serviciosDetalle || (servicio ? servicio.nombre : 'Servicio Personalizado'),
      esPersonalizado: datos.esPersonalizado || false,
    },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
    },
  });

  // Registrar la fase inicial en el historial
  await faseService.registrarFase(cita.id, FASES_POR_DEFECTO[0]);

  // Reservar accesorios si se proporcionaron
  if (datos.accesorios && datos.accesorios.length > 0) {
    for (const acc of datos.accesorios) {
      // Reservar stock
      await accesorioService.reservarStock(acc.accesorioId, acc.cantidad);

      // Crear relación CitaAccesorio
      await prisma.citaAccesorio.create({
        data: {
          citaId: cita.id,
          accesorioId: acc.accesorioId,
          cantidad: acc.cantidad,
        },
      });
    }
  }

  // Retornar cita completa
  return prisma.cita.findUnique({
    where: { id: cita.id },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
      accesorios: { include: { accesorio: true } },
      faseHistorial: true,
    },
  });
}

/**
 * Listar citas con filtros opcionales.
 */
async function listar({ estado, tecnicoId, clienteId, page = 1, limit = 20 }) {
  const where = {};
  if (estado) where.estado = estado;
  if (tecnicoId) where.tecnicoId = parseInt(tecnicoId);
  if (clienteId) where.clienteId = parseInt(clienteId);

  const [citas, total] = await Promise.all([
    prisma.cita.findMany({
      where,
      include: {
        cliente: { select: { id: true, nombre: true, contactos: true } },
        vehiculo: true,
        servicio: true,
        tecnico: { select: { id: true, nombre: true } },
        accesorios: { include: { accesorio: true } },
      },
      orderBy: { horaAgendada: 'desc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.cita.count({ where }),
  ]);

  // Enriquecer con cálculos de tiempo
  const citasConTiempo = citas.map(cita => ({
    ...cita,
    tiempoTranscurrido: calcularTiempoTranscurrido(cita.horaInicioReal, cita.horaFinReal),
    tiempoRestante: calcularTiempoRestante(cita.horaInicioReal, cita.duracionEstimadaMinutos),
    horaEstimadaEntrega: calcularHoraEstimadaEntrega(cita.horaInicioReal, cita.horaAgendada, cita.duracionEstimadaMinutos),
  }));

  return { citas: citasConTiempo, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Obtener citas sin asignar (AGENDADAS sin técnico).
 */
async function listarSinAsignar() {
  return prisma.cita.findMany({
    where: {
      estado: 'AGENDADA',
      tecnicoId: null,
    },
    include: {
      cliente: { select: { id: true, nombre: true } },
      vehiculo: true,
      servicio: true,
    },
    orderBy: { horaAgendada: 'asc' },
  });
}

/**
 * Obtener citas pospuestas (bandeja de reasignación).
 */
async function listarPospuestas() {
  return prisma.cita.findMany({
    where: { estado: 'POSPUESTA' },
    include: {
      cliente: { select: { id: true, nombre: true, contactos: true } },
      vehiculo: true,
      servicio: true,
      accesorios: { include: { accesorio: true } },
    },
    orderBy: { updatedAt: 'desc' },
  });
}

/**
 * Obtener detalle completo de una cita.
 */
async function obtenerPorId(id) {
  const cita = await prisma.cita.findUnique({
    where: { id: parseInt(id) },
    include: {
      cliente: { select: { id: true, nombre: true, contactos: true } },
      vehiculo: true,
      servicio: true,
      tecnico: { select: { id: true, nombre: true } },
      accesorios: { include: { accesorio: true } },
      faseHistorial: { orderBy: { timestamp: 'asc' } },
      evidencias: { orderBy: { createdAt: 'asc' } },
    },
  });

  if (!cita) throw new NotFoundError('Cita');

  return {
    ...cita,
    tiempoTranscurrido: calcularTiempoTranscurrido(cita.horaInicioReal, cita.horaFinReal),
    tiempoRestante: calcularTiempoRestante(cita.horaInicioReal, cita.duracionEstimadaMinutos),
    horaEstimadaEntrega: calcularHoraEstimadaEntrega(cita.horaInicioReal, cita.horaAgendada, cita.duracionEstimadaMinutos),
  };
}

/**
 * Actualizar duración estimada (técnico ajusta tras inspección).
 */
async function actualizarDuracion(citaId, tecnicoId, nuevaDuracion) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new NotFoundError('Cita');
  if (cita.tecnicoId !== tecnicoId) {
    throw new ForbiddenError('Solo el técnico asignado puede ajustar la duración');
  }

  return prisma.cita.update({
    where: { id: citaId },
    data: { duracionEstimadaMinutos: parseInt(nuevaDuracion) },
    include: {
      cliente: true,
      vehiculo: true,
      servicio: true,
    },
  });
}

/**
 * Subir foto de inicio/fin a una cita.
 */
async function subirFoto(citaId, tipo, filePath) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new NotFoundError('Cita');

  if (tipo === 'inicio') {
    return prisma.cita.update({
      where: { id: citaId },
      data: { fotoInicio: filePath },
    });
  } else if (tipo === 'fin') {
    return prisma.cita.update({
      where: { id: citaId },
      data: { fotoFin: filePath },
    });
  } else {
    throw new ValidationError('Tipo de foto inválido. Use "inicio" o "fin"');
  }
}

async function subirEvidencia(citaId, tipo, fase, filePath, mimetype) {
  const cita = await prisma.cita.findUnique({ where: { id: citaId } });
  if (!cita) throw new NotFoundError('Cita');

  const esVideo = mimetype.startsWith('video/');
  const archivoTipo = esVideo ? 'VIDEO' : 'IMAGEN';

  // Contar evidencias existentes del mismo tipo
  const evidenciasExistentes = await prisma.evidencia.findMany({
    where: { citaId, tipo }
  });

  const numFotos = evidenciasExistentes.filter(e => e.archivoTipo === 'IMAGEN').length;
  const numVideos = evidenciasExistentes.filter(e => e.archivoTipo === 'VIDEO').length;

  if (esVideo) {
    if (numVideos >= 1) {
      throw new ValidationError(`Solo se permite subir como máximo 1 video para el estado de tipo ${tipo}`);
    }
  } else {
    if (numFotos >= 5) {
      throw new ValidationError(`Solo se permiten subir como máximo 5 fotos para el estado de tipo ${tipo}`);
    }
  }

  return prisma.evidencia.create({
    data: {
      citaId,
      tipo,
      fase,
      archivoUrl: filePath,
      archivoTipo
    }
  });
}

module.exports = {
  crear,
  listar,
  listarSinAsignar,
  listarPospuestas,
  obtenerPorId,
  actualizarDuracion,
  subirFoto,
  subirEvidencia,
};

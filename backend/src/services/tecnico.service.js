const prisma = require('../config/database');
const { hashPassword } = require('./auth.service');
const { NotFoundError, ValidationError } = require('../utils/errors');

/**
 * Listar técnicos. Opcionalmente filtrar por estado.
 */
async function listar({ estado, page = 1, limit = 20 }) {
  const where = estado ? { estado } : {};

  const [tecnicos, total] = await Promise.all([
    prisma.tecnico.findMany({
      where,
      select: {
        id: true,
        nombre: true,
        usuario: true,
        estado: true,
        createdAt: true,
        _count: { select: { citas: true } },
      },
      orderBy: { createdAt: 'desc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.tecnico.count({ where }),
  ]);

  return { tecnicos, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Obtener técnico por ID.
 */
async function obtenerPorId(id) {
  const tecnico = await prisma.tecnico.findUnique({
    where: { id: parseInt(id) },
    select: {
      id: true,
      nombre: true,
      usuario: true,
      estado: true,
      createdAt: true,
    },
  });
  if (!tecnico) throw new NotFoundError('Técnico');
  return tecnico;
}

/**
 * Crear un técnico (solo admin).
 */
async function crear(datos) {
  // Verificar usuario único
  const existente = await prisma.tecnico.findUnique({ where: { usuario: datos.usuario } });
  if (existente) throw new ValidationError('Ya existe un técnico con ese usuario');

  const passwordHasheada = await hashPassword(datos.password);

  return prisma.tecnico.create({
    data: {
      nombre: datos.nombre,
      usuario: datos.usuario,
      password: passwordHasheada,
      estado: 'ACTIVO',
    },
    select: {
      id: true,
      nombre: true,
      usuario: true,
      estado: true,
      createdAt: true,
    },
  });
}

/**
 * Editar datos de un técnico.
 */
async function actualizar(id, datos) {
  const tecnico = await prisma.tecnico.findUnique({ where: { id: parseInt(id) } });
  if (!tecnico) throw new NotFoundError('Técnico');

  const updateData = {};
  if (datos.nombre) updateData.nombre = datos.nombre;
  if (datos.usuario) {
    const existente = await prisma.tecnico.findUnique({ where: { usuario: datos.usuario } });
    if (existente && existente.id !== parseInt(id)) {
      throw new ValidationError('Ya existe otro técnico con ese usuario');
    }
    updateData.usuario = datos.usuario;
  }
  if (datos.password) {
    updateData.password = await hashPassword(datos.password);
  }

  return prisma.tecnico.update({
    where: { id: parseInt(id) },
    data: updateData,
    select: {
      id: true,
      nombre: true,
      usuario: true,
      estado: true,
      createdAt: true,
    },
  });
}

/**
 * Cambiar estado del técnico (ACTIVO ↔ INACTIVO).
 * Soft delete: nunca se borra, solo se cambia el estado.
 * Al marcar INACTIVO: no puede loguearse, no aparece en lista de disponibles.
 * Su historial de citas pasadas permanece intacto.
 */
async function cambiarEstado(id, nuevoEstado) {
  const tecnico = await prisma.tecnico.findUnique({ where: { id: parseInt(id) } });
  if (!tecnico) throw new NotFoundError('Técnico');

  if (!['ACTIVO', 'INACTIVO'].includes(nuevoEstado)) {
    throw new ValidationError('Estado inválido. Debe ser ACTIVO o INACTIVO');
  }

  return prisma.tecnico.update({
    where: { id: parseInt(id) },
    data: { estado: nuevoEstado },
    select: {
      id: true,
      nombre: true,
      usuario: true,
      estado: true,
      createdAt: true,
    },
  });
}

/**
 * Obtener el estado actual de un técnico (libre, en curso, pausado)
 * basado en sus citas activas.
 */
async function obtenerDisponibilidad(id) {
  const tecnico = await prisma.tecnico.findUnique({ where: { id: parseInt(id) } });
  if (!tecnico) throw new NotFoundError('Técnico');

  const citasActivas = await prisma.cita.findMany({
    where: {
      tecnicoId: parseInt(id),
      estado: { in: ['ACEPTADA', 'EN_CURSO', 'PAUSADA'] },
    },
    select: { id: true, estado: true },
  });

  const enCurso = citasActivas.find(c => c.estado === 'EN_CURSO');
  const pausada = citasActivas.find(c => c.estado === 'PAUSADA');

  let disponibilidad = 'LIBRE';
  if (enCurso) disponibilidad = 'EN_CURSO';
  else if (pausada) disponibilidad = 'PAUSADO';
  else if (citasActivas.length > 0) disponibilidad = 'OCUPADO';

  return {
    tecnicoId: parseInt(id),
    nombre: tecnico.nombre,
    estado: tecnico.estado,
    disponibilidad,
    citasActivas: citasActivas.length,
    tieneCitaPausada: !!pausada,
  };
}

module.exports = {
  listar,
  obtenerPorId,
  crear,
  actualizar,
  cambiarEstado,
  obtenerDisponibilidad,
};

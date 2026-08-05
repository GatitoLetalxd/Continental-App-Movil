const prisma = require('../config/database');
const { NotFoundError, ValidationError } = require('../utils/errors');

/**
 * Listar servicios. Opcionalmente solo activos.
 */
async function listar({ soloActivos = false, page = 1, limit = 50 }) {
  const where = soloActivos ? { activo: true } : {};

  const [servicios, total] = await Promise.all([
    prisma.servicio.findMany({
      where,
      orderBy: { nombre: 'asc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.servicio.count({ where }),
  ]);

  return { servicios, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Obtener servicio por ID.
 */
async function obtenerPorId(id) {
  const servicio = await prisma.servicio.findUnique({
    where: { id: parseInt(id) },
  });
  if (!servicio) throw new NotFoundError('Servicio');
  return servicio;
}

/**
 * Crear un servicio (admin).
 */
async function crear(datos) {
  return prisma.servicio.create({
    data: {
      nombre: datos.nombre,
      descripcion: datos.descripcion || null,
      duracionEstimadaMinutos: parseInt(datos.duracionEstimadaMinutos),
      precio: parseFloat(datos.precio),
      activo: true,
    },
  });
}

/**
 * Editar un servicio.
 */
async function actualizar(id, datos) {
  const servicio = await prisma.servicio.findUnique({ where: { id: parseInt(id) } });
  if (!servicio) throw new NotFoundError('Servicio');

  const updateData = {};
  if (datos.nombre) updateData.nombre = datos.nombre;
  if (datos.descripcion !== undefined) updateData.descripcion = datos.descripcion;
  if (datos.duracionEstimadaMinutos) updateData.duracionEstimadaMinutos = parseInt(datos.duracionEstimadaMinutos);
  if (datos.precio !== undefined) updateData.precio = parseFloat(datos.precio);

  return prisma.servicio.update({
    where: { id: parseInt(id) },
    data: updateData,
  });
}

/**
 * Activar/desactivar servicio. No se borra, solo se cambia el flag.
 */
async function cambiarEstado(id, activo) {
  const servicio = await prisma.servicio.findUnique({ where: { id: parseInt(id) } });
  if (!servicio) throw new NotFoundError('Servicio');

  return prisma.servicio.update({
    where: { id: parseInt(id) },
    data: { activo: Boolean(activo) },
  });
}

module.exports = {
  listar,
  obtenerPorId,
  crear,
  actualizar,
  cambiarEstado,
};

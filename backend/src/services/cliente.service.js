const prisma = require('../config/database');
const { hashPassword } = require('./auth.service');
const { NotFoundError, ValidationError } = require('../utils/errors');
const { MAX_CONTACTOS_CLIENTE } = require('../config/constants');

/**
 * Listar todos los clientes con sus vehículos.
 */
async function listar({ busqueda, page = 1, limit = 20 }) {
  const where = busqueda
    ? { nombre: { contains: busqueda, mode: 'insensitive' } }
    : {};

  const [clientes, total] = await Promise.all([
    prisma.cliente.findMany({
      where,
      include: { vehiculos: true },
      orderBy: { createdAt: 'desc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.cliente.count({ where }),
  ]);

  return { clientes, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Obtener un cliente por ID con sus vehículos.
 */
async function obtenerPorId(id) {
  const cliente = await prisma.cliente.findUnique({
    where: { id: parseInt(id) },
    include: { vehiculos: true },
  });
  if (!cliente) throw new NotFoundError('Cliente');
  return cliente;
}

/**
 * Actualizar perfil del cliente.
 */
async function actualizar(id, datos) {
  const cliente = await prisma.cliente.findUnique({ where: { id: parseInt(id) } });
  if (!cliente) throw new NotFoundError('Cliente');

  const updateData = {};

  if (datos.nombre) updateData.nombre = datos.nombre;

  if (datos.contactos) {
    if (datos.contactos.length > MAX_CONTACTOS_CLIENTE) {
      throw new ValidationError(`Máximo ${MAX_CONTACTOS_CLIENTE} contactos permitidos`);
    }
    updateData.contactos = datos.contactos;
  }

  if (datos.pin) {
    if (!/^\d{4,8}$/.test(datos.pin)) {
      throw new ValidationError('El PIN debe ser numérico de 4 a 8 dígitos');
    }
    updateData.pin = await hashPassword(datos.pin);
  }

  return prisma.cliente.update({
    where: { id: parseInt(id) },
    data: updateData,
    include: { vehiculos: true },
  });
}

module.exports = {
  listar,
  obtenerPorId,
  actualizar,
};

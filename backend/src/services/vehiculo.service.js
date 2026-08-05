const prisma = require('../config/database');
const { NotFoundError, ValidationError } = require('../utils/errors');

/**
 * Listar vehículos de un cliente.
 */
async function listarPorCliente(clienteId) {
  return prisma.vehiculo.findMany({
    where: { clienteId: parseInt(clienteId) },
    orderBy: { id: 'desc' },
  });
}

/**
 * Obtener un vehículo por ID.
 */
async function obtenerPorId(id) {
  const vehiculo = await prisma.vehiculo.findUnique({
    where: { id: parseInt(id) },
    include: { cliente: { select: { id: true, nombre: true } } },
  });
  if (!vehiculo) throw new NotFoundError('Vehículo');
  return vehiculo;
}

/**
 * Crear un vehículo para un cliente.
 */
async function crear(clienteId, datos) {
  // Verificar que el cliente existe
  const cliente = await prisma.cliente.findUnique({ where: { id: parseInt(clienteId) } });
  if (!cliente) throw new NotFoundError('Cliente');

  return prisma.vehiculo.create({
    data: {
      marca: datos.marca,
      modelo: datos.modelo,
      anio: parseInt(datos.anio),
      placa: datos.placa || '',
      notas: datos.notas || null,
      clienteId: parseInt(clienteId),
    },
  });
}

/**
 * Actualizar un vehículo.
 */
async function actualizar(id, datos) {
  const vehiculo = await prisma.vehiculo.findUnique({ where: { id: parseInt(id) } });
  if (!vehiculo) throw new NotFoundError('Vehículo');

  const updateData = {};
  if (datos.marca) updateData.marca = datos.marca;
  if (datos.modelo) updateData.modelo = datos.modelo;
  if (datos.anio) updateData.anio = parseInt(datos.anio);
  if (datos.placa) updateData.placa = datos.placa;
  if (datos.notas !== undefined) updateData.notas = datos.notas;

  return prisma.vehiculo.update({
    where: { id: parseInt(id) },
    data: updateData,
  });
}

/**
 * Eliminar un vehículo (solo si no tiene citas asociadas).
 */
async function eliminar(id) {
  const vehiculo = await prisma.vehiculo.findUnique({
    where: { id: parseInt(id) },
    include: { citas: { take: 1 } },
  });
  if (!vehiculo) throw new NotFoundError('Vehículo');

  if (vehiculo.citas.length > 0) {
    throw new ValidationError('No se puede eliminar un vehículo con citas asociadas');
  }

  return prisma.vehiculo.delete({ where: { id: parseInt(id) } });
}

module.exports = {
  listarPorCliente,
  obtenerPorId,
  crear,
  actualizar,
  eliminar,
};

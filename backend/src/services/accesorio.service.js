const prisma = require('../config/database');
const { NotFoundError, ValidationError } = require('../utils/errors');

/**
 * Listar accesorios con indicador de stock bajo.
 */
async function listar({ page = 1, limit = 50 }) {
  const [accesorios, total] = await Promise.all([
    prisma.accesorio.findMany({
      orderBy: { nombre: 'asc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.accesorio.count(),
  ]);

  // Agregar indicador de stock bajo
  const conAlerta = accesorios.map(acc => ({
    ...acc,
    stockBajo: acc.stock <= acc.stockMinimoAlerta,
  }));

  return { accesorios: conAlerta, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Obtener accesorio por ID.
 */
async function obtenerPorId(id) {
  const accesorio = await prisma.accesorio.findUnique({
    where: { id: parseInt(id) },
  });
  if (!accesorio) throw new NotFoundError('Accesorio');
  return {
    ...accesorio,
    stockBajo: accesorio.stock <= accesorio.stockMinimoAlerta,
  };
}

/**
 * Crear un accesorio (admin).
 */
async function crear(datos) {
  return prisma.accesorio.create({
    data: {
      nombre: datos.nombre,
      descripcion: datos.descripcion || null,
      precio: parseFloat(datos.precio),
      stock: parseInt(datos.stock) || 0,
      stockMinimoAlerta: parseInt(datos.stockMinimoAlerta) || 5,
    },
  });
}

/**
 * Editar un accesorio (incluyendo ajustar stock).
 */
async function actualizar(id, datos) {
  const accesorio = await prisma.accesorio.findUnique({ where: { id: parseInt(id) } });
  if (!accesorio) throw new NotFoundError('Accesorio');

  const updateData = {};
  if (datos.nombre) updateData.nombre = datos.nombre;
  if (datos.descripcion !== undefined) updateData.descripcion = datos.descripcion;
  if (datos.precio !== undefined) updateData.precio = parseFloat(datos.precio);
  if (datos.stock !== undefined) updateData.stock = parseInt(datos.stock);
  if (datos.stockMinimoAlerta !== undefined) updateData.stockMinimoAlerta = parseInt(datos.stockMinimoAlerta);

  return prisma.accesorio.update({
    where: { id: parseInt(id) },
    data: updateData,
  });
}

/**
 * Obtener accesorios con stock bajo (umbral configurable por accesorio).
 */
async function obtenerAlertasStock() {
  const accesorios = await prisma.accesorio.findMany({
    orderBy: { stock: 'asc' },
  });

  return accesorios
    .filter(acc => acc.stock <= acc.stockMinimoAlerta)
    .map(acc => ({
      ...acc,
      stockBajo: true,
      deficit: acc.stockMinimoAlerta - acc.stock,
    }));
}

/**
 * Reservar stock de accesorios para una cita.
 * Descuenta del stock disponible.
 */
async function reservarStock(accesorioId, cantidad) {
  const accesorio = await prisma.accesorio.findUnique({ where: { id: parseInt(accesorioId) } });
  if (!accesorio) throw new NotFoundError('Accesorio');

  if (accesorio.stock < cantidad) {
    throw new ValidationError(`Stock insuficiente para "${accesorio.nombre}". Disponible: ${accesorio.stock}, solicitado: ${cantidad}`);
  }

  return prisma.accesorio.update({
    where: { id: parseInt(accesorioId) },
    data: { stock: accesorio.stock - cantidad },
  });
}

/**
 * Liberar stock de accesorios (al cancelar una cita).
 * Devuelve las cantidades al stock.
 */
async function liberarStock(accesorioId, cantidad) {
  return prisma.accesorio.update({
    where: { id: parseInt(accesorioId) },
    data: { stock: { increment: cantidad } },
  });
}

module.exports = {
  listar,
  obtenerPorId,
  crear,
  actualizar,
  obtenerAlertasStock,
  reservarStock,
  liberarStock,
};

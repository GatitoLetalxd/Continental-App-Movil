const bcrypt = require('bcryptjs');
const jwt = require('jsonwebtoken');
const prisma = require('../config/database');
const { UnauthorizedError, ValidationError, ConflictError } = require('../utils/errors');
const { ROLES, MAX_CONTACTOS_CLIENTE } = require('../config/constants');

const SALT_ROUNDS = 10;

/**
 * Login unificado: busca en Admin → Técnico → Cliente.
 * Retorna JWT con { id, rol, nombre }.
 */
async function login(usuario, password) {
  // 1. Buscar en Admin
  const admin = await prisma.admin.findUnique({ where: { usuario } });
  if (admin) {
    const valid = await bcrypt.compare(password, admin.password);
    if (!valid) throw new UnauthorizedError('Credenciales incorrectas');
    return generarToken({ id: admin.id, rol: ROLES.ADMIN, nombre: admin.nombre });
  }

  // 2. Buscar en Técnico
  const tecnico = await prisma.tecnico.findUnique({ where: { usuario } });
  if (tecnico) {
    if (tecnico.estado === 'INACTIVO') {
      throw new UnauthorizedError('Tu cuenta de técnico está inactiva. Contacta al administrador.');
    }
    const valid = await bcrypt.compare(password, tecnico.password);
    if (!valid) throw new UnauthorizedError('Credenciales incorrectas');
    return generarToken({ id: tecnico.id, rol: ROLES.TECNICO, nombre: tecnico.nombre });
  }

  // 3. Buscar en Cliente: primero por número de contacto (único), luego por nombre
  // Si se usa nombre y hay más de un cliente con ese nombre, usar número de contacto
  let clientesEncontrados = [];

  // Intentar match de número de teléfono exacto usando el filtro de array nativo de Postgres
  const clientePorTelefono = await prisma.cliente.findMany({
    where: { contactos: { has: usuario } },
  });

  if (clientePorTelefono.length > 0) {
    clientesEncontrados = clientePorTelefono;
  } else {
    // Fallback: buscar por nombre
    clientesEncontrados = await prisma.cliente.findMany({ where: { nombre: usuario } });
    if (clientesEncontrados.length > 1) {
      throw new UnauthorizedError(
        'Hay más de un cliente con ese nombre. Por favor, ingrese su número de teléfono como usuario para acceder de forma segura.'
      );
    }
  }

  for (const cliente of clientesEncontrados) {
    const valid = await bcrypt.compare(password, cliente.pin);
    if (valid) {
      return generarToken({ id: cliente.id, rol: ROLES.CLIENTE, nombre: cliente.nombre });
    }
  }

  throw new UnauthorizedError('Credenciales incorrectas');
}

/**
 * Registro de cliente: nombre + PIN numérico + contactos.
 */
async function registrarCliente({ nombre, pin, contactos }) {
  if (!nombre || !pin) {
    throw new ValidationError('Nombre y PIN son obligatorios');
  }

  if (!/^\d{4,8}$/.test(pin)) {
    throw new ValidationError('El PIN debe ser numérico de 4 a 8 dígitos');
  }

  if (contactos && contactos.length > MAX_CONTACTOS_CLIENTE) {
    throw new ValidationError(`Máximo ${MAX_CONTACTOS_CLIENTE} contactos permitidos`);
  }

  const pinHasheado = await bcrypt.hash(pin, SALT_ROUNDS);

  const cliente = await prisma.cliente.create({
    data: {
      nombre,
      pin: pinHasheado,
      contactos: contactos || [],
    },
  });

  return generarToken({ id: cliente.id, rol: ROLES.CLIENTE, nombre: cliente.nombre });
}

/**
 * Obtener datos del usuario autenticado según su rol.
 */
async function obtenerPerfil(userId, rol) {
  switch (rol) {
    case ROLES.ADMIN:
      return prisma.admin.findUnique({
        where: { id: userId },
        select: { id: true, usuario: true, nombre: true, createdAt: true },
      });
    case ROLES.TECNICO:
      return prisma.tecnico.findUnique({
        where: { id: userId },
        select: { id: true, nombre: true, usuario: true, estado: true, createdAt: true },
      });
    case ROLES.CLIENTE:
      // No se puede usar select + include simultáneamente en Prisma — usar solo select con vehiculos incluido
      return prisma.cliente.findUnique({
        where: { id: userId },
        select: {
          id: true,
          nombre: true,
          contactos: true,
          createdAt: true,
          vehiculos: {
            select: { id: true, marca: true, modelo: true, anio: true, placa: true, notas: true },
          },
        },
      });
    default:
      throw new UnauthorizedError('Rol desconocido');
  }
}

/**
 * Genera un JWT con payload { id, rol, nombre }.
 */
function generarToken(payload) {
  const token = jwt.sign(payload, process.env.JWT_SECRET, {
    expiresIn: process.env.JWT_EXPIRES_IN || '7d',
  });
  return { token, user: payload };
}

/**
 * Hashea una contraseña/PIN.
 */
async function hashPassword(password) {
  return bcrypt.hash(password, SALT_ROUNDS);
}

module.exports = {
  login,
  registrarCliente,
  obtenerPerfil,
  hashPassword,
};

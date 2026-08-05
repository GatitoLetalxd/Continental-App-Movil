const jwt = require('jsonwebtoken');
const { UnauthorizedError, ForbiddenError } = require('../utils/errors');

/**
 * Middleware: Verifica el token JWT en el header Authorization.
 * Agrega req.user = { id, rol, nombre } al request.
 */
function verifyToken(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader || !authHeader.startsWith('Bearer ')) {
    return next(new UnauthorizedError('Token no proporcionado'));
  }

  const token = authHeader.split(' ')[1];

  try {
    const decoded = jwt.verify(token, process.env.JWT_SECRET);
    req.user = {
      id: decoded.id,
      rol: decoded.rol,
      nombre: decoded.nombre,
    };
    next();
  } catch (error) {
    return next(new UnauthorizedError('Token inválido o expirado'));
  }
}

/**
 * Factory: Crea middleware que verifica si el rol del usuario está permitido.
 * @param  {...string} rolesPermitidos - Roles que pueden acceder al recurso.
 * 
 * Uso: requireRole('ADMIN', 'TECNICO')
 */
function requireRole(...rolesPermitidos) {
  return (req, res, next) => {
    if (!req.user) {
      return next(new UnauthorizedError('Debe autenticarse primero'));
    }
    if (!rolesPermitidos.includes(req.user.rol)) {
      return next(new ForbiddenError(`Acceso restringido. Roles permitidos: ${rolesPermitidos.join(', ')}`));
    }
    next();
  };
}

module.exports = {
  verifyToken,
  requireRole,
};

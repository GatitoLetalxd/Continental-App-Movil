/**
 * Clases de error custom para respuestas HTTP consistentes.
 */

class AppError extends Error {
  constructor(message, statusCode = 500) {
    super(message);
    this.statusCode = statusCode;
    this.isOperational = true;
    Error.captureStackTrace(this, this.constructor);
  }
}

class NotFoundError extends AppError {
  constructor(entity = 'Recurso') {
    super(`${entity} no encontrado`, 404);
  }
}

class ValidationError extends AppError {
  constructor(message = 'Datos inválidos') {
    super(message, 400);
  }
}

class UnauthorizedError extends AppError {
  constructor(message = 'No autorizado') {
    super(message, 401);
  }
}

class ForbiddenError extends AppError {
  constructor(message = 'Acceso denegado') {
    super(message, 403);
  }
}

class ConflictError extends AppError {
  constructor(message = 'Conflicto con el estado actual') {
    super(message, 409);
  }
}

module.exports = {
  AppError,
  NotFoundError,
  ValidationError,
  UnauthorizedError,
  ForbiddenError,
  ConflictError,
};

const { ValidationError } = require('../utils/errors');

/**
 * Factory: Crea middleware de validación usando un schema de Joi.
 * @param {object} schema - Schema de Joi para validar req.body.
 * @param {string} source - Parte del request a validar: 'body', 'query', 'params'.
 */
function validate(schema, source = 'body') {
  return (req, res, next) => {
    const { error, value } = schema.validate(req[source], {
      abortEarly: false,
      stripUnknown: true,
    });

    if (error) {
      const mensajes = error.details.map(d => d.message).join('. ');
      return next(new ValidationError(mensajes));
    }

    // Reemplazar con valores limpiados/transformados
    req[source] = value;
    next();
  };
}

module.exports = { validate };

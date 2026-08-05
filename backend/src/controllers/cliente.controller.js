const clienteService = require('../services/cliente.service');

async function listar(req, res, next) {
  try {
    const data = await clienteService.listar(req.query);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await clienteService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function actualizar(req, res, next) {
  try {
    // Un cliente solo puede editar su propio perfil
    const id = req.user.rol === 'CLIENTE' ? req.user.id : req.params.id;
    const data = await clienteService.actualizar(id, req.body);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

module.exports = { listar, obtenerPorId, actualizar };

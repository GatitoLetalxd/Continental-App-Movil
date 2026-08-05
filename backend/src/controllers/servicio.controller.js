const servicioService = require('../services/servicio.service');

async function listar(req, res, next) {
  try {
    const data = await servicioService.listar(req.query);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await servicioService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function crear(req, res, next) {
  try {
    const data = await servicioService.crear(req.body);
    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function actualizar(req, res, next) {
  try {
    const data = await servicioService.actualizar(req.params.id, req.body);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function cambiarEstado(req, res, next) {
  try {
    const data = await servicioService.cambiarEstado(req.params.id, req.body.activo);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

module.exports = { listar, obtenerPorId, crear, actualizar, cambiarEstado };

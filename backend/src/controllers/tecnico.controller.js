const tecnicoService = require('../services/tecnico.service');

async function listar(req, res, next) {
  try {
    const data = await tecnicoService.listar(req.query);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await tecnicoService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function crear(req, res, next) {
  try {
    const data = await tecnicoService.crear(req.body);
    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function actualizar(req, res, next) {
  try {
    const data = await tecnicoService.actualizar(req.params.id, req.body);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function cambiarEstado(req, res, next) {
  try {
    const data = await tecnicoService.cambiarEstado(req.params.id, req.body.estado);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerDisponibilidad(req, res, next) {
  try {
    const data = await tecnicoService.obtenerDisponibilidad(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

module.exports = { listar, obtenerPorId, crear, actualizar, cambiarEstado, obtenerDisponibilidad };

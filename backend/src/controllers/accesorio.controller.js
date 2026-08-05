const accesorioService = require('../services/accesorio.service');

async function listar(req, res, next) {
  try {
    const data = await accesorioService.listar(req.query);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await accesorioService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function crear(req, res, next) {
  try {
    const data = await accesorioService.crear(req.body);
    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function actualizar(req, res, next) {
  try {
    const data = await accesorioService.actualizar(req.params.id, req.body);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function alertasStock(req, res, next) {
  try {
    const data = await accesorioService.obtenerAlertasStock();
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

module.exports = { listar, obtenerPorId, crear, actualizar, alertasStock };

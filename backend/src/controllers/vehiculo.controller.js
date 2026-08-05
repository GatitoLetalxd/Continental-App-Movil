const vehiculoService = require('../services/vehiculo.service');

async function listarPorCliente(req, res, next) {
  try {
    const clienteId = req.user.rol === 'CLIENTE' ? req.user.id : req.params.clienteId;
    const data = await vehiculoService.listarPorCliente(clienteId);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await vehiculoService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function crear(req, res, next) {
  try {
    const clienteId = req.user.rol === 'CLIENTE' ? req.user.id : req.body.clienteId;
    const data = await vehiculoService.crear(clienteId, req.body);
    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function actualizar(req, res, next) {
  try {
    const data = await vehiculoService.actualizar(req.params.id, req.body);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function eliminar(req, res, next) {
  try {
    await vehiculoService.eliminar(req.params.id);
    res.json({ success: true, message: 'Vehículo eliminado' });
  } catch (error) {
    next(error);
  }
}

module.exports = { listarPorCliente, obtenerPorId, crear, actualizar, eliminar };

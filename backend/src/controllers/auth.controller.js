const authService = require('../services/auth.service');

async function login(req, res, next) {
  try {
    const { usuario, password } = req.body;
    const result = await authService.login(usuario, password);
    res.json({ success: true, data: result });
  } catch (error) {
    next(error);
  }
}

async function registrar(req, res, next) {
  try {
    const result = await authService.registrarCliente(req.body);
    res.status(201).json({ success: true, data: result });
  } catch (error) {
    next(error);
  }
}

async function perfil(req, res, next) {
  try {
    const data = await authService.obtenerPerfil(req.user.id, req.user.rol);
    res.json({ success: true, data: { ...data, rol: req.user.rol } });
  } catch (error) {
    next(error);
  }
}

module.exports = { login, registrar, perfil };

const router = require('express').Router();
const authController = require('../controllers/auth.controller');
const { verifyToken } = require('../middleware/auth');

// Login unificado (los 3 roles)
router.post('/login', authController.login);

// Registro de clientes
router.post('/register', authController.registrar);

// Perfil del usuario autenticado
router.get('/me', verifyToken, authController.perfil);

module.exports = router;

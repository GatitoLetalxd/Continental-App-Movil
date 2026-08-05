const router = require('express').Router();
const clienteController = require('../controllers/cliente.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);

// Admin y Técnico pueden listar y ver clientes
router.get('/', requireRole('ADMIN', 'TECNICO'), clienteController.listar);
router.get('/:id', requireRole('ADMIN', 'CLIENTE'), clienteController.obtenerPorId);

// Cliente puede editar su perfil, admin puede editar cualquiera
router.put('/:id', requireRole('ADMIN', 'CLIENTE'), clienteController.actualizar);

module.exports = router;

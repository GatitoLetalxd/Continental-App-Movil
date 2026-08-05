const router = require('express').Router();
const tecnicoController = require('../controllers/tecnico.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);

// Todos los autenticados pueden ver la lista de técnicos activos
router.get('/', tecnicoController.listar);
router.get('/:id', tecnicoController.obtenerPorId);
router.get('/:id/disponibilidad', tecnicoController.obtenerDisponibilidad);

// Solo admin puede crear, editar y cambiar estado
router.post('/', requireRole('ADMIN'), tecnicoController.crear);
router.put('/:id', requireRole('ADMIN'), tecnicoController.actualizar);
router.patch('/:id/estado', requireRole('ADMIN'), tecnicoController.cambiarEstado);

module.exports = router;

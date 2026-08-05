const router = require('express').Router();
const servicioController = require('../controllers/servicio.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);

// Todos pueden ver servicios
router.get('/', servicioController.listar);
router.get('/:id', servicioController.obtenerPorId);

// Solo admin puede crear, editar y cambiar estado
router.post('/', requireRole('ADMIN'), servicioController.crear);
router.put('/:id', requireRole('ADMIN'), servicioController.actualizar);
router.patch('/:id/estado', requireRole('ADMIN'), servicioController.cambiarEstado);

module.exports = router;

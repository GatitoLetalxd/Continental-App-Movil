const router = require('express').Router();
const vehiculoController = require('../controllers/vehiculo.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);

// Listar vehículos de un cliente
router.get('/cliente/:clienteId', requireRole('ADMIN', 'TECNICO'), vehiculoController.listarPorCliente);
router.get('/mis-vehiculos', requireRole('CLIENTE'), vehiculoController.listarPorCliente);

// CRUD
router.get('/:id', requireRole('ADMIN', 'TECNICO', 'CLIENTE'), vehiculoController.obtenerPorId);
router.post('/', requireRole('ADMIN', 'TECNICO', 'CLIENTE'), vehiculoController.crear);
router.put('/:id', requireRole('ADMIN', 'CLIENTE'), vehiculoController.actualizar);
router.delete('/:id', requireRole('ADMIN', 'CLIENTE'), vehiculoController.eliminar);

module.exports = router;

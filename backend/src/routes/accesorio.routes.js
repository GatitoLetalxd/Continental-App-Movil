const router = require('express').Router();
const accesorioController = require('../controllers/accesorio.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);

// Todos los autenticados pueden ver accesorios
router.get('/', accesorioController.listar);
router.get('/alertas', requireRole('ADMIN'), accesorioController.alertasStock);
router.get('/:id', accesorioController.obtenerPorId);

// Solo admin puede crear y editar
router.post('/', requireRole('ADMIN'), accesorioController.crear);
router.put('/:id', requireRole('ADMIN'), accesorioController.actualizar);

module.exports = router;

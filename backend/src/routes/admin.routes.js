const router = require('express').Router();
const adminController = require('../controllers/admin.controller');
const { verifyToken, requireRole } = require('../middleware/auth');

router.use(verifyToken);
router.use(requireRole('ADMIN'));

// Dashboard operativo
router.get('/dashboard', adminController.dashboard);

// Reportes
router.get('/reportes/productividad', adminController.reporteProductividad);
router.get('/reportes/servicios', adminController.reporteServicios);

module.exports = router;

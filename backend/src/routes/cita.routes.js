const router = require('express').Router();
const citaController = require('../controllers/cita.controller');
const { verifyToken, requireRole } = require('../middleware/auth');
const { uploadEvidencia, uploadEvidenciaGeneral } = require('../middleware/upload');

router.use(verifyToken);

// ─── CRUD ────────────────────────────────────────────────
router.post('/', requireRole('ADMIN', 'TECNICO', 'CLIENTE'), citaController.crear);
router.get('/', citaController.listar);
router.get('/sin-asignar', requireRole('ADMIN', 'TECNICO'), citaController.listarSinAsignar);
router.get('/pospuestas', requireRole('ADMIN', 'TECNICO'), citaController.listarPospuestas);
router.get('/fases-defecto', citaController.obtenerFasesPorDefecto);
router.get('/:id', citaController.obtenerPorId);

// ─── Transiciones de estado ──────────────────────────────
router.patch('/:id/aceptar', requireRole('TECNICO'), citaController.aceptar);
router.patch('/:id/iniciar', requireRole('TECNICO'), citaController.iniciar);
router.patch('/:id/pausar', requireRole('TECNICO'), citaController.pausar);
router.patch('/:id/reanudar', requireRole('TECNICO'), citaController.reanudar);
router.patch('/:id/finalizar', requireRole('TECNICO'), citaController.finalizar);
router.patch('/:id/duracion', requireRole('TECNICO', 'ADMIN'), citaController.actualizarDuracion);
router.patch('/:id/entregar', requireRole('ADMIN', 'TECNICO'), citaController.entregar);
router.patch('/:id/posponer', requireRole('TECNICO'), citaController.posponer);
router.patch('/:id/cancelar', requireRole('ADMIN', 'TECNICO'), citaController.cancelar);
router.patch('/:id/reasignar', requireRole('ADMIN', 'TECNICO'), citaController.reasignar);

// ─── Fotos de evidencia ──────────────────────────────────
router.post('/:id/foto/:tipo', requireRole('TECNICO'), uploadEvidencia.single('foto'), citaController.subirFoto);
router.post('/:id/evidencia', requireRole('TECNICO'), uploadEvidenciaGeneral.single('archivo'), citaController.subirEvidencia);

// ─── Fases ───────────────────────────────────────────────
router.patch('/:id/fase', requireRole('TECNICO'), citaController.cambiarFase);
router.patch('/:id/fase/avanzar', requireRole('TECNICO'), citaController.avanzarFase);
router.patch('/:id/fase/resetear', requireRole('TECNICO'), citaController.resetearFase);
router.get('/:id/fases', citaController.obtenerHistorialFases);

// ─── WhatsApp ────────────────────────────────────────────
router.get('/:id/whatsapp/:tipo', requireRole('ADMIN', 'TECNICO'), citaController.obtenerLinkWhatsApp);

module.exports = router;

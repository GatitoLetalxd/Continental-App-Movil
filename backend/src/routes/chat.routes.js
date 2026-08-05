const router = require('express').Router();
const chatController = require('../controllers/chat.controller');
const { verifyToken, requireRole } = require('../middleware/auth');
const { uploadChat } = require('../middleware/upload');

router.use(verifyToken);

// Solo cliente y técnico tienen acceso al chat (NO admin en v1)
router.get('/:citaId/mensajes', requireRole('CLIENTE', 'TECNICO'), chatController.listarMensajes);
router.post('/:citaId/mensaje', requireRole('CLIENTE', 'TECNICO'), chatController.enviarMensaje);
router.post('/:citaId/multimedia', requireRole('CLIENTE', 'TECNICO'), uploadChat.single('archivo'), chatController.enviarMultimedia);

// Descarga de archivos multimedia del chat
router.get('/:citaId/archivo/*filePath', requireRole('CLIENTE', 'TECNICO'), chatController.descargarArchivo);

module.exports = router;

const chatService = require('../services/chat.service');
const path = require('path');

async function enviarMensaje(req, res, next) {
  try {
    const autorTipo = req.user.rol === 'CLIENTE' ? 'CLIENTE' : 'TECNICO';
    const data = await chatService.enviarMensajeTexto(
      req.params.citaId,
      autorTipo,
      req.user.id,
      req.body.contenido
    );

    // Emitir evento Socket.io
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.citaId}`).emit('chat:nuevoMensaje', {
        citaId: parseInt(req.params.citaId),
        mensaje: data,
      });
    }

    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function enviarMultimedia(req, res, next) {
  try {
    if (!req.file) {
      return res.status(400).json({ success: false, message: 'Debe seleccionar un archivo' });
    }

    const autorTipo = req.user.rol === 'CLIENTE' ? 'CLIENTE' : 'TECNICO';
    const esVideo = /video/.test(req.file.mimetype);
    const tipo = esVideo ? 'VIDEO' : 'IMAGEN';

    const data = await chatService.enviarMensajeMultimedia(
      req.params.citaId,
      autorTipo,
      req.user.id,
      tipo,
      req.file.path,
      req.file.size
    );

    // Emitir evento Socket.io
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.citaId}`).emit('chat:nuevoMensaje', {
        citaId: parseInt(req.params.citaId),
        mensaje: data,
      });
    }

    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function listarMensajes(req, res, next) {
  try {
    const autorTipo = req.user.rol === 'CLIENTE' ? 'CLIENTE' : 'TECNICO';
    const data = await chatService.listarMensajes(
      req.params.citaId,
      autorTipo,
      req.user.id,
      req.query
    );
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function descargarArchivo(req, res, next) {
  try {
    const autorTipo = req.user.rol === 'CLIENTE' ? 'CLIENTE' : 'TECNICO';
    await chatService.verificarAccesoChat(req.params.citaId, autorTipo, req.user.id);

    const rawPath = req.params.filePath || req.query.path;
    if (!rawPath) {
      return res.status(400).json({ success: false, message: 'Ruta de archivo no proporcionada' });
    }

    const uploadsDir = path.resolve(process.env.UPLOAD_DIR || 'uploads');
    const filePath = path.resolve(rawPath);

    // Seguridad: Prevenir Path Traversal fuera del directorio de uploads
    if (!filePath.startsWith(uploadsDir)) {
      return res.status(403).json({ success: false, message: 'Acceso no permitido fuera del directorio de archivos' });
    }

    res.download(filePath);
  } catch (error) {
    next(error);
  }
}

module.exports = { enviarMensaje, enviarMultimedia, listarMensajes, descargarArchivo };

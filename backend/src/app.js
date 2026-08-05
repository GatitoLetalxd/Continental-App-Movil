require('dotenv').config();
process.env.TZ = 'America/Lima';
const express = require('express');
const cors = require('cors');
const http = require('http');
const path = require('path');
const cron = require('node-cron');
const { setupSocketIO } = require('./socket/index');

// ─── App Express ─────────────────────────────────────────
const app = express();
const server = http.createServer(app);

// ─── Socket.io ───────────────────────────────────────────
const io = setupSocketIO(server);
app.set('io', io); // Disponible en req.app.get('io') para controladores

// ─── Middleware global ───────────────────────────────────
app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

// Servir archivos estáticos de uploads
app.use('/uploads', express.static(path.join(__dirname, '..', 'uploads')));

// ─── Rutas ───────────────────────────────────────────────
app.use('/api/auth', require('./routes/auth.routes'));
app.use('/api/clientes', require('./routes/cliente.routes'));
app.use('/api/vehiculos', require('./routes/vehiculo.routes'));
app.use('/api/tecnicos', require('./routes/tecnico.routes'));
app.use('/api/servicios', require('./routes/servicio.routes'));
app.use('/api/accesorios', require('./routes/accesorio.routes'));
app.use('/api/citas', require('./routes/cita.routes'));
app.use('/api/chat', require('./routes/chat.routes'));
app.use('/api/admin', require('./routes/admin.routes'));

// ─── Health check ────────────────────────────────────────
app.get('/api/health', (req, res) => {
  res.json({
    status: 'ok',
    timestamp: new Date().toISOString(),
    uptime: process.uptime(),
  });
});

// ─── Error handler global ────────────────────────────────
app.use((err, req, res, next) => {
  // Error de Multer
  if (err.code === 'LIMIT_FILE_SIZE') {
    return res.status(400).json({
      success: false,
      message: 'El archivo excede el tamaño máximo permitido',
    });
  }

  const statusCode = err.statusCode || 500;
  const message = err.isOperational ? err.message : 'Error interno del servidor';

  if (!err.isOperational) {
    console.error('[ERROR]', err);
  }

  res.status(statusCode).json({
    success: false,
    message,
    ...(process.env.NODE_ENV === 'development' && { stack: err.stack }),
  });
});

// ─── Cron Jobs ───────────────────────────────────────────
const { ejecutarBackup } = require('./jobs/backupDb');
const { ejecutar: limpiarFotos } = require('./jobs/cleanupPhotos');
const { ejecutar: limpiarChat } = require('./jobs/cleanupChat');

// Backup diario a las 2:00 AM
cron.schedule('0 2 * * *', () => {
  console.log('[Cron] Ejecutando backup de base de datos...');
  ejecutarBackup().catch(err => console.error('[Cron] Error en backup:', err.message));
});

// Limpieza de fotos a las 3:00 AM
cron.schedule('0 3 * * *', () => {
  console.log('[Cron] Ejecutando limpieza de fotos de evidencia...');
  limpiarFotos();
});

// Limpieza de chat multimedia a las 3:30 AM
cron.schedule('30 3 * * *', () => {
  console.log('[Cron] Ejecutando limpieza de multimedia de chat...');
  limpiarChat();
});

// ─── Iniciar servidor ────────────────────────────────────
const PORT = process.env.PORT || 3000;

server.listen(PORT, () => {
  console.log(`
╔══════════════════════════════════════════════════════╗
║   🔧 Continental Taller - Backend API               ║
║   📡 Servidor:  http://localhost:${PORT}               ║
║   🔌 Socket.io: ws://localhost:${PORT}                 ║
║   🌍 Dominio:   ${process.env.DOMAIN || 'N/A'}              ║
║   📦 Entorno:   ${process.env.NODE_ENV || 'development'}                  ║
╚══════════════════════════════════════════════════════╝
  `);
});

module.exports = { app, server, io };

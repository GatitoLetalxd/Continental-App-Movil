const { Server } = require('socket.io');
const jwt = require('jsonwebtoken');

/**
 * Configura Socket.io con autenticación JWT y rooms.
 * 
 * Rooms:
 * - cita:<citaId>  → Para updates de una cita específica (estado, fase, chat)
 * - tecnico:<id>   → Para notificaciones al técnico
 * - admin          → Para notificaciones al admin
 */
function setupSocketIO(httpServer) {
  const io = new Server(httpServer, {
    cors: {
      origin: '*', // En producción, restringir al dominio
      methods: ['GET', 'POST'],
    },
  });

  // Middleware de autenticación por JWT en el handshake
  io.use((socket, next) => {
    const token = socket.handshake.auth.token || socket.handshake.query.token;
    if (!token) {
      return next(new Error('Token no proporcionado'));
    }

    try {
      const decoded = jwt.verify(token, process.env.JWT_SECRET);
      socket.user = decoded;
      next();
    } catch (err) {
      return next(new Error('Token inválido'));
    }
  });

  io.on('connection', (socket) => {
    const { id, rol, nombre } = socket.user;
    console.log(`[Socket.io] Conectado: ${nombre} (${rol}) - Socket ID: ${socket.id}`);

    // Unir a rooms según el rol
    if (rol === 'ADMIN') {
      socket.join('admin');
    } else if (rol === 'TECNICO') {
      socket.join(`tecnico:${id}`);
    }

    // El cliente/técnico se une a la room de una cita específica
    socket.on('cita:join', (citaId) => {
      socket.join(`cita:${citaId}`);
      console.log(`[Socket.io] ${nombre} se unió a cita:${citaId}`);
    });

    socket.on('cita:leave', (citaId) => {
      socket.leave(`cita:${citaId}`);
      console.log(`[Socket.io] ${nombre} salió de cita:${citaId}`);
    });

    // Chat en tiempo real vía Socket.io (alternativa a REST)
    socket.on('chat:enviarMensaje', async (data) => {
      try {
        const chatService = require('../services/chat.service');
        const autorTipo = rol === 'CLIENTE' ? 'CLIENTE' : 'TECNICO';
        
        const mensaje = await chatService.enviarMensajeTexto(
          data.citaId,
          autorTipo,
          id,
          data.contenido
        );

        // Emitir a todos los que están en la room de la cita
        io.to(`cita:${data.citaId}`).emit('chat:nuevoMensaje', {
          citaId: data.citaId,
          mensaje,
        });
      } catch (err) {
        socket.emit('error', { message: err.message });
      }
    });

    socket.on('disconnect', () => {
      console.log(`[Socket.io] Desconectado: ${nombre} (${rol})`);
    });
  });

  return io;
}

module.exports = { setupSocketIO };

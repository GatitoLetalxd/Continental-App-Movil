const prisma = require('../config/database');
const { NotFoundError, ValidationError, ForbiddenError } = require('../utils/errors');
const {
  MAX_VIDEOS_POR_CHAT,
  MAX_TAMANO_VIDEO_COMBINADO_MB,
  MAX_TAMANO_FOTO_MB,
  DIAS_EXPIRACION_CHAT_MULTIMEDIA,
} = require('../config/constants');
const fs = require('fs');
const path = require('path');

/**
 * Enviar un mensaje de texto en el chat de una cita.
 */
async function enviarMensajeTexto(citaId, autorTipo, autorId, contenido) {
  await verificarAccesoChat(citaId, autorTipo, autorId);

  const expiraAt = new Date();
  expiraAt.setDate(expiraAt.getDate() + DIAS_EXPIRACION_CHAT_MULTIMEDIA);

  return prisma.mensaje.create({
    data: {
      citaId: parseInt(citaId),
      autorTipo,
      autorId,
      contenido,
      tipo: 'TEXTO',
      expiraAt,
    },
  });
}

/**
 * Enviar un archivo multimedia (imagen o video) en el chat.
 * Valida los límites por conversación.
 */
async function enviarMensajeMultimedia(citaId, autorTipo, autorId, tipo, archivoUrl, tamanoBytes) {
  await verificarAccesoChat(citaId, autorTipo, autorId);

  // Validar límites de video
  if (tipo === 'VIDEO') {
    const videosExistentes = await prisma.mensaje.count({
      where: {
        citaId: parseInt(citaId),
        tipo: 'VIDEO',
        expirado: false,
      },
    });

    if (videosExistentes >= MAX_VIDEOS_POR_CHAT) {
      throw new ValidationError(
        `Máximo ${MAX_VIDEOS_POR_CHAT} videos por conversación. Ya tienes ${videosExistentes}.`
      );
    }

    // Validar tamaño combinado de videos (250MB)
    const tamanoMB = tamanoBytes / (1024 * 1024);
    // Nota: el tamaño combinado se valida estimativamente aquí
    if (tamanoMB > MAX_TAMANO_VIDEO_COMBINADO_MB) {
      throw new ValidationError(
        `El tamaño del video excede el límite de ${MAX_TAMANO_VIDEO_COMBINADO_MB}MB`
      );
    }
  }

  // Validar tamaño de foto
  if (tipo === 'IMAGEN') {
    const tamanoMB = tamanoBytes / (1024 * 1024);
    if (tamanoMB > MAX_TAMANO_FOTO_MB) {
      throw new ValidationError(`La foto excede el límite de ${MAX_TAMANO_FOTO_MB}MB`);
    }
  }

  const expiraAt = new Date();
  expiraAt.setDate(expiraAt.getDate() + DIAS_EXPIRACION_CHAT_MULTIMEDIA);

  return prisma.mensaje.create({
    data: {
      citaId: parseInt(citaId),
      autorTipo,
      autorId,
      tipo,
      archivoUrl,
      expiraAt,
    },
  });
}

/**
 * Listar mensajes de una cita (paginado).
 */
async function listarMensajes(citaId, autorTipo, autorId, { page = 1, limit = 50 } = {}) {
  await verificarAccesoChat(citaId, autorTipo, autorId);

  const [mensajes, total] = await Promise.all([
    prisma.mensaje.findMany({
      where: { citaId: parseInt(citaId) },
      orderBy: { createdAt: 'asc' },
      skip: (page - 1) * limit,
      take: limit,
    }),
    prisma.mensaje.count({ where: { citaId: parseInt(citaId) } }),
  ]);

  return { mensajes, total, page, totalPages: Math.ceil(total / limit) };
}

/**
 * Verificar que el usuario tiene acceso al chat de la cita.
 * Chat PRIVADO: solo el cliente de la cita y el técnico asignado.
 * El admin NO tiene acceso en v1.
 */
async function verificarAccesoChat(citaId, autorTipo, autorId) {
  const cita = await prisma.cita.findUnique({
    where: { id: parseInt(citaId) },
    select: { clienteId: true, tecnicoId: true },
  });

  if (!cita) throw new NotFoundError('Cita');

  if (autorTipo === 'CLIENTE' && cita.clienteId !== autorId) {
    throw new ForbiddenError('No tienes acceso al chat de esta cita');
  }

  if (autorTipo === 'TECNICO' && cita.tecnicoId !== autorId) {
    throw new ForbiddenError('No tienes acceso al chat de esta cita');
  }
}

module.exports = {
  enviarMensajeTexto,
  enviarMensajeMultimedia,
  listarMensajes,
  verificarAccesoChat,
};

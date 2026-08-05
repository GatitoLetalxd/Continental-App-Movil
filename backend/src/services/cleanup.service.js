const prisma = require('../config/database');
const fs = require('fs');
const path = require('path');
const {
  DIAS_EXPIRACION_FOTOS_EVIDENCIA,
  DIAS_EXPIRACION_CHAT_MULTIMEDIA,
} = require('../config/constants');

const UPLOAD_DIR = process.env.UPLOAD_DIR || 'uploads';

/**
 * Limpia fotos de evidencia con más de 30 días de antigüedad.
 * Elimina el archivo del disco y marca fotoExpirada = true en la BD.
 */
async function limpiarFotosEvidencia() {
  const fechaLimite = new Date();
  fechaLimite.setDate(fechaLimite.getDate() - DIAS_EXPIRACION_FOTOS_EVIDENCIA);

  console.log(`[Cleanup] Buscando fotos de evidencia anteriores a ${fechaLimite.toISOString()}`);

  // Fotos de inicio expiradas
  const citasConFotoInicio = await prisma.cita.findMany({
    where: {
      fotoInicio: { not: null },
      fotoInicioExpirada: false,
      createdAt: { lt: fechaLimite },
    },
    select: { id: true, fotoInicio: true },
  });

  for (const cita of citasConFotoInicio) {
    try {
      const filePath = path.resolve(cita.fotoInicio);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
        console.log(`[Cleanup] Eliminada foto inicio: ${filePath}`);
      }
      await prisma.cita.update({
        where: { id: cita.id },
        data: { fotoInicioExpirada: true },
      });
    } catch (err) {
      console.error(`[Cleanup] Error al eliminar foto inicio de cita ${cita.id}:`, err.message);
    }
  }

  // Fotos de fin expiradas
  const citasConFotoFin = await prisma.cita.findMany({
    where: {
      fotoFin: { not: null },
      fotoFinExpirada: false,
      createdAt: { lt: fechaLimite },
    },
    select: { id: true, fotoFin: true },
  });

  for (const cita of citasConFotoFin) {
    try {
      const filePath = path.resolve(cita.fotoFin);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
        console.log(`[Cleanup] Eliminada foto fin: ${filePath}`);
      }
      await prisma.cita.update({
        where: { id: cita.id },
        data: { fotoFinExpirada: true },
      });
    } catch (err) {
      console.error(`[Cleanup] Error al eliminar foto fin de cita ${cita.id}:`, err.message);
    }
  }

  const totalEliminadas = citasConFotoInicio.length + citasConFotoFin.length;
  console.log(`[Cleanup] Fotos de evidencia procesadas: ${totalEliminadas}`);
  return totalEliminadas;
}

/**
 * Limpia multimedia de chat con más de 7 días de antigüedad.
 * Elimina el archivo del disco y marca expirado = true en la BD.
 */
async function limpiarMultimediaChat() {
  const ahora = new Date();

  console.log(`[Cleanup] Buscando multimedia de chat expirada`);

  const mensajesExpirados = await prisma.mensaje.findMany({
    where: {
      tipo: { in: ['IMAGEN', 'VIDEO'] },
      archivoUrl: { not: null },
      expirado: false,
      expiraAt: { lt: ahora },
    },
    select: { id: true, archivoUrl: true },
  });

  for (const mensaje of mensajesExpirados) {
    try {
      const filePath = path.resolve(mensaje.archivoUrl);
      if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
        console.log(`[Cleanup] Eliminado archivo chat: ${filePath}`);
      }
      await prisma.mensaje.update({
        where: { id: mensaje.id },
        data: { expirado: true, archivoUrl: null },
      });
    } catch (err) {
      console.error(`[Cleanup] Error al eliminar archivo de mensaje ${mensaje.id}:`, err.message);
    }
  }

  console.log(`[Cleanup] Multimedia de chat procesada: ${mensajesExpirados.length}`);
  return mensajesExpirados.length;
}

module.exports = {
  limpiarFotosEvidencia,
  limpiarMultimediaChat,
};

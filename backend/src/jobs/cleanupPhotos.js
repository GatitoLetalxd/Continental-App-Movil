const cleanupService = require('../services/cleanup.service');

/**
 * Job: Limpieza diaria de fotos de evidencia con más de 30 días.
 */
async function ejecutar() {
  console.log('[CleanupPhotos] Iniciando limpieza de fotos de evidencia...');
  try {
    const eliminadas = await cleanupService.limpiarFotosEvidencia();
    console.log(`[CleanupPhotos] Finalizado. Fotos procesadas: ${eliminadas}`);
  } catch (error) {
    console.error('[CleanupPhotos] Error:', error.message);
  }
}

module.exports = { ejecutar };

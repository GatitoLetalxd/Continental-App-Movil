const cleanupService = require('../services/cleanup.service');

/**
 * Job: Limpieza diaria de multimedia de chat con más de 7 días.
 */
async function ejecutar() {
  console.log('[CleanupChat] Iniciando limpieza de multimedia de chat...');
  try {
    const eliminados = await cleanupService.limpiarMultimediaChat();
    console.log(`[CleanupChat] Finalizado. Archivos procesados: ${eliminados}`);
  } catch (error) {
    console.error('[CleanupChat] Error:', error.message);
  }
}

module.exports = { ejecutar };

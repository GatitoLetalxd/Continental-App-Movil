const { exec } = require('child_process');
const path = require('path');
const fs = require('fs');

/**
 * Ejecuta pg_dump para hacer backup de la base de datos.
 * Se ejecuta como cron job diario.
 */
async function ejecutarBackup() {
  const backupDir = path.join(process.cwd(), 'backups');
  if (!fs.existsSync(backupDir)) {
    fs.mkdirSync(backupDir, { recursive: true });
  }

  const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
  const filename = `backup-${timestamp}.sql`;
  const filepath = path.join(backupDir, filename);

  // Extraer datos de conexión del DATABASE_URL
  const dbUrl = process.env.DATABASE_URL;
  const match = dbUrl.match(/postgresql:\/\/([^:]+):([^@]+)@([^:]+):(\d+)\/([^?]+)/);
  
  if (!match) {
    console.error('[Backup] No se pudo parsear DATABASE_URL');
    return;
  }

  const [, user, password, host, port, database] = match;

  const cmd = `PGPASSWORD="${password}" pg_dump -h ${host} -p ${port} -U ${user} -d ${database} -f "${filepath}"`;

  return new Promise((resolve, reject) => {
    console.log(`[Backup] Iniciando backup: ${filename}`);
    
    exec(cmd, (error, stdout, stderr) => {
      if (error) {
        console.error(`[Backup] Error: ${error.message}`);
        reject(error);
        return;
      }
      if (stderr && !stderr.includes('pg_dump')) {
        console.warn(`[Backup] Advertencia: ${stderr}`);
      }

      // Mantener solo los últimos 7 backups
      const backups = fs.readdirSync(backupDir)
        .filter(f => f.startsWith('backup-') && f.endsWith('.sql'))
        .sort()
        .reverse();

      if (backups.length > 7) {
        for (const old of backups.slice(7)) {
          fs.unlinkSync(path.join(backupDir, old));
          console.log(`[Backup] Eliminado backup antiguo: ${old}`);
        }
      }

      console.log(`[Backup] Completado: ${filename}`);
      resolve(filepath);
    });
  });
}

module.exports = { ejecutarBackup };

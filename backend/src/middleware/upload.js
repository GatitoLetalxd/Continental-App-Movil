const multer = require('multer');
const path = require('path');
const fs = require('fs');
const { ValidationError } = require('../utils/errors');

const UPLOAD_DIR = process.env.UPLOAD_DIR || 'uploads';

// Asegurar que existan los directorios
const ensureDir = (dir) => {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
};

// ─── Storage para fotos de evidencia ──────────────────────
const evidenciaStorage = multer.diskStorage({
  destination: (req, file, cb) => {
    const dir = path.join(UPLOAD_DIR, 'evidencia');
    ensureDir(dir);
    cb(null, dir);
  },
  filename: (req, file, cb) => {
    const uniqueSuffix = `${Date.now()}-${Math.round(Math.random() * 1E9)}`;
    const ext = path.extname(file.originalname);
    cb(null, `evidencia-${uniqueSuffix}${ext}`);
  },
});

// ─── Storage para multimedia del chat ─────────────────────
const chatStorage = multer.diskStorage({
  destination: (req, file, cb) => {
    const dir = path.join(UPLOAD_DIR, 'chat');
    ensureDir(dir);
    cb(null, dir);
  },
  filename: (req, file, cb) => {
    const uniqueSuffix = `${Date.now()}-${Math.round(Math.random() * 1E9)}`;
    const ext = path.extname(file.originalname);
    cb(null, `chat-${uniqueSuffix}${ext}`);
  },
});

// ─── Filtros de tipo de archivo ───────────────────────────
const imageFilter = (req, file, cb) => {
  const allowedTypes = /jpeg|jpg|png|webp/;
  const extname = allowedTypes.test(path.extname(file.originalname).toLowerCase());
  const mimetype = allowedTypes.test(file.mimetype);

  if (extname && mimetype) {
    cb(null, true);
  } else {
    cb(new ValidationError('Solo se permiten imágenes (jpg, png, webp)'));
  }
};

const mediaFilter = (req, file, cb) => {
  const imageTypes = /jpeg|jpg|png|webp/;
  const videoTypes = /mp4|mov|quicktime/;
  const ext = path.extname(file.originalname).toLowerCase();
  
  const isImage = imageTypes.test(ext) || imageTypes.test(file.mimetype);
  const isVideo = videoTypes.test(ext) || /video/.test(file.mimetype);

  if (isImage || isVideo) {
    cb(null, true);
  } else {
    cb(new ValidationError('Solo se permiten imágenes (jpg, png, webp) y videos (mp4, mov)'));
  }
};

// ─── Instancias de Multer ─────────────────────────────────

// Para fotos de evidencia (inicio/fin de cita): max 50MB
const uploadEvidencia = multer({
  storage: evidenciaStorage,
  fileFilter: imageFilter,
  limits: { fileSize: 50 * 1024 * 1024 }, // 50MB
});

// Para chat multimedia: max 250MB por archivo (se valida el total combinado en el service)
const uploadChat = multer({
  storage: chatStorage,
  fileFilter: mediaFilter,
  limits: { fileSize: 250 * 1024 * 1024 }, // 250MB max por archivo individual
});

// Para evidencias generales de citas (fotos y videos): max 200MB (según requerimiento de fase)
const uploadEvidenciaGeneral = multer({
  storage: evidenciaStorage,
  fileFilter: mediaFilter,
  limits: { fileSize: 200 * 1024 * 1024 }, // 200MB
});

module.exports = {
  uploadEvidencia,
  uploadChat,
  uploadEvidenciaGeneral,
  UPLOAD_DIR,
};

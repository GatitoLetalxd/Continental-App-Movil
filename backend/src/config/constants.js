// ─── Constantes de la aplicación ─────────────────────────────
// Fuente única de verdad para valores configurables.
// NO hardcodear estos valores en otros archivos.

// Fases por defecto de una cita (plantilla)
const FASES_POR_DEFECTO = ['Inspección', 'Desmontaje', 'Instalación', 'Finalizado'];

// Longitud máxima del nombre de fase (texto libre)
const MAX_CARACTERES_FASE = 30;

// Anticipación mínima para citas remotas (minutos)
const MIN_ANTICIPACION_REMOTA = parseInt(process.env.MIN_ANTICIPACION_REMOTA || '120', 10);

// Expiración de archivos
const DIAS_EXPIRACION_FOTOS_EVIDENCIA = 30;
const DIAS_EXPIRACION_CHAT_MULTIMEDIA = 7;

// Límites de archivos de chat
const MAX_VIDEOS_POR_CHAT = 2;
const MAX_TAMANO_VIDEO_COMBINADO_MB = 250;
const MAX_TAMANO_FOTO_MB = 50;

// Contactos máximos por cliente
const MAX_CONTACTOS_CLIENTE = 2;

// ─── Plantillas de mensajes WhatsApp ─────────────────────────
// Se usan para generar links wa.me con mensaje preescrito.
// Variables disponibles: {nombre}, {placa}, {servicio}, {nuevaHora}

const PLANTILLAS_WHATSAPP = {
  CAMBIO_HORA_ESTIMADA: 
    'Hola {nombre}, le informamos que la hora estimada de entrega de su vehículo {vehiculo} ha sido actualizada a {nuevaHora}. Disculpe las molestias. — Taller Continental',

  CITA_POSPUESTA: 
    'Hola {nombre}, lamentamos informarle que su cita para el servicio "{servicio}" ha sido pospuesta. Nos pondremos en contacto para reprogramarla con prioridad. Disculpe las molestias. — Taller Continental',

  LISTA_PARA_ENTREGA: 
    'Hola {nombre}, ¡su vehículo {vehiculo} ya está listo para retirar! Le esperamos en el taller. ¡Gracias por su confianza! — Taller Continental',
};

// Transiciones válidas de la máquina de estados de cita
const TRANSICIONES_CITA = {
  AGENDADA:    ['ACEPTADA', 'CANCELADA'],
  ACEPTADA:    ['EN_CURSO', 'POSPUESTA', 'CANCELADA'],
  EN_CURSO:    ['PAUSADA', 'FINALIZADA', 'POSPUESTA', 'CANCELADA'],
  PAUSADA:     ['EN_CURSO', 'POSPUESTA', 'CANCELADA'],
  FINALIZADA:  ['ENTREGADA'],
  ENTREGADA:   [],
  POSPUESTA:   ['ACEPTADA'],  // Reasignación: vuelve a ACEPTADA con nuevo técnico
  CANCELADA:   [],
};

// Roles del sistema
const ROLES = {
  ADMIN: 'ADMIN',
  TECNICO: 'TECNICO',
  CLIENTE: 'CLIENTE',
};

module.exports = {
  FASES_POR_DEFECTO,
  MAX_CARACTERES_FASE,
  MIN_ANTICIPACION_REMOTA,
  DIAS_EXPIRACION_FOTOS_EVIDENCIA,
  DIAS_EXPIRACION_CHAT_MULTIMEDIA,
  MAX_VIDEOS_POR_CHAT,
  MAX_TAMANO_VIDEO_COMBINADO_MB,
  MAX_TAMANO_FOTO_MB,
  MAX_CONTACTOS_CLIENTE,
  PLANTILLAS_WHATSAPP,
  TRANSICIONES_CITA,
  ROLES,
};

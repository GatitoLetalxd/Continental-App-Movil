const prisma = require('../config/database');
const { PLANTILLAS_WHATSAPP } = require('../config/constants');
const { NotFoundError, ValidationError } = require('../utils/errors');
const { calcularHoraEstimadaEntrega, formatearHora } = require('../utils/time');

/**
 * Genera un link wa.me con mensaje preescrito.
 * 
 * El sistema NO envía mensajes automáticamente.
 * Solo genera el link para que un humano (técnico/admin) lo abra y envíe manualmente.
 * 
 * @param {string} telefono - Número con código de país (ej: "584121234567")
 * @param {string} plantilla - Texto con variables {nombre}, {placa}, {servicio}, {nuevaHora}
 * @param {object} datos - Valores para las variables
 * @returns {string} Link completo wa.me
 */
function generarLinkWa(telefono, plantilla, datos = {}) {
  let mensaje = plantilla;

  // Reemplazar variables
  for (const [key, value] of Object.entries(datos)) {
    mensaje = mensaje.replace(new RegExp(`\\{${key}\\}`, 'g'), value);
  }

  // Limpiar el número de teléfono (solo dígitos)
  const numeroLimpio = telefono.replace(/\D/g, '');

  // URL encode del mensaje
  const mensajeCodificado = encodeURIComponent(mensaje);

  return `https://wa.me/${numeroLimpio}?text=${mensajeCodificado}`;
}

/**
 * Genera el link wa.me para un evento específico de una cita.
 * 
 * Tipos soportados:
 * - "cambio_hora": Cambio de hora estimada de entrega
 * - "pospuesta": Cita pospuesta
 * - "lista_entrega": Cita lista para entrega
 */
async function generarLinkParaCita(citaId, tipo) {
  const cita = await prisma.cita.findUnique({
    where: { id: parseInt(citaId) },
    include: {
      cliente: { select: { nombre: true, contactos: true } },
      vehiculo: { select: { marca: true, modelo: true, placa: true } },
      servicio: { select: { nombre: true } },
    },
  });

  if (!cita) throw new NotFoundError('Cita');

  // Obtener el primer contacto del cliente
  if (!cita.cliente.contactos || cita.cliente.contactos.length === 0) {
    throw new ValidationError('El cliente no tiene números de contacto registrados');
  }
  const telefono = cita.cliente.contactos[0];

  const vehiculoInfo = cita.vehiculo.placa && cita.vehiculo.placa.trim().length > 0
    ? `${cita.vehiculo.marca} ${cita.vehiculo.modelo} (${cita.vehiculo.placa})`
    : `${cita.vehiculo.marca} ${cita.vehiculo.modelo}`;

  // Datos para las plantillas
  const nombreServicio = cita.servicio ? cita.servicio.nombre : (cita.serviciosDetalle || 'Servicio Personalizado');
  const datos = {
    nombre: cita.cliente.nombre,
    vehiculo: vehiculoInfo,
    servicio: nombreServicio,
    nuevaHora: formatearHora(
      calcularHoraEstimadaEntrega(cita.horaInicioReal, cita.horaAgendada, cita.duracionEstimadaMinutos)
    ),
  };

  let plantilla;
  switch (tipo) {
    case 'cambio_hora':
      plantilla = PLANTILLAS_WHATSAPP.CAMBIO_HORA_ESTIMADA;
      break;
    case 'pospuesta':
      plantilla = PLANTILLAS_WHATSAPP.CITA_POSPUESTA;
      break;
    case 'lista_entrega':
      plantilla = PLANTILLAS_WHATSAPP.LISTA_PARA_ENTREGA;
      break;
    default:
      throw new ValidationError('Tipo de mensaje no válido. Use: cambio_hora, pospuesta, lista_entrega');
  }

  // Generar links para todos los contactos del cliente
  const links = cita.cliente.contactos.map(contacto => ({
    telefono: contacto,
    link: generarLinkWa(contacto, plantilla, datos),
  }));

  return {
    tipo,
    cita: {
      id: cita.id,
      cliente: cita.cliente.nombre,
      vehiculo: vehiculoInfo,
      servicio: nombreServicio,
    },
    links,
  };
}

module.exports = {
  generarLinkWa,
  generarLinkParaCita,
};

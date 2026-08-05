const citaService = require('../services/cita.service');
const citaEstadoService = require('../services/citaEstado.service');
const faseService = require('../services/fase.service');
const whatsappService = require('../services/whatsapp.service');

// ─── CRUD ────────────────────────────────────────────────

async function crear(req, res, next) {
  try {
    const datos = { ...req.body };
    // Si es un cliente creando su propia cita
    if (req.user.rol === 'CLIENTE') {
      datos.clienteId = req.user.id;
    }
    // Agregar creadorId para saber qué técnico o admin está creando la cita
    datos.creadorId = req.user.id;

    const data = await citaService.crear(datos, req.user.rol);

    // Emitir eventos Socket.io si la cita fue auto-aceptada
    if (req.user.rol === 'TECNICO' && datos.autoAceptar) {
      const io = req.app.get('io');
      if (io) {
        io.to(`cita:${data.id}`).emit('cita:estadoCambiado', {
          citaId: data.id,
          nuevoEstado: 'ACEPTADA',
          timestamp: new Date(),
        });
        io.to('admin').emit('tecnico:disponibilidadCambiada', {
          tecnicoId: req.user.id,
          timestamp: new Date(),
        });
      }
    }

    res.status(201).json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function listar(req, res, next) {
  try {
    const filtros = { ...req.query };
    // Un cliente solo ve sus citas
    if (req.user.rol === 'CLIENTE') {
      filtros.clienteId = req.user.id;
    }
    // Un técnico ve SOLO las suyas por defecto.
    // El parámetro 'todas' viene como string desde la query URL.
    // Solo lo omitimos si explícitamente se envía 'todas=true'.
    if (req.user.rol === 'TECNICO' && filtros.todas !== 'true') {
      filtros.tecnicoId = req.user.id;
    }
    const data = await citaService.listar(filtros);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function listarSinAsignar(req, res, next) {
  try {
    const data = await citaService.listarSinAsignar();
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function listarPospuestas(req, res, next) {
  try {
    const data = await citaService.listarPospuestas();
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerPorId(req, res, next) {
  try {
    const data = await citaService.obtenerPorId(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

// ─── Transiciones de estado ──────────────────────────────

async function aceptar(req, res, next) {
  try {
    const data = await citaEstadoService.aceptar(parseInt(req.params.id), req.user.id);
    // Emitir evento Socket.io
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'ACEPTADA',
        timestamp: new Date(),
      });
      io.to('admin').emit('tecnico:disponibilidadCambiada', {
        tecnicoId: req.user.id,
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function iniciar(req, res, next) {
  try {
    const data = await citaEstadoService.iniciar(parseInt(req.params.id), req.user.id);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'EN_CURSO',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function pausar(req, res, next) {
  try {
    const data = await citaEstadoService.pausar(parseInt(req.params.id), req.user.id);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'PAUSADA',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function reanudar(req, res, next) {
  try {
    const data = await citaEstadoService.reanudar(parseInt(req.params.id), req.user.id);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'EN_CURSO',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function finalizar(req, res, next) {
  try {
    const data = await citaEstadoService.finalizar(parseInt(req.params.id), req.user.id);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'FINALIZADA',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function entregar(req, res, next) {
  try {
    const data = await citaEstadoService.entregar(parseInt(req.params.id));
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'ENTREGADA',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function posponer(req, res, next) {
  try {
    const data = await citaEstadoService.posponer(
      parseInt(req.params.id),
      req.user.id,
      req.body.motivo
    );
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'POSPUESTA',
        timestamp: new Date(),
      });
      io.to('admin').emit('cita:pospuesta', {
        citaId: parseInt(req.params.id),
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function cancelar(req, res, next) {
  try {
    const data = await citaEstadoService.cancelar(
      parseInt(req.params.id),
      req.user.id,
      req.user.rol
    );
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'CANCELADA',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function reasignar(req, res, next) {
  try {
    const data = await citaEstadoService.reasignar(
      parseInt(req.params.id),
      parseInt(req.body.tecnicoId),
      req.body.nuevaHoraAgendada
    );
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:estadoCambiado', {
        citaId: parseInt(req.params.id),
        nuevoEstado: 'ACEPTADA',
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

// ─── Duración ────────────────────────────────────────────

async function actualizarDuracion(req, res, next) {
  try {
    const data = await citaService.actualizarDuracion(
      parseInt(req.params.id),
      req.user.id,
      req.body.duracionEstimadaMinutos
    );
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

// ─── Fotos ───────────────────────────────────────────────

async function subirFoto(req, res, next) {
  try {
    if (!req.file) {
      return res.status(400).json({ success: false, message: 'Debe seleccionar una foto' });
    }
    const data = await citaService.subirFoto(
      parseInt(req.params.id),
      req.params.tipo, // 'inicio' o 'fin'
      req.file.path
    );
    res.json({ success: true, data, archivoUrl: req.file.path });
  } catch (error) {
    next(error);
  }
}

async function subirEvidencia(req, res, next) {
  try {
    if (!req.file) {
      return res.status(400).json({ success: false, message: 'Debe seleccionar un archivo' });
    }
    const { tipo, fase } = req.body;
    if (!['INICIO', 'AVANCE', 'FIN'].includes(tipo)) {
      return res.status(400).json({ success: false, message: 'Tipo de evidencia no válido. Use INICIO, AVANCE o FIN' });
    }

    const data = await citaService.subirEvidencia(
      parseInt(req.params.id),
      tipo,
      fase || null,
      req.file.path,
      req.file.mimetype
    );

    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:evidenciaSubida', {
        citaId: parseInt(req.params.id),
        evidencia: data,
      });
    }

    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

// ─── Fases ───────────────────────────────────────────────

async function cambiarFase(req, res, next) {
  try {
    const data = await faseService.cambiarFase(req.params.id, req.body.nombreFase);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:faseCambiada', {
        citaId: parseInt(req.params.id),
        nuevaFase: req.body.nombreFase,
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function avanzarFase(req, res, next) {
  try {
    const data = await faseService.avanzarFase(req.params.id);
    const io = req.app.get('io');
    if (io) {
      io.to(`cita:${req.params.id}`).emit('cita:faseCambiada', {
        citaId: parseInt(req.params.id),
        nuevaFase: data.faseActual,
        timestamp: new Date(),
      });
    }
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function resetearFase(req, res, next) {
  try {
    const data = await faseService.resetearFasePorDefecto(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerHistorialFases(req, res, next) {
  try {
    const data = await faseService.obtenerHistorial(req.params.id);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

async function obtenerFasesPorDefecto(req, res, next) {
  try {
    const data = faseService.obtenerFasesPorDefecto();
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

// ─── WhatsApp ────────────────────────────────────────────

async function obtenerLinkWhatsApp(req, res, next) {
  try {
    const data = await whatsappService.generarLinkParaCita(req.params.id, req.params.tipo);
    res.json({ success: true, data });
  } catch (error) {
    next(error);
  }
}

module.exports = {
  crear,
  listar,
  listarSinAsignar,
  listarPospuestas,
  obtenerPorId,
  aceptar,
  iniciar,
  pausar,
  reanudar,
  finalizar,
  entregar,
  posponer,
  cancelar,
  reasignar,
  actualizarDuracion,
  subirFoto,
  subirEvidencia,
  cambiarFase,
  avanzarFase,
  resetearFase,
  obtenerHistorialFases,
  obtenerFasesPorDefecto,
  obtenerLinkWhatsApp,
};

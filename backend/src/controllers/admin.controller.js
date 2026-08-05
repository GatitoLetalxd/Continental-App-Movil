const prisma = require('../config/database');
const tecnicoService = require('../services/tecnico.service');

/**
 * Dashboard operativo: estado en tiempo real de cada técnico.
 */
async function dashboard(req, res, next) {
  try {
    const tecnicos = await prisma.tecnico.findMany({
      where: { estado: 'ACTIVO' },
      select: { id: true, nombre: true },
    });

    const estadoTecnicos = await Promise.all(
      tecnicos.map(t => tecnicoService.obtenerDisponibilidad(t.id))
    );

    // Contar citas por estado
    const conteos = await prisma.cita.groupBy({
      by: ['estado'],
      _count: { id: true },
    });

    const resumen = {
      tecnicos: estadoTecnicos,
      citasPorEstado: conteos.reduce((acc, c) => {
        acc[c.estado] = c._count.id;
        return acc;
      }, {}),
      citasPospuestas: await prisma.cita.count({ where: { estado: 'POSPUESTA' } }),
      citasSinAsignar: await prisma.cita.count({ where: { estado: 'AGENDADA', tecnicoId: null } }),
    };

    res.json({ success: true, data: resumen });
  } catch (error) {
    next(error);
  }
}

/**
 * Reporte de productividad por técnico.
 */
async function reporteProductividad(req, res, next) {
  try {
    const { fechaInicio, fechaFin } = req.query;
    const where = { estado: 'ENTREGADA' };

    if (fechaInicio || fechaFin) {
      where.horaFinReal = {};
      if (fechaInicio) where.horaFinReal.gte = new Date(fechaInicio);
      if (fechaFin) where.horaFinReal.lte = new Date(fechaFin);
    }

    const tecnicos = await prisma.tecnico.findMany({
      select: {
        id: true,
        nombre: true,
        estado: true,
        citas: {
          where,
          select: {
            id: true,
            horaInicioReal: true,
            horaFinReal: true,
            duracionEstimadaMinutos: true,
            servicio: { select: { nombre: true } },
          },
        },
      },
    });

    const reporte = tecnicos.map(tecnico => {
      const citasFinalizadas = tecnico.citas.length;
      const tiempos = tecnico.citas
        .filter(c => c.horaInicioReal && c.horaFinReal)
        .map(c => (new Date(c.horaFinReal) - new Date(c.horaInicioReal)) / (1000 * 60));
      
      const tiempoPromedio = tiempos.length > 0
        ? Math.round(tiempos.reduce((a, b) => a + b, 0) / tiempos.length)
        : 0;

      // Tiempo promedio por tipo de servicio
      const porServicio = {};
      for (const cita of tecnico.citas) {
        const nombre = cita.servicio ? cita.servicio.nombre : (cita.serviciosDetalle || 'Servicio Personalizado');
        if (!porServicio[nombre]) porServicio[nombre] = { total: 0, count: 0 };
        if (cita.horaInicioReal && cita.horaFinReal) {
          porServicio[nombre].total += (new Date(cita.horaFinReal) - new Date(cita.horaInicioReal)) / (1000 * 60);
          porServicio[nombre].count += 1;
        }
      }

      const tiempoPorServicio = Object.entries(porServicio).map(([servicio, data]) => ({
        servicio,
        promedio: Math.round(data.total / data.count),
        cantidad: data.count,
      }));

      return {
        tecnicoId: tecnico.id,
        nombre: tecnico.nombre,
        estado: tecnico.estado,
        citasFinalizadas,
        tiempoPromedioMinutos: tiempoPromedio,
        tiempoPorServicio,
      };
    });

    res.json({ success: true, data: reporte });
  } catch (error) {
    next(error);
  }
}

/**
 * Reporte de servicios más solicitados.
 */
async function reporteServicios(req, res, next) {
  try {
    const { fechaInicio, fechaFin } = req.query;
    const where = {};

    if (fechaInicio || fechaFin) {
      where.createdAt = {};
      if (fechaInicio) where.createdAt.gte = new Date(fechaInicio);
      if (fechaFin) where.createdAt.lte = new Date(fechaFin);
    }

    const servicios = await prisma.cita.groupBy({
      by: ['servicioId'],
      where,
      _count: { id: true },
      orderBy: { _count: { id: 'desc' } },
    });

    // Enriquecer con datos del servicio
    const resultado = await Promise.all(
      servicios.map(async (s) => {
        if (!s.servicioId) {
          return {
            servicioId: null,
            nombre: 'Servicios Personalizados',
            precio: 0,
            cantidadSolicitada: s._count.id,
          };
        }
        const servicio = await prisma.servicio.findUnique({
          where: { id: s.servicioId },
          select: { nombre: true, precio: true },
        });
        return {
          servicioId: s.servicioId,
          nombre: servicio?.nombre || 'Desconocido',
          precio: servicio?.precio,
          cantidadSolicitada: s._count.id,
        };
      })
    );

    res.json({ success: true, data: resultado });
  } catch (error) {
    next(error);
  }
}

module.exports = { dashboard, reporteProductividad, reporteServicios };

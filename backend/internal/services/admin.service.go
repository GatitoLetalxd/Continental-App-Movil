package services

import (
	"math"
	"sort"
	"time"

	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type AdminService struct {
	db             *gorm.DB
	tecnicoService *TecnicoService
}

func NewAdminService() *AdminService {
	return &AdminService{
		db:             database.DB,
		tecnicoService: NewTecnicoService(),
	}
}

type DashboardResponse struct {
	Tecnicos        []*DisponibilidadTecnicoResponse `json:"tecnicos"`
	CitasPorEstado  map[string]int64                 `json:"citasPorEstado"`
	CitasPospuestas int64                            `json:"citasPospuestas"`
	CitasSinAsignar int64                            `json:"citasSinAsignar"`
}

type TiempoServicioReporte struct {
	Servicio string `json:"servicio"`
	Promedio int    `json:"promedio"`
	Cantidad int    `json:"cantidad"`
}

type ReporteProductividadItem struct {
	TecnicoID             int                     `json:"tecnicoId"`
	Nombre                string                  `json:"nombre"`
	Estado                string                  `json:"estado"`
	CitasFinalizadas      int                     `json:"citasFinalizadas"`
	TiempoPromedioMinutos int                     `json:"tiempoPromedioMinutos"`
	TiempoPorServicio     []TiempoServicioReporte `json:"tiempoPorServicio"`
}

type ReporteServicioItem struct {
	ServicioID         *int     `json:"servicioId"`
	Nombre             string   `json:"nombre"`
	Precio             *float64 `json:"precio"`
	CantidadSolicitada int64    `json:"cantidadSolicitada"`
}

func (s *AdminService) Dashboard() (*DashboardResponse, error) {
	var tecnicos []models.Tecnico
	if err := s.db.Where("estado = ?", models.EstadoTecnicoActivo).Find(&tecnicos).Error; err != nil {
		return nil, utils.NewAppError("Error al consultar técnicos: "+err.Error(), 500)
	}

	estadosTecnicos := make([]*DisponibilidadTecnicoResponse, 0, len(tecnicos))
	for _, t := range tecnicos {
		disp, err := s.tecnicoService.ObtenerDisponibilidad(t.ID)
		if err == nil {
			estadosTecnicos = append(estadosTecnicos, disp)
		}
	}

	type ConteoEstado struct {
		Estado string
		Total  int64
	}
	var conteos []ConteoEstado
	s.db.Model(&models.Cita{}).Select("estado, count(id) as total").Group("estado").Scan(&conteos)

	citasPorEstado := make(map[string]int64)
	for _, c := range conteos {
		citasPorEstado[c.Estado] = c.Total
	}

	var citasPospuestas int64
	s.db.Model(&models.Cita{}).Where("estado = ?", models.EstadoCitaPospuesta).Count(&citasPospuestas)

	var citasSinAsignar int64
	s.db.Model(&models.Cita{}).Where("estado = ? AND tecnico_id IS NULL", models.EstadoCitaAgendada).Count(&citasSinAsignar)

	return &DashboardResponse{
		Tecnicos:        estadosTecnicos,
		CitasPorEstado:  citasPorEstado,
		CitasPospuestas: citasPospuestas,
		CitasSinAsignar: citasSinAsignar,
	}, nil
}

func (s *AdminService) ReporteProductividad(fechaInicio, fechaFin string) ([]ReporteProductividadItem, error) {
	var tecnicos []models.Tecnico
	if err := s.db.Find(&tecnicos).Error; err != nil {
		return nil, utils.NewAppError("Error al consultar técnicos: "+err.Error(), 500)
	}

	reporte := make([]ReporteProductividadItem, 0, len(tecnicos))

	for _, t := range tecnicos {
		query := s.db.Where("tecnico_id = ? AND estado = 'ENTREGADA'", t.ID)

		if fechaInicio != "" {
			if parsed, err := time.Parse("2006-01-02", fechaInicio); err == nil {
				query = query.Where("hora_fin_real >= ?", parsed)
			}
		}
		if fechaFin != "" {
			if parsed, err := time.Parse("2006-01-02", fechaFin); err == nil {
				query = query.Where("hora_fin_real <= ?", parsed.Add(24*time.Hour))
			}
		}

		var citas []models.Cita
		query.Preload("Servicio").Find(&citas)

		type svcData struct {
			total float64
			count int
		}
		porServicio := make(map[string]*svcData)
		var sumaTiempos float64
		var countTiempos int

		for _, c := range citas {
			if c.HoraInicioReal != nil && c.HoraFinReal != nil {
				diffMin := c.HoraFinReal.Sub(*c.HoraInicioReal).Minutes()
				sumaTiempos += diffMin
				countTiempos++

				nombreSvc := "Servicio Personalizado"
				if c.Servicio != nil && c.Servicio.Nombre != "" {
					nombreSvc = c.Servicio.Nombre
				} else if c.ServiciosDetalle != nil && *c.ServiciosDetalle != "" {
					nombreSvc = *c.ServiciosDetalle
				}

				if _, ok := porServicio[nombreSvc]; !ok {
					porServicio[nombreSvc] = &svcData{}
				}
				porServicio[nombreSvc].total += diffMin
				porServicio[nombreSvc].count++
			}
		}

		tiempoPromedio := 0
		if countTiempos > 0 {
			tiempoPromedio = int(math.Round(sumaTiempos / float64(countTiempos)))
		}

		tiempoPorServicioList := make([]TiempoServicioReporte, 0, len(porServicio))
		for svcName, data := range porServicio {
			prom := 0
			if data.count > 0 {
				prom = int(math.Round(data.total / float64(data.count)))
			}
			tiempoPorServicioList = append(tiempoPorServicioList, TiempoServicioReporte{
				Servicio: svcName,
				Promedio: prom,
				Cantidad: data.count,
			})
		}
		sort.Slice(tiempoPorServicioList, func(i, j int) bool {
			return tiempoPorServicioList[i].Servicio < tiempoPorServicioList[j].Servicio
		})

		reporte = append(reporte, ReporteProductividadItem{
			TecnicoID:             t.ID,
			Nombre:                t.Nombre,
			Estado:                t.Estado,
			CitasFinalizadas:      len(citas),
			TiempoPromedioMinutos: tiempoPromedio,
			TiempoPorServicio:     tiempoPorServicioList,
		})
	}

	return reporte, nil
}

func (s *AdminService) ReporteServicios(fechaInicio, fechaFin string) ([]ReporteServicioItem, error) {
	type AgrupadoServicio struct {
		ServicioID *int  `gorm:"column:servicio_id"`
		Total      int64 `gorm:"column:total"`
	}

	query := s.db.Model(&models.Cita{}).Select("servicio_id, count(id) as total")

	if fechaInicio != "" {
		if parsed, err := time.Parse("2006-01-02", fechaInicio); err == nil {
			query = query.Where("created_at >= ?", parsed)
		}
	}
	if fechaFin != "" {
		if parsed, err := time.Parse("2006-01-02", fechaFin); err == nil {
			query = query.Where("created_at <= ?", parsed.Add(24*time.Hour))
		}
	}

	var agrupados []AgrupadoServicio
	query.Group("servicio_id").Order("total DESC").Scan(&agrupados)

	resultado := make([]ReporteServicioItem, 0, len(agrupados))
	for _, item := range agrupados {
		if item.ServicioID == nil {
			resultado = append(resultado, ReporteServicioItem{
				ServicioID:         nil,
				Nombre:             "Servicios Personalizados",
				Precio:             nil,
				CantidadSolicitada: item.Total,
			})
			continue
		}

		var servicio models.Servicio
		if err := s.db.First(&servicio, *item.ServicioID).Error; err == nil {
			precio := servicio.Precio
			resultado = append(resultado, ReporteServicioItem{
				ServicioID:         item.ServicioID,
				Nombre:             servicio.Nombre,
				Precio:             &precio,
				CantidadSolicitada: item.Total,
			})
		} else {
			resultado = append(resultado, ReporteServicioItem{
				ServicioID:         item.ServicioID,
				Nombre:             "Desconocido",
				Precio:             nil,
				CantidadSolicitada: item.Total,
			})
		}
	}

	return resultado, nil
}

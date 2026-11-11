package models

import "time"

const (
	EstadoCitaAgendada   = "AGENDADA"
	EstadoCitaAceptada   = "ACEPTADA"
	EstadoCitaEnCurso    = "EN_CURSO"
	EstadoCitaPausada    = "PAUSADA"
	EstadoCitaFinalizada = "FINALIZADA"
	EstadoCitaEntregada  = "ENTREGADA"
	EstadoCitaPospuesta  = "POSPUESTA"
	EstadoCitaCancelada  = "CANCELADA"
)

type Cita struct {
	ID                      int        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ClienteID               int        `gorm:"not null;column:cliente_id;index" json:"clienteId"`
	VehiculoID              int        `gorm:"not null;column:vehiculo_id" json:"vehiculoId"`
	TecnicoID               *int       `gorm:"column:tecnico_id;index" json:"tecnicoId"`
	ServicioID              *int       `gorm:"column:servicio_id" json:"servicioId"`
	Estado                  string     `gorm:"default:'AGENDADA';column:estado;index" json:"estado"`
	HoraAgendada            time.Time  `gorm:"not null;column:hora_agendada;index" json:"horaAgendada"`
	DuracionEstimadaMinutos int        `gorm:"not null;column:duracion_estimada_minutos" json:"duracionEstimadaMinutos"`
	HoraInicioReal          *time.Time `gorm:"column:hora_inicio_real" json:"horaInicioReal"`
	HoraFinReal             *time.Time `gorm:"column:hora_fin_real" json:"horaFinReal"`
	FaseActual              *string    `gorm:"default:'Inspección';column:fase_actual" json:"faseActual"`
	FotoInicio              *string    `gorm:"column:foto_inicio" json:"fotoInicio"`
	FotoFin                 *string    `gorm:"column:foto_fin" json:"fotoFin"`
	FotoInicioExpirada      bool       `gorm:"default:false;column:foto_inicio_expirada" json:"fotoInicioExpirada"`
	FotoFinExpirada         bool       `gorm:"default:false;column:foto_fin_expirada" json:"fotoFinExpirada"`
	MotivoPosposicion       *string    `gorm:"column:motivo_posposicion" json:"motivoPosposicion"`
	EsInmediata             bool       `gorm:"default:false;column:es_inmediata" json:"esInmediata"`
	Adicionales             *string    `gorm:"column:adicionales" json:"adicionales"`
	ServiciosDetalle        *string    `gorm:"column:servicios_detalle" json:"serviciosDetalle"`
	EsPersonalizado         bool       `gorm:"default:false;column:es_personalizado" json:"esPersonalizado"`
	CreatedAt               time.Time  `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	UpdatedAt               time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updatedAt"`

	// Relaciones
	Cliente       *Cliente            `gorm:"foreignKey:ClienteID" json:"cliente,omitempty"`
	Vehiculo      *Vehiculo           `gorm:"foreignKey:VehiculoID" json:"vehiculo,omitempty"`
	Tecnico       *Tecnico            `gorm:"foreignKey:TecnicoID" json:"tecnico,omitempty"`
	Servicio      *Servicio           `gorm:"foreignKey:ServicioID" json:"servicio,omitempty"`
	Accesorios    []CitaAccesorio     `gorm:"foreignKey:CitaID" json:"accesorios,omitempty"`
	FaseHistorial []CitaFaseHistorial `gorm:"foreignKey:CitaID" json:"faseHistorial,omitempty"`
	Mensajes      []Mensaje           `gorm:"foreignKey:CitaID" json:"mensajes,omitempty"`
	Evidencias    []Evidencia         `gorm:"foreignKey:CitaID" json:"evidencias,omitempty"`

	// Campos dinámicos calculados (no persistidos)
	TiempoTranscurrido  int        `gorm:"-" json:"tiempoTranscurrido"`
	TiempoRestante       int        `gorm:"-" json:"tiempoRestante"`
	HoraEstimadaEntrega *time.Time `gorm:"-" json:"horaEstimadaEntrega,omitempty"`
}

func (Cita) TableName() string {
	return "citas"
}

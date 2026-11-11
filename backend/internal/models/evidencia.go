package models

import "time"

const (
	EvidenciaTipoInicio = "INICIO"
	EvidenciaTipoAvance = "AVANCE"
	EvidenciaTipoFin    = "FIN"

	ArchivoTipoImagen = "IMAGEN"
	ArchivoTipoVideo  = "VIDEO"
)

type Evidencia struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CitaID      int       `gorm:"not null;column:cita_id;index" json:"citaId"`
	Tipo        string    `gorm:"not null;column:tipo" json:"tipo"`
	ArchivoURL  string    `gorm:"not null;column:archivo_url" json:"archivoUrl"`
	ArchivoTipo string    `gorm:"not null;column:archivo_type" json:"archivoTipo"`
	Fase        *string   `gorm:"column:fase" json:"fase"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	Cita *Cita `gorm:"foreignKey:CitaID" json:"cita,omitempty"`
}

func (Evidencia) TableName() string {
	return "evidencias"
}

package models

import "time"

type Servicio struct {
	ID                      int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Nombre                  string    `gorm:"not null;column:nombre" json:"nombre"`
	Descripcion             *string   `gorm:"column:descripcion" json:"descripcion"`
	DuracionEstimadaMinutos int       `gorm:"not null;column:duracion_estimada_minutos" json:"duracionEstimadaMinutos"`
	Precio                  float64   `gorm:"type:real;not null;column:precio" json:"precio"`
	Activo                  bool      `gorm:"default:true;column:activo" json:"activo"`
	CreatedAt               time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	Citas []Cita `gorm:"foreignKey:ServicioID" json:"citas,omitempty"`
}

func (Servicio) TableName() string {
	return "servicios"
}

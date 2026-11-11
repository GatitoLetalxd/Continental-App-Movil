package models

import "time"

type CitaFaseHistorial struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CitaID     int       `gorm:"not null;column:cita_id;index" json:"citaId"`
	NombreFase string    `gorm:"not null;column:nombre_fase" json:"nombreFase"`
	Timestamp  time.Time `gorm:"column:timestamp;autoCreateTime" json:"timestamp"`
	Cita       *Cita     `gorm:"foreignKey:CitaID" json:"cita,omitempty"`
}

func (CitaFaseHistorial) TableName() string {
	return "cita_fase_historial"
}

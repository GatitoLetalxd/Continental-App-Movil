package models

import "time"

const (
	EstadoTecnicoActivo   = "ACTIVO"
	EstadoTecnicoInactivo = "INACTIVO"
)

type Tecnico struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Nombre    string    `gorm:"not null;column:nombre" json:"nombre"`
	Usuario   string    `gorm:"uniqueIndex;not null;column:usuario" json:"usuario"`
	Password  string    `gorm:"not null;column:password" json:"-"`
	Estado    string    `gorm:"default:'ACTIVO';column:estado" json:"estado"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	Citas []Cita `gorm:"foreignKey:TecnicoID" json:"citas,omitempty"`
}

func (Tecnico) TableName() string {
	return "tecnicos"
}

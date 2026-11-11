package models

import "time"

type Cliente struct {
	ID        int         `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Nombre    string      `gorm:"not null;column:nombre" json:"nombre"`
	Pin       string      `gorm:"not null;column:pin" json:"-"`
	Contactos StringSlice `gorm:"type:text;column:contactos" json:"contactos"`
	CreatedAt time.Time   `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	Vehiculos []Vehiculo `gorm:"foreignKey:ClienteID" json:"vehiculos,omitempty"`
	Citas     []Cita     `gorm:"foreignKey:ClienteID" json:"citas,omitempty"`
}

func (Cliente) TableName() string {
	return "clientes"
}

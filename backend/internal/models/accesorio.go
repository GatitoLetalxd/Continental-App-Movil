package models

import "time"

type Accesorio struct {
	ID                int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Nombre            string    `gorm:"not null;column:nombre" json:"nombre"`
	Descripcion       *string   `gorm:"column:descripcion" json:"descripcion"`
	Precio            float64   `gorm:"type:real;not null;column:precio" json:"precio"`
	Stock             int       `gorm:"default:0;column:stock" json:"stock"`
	StockMinimoAlerta int       `gorm:"default:5;column:stock_minimo_alerta" json:"stockMinimoAlerta"`
	CreatedAt         time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`

	CitaAccesorios []CitaAccesorio `gorm:"foreignKey:AccesorioID" json:"citaAccesorios,omitempty"`

	// Campos calculados para respuestas API
	StockBajo bool `gorm:"-" json:"stockBajo,omitempty"`
	Deficit   int  `gorm:"-" json:"deficit,omitempty"`
}

func (Accesorio) TableName() string {
	return "accesorios"
}

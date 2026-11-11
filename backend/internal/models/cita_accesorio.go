package models

type CitaAccesorio struct {
	ID          int        `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CitaID      int        `gorm:"not null;column:cita_id;uniqueIndex:idx_cita_accesorio" json:"citaId"`
	AccesorioID int        `gorm:"not null;column:accesorio_id;uniqueIndex:idx_cita_accesorio" json:"accesorioId"`
	Cantidad    int        `gorm:"default:1;column:cantidad" json:"cantidad"`
	Cita        *Cita      `gorm:"foreignKey:CitaID" json:"cita,omitempty"`
	Accesorio   *Accesorio `gorm:"foreignKey:AccesorioID" json:"accesorio,omitempty"`
}

func (CitaAccesorio) TableName() string {
	return "cita_accesorios"
}

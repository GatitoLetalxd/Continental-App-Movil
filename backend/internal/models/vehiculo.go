package models

type Vehiculo struct {
	ID        int      `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Marca     string   `gorm:"not null;column:marca" json:"marca"`
	Modelo    string   `gorm:"not null;column:modelo" json:"modelo"`
	Anio      int      `gorm:"not null;column:anio" json:"anio"`
	Placa     string   `gorm:"not null;column:placa" json:"placa"`
	Notas     *string  `gorm:"column:notas" json:"notas"`
	ClienteID int      `gorm:"not null;column:cliente_id" json:"clienteId"`
	Cliente   *Cliente `gorm:"foreignKey:ClienteID" json:"cliente,omitempty"`
	Citas     []Cita   `gorm:"foreignKey:VehiculoID" json:"citas,omitempty"`
}

func (Vehiculo) TableName() string {
	return "vehiculos"
}

package models

import "time"

const (
	AutorTipoCliente = "CLIENTE"
	AutorTipoTecnico = "TECNICO"

	TipoMensajeTexto  = "TEXTO"
	TipoMensajeImagen = "IMAGEN"
	TipoMensajeVideo  = "VIDEO"
)

type Mensaje struct {
	ID         int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	CitaID     int       `gorm:"not null;column:cita_id;index" json:"citaId"`
	AutorTipo  string    `gorm:"not null;column:autor_tipo" json:"autorTipo"`
	AutorID    int       `gorm:"not null;column:autor_id" json:"autorId"`
	Contenido  *string   `gorm:"column:contenido" json:"contenido"`
	Tipo       string    `gorm:"default:'TEXTO';column:tipo" json:"tipo"`
	ArchivoURL *string   `gorm:"column:archivo_url" json:"archivoUrl"`
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
	ExpiraAt   time.Time `gorm:"column:expira_at;not null" json:"expiraAt"`
	Expirado   bool      `gorm:"default:false;column:expirado" json:"expirado"`

	Cita *Cita `gorm:"foreignKey:CitaID" json:"cita,omitempty"`
}

func (Mensaje) TableName() string {
	return "mensajes"
}

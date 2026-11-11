package models

import "time"

type Admin struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Usuario   string    `gorm:"uniqueIndex;not null;column:usuario" json:"usuario"`
	Password  string    `gorm:"not null;column:password" json:"-"`
	Nombre    string    `gorm:"default:'Administrador';column:nombre" json:"nombre"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"createdAt"`
}

func (Admin) TableName() string {
	return "admins"
}

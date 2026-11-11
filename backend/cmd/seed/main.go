package main

import (
	"fmt"
	"log"

	"continental-backend/internal/config"
	"continental-backend/internal/database"
	"continental-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	fmt.Println("🌱 Iniciando seed para SQLite...")

	cfg := config.LoadConfig()
	db, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Error conectando a SQLite: %v", err)
	}

	// 1. Admin
	adminHash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), 10)
	var admin models.Admin
	if err := db.Where("usuario = ?", "admin").First(&admin).Error; err != nil {
		admin = models.Admin{
			Usuario:  "admin",
			Password: string(adminHash),
			Nombre:   "Administrador Continental",
		}
		db.Create(&admin)
		fmt.Printf("✅ Admin creado: %s (contraseña: admin123)\n", admin.Usuario)
	}

	// 2. Técnicos
	tecHash, _ := bcrypt.GenerateFromPassword([]byte("tec123"), 10)
	tecnicos := []models.Tecnico{
		{Nombre: "Carlos Rodríguez", Usuario: "carlos", Password: string(tecHash), Estado: "ACTIVO"},
		{Nombre: "Miguel Torres", Usuario: "miguel", Password: string(tecHash), Estado: "ACTIVO"},
	}
	for _, t := range tecnicos {
		var existente models.Tecnico
		if err := db.Where("usuario = ?", t.Usuario).First(&existente).Error; err != nil {
			db.Create(&t)
			fmt.Printf("✅ Técnico creado: %s (usuario: %s, contraseña: tec123)\n", t.Nombre, t.Usuario)
		}
	}

	// 3. Servicios
	servicios := []models.Servicio{
		{Nombre: "Tapizado completo de asientos", Descripcion: ptrStr("Retapizado completo de todos los asientos del vehículo con material a elección del cliente."), DuracionEstimadaMinutos: 480, Precio: 150.00, Activo: true},
		{Nombre: "Tapizado de puertas", Descripcion: ptrStr("Retapizado de paneles de puertas con material coordinado."), DuracionEstimadaMinutos: 240, Precio: 80.00, Activo: true},
		{Nombre: "Instalación de sistema de audio", Descripcion: ptrStr("Instalación completa de sistema de audio automotriz: reproductor, amplificador y parlantes."), DuracionEstimadaMinutos: 180, Precio: 60.00, Activo: true},
		{Nombre: "Instalación eléctrica de luces LED", Descripcion: ptrStr("Instalación de luces LED interiores y exteriores."), DuracionEstimadaMinutos: 120, Precio: 45.00, Activo: true},
		{Nombre: "Instalación de alarma", Descripcion: ptrStr("Instalación de sistema de alarma con control remoto."), DuracionEstimadaMinutos: 90, Precio: 35.00, Activo: true},
		{Nombre: "Instalación de accesorios generales", Descripcion: ptrStr("Instalación de accesorios varios: protectores, guardafangos, estribos, etc."), DuracionEstimadaMinutos: 60, Precio: 25.00, Activo: true},
	}
	for _, s := range servicios {
		var existente models.Servicio
		if err := db.Where("nombre = ?", s.Nombre).First(&existente).Error; err != nil {
			db.Create(&s)
		}
	}
	fmt.Printf("✅ %d servicios verificados/creados\n", len(servicios))

	// 4. Accesorios
	accesorios := []models.Accesorio{
		{Nombre: "Cuero sintético negro", Descripcion: ptrStr("Rollo de cuero sintético, color negro"), Precio: 25.00, Stock: 15, StockMinimoAlerta: 3},
		{Nombre: "Cuero sintético beige", Descripcion: ptrStr("Rollo de cuero sintético, color beige"), Precio: 25.00, Stock: 10, StockMinimoAlerta: 3},
		{Nombre: "Espuma de alta densidad", Descripcion: ptrStr("Plancha de espuma para asientos"), Precio: 8.00, Stock: 20, StockMinimoAlerta: 5},
		{Nombre: "Kit parlantes 6x9", Descripcion: ptrStr("Par de parlantes 6x9 pulgadas"), Precio: 35.00, Stock: 8, StockMinimoAlerta: 2},
		{Nombre: "Amplificador 4 canales", Descripcion: ptrStr("Amplificador de audio 4 canales 1200W"), Precio: 45.00, Stock: 5, StockMinimoAlerta: 2},
		{Nombre: "Cinta LED RGB 5m", Descripcion: ptrStr("Cinta LED RGB de 5 metros con controlador"), Precio: 12.00, Stock: 12, StockMinimoAlerta: 3},
		{Nombre: "Alarma con sensor", Descripcion: ptrStr("Sistema de alarma con sensor de movimiento"), Precio: 28.00, Stock: 6, StockMinimoAlerta: 2},
		{Nombre: "Cable automotriz calibre 12", Descripcion: ptrStr("Rollo de 10m de cable automotriz"), Precio: 5.00, Stock: 30, StockMinimoAlerta: 10},
	}
	for _, a := range accesorios {
		var existente models.Accesorio
		if err := db.Where("nombre = ?", a.Nombre).First(&existente).Error; err != nil {
			db.Create(&a)
		}
	}
	fmt.Printf("✅ %d accesorios verificados/creados\n", len(accesorios))

	// 5. Cliente de prueba
	clientePin, _ := bcrypt.GenerateFromPassword([]byte("1234"), 10)
	var cliente models.Cliente
	if err := db.Where("nombre = ?", "Juan Pérez").First(&cliente).Error; err != nil {
		cliente = models.Cliente{
			Nombre:    "Juan Pérez",
			Pin:       string(clientePin),
			Contactos: []string{"584121234567", "584129876543"},
		}
		db.Create(&cliente)
		fmt.Printf("✅ Cliente creado: %s (PIN: 1234)\n", cliente.Nombre)

		// Vehículo de prueba
		vehiculo := models.Vehiculo{
			Marca:     "Toyota",
			Modelo:    "Corolla",
			Anio:      2020,
			Placa:     "ABC123",
			Notas:     ptrStr("Sin modificaciones previas"),
			ClienteID: cliente.ID,
		}
		db.Create(&vehiculo)
		fmt.Println("✅ Vehículo de prueba creado: Toyota Corolla 2020")
	}

	fmt.Println("\n🎉 Seed completado exitosamente en SQLite!")
}

func ptrStr(s string) *string {
	return &s
}

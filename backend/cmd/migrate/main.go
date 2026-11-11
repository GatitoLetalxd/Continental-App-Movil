package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"continental-backend/internal/config"
	"continental-backend/internal/database"
	"continental-backend/internal/models"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	fmt.Println("🚀 [Migrate] Iniciando migración de datos PostgreSQL → SQLite...")

	cfg := config.LoadConfig()

	// 1. Conectar a SQLite
	sqliteDB, err := database.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Error conectando a SQLite: %v", err)
	}

	// 2. Conectar a PostgreSQL con database/sql
	pgDB, err := sql.Open("pgx", cfg.PGDatabaseURL)
	if err != nil {
		log.Fatalf("❌ Error conectando a PostgreSQL: %v", err)
	}
	defer pgDB.Close()

	if err := pgDB.Ping(); err != nil {
		log.Fatalf("❌ Error al hacer ping a PostgreSQL: %v", err)
	}
	fmt.Println("✅ Conexión exitosa a PostgreSQL y SQLite")

	// 3. Migrar Admins
	var totalAdmins int
	rows, err := pgDB.Query(`SELECT id, usuario, password, nombre, created_at FROM admins ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo admins: %v", err)
	}
	for rows.Next() {
		var a models.Admin
		if err := rows.Scan(&a.ID, &a.Usuario, &a.Password, &a.Nombre, &a.CreatedAt); err != nil {
			log.Fatalf("Error escaneando admin: %v", err)
		}
		if err := sqliteDB.Save(&a).Error; err != nil {
			log.Fatalf("Error guardando admin en SQLite: %v", err)
		}
		totalAdmins++
	}
	rows.Close()
	fmt.Printf("✅ Admins migrados: %d\n", totalAdmins)

	// 4. Migrar Tecnicos
	var totalTecnicos int
	rows, err = pgDB.Query(`SELECT id, nombre, usuario, password, estado, created_at FROM tecnicos ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo tecnicos: %v", err)
	}
	for rows.Next() {
		var t models.Tecnico
		if err := rows.Scan(&t.ID, &t.Nombre, &t.Usuario, &t.Password, &t.Estado, &t.CreatedAt); err != nil {
			log.Fatalf("Error escaneando tecnico: %v", err)
		}
		if err := sqliteDB.Save(&t).Error; err != nil {
			log.Fatalf("Error guardando tecnico en SQLite: %v", err)
		}
		totalTecnicos++
	}
	rows.Close()
	fmt.Printf("✅ Técnicos migrados: %d\n", totalTecnicos)

	// 5. Migrar Clientes (usando array_to_json para contactos)
	var totalClientes int
	rows, err = pgDB.Query(`SELECT id, nombre, pin, COALESCE(array_to_json(contactos)::text, '[]'), created_at FROM clientes ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo clientes: %v", err)
	}
	for rows.Next() {
		var c models.Cliente
		var contactosJSON string
		if err := rows.Scan(&c.ID, &c.Nombre, &c.Pin, &contactosJSON, &c.CreatedAt); err != nil {
			log.Fatalf("Error escaneando cliente: %v", err)
		}
		var arr []string
		if err := json.Unmarshal([]byte(contactosJSON), &arr); err == nil {
			c.Contactos = arr
		} else {
			c.Contactos = []string{}
		}
		if err := sqliteDB.Save(&c).Error; err != nil {
			log.Fatalf("Error guardando cliente en SQLite: %v", err)
		}
		totalClientes++
	}
	rows.Close()
	fmt.Printf("✅ Clientes migrados: %d\n", totalClientes)

	// 6. Migrar Vehiculos
	var totalVehiculos int
	rows, err = pgDB.Query(`SELECT id, marca, modelo, anio, placa, notas, cliente_id FROM vehiculos ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo vehiculos: %v", err)
	}
	for rows.Next() {
		var v models.Vehiculo
		if err := rows.Scan(&v.ID, &v.Marca, &v.Modelo, &v.Anio, &v.Placa, &v.Notas, &v.ClienteID); err != nil {
			log.Fatalf("Error escaneando vehiculo: %v", err)
		}
		if err := sqliteDB.Save(&v).Error; err != nil {
			log.Fatalf("Error guardando vehiculo en SQLite: %v", err)
		}
		totalVehiculos++
	}
	rows.Close()
	fmt.Printf("✅ Vehículos migrados: %d\n", totalVehiculos)

	// 7. Migrar Servicios
	var totalServicios int
	rows, err = pgDB.Query(`SELECT id, nombre, descripcion, duracion_estimada_minutos, precio, activo, created_at FROM servicios ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo servicios: %v", err)
	}
	for rows.Next() {
		var s models.Servicio
		if err := rows.Scan(&s.ID, &s.Nombre, &s.Descripcion, &s.DuracionEstimadaMinutos, &s.Precio, &s.Activo, &s.CreatedAt); err != nil {
			log.Fatalf("Error escaneando servicio: %v", err)
		}
		if err := sqliteDB.Save(&s).Error; err != nil {
			log.Fatalf("Error guardando servicio en SQLite: %v", err)
		}
		totalServicios++
	}
	rows.Close()
	fmt.Printf("✅ Servicios migrados: %d\n", totalServicios)

	// 8. Migrar Accesorios
	var totalAccesorios int
	rows, err = pgDB.Query(`SELECT id, nombre, descripcion, precio, stock, stock_minimo_alerta, created_at FROM accesorios ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo accesorios: %v", err)
	}
	for rows.Next() {
		var a models.Accesorio
		if err := rows.Scan(&a.ID, &a.Nombre, &a.Descripcion, &a.Precio, &a.Stock, &a.StockMinimoAlerta, &a.CreatedAt); err != nil {
			log.Fatalf("Error escaneando accesorio: %v", err)
		}
		if err := sqliteDB.Save(&a).Error; err != nil {
			log.Fatalf("Error guardando accesorio en SQLite: %v", err)
		}
		totalAccesorios++
	}
	rows.Close()
	fmt.Printf("✅ Accesorios migrados: %d\n", totalAccesorios)

	// 9. Migrar Citas
	var totalCitas int
	rows, err = pgDB.Query(`SELECT id, cliente_id, vehiculo_id, tecnico_id, servicio_id, estado, hora_agendada, duracion_estimada_minutos, hora_inicio_real, hora_fin_real, fase_actual, foto_inicio, foto_fin, foto_inicio_expirada, foto_fin_expirada, motivo_posposicion, es_inmediata, adicionales, servicios_detalle, es_personalizado, created_at, updated_at FROM citas ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo citas: %v", err)
	}
	for rows.Next() {
		var c models.Cita
		if err := rows.Scan(
			&c.ID, &c.ClienteID, &c.VehiculoID, &c.TecnicoID, &c.ServicioID,
			&c.Estado, &c.HoraAgendada, &c.DuracionEstimadaMinutos,
			&c.HoraInicioReal, &c.HoraFinReal, &c.FaseActual,
			&c.FotoInicio, &c.FotoFin, &c.FotoInicioExpirada, &c.FotoFinExpirada,
			&c.MotivoPosposicion, &c.EsInmediata, &c.Adicionales,
			&c.ServiciosDetalle, &c.EsPersonalizado, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			log.Fatalf("Error escaneando cita: %v", err)
		}
		if err := sqliteDB.Save(&c).Error; err != nil {
			log.Fatalf("Error guardando cita en SQLite: %v", err)
		}
		totalCitas++
	}
	rows.Close()
	fmt.Printf("✅ Citas migradas: %d\n", totalCitas)

	// 10. Migrar CitaAccesorios
	var totalCitaAccesorios int
	rows, err = pgDB.Query(`SELECT id, cita_id, accesorio_id, cantidad FROM cita_accesorios ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo cita_accesorios: %v", err)
	}
	for rows.Next() {
		var ca models.CitaAccesorio
		if err := rows.Scan(&ca.ID, &ca.CitaID, &ca.AccesorioID, &ca.Cantidad); err != nil {
			log.Fatalf("Error escaneando cita_accesorio: %v", err)
		}
		if err := sqliteDB.Save(&ca).Error; err != nil {
			log.Fatalf("Error guardando cita_accesorio en SQLite: %v", err)
		}
		totalCitaAccesorios++
	}
	rows.Close()
	fmt.Printf("✅ CitaAccesorios migrados: %d\n", totalCitaAccesorios)

	// 11. Migrar CitaFaseHistorial
	var totalFases int
	rows, err = pgDB.Query(`SELECT id, cita_id, nombre_fase, timestamp FROM cita_fase_historial ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo cita_fase_historial: %v", err)
	}
	for rows.Next() {
		var f models.CitaFaseHistorial
		if err := rows.Scan(&f.ID, &f.CitaID, &f.NombreFase, &f.Timestamp); err != nil {
			log.Fatalf("Error escaneando fase: %v", err)
		}
		if err := sqliteDB.Save(&f).Error; err != nil {
			log.Fatalf("Error guardando fase en SQLite: %v", err)
		}
		totalFases++
	}
	rows.Close()
	fmt.Printf("✅ CitaFaseHistorial migrados: %d\n", totalFases)

	// 12. Migrar Mensajes
	var totalMensajes int
	rows, err = pgDB.Query(`SELECT id, cita_id, autor_tipo, autor_id, contenido, tipo, archivo_url, created_at, expira_at, expirado FROM mensajes ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo mensajes: %v", err)
	}
	for rows.Next() {
		var m models.Mensaje
		if err := rows.Scan(&m.ID, &m.CitaID, &m.AutorTipo, &m.AutorID, &m.Contenido, &m.Tipo, &m.ArchivoURL, &m.CreatedAt, &m.ExpiraAt, &m.Expirado); err != nil {
			log.Fatalf("Error escaneando mensaje: %v", err)
		}
		if err := sqliteDB.Save(&m).Error; err != nil {
			log.Fatalf("Error guardando mensaje en SQLite: %v", err)
		}
		totalMensajes++
	}
	rows.Close()
	fmt.Printf("✅ Mensajes migrados: %d\n", totalMensajes)

	// 13. Migrar Evidencias
	var totalEvidencias int
	rows, err = pgDB.Query(`SELECT id, cita_id, tipo, archivo_url, archivo_type, fase, created_at FROM evidencias ORDER BY id ASC`)
	if err != nil {
		log.Fatalf("Error leyendo evidencias: %v", err)
	}
	for rows.Next() {
		var e models.Evidencia
		if err := rows.Scan(&e.ID, &e.CitaID, &e.Tipo, &e.ArchivoURL, &e.ArchivoTipo, &e.Fase, &e.CreatedAt); err != nil {
			log.Fatalf("Error escaneando evidencia: %v", err)
		}
		if err := sqliteDB.Save(&e).Error; err != nil {
			log.Fatalf("Error guardando evidencia en SQLite: %v", err)
		}
		totalEvidencias++
	}
	rows.Close()
	fmt.Printf("✅ Evidencias migradas: %d\n", totalEvidencias)

	fmt.Println("\n🎉 ¡Migración de PostgreSQL a SQLite completada exitosamente!")
}

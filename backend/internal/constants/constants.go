package constants

// Fases por defecto de una cita (plantilla)
var FasesPorDefecto = []string{"Inspección", "Desmontaje", "Instalación", "Finalizado"}

// Límites y configuraciones
const (
	MaxCaracteresFase           = 30
	DiasExpiracionFotosEvidencia = 30
	DiasExpiracionChatMultimedia = 7
	MaxVideosPorChat            = 2
	MaxTamanoVideoCombinadoMB   = 250
	MaxTamanoFotoMB             = 50
	MaxContactosCliente         = 2
)

// Roles del sistema
const (
	RoleAdmin   = "ADMIN"
	RoleTecnico = "TECNICO"
	RoleCliente = "CLIENTE"
)

// Plantillas de mensajes WhatsApp (generación de wa.me)
var PlantillasWhatsApp = map[string]string{
	"CAMBIO_HORA_ESTIMADA": "Hola {nombre}, le informamos que la hora estimada de entrega de su vehículo {vehiculo} ha sido actualizada a {nuevaHora}. Disculpe las molestias. — Taller Continental",
	"CITA_POSPUESTA":       "Hola {nombre}, lamentamos informarle que su cita para el servicio \"{servicio}\" ha sido pospuesta. Nos pondremos en contacto para reprogramarla con prioridad. Disculpe las molestias. — Taller Continental",
	"LISTA_PARA_ENTREGA":   "Hola {nombre}, ¡su vehículo {vehiculo} ya está listo para retirar! Le esperamos en el taller. ¡Gracias por su confianza! — Taller Continental",
}

// Transiciones válidas de la máquina de estados de cita
var TransicionesCita = map[string][]string{
	"AGENDADA":   {"ACEPTADA", "CANCELADA"},
	"ACEPTADA":   {"EN_CURSO", "POSPUESTA", "CANCELADA"},
	"EN_CURSO":   {"PAUSADA", "FINALIZADA", "POSPUESTA", "CANCELADA"},
	"PAUSADA":    {"EN_CURSO", "POSPUESTA", "CANCELADA"},
	"FINALIZADA": {"ENTREGADA"},
	"ENTREGADA":  {},
	"POSPUESTA":  {"ACEPTADA"}, // Reasignación
	"CANCELADA":  {},
}

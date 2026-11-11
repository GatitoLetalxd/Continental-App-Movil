package services

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"continental-backend/internal/constants"
	"continental-backend/internal/database"
	"continental-backend/internal/models"
	"continental-backend/internal/utils"

	"gorm.io/gorm"
)

type WhatsAppService struct {
	db *gorm.DB
}

func NewWhatsAppService() *WhatsAppService {
	return &WhatsAppService{db: database.DB}
}

type WhatsAppLinkItem struct {
	Telefono string `json:"telefono"`
	Link     string `json:"link"`
}

type GenerarLinkWhatsAppResponse struct {
	Tipo string `json:"tipo"`
	Cita struct {
		ID       int    `json:"id"`
		Cliente  string `json:"cliente"`
		Vehiculo string `json:"vehiculo"`
		Servicio string `json:"servicio"`
	} `json:"cita"`
	Links []WhatsAppLinkItem `json:"links"`
}

func (s *WhatsAppService) GenerarLinkWa(telefono, plantilla string, datos map[string]string) string {
	mensaje := plantilla
	for k, v := range datos {
		mensaje = strings.ReplaceAll(mensaje, fmt.Sprintf("{%s}", k), v)
	}

	re := regexp.MustCompile(`\D`)
	numeroLimpio := re.ReplaceAllString(telefono, "")
	mensajeCodificado := url.QueryEscape(mensaje)

	return fmt.Sprintf("https://wa.me/%s?text=%s", numeroLimpio, mensajeCodificado)
}

func (s *WhatsAppService) GenerarLinkParaCita(citaId int, tipo string) (*GenerarLinkWhatsAppResponse, error) {
	var cita models.Cita
	err := s.db.Preload("Cliente").Preload("Vehiculo").Preload("Servicio").First(&cita, citaId).Error
	if err != nil {
		return nil, utils.NewNotFoundError("Cita")
	}

	if cita.Cliente == nil || len(cita.Cliente.Contactos) == 0 {
		return nil, utils.NewValidationError("El cliente no tiene números de contacto registrados")
	}

	vehiculoInfo := "Vehículo"
	if cita.Vehiculo != nil {
		vehiculoInfo = fmt.Sprintf("%s %s", cita.Vehiculo.Marca, cita.Vehiculo.Modelo)
		if strings.TrimSpace(cita.Vehiculo.Placa) != "" {
			vehiculoInfo = fmt.Sprintf("%s %s (%s)", cita.Vehiculo.Marca, cita.Vehiculo.Modelo, cita.Vehiculo.Placa)
		}
	}

	nombreServicio := "Servicio Personalizado"
	if cita.Servicio != nil && cita.Servicio.Nombre != "" {
		nombreServicio = cita.Servicio.Nombre
	} else if cita.ServiciosDetalle != nil && *cita.ServiciosDetalle != "" {
		nombreServicio = *cita.ServiciosDetalle
	}

	horaEntrega := utils.CalcularHoraEstimadaEntrega(cita.HoraInicioReal, cita.HoraAgendada, cita.DuracionEstimadaMinutos)
	nuevaHora := utils.FormatearHora(horaEntrega)

	datos := map[string]string{
		"nombre":    cita.Cliente.Nombre,
		"vehiculo":  vehiculoInfo,
		"servicio":  nombreServicio,
		"nuevaHora": nuevaHora,
	}

	var plantilla string
	switch tipo {
	case "cambio_hora":
		plantilla = constants.PlantillasWhatsApp["CAMBIO_HORA_ESTIMADA"]
	case "pospuesta":
		plantilla = constants.PlantillasWhatsApp["CITA_POSPUESTA"]
	case "lista_entrega":
		plantilla = constants.PlantillasWhatsApp["LISTA_PARA_ENTREGA"]
	default:
		return nil, utils.NewValidationError("Tipo de mensaje no válido. Use: cambio_hora, pospuesta, lista_entrega")
	}

	links := make([]WhatsAppLinkItem, 0, len(cita.Cliente.Contactos))
	for _, contacto := range cita.Cliente.Contactos {
		link := s.GenerarLinkWa(contacto, plantilla, datos)
		links = append(links, WhatsAppLinkItem{
			Telefono: contacto,
			Link:     link,
		})
	}

	resp := &GenerarLinkWhatsAppResponse{
		Tipo: tipo,
		Links: links,
	}
	resp.Cita.ID = cita.ID
	resp.Cita.Cliente = cita.Cliente.Nombre
	resp.Cita.Vehiculo = vehiculoInfo
	resp.Cita.Servicio = nombreServicio

	return resp, nil
}

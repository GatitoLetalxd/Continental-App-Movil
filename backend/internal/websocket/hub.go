package websocket

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"continental-backend/internal/config"
	"continental-backend/internal/constants"
	"continental-backend/internal/middleware"
	"continental-backend/internal/services"

	"github.com/golang-jwt/jwt/v5"
	socketio "github.com/googollee/go-socket.io"
	"github.com/googollee/go-socket.io/engineio"
	"github.com/googollee/go-socket.io/engineio/transport"
	"github.com/googollee/go-socket.io/engineio/transport/polling"
	"github.com/googollee/go-socket.io/engineio/transport/websocket"
)

func parseCitaID(val interface{}) int {
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		id, _ := strconv.Atoi(v)
		return id
	default:
		return 0
	}
}

type SocketHub struct {
	Server      *socketio.Server
	chatService *services.ChatService
}

var Hub *SocketHub

func InitSocketHub() (*SocketHub, error) {
	server := socketio.NewServer(&engineio.Options{
		Transports: []transport.Transport{
			&polling.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
			&websocket.Transport{
				CheckOrigin: func(r *http.Request) bool { return true },
			},
		},
	})

	hub := &SocketHub{
		Server:      server,
		chatService: services.NewChatService(),
	}

	server.OnConnect("/", func(s socketio.Conn) error {
		// Extraer token de múltiples fuentes: query params o headers HTTP
		url := s.URL()
		token := url.Query().Get("token")
		if token == "" {
			token = url.Query().Get("auth")
		}
		if token == "" {
			authHeader := s.RemoteHeader().Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else if authHeader != "" {
				token = authHeader
			}
		}
		if token == "" {
			token = s.RemoteHeader().Get("token")
		}
		if token == "" {
			token = s.RemoteHeader().Get("x-token")
		}

		if token == "" {
			log.Println("[Socket.io] Conexión rechazada: token no proporcionado")
			return fmt.Errorf("token no proporcionado")
		}

		// Validar JWT
		claims := &middleware.JWTClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !parsedToken.Valid {
			log.Printf("[Socket.io] Conexión rechazada: token inválido (%v)", err)
			return fmt.Errorf("token inválido")
		}

		s.SetContext(claims)
		log.Printf("[Socket.io] Conectado: %s (%s) - Socket ID: %s", claims.Nombre, claims.Rol, s.ID())

		// Unir a rooms según rol
		if claims.Rol == constants.RoleAdmin {
			s.Join("admin")
		} else if claims.Rol == constants.RoleTecnico {
			s.Join(fmt.Sprintf("tecnico:%d", claims.ID))
		}

		return nil
	})

	server.OnEvent("/", "cita:join", func(s socketio.Conn, rawCitaId interface{}) {
		citaId := parseCitaID(rawCitaId)
		if citaId == 0 {
			return
		}
		room := fmt.Sprintf("cita:%d", citaId)
		s.Join(room)
		claims, _ := s.Context().(*middleware.JWTClaims)
		nombre := "Usuario"
		if claims != nil {
			nombre = claims.Nombre
		}
		log.Printf("[Socket.io] %s se unió a %s", nombre, room)
	})

	server.OnEvent("/", "cita:leave", func(s socketio.Conn, rawCitaId interface{}) {
		citaId := parseCitaID(rawCitaId)
		if citaId == 0 {
			return
		}
		room := fmt.Sprintf("cita:%d", citaId)
		s.Leave(room)
		claims, _ := s.Context().(*middleware.JWTClaims)
		nombre := "Usuario"
		if claims != nil {
			nombre = claims.Nombre
		}
		log.Printf("[Socket.io] %s salió de %s", nombre, room)
	})

	type EnviarMensajeData struct {
		CitaID    interface{} `json:"citaId"`
		Contenido string      `json:"contenido"`
	}

	server.OnEvent("/", "chat:enviarMensaje", func(s socketio.Conn, data EnviarMensajeData) {
		claims, ok := s.Context().(*middleware.JWTClaims)
		if !ok || claims == nil {
			s.Emit("error", map[string]string{"message": "No autenticado"})
			return
		}

		citaId := parseCitaID(data.CitaID)
		if citaId == 0 {
			s.Emit("error", map[string]string{"message": "citaId inválido"})
			return
		}

		autorTipo := constants.RoleCliente
		if claims.Rol != constants.RoleCliente {
			autorTipo = constants.RoleTecnico
		}

		msg, err := hub.chatService.EnviarMensajeTexto(citaId, autorTipo, claims.ID, data.Contenido)
		if err != nil {
			s.Emit("error", map[string]string{"message": err.Error()})
			return
		}

		hub.Server.BroadcastToRoom("/", fmt.Sprintf("cita:%d", citaId), "chat:nuevoMensaje", map[string]interface{}{
			"citaId":  citaId,
			"mensaje": msg,
		})
	})

	server.OnError("/", func(s socketio.Conn, e error) {
		if !strings.Contains(e.Error(), "connection closed") {
			log.Printf("[Socket.io] Error en socket: %v", e)
		}
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		claims, _ := s.Context().(*middleware.JWTClaims)
		nombre := "Desconocido"
		rol := ""
		if claims != nil {
			nombre = claims.Nombre
			rol = claims.Rol
		}
		log.Printf("[Socket.io] Desconectado: %s (%s) - Motivo: %s", nombre, rol, reason)
	})

	Hub = hub
	return hub, nil
}

// Métodos de emisión convenientes para los handlers
func (h *SocketHub) EmitirEstadoCambiado(citaId int, nuevoEstado string) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", fmt.Sprintf("cita:%d", citaId), "cita:estadoCambiado", map[string]interface{}{
			"citaId":      citaId,
			"nuevoEstado": nuevoEstado,
			"timestamp":   time.Now().Format(time.RFC3339),
		})
	}
}

func (h *SocketHub) EmitirFaseCambiada(citaId int, nuevaFase string) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", fmt.Sprintf("cita:%d", citaId), "cita:faseCambiada", map[string]interface{}{
			"citaId":    citaId,
			"nuevaFase": nuevaFase,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (h *SocketHub) EmitirEvidenciaSubida(citaId int, evidencia any) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", fmt.Sprintf("cita:%d", citaId), "cita:evidenciaSubida", map[string]interface{}{
			"citaId":    citaId,
			"evidencia": evidencia,
		})
	}
}

func (h *SocketHub) EmitirDisponibilidadCambiada(tecnicoId int) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", "admin", "tecnico:disponibilidadCambiada", map[string]interface{}{
			"tecnicoId": tecnicoId,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (h *SocketHub) EmitirCitaPospuesta(citaId int) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", "admin", "cita:pospuesta", map[string]interface{}{
			"citaId":    citaId,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func (h *SocketHub) EmitirNuevoMensajeChat(citaId int, mensaje any) {
	if h != nil && h.Server != nil {
		h.Server.BroadcastToRoom("/", fmt.Sprintf("cita:%d", citaId), "chat:nuevoMensaje", map[string]interface{}{
			"citaId":  citaId,
			"mensaje": mensaje,
		})
	}
}

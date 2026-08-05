# Especificación de Construcción — App de Gestión de Taller

> Este documento es un prompt de construcción para un agente de IA (ej. Claude Code). Contiene todas las decisiones de producto ya cerradas. El agente debe seguir esta especificación como fuente de verdad, no inventar comportamiento no descrito aquí sin antes preguntar.

---

## 1. Resumen del proyecto

Aplicación de gestión operativa para un taller (tapizado, electricidad automotriz, accesorios) con 3 roles: **Técnico**, **Cliente**, **Administrador**. El sistema coordina citas, seguimiento en tiempo real del trabajo, comunicación cliente-técnico, e inventario básico.

El diseño de UI/UX se construirá con un sistema de diseño externo ("UI UX Pro Max") — este documento no prescribe estilos visuales, solo estructura, datos y comportamiento. El agente debe dejar la capa visual desacoplada de la lógica para poder aplicar el sistema de diseño sin reescribir lógica de negocio.

---

## 2. Stack técnico (decisiones finales, no renegociables sin aprobación)

| Capa | Tecnología |
|---|---|
| Frontend móvil (Android) | Flutter, compilado a `.apk` |
| Frontend web (PWA, iOS/PC) | Flutter Web, mismo código base |
| Comunicación en tiempo real | Socket.io (cliente Flutter ↔ servidor Node.js) |
| Backend | Node.js + Express, arquitectura en capas (Rutas → Controladores → Servicios → Modelos) |
| ORM | Prisma |
| Base de datos | PostgreSQL |
| Reverse proxy / TLS | Nginx + Let's Encrypt (ya configurado en el VPS) |
| Almacenamiento de archivos | Sistema de archivos local del VPS (Contabo), con job de expiración automática |
| Notificaciones externas | Enlaces `wa.me` con mensaje preescrito, envío **manual** (el sistema genera el link, un humano hace clic) |
| Dominio | Subdominio bajo `*.moondev.online` |
| Backups | `pg_dump` automatizado por cron (diario) |
| Despliegue | Directo a producción, con una semana de pruebas internas antes de onboarding de clientes reales |

**Importante:** NO integrar ninguna API oficial ni no oficial de WhatsApp para envío automático de mensajes. El único mecanismo de "notificación externa" es generar un link `https://wa.me/<numero>?text=<mensaje_urlencoded>` que un humano (técnico o admin) abre y envía manualmente.

---

## 3. Roles y autenticación

Todos los roles comparten el mismo sistema de login (un solo formulario, el backend determina el rol tras autenticar).

### 3.1 Cliente
- Registro con: nombre completo + PIN numérico simple (sin verificación por SMS/WhatsApp/email).
- Puede tener **múltiples vehículos**.
- Un vehículo pertenece a **un único cliente** (no compartido entre cuentas).
- Puede registrar **hasta 2 números de contacto**.

### 3.2 Técnico
- Login con usuario + contraseña/PIN simple (definido por el admin al crearlo).
- Estados posibles: `activo`, `inactivo`.
- **Soft delete obligatorio**: nunca se borra un técnico de la base de datos. Al "eliminar" desde el panel admin, se marca `inactivo`:
  - No puede loguearse.
  - No aparece en la lista de técnicos disponibles para nuevas citas.
  - Su historial de citas pasadas permanece intacto y visible en reportes.
  - Puede reactivarse (`activo` de nuevo) sin perder historial.

### 3.3 Administrador
- Un solo usuario admin (el dueño del taller / el propio desarrollador), sin necesidad de multi-admin en v1.
- Login con usuario + contraseña.
- Acceso completo de lectura/escritura a todas las entidades.

---

## 4. Modelo de datos (entidades principales)

Diseñar el schema de Prisma con al menos estas entidades y relaciones. El agente puede ajustar nombres de campos a convención pero debe respetar la cardinalidad descrita.

### `Cliente`
- id, nombre, pin (hasheado), contactos (1 a 2 teléfonos), createdAt.
- Relación: 1 Cliente → N Vehiculo.

### `Vehiculo`
- id, marca, modelo, año, placa, notas (alertas de modificaciones previas), clienteId (FK, obligatorio, un solo cliente).

### `Tecnico`
- id, nombre, usuario, password (hasheada), estado (`activo` | `inactivo`), createdAt.
- Relación: 1 Tecnico → N Cita (histórico, incluso si el técnico pasa a inactivo).

### `Servicio`
- id, nombre, descripcion, duracionEstimadaMinutos (base), precio, activo (boolean, admin puede desactivar sin borrar).
- CRUD completo gestionado por el Administrador (crear, editar, desactivar).

### `Accesorio`
- id, nombre, descripcion, precio, stock, stockMinimoAlerta.
- CRUD completo por el Administrador.

### `Cita`
- id, clienteId, vehiculoId, tecnicoId (nullable si aún no asignada), servicioId.
- estado: ver máquina de estados en sección 5.
- horaAgendada (timestamp), duracionEstimadaMinutos (editable por el técnico tras inspección).
- horaInicioReal (timestamp, se guarda al presionar "Iniciar" — NO usar contador en memoria, calcular tiempo transcurrido siempre por diferencia de timestamps).
- horaFinReal (timestamp).
- faseActual (texto libre, ver sección 4.1).
- accesorios (relación N:M con Accesorio, vía tabla intermedia `CitaAccesorio` con cantidad).
- fotoInicio, fotoFin (URLs/paths a archivos, ver sección 7).

### `CitaFaseHistorial`
- id, citaId, nombreFase, timestamp.
- Registra cada cambio de fase (incluida la fase "por defecto" original), para que los reportes de productividad no dependan solo del texto libre actual.

### `Mensaje` (chat cliente-técnico)
- id, citaId, autorTipo (`cliente` | `tecnico`), contenido (texto, o referencia a archivo adjunto), tipo (`texto` | `imagen` | `video`), createdAt, expiraAt (createdAt + 7 días).

### 4.1 Fases de una cita
- Fases por defecto (plantilla, no tabla separada, puede ser un enum o lista fija en config): `Inspección` → `Desmontaje` → `Instalación` → `Finalizado`.
- El técnico puede sobrescribir el nombre de la fase actual con texto libre (máx. 30 caracteres) para reflejar mejor lo que está haciendo.
- Cada cambio de fase (por defecto o renombrada) se registra en `CitaFaseHistorial` para no perder trazabilidad aunque el texto visible cambie.

---

## 5. Máquina de estados de la Cita

Estados: `Agendada` → `Aceptada` → `EnCurso` → `Pausada` → `Finalizada` → `Entregada`
Estados alternativos: `Pospuesta`, `Cancelada`.

Reglas de transición:
- `Agendada`: creada por el cliente (o por el admin/técnico si el cliente está presencialmente en el taller).
- `Agendada` → `Aceptada`: un técnico disponible la acepta manualmente. **No hay auto-asignación.** Un técnico puede ver citas sin asignar y aceptarlas.
- `Aceptada` → `EnCurso`: el técnico presiona "Iniciar". Se guarda `horaInicioReal`.
- `EnCurso` → `Pausada`: el técnico presiona "Pausar" (ej. esperando secado de materiales). El técnico **puede** tomar otra cita mientras está en pausa, pero el sistema debe mostrar una advertencia ("Tienes una cita en pausa, cuida tus tiempos para entregar ambas a tiempo").
- `Pausada` → `EnCurso`: reanudar.
- `EnCurso` → `Finalizada`: el técnico presiona "Finalizar", obligatorio subir `fotoFin` (cámara o galería, ambas opciones habilitadas). Se guarda `horaFinReal`.
- `Finalizada` → `Entregada`: marcado manual cuando el cliente retira el vehículo (puede ser el mismo técnico o el admin quien lo marque).
- Cualquier estado activo (`Aceptada`, `EnCurso`, `Pausada`) → `Pospuesta`: **solo el técnico puede posponer**, nunca el cliente directamente (el cliente debe comunicarse con el técnico si tiene una situación). Al posponer:
  - Los accesorios reservados para la cita **permanecen reservados** (no se liberan stock).
  - El admin ve la cita en una bandeja de "pendientes de reasignar" y la reasigna **manualmente** a otro técnico o fecha. Un técnico también puede ver y aceptar directamente una cita pospuesta sin depender del admin.
- Cualquier estado activo → `Cancelada`: **solo el técnico o el admin puede cancelar**, nunca el cliente. Al cancelar, los accesorios reservados **se liberan** de vuelta al stock.

### 5.1 Cálculo de tiempo (importante)
No usar un cronómetro corriendo en memoria del servidor. El tiempo transcurrido/estimado se calcula siempre comparando `horaInicioReal` (o `now()`) contra `horaAgendada + duracionEstimadaMinutos`. Esto permite que si el servidor se reinicia, no se pierda ni se desincronice ningún conteo — todo se recalcula desde los timestamps guardados en base de datos.

### 5.2 Citas inmediatas
Si un cliente está físicamente en el taller y hay un técnico disponible, se permite agendar una cita con `horaAgendada = now()` (sin anticipación mínima en ese caso específico). Para agendamiento remoto (no presencial) sí debe exigirse una anticipación mínima configurable (definir un valor por defecto razonable, ej. 2 horas, y dejarlo como constante fácil de cambiar).

---

## 6. Tiempo real y notificaciones

### 6.1 Eventos Socket.io (app abierta)
- Cambio de estado de cita (para que el cliente vea el status tracker actualizarse en vivo).
- Cambio de fase (`faseActual`).
- Nuevo mensaje en el chat de una cita.
- Actualización de disponibilidad de técnico (para el dashboard del admin).

### 6.2 Notificaciones push (FCM) — app cerrada
Usar para eventos críticos que el usuario debe recibir aunque no tenga la app abierta:
- Técnico: nueva cita asignada/reasignada a él.
- Cliente: cambio de fase importante, cita finalizada / lista para entrega.
- Admin: cita pospuesta que requiere reasignación.

No usar push para eventos de bajo valor (ej. cada micro-cambio de cronómetro).

### 6.3 Notificación externa vía WhatsApp (manual)
Para los siguientes eventos, el sistema debe generar un botón/link `wa.me` con mensaje preescrito, listo para que un humano lo envíe con un clic:
- Cambio de hora estimada de entrega tras ajuste del técnico.
- Cita pospuesta (mensaje empático ofreciendo reprogramar con prioridad).
- Cita lista para entrega.

Definir las plantillas de texto de estos 3 mensajes como constantes editables (no hardcodeadas en múltiples lugares).

---

## 7. Manejo de archivos

### 7.1 Fotos de evidencia (inicio/fin de trabajo)
- Almacenamiento: sistema de archivos local del VPS (no R2, ya que hay espacio suficiente).
- Origen permitido: cámara integrada **o** galería (ambas opciones habilitadas en la UI).
- **Job de expiración**: cron diario que borra fotos con más de 30 días de antigüedad y limpia la referencia en base de datos (o marca `fotoExpirada = true` si se prefiere conservar el registro sin el archivo).

### 7.2 Chat multimedia
- Máximo por cliente/conversación: 2 videos, con un tope combinado de 250MB entre ambos.
- Fotos: máximo 50MB cada una.
- Expiración: 7 días después de enviado, luego se elimina el archivo del disco.
- Botón "Guardar en el dispositivo" visible en cada imagen/video del chat, para que el usuario lo respalde antes de que expire.
- El chat es **privado**: solo el cliente y el técnico asignado a esa cita lo ven. El admin no tiene acceso a las conversaciones salvo que se agregue esa función explícitamente en el futuro (no incluir en v1).

---

## 8. Rol Técnico — Funciones (App Android/Web)

- Vista de agenda con bloques de tiempo del día.
- Lista de citas sin asignar disponibles para aceptar.
- Botones de estado: Aceptar → Iniciar → Pausar/Reanudar → Finalizar.
- Slider para ajustar `duracionEstimadaMinutos` tras inspección inicial.
- Botón "Posponer" con confirmación y campo opcional de motivo.
- Botón "Cancelar" con confirmación (libera accesorios reservados).
- Captura de foto (cámara o galería) obligatoria antes de Iniciar y antes de Finalizar.
- Edición del nombre de la fase actual (texto libre, máx. 30 caracteres), con opción de volver a la plantilla por defecto.
- Chat con el cliente de la cita activa.

## 9. Rol Cliente — Funciones (App Web/PWA)

- Registro simple (nombre + PIN), gestión de hasta 2 contactos.
- Gestión de perfil de vehículo(s): marca, modelo, año, placa, notas.
- Agendar cita: elegir servicio + agregar accesorios al carrito (solo cotización, **sin pago dentro de la app**).
- Ver línea de tiempo del proyecto (status tracker) con las fases nombradas (plantilla o texto libre del técnico).
- Chat multimedia con el técnico asignado.
- El cliente **no puede** cancelar ni posponer citas desde la app — debe comunicarse con el técnico o el taller directamente para cualquier cambio.

## 10. Rol Administrador — Funciones (Panel Web PC)

- Dashboard operativo: estado en tiempo real de cada técnico (libre / en curso / pausado).
- Bandeja de citas pospuestas pendientes de reasignar (reasignación siempre manual).
- CRUD de técnicos (crear, editar, marcar inactivo/activo — nunca borrar).
- CRUD de servicios (crear, editar, activar/desactivar).
- CRUD de accesorios (crear, editar, ajustar stock).
- Alertas de stock bajo (umbral configurable por accesorio, `stockMinimoAlerta`).
- Reportes: productividad por técnico (citas finalizadas, tiempo promedio por servicio) y servicios más solicitados (por período de tiempo, filtrable).

---

## 11. Reglas de negocio — resumen de casos borde

1. Reasignación de citas: **siempre manual**, nunca automática. Un técnico también puede auto-asignarse una cita pospuesta o sin asignar.
2. Accesorios reservados: se **liberan** al cancelar, se **mantienen reservados** al posponer.
3. Un técnico puede tener una cita en pausa y otra en curso simultáneamente, con advertencia visual de que debe cuidar sus tiempos.
4. Citas inmediatas (sin anticipación) solo permitidas si el cliente está presencialmente en el taller y hay técnico disponible.
5. El cliente nunca puede cancelar/posponer directamente — toda gestión de cambios pasa por el técnico o el admin.
6. Nada se borra físicamente de la base de datos salvo archivos expirados (fotos >30 días, multimedia de chat >7 días). Entidades de negocio (técnicos, citas, clientes) usan soft delete o simplemente cambian de estado.

---

## 12. Infraestructura y despliegue

- Backend Node.js corriendo detrás de Nginx (ya configurado con SSL vía Let's Encrypt) bajo un subdominio de `moondev.online`.
- Base de datos PostgreSQL en el mismo VPS Contabo, gestionada vía Prisma.
- Configurar Nginx con los headers de upgrade necesarios para que las conexiones WebSocket de Socket.io funcionen correctamente detrás del reverse proxy (`Upgrade`, `Connection`, `proxy_http_version 1.1`).
- Cron jobs necesarios:
  - `pg_dump` diario de la base de datos (backup manual, sin automatización de restauración — el admin decide cuándo restaurar).
  - Limpieza diaria de fotos de evidencia con más de 30 días.
  - Limpieza diaria de multimedia de chat con más de 7 días.
- Plan de contingencia: si el VPS cae, el taller opera manualmente sin la app (no requiere fallback digital automatizado).
- Despliegue directo a producción; una semana de pruebas internas antes de dar acceso a clientes reales.

---

## 13. Orden sugerido de implementación para el agente

1. Setup de proyecto: backend Express + Prisma + Postgres, schema completo de la sección 4.
2. Autenticación de los 3 roles (login unificado, sesiones/JWT).
3. CRUD de Técnico, Servicio, Accesorio (panel admin, sin UI aún, probar vía API).
4. Máquina de estados de Cita completa (sección 5) con tests de las transiciones válidas/inválidas.
5. Sistema de fases y su historial (sección 4.1).
6. Integración de Socket.io para eventos en tiempo real (sección 6.1).
7. Sistema de archivos: subida de fotos de evidencia + jobs de expiración (sección 7.1).
8. Chat multimedia con límites y expiración (sección 7.2).
9. Generación de links `wa.me` con plantillas (sección 6.3).
10. Frontend Flutter: pantallas por rol (secciones 8, 9, 10), aplicando el sistema de diseño UI UX Pro Max sobre la lógica ya construida.
11. Reportes del panel admin (sección 10).
12. Cron jobs de backup y limpieza (sección 12).
13. Despliegue a producción y pruebas internas de una semana.

---

## 14. Fuera de alcance para v1 (no construir salvo que se pida explícitamente)

- Pagos dentro de la app.
- API oficial/no oficial de WhatsApp para envío automático.
- Multi-admin.
- Acceso del admin a los chats cliente-técnico.
- Modo offline completo para el técnico (solo reconexión tras corte, no operación offline extendida).
- Restauración automática de backups.

# Requerimientos — Continental Taller App

> Documento de requerimientos del sistema de gestión operativa para **Taller Continental** (tapizado, electricidad automotriz y accesorios). Incluye requerimientos funcionales implementados, no funcionales e ideas para versiones futuras.

---

## Índice

1. [Requerimientos Funcionales Implementados](#1-requerimientos-funcionales-implementados)
2. [Requerimientos No Funcionales](#2-requerimientos-no-funcionales)
3. [Reglas de Negocio Críticas](#3-reglas-de-negocio-críticas)
4. [Fuera de Alcance en v1 (por implementar a futuro)](#4-fuera-de-alcance-en-v1-por-implementar-a-futuro)

---

## 1. Requerimientos Funcionales Implementados

### 1.1 Autenticación y Roles

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-01 | Login unificado para los 3 roles desde un único formulario; el backend determina el rol automáticamente | ✅ |
| RF-02 | **Cliente**: registro con nombre completo + PIN numérico (4 dígitos), sin verificación por SMS ni email | ✅ |
| RF-03 | **Técnico**: login con usuario + contraseña; creado por el Administrador | ✅ |
| RF-04 | **Administrador**: login con usuario + contraseña; acceso total al sistema | ✅ |
| RF-05 | Tokens JWT con expiración, incluidos en cada request autenticado | ✅ |
| RF-06 | Middleware de autorización por rol (`requireRole`) en cada ruta sensible | ✅ |
| RF-07 | Técnico en estado `INACTIVO` no puede iniciar sesión | ✅ |

---

### 1.2 Gestión de Clientes y Vehículos

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-08 | El cliente puede registrar hasta **2 números de contacto** (teléfono) | ✅ |
| RF-09 | Login de cliente posible con número de contacto o nombre | ✅ |
| RF-10 | El cliente puede gestionar múltiples vehículos (marca, modelo, año, placa, notas de modificaciones previas) | ✅ |
| RF-11 | Un vehículo pertenece a un único cliente (no compartido entre cuentas) | ✅ |
| RF-12 | CRUD de vehículos disponible para el cliente y el admin | ✅ |

---

### 1.3 Gestión de Citas

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-13 | El cliente puede agendar una cita seleccionando servicio y accesorios (cotización, sin pago en la app) | ✅ |
| RF-14 | Soporte de **citas inmediatas** (`esInmediata = true`) cuando el cliente está físicamente en el taller | ✅ |
| RF-15 | Agendamiento remoto requiere anticipación mínima configurable (por defecto: 2 horas) | ✅ |
| RF-16 | El técnico o admin puede crear citas en nombre del cliente | ✅ |
| RF-17 | Citas sin técnico asignado accesibles para que cualquier técnico disponible las tome | ✅ |
| RF-18 | El técnico puede ajustar `duracionEstimadaMinutos` tras inspección inicial | ✅ |
| RF-19 | La cita registra `horaInicioReal` y `horaFinReal`; el tiempo transcurrido se calcula por diferencia de timestamps, nunca por contador en memoria | ✅ |
| RF-20 | Campo `adicionales` y `serviciosDetalle` para notas extra en la cita | ✅ |
| RF-21 | Soporte de servicio personalizado (`esPersonalizado = true`) sin selección de catálogo | ✅ |
| RF-22 | El admin y el técnico pueden reasignar manualmente una cita a otro técnico | ✅ |
| RF-23 | Generación de link `wa.me` con mensaje preescrito para notificaciones manuales vía WhatsApp | ✅ |

---

### 1.4 Máquina de Estados de la Cita

| ID | Transición | Quién puede ejecutarla | Estado |
|----|-----------|------------------------|--------|
| RF-24 | `AGENDADA` → `ACEPTADA` | Técnico (acepta una cita libre) | ✅ |
| RF-25 | `ACEPTADA` → `EN_CURSO` | Técnico (presiona "Iniciar") | ✅ |
| RF-26 | `EN_CURSO` → `PAUSADA` | Técnico | ✅ |
| RF-27 | `PAUSADA` → `EN_CURSO` | Técnico (reanudar) | ✅ |
| RF-28 | `EN_CURSO` → `FINALIZADA` | Técnico (requiere subir foto de fin) | ✅ |
| RF-29 | `FINALIZADA` → `ENTREGADA` | Técnico o Admin | ✅ |
| RF-30 | Cualquier estado activo → `POSPUESTA` | Solo el Técnico | ✅ |
| RF-31 | Cualquier estado activo → `CANCELADA` | Técnico o Admin (nunca el cliente) | ✅ |
| RF-32 | Al posponer: accesorios reservados permanecen sin liberar | ✅ |
| RF-33 | Al cancelar: accesorios reservados se devuelven al stock | ✅ |
| RF-34 | Advertencia visual cuando el técnico tiene una cita en pausa y acepta otra | ✅ |

---

### 1.5 Fases del Trabajo

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-35 | Fases por defecto: `Inspección` → `Desmontaje` → `Instalación` → `Finalizado` | ✅ |
| RF-36 | El técnico puede renombrar la fase actual con texto libre (máx. 30 caracteres) | ✅ |
| RF-37 | El técnico puede avanzar a la siguiente fase de la plantilla, o resetear a la primera | ✅ |
| RF-38 | Cada cambio de fase se registra en `CitaFaseHistorial` con timestamp (trazabilidad total) | ✅ |
| RF-39 | El cliente ve la línea de tiempo de fases en tiempo real vía Socket.io | ✅ |

---

### 1.6 Chat Cliente-Técnico

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-40 | Chat privado por cita entre el cliente y el técnico asignado | ✅ |
| RF-41 | Soporte de mensajes de texto, imagen y video | ✅ |
| RF-42 | Límite: máx. 2 videos por conversación, tope combinado de 250 MB | ✅ |
| RF-43 | Imágenes: máx. 50 MB por archivo | ✅ |
| RF-44 | Mensajes con archivo expiran a los 7 días (archivo borrado del disco, registro marcado como expirado) | ✅ |
| RF-45 | Botón para guardar imagen/video en el dispositivo antes de que expire | ✅ |
| RF-46 | Mensajes llegan en tiempo real vía Socket.io mientras la app está abierta | ✅ |

---

### 1.7 Gestión de Accesorios e Inventario

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-47 | CRUD completo de accesorios por el Administrador | ✅ |
| RF-48 | Cada accesorio tiene `stock` y `stockMinimoAlerta` configurables | ✅ |
| RF-49 | El stock se libera automáticamente cuando una cita es cancelada | ✅ |
| RF-50 | Alerta de stock bajo cuando el stock cae por debajo del umbral configurado | ✅ |
| RF-51 | Relación N:M entre `Cita` y `Accesorio` vía tabla `CitaAccesorio` con cantidad | ✅ |

---

### 1.8 Panel de Administración

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-52 | Dashboard con estado en tiempo real de cada técnico (libre / en curso / pausado) | ✅ |
| RF-53 | Bandeja de citas pospuestas pendientes de reasignación | ✅ |
| RF-54 | CRUD de técnicos: crear, editar, marcar inactivo/activo (soft delete, nunca borrar físicamente) | ✅ |
| RF-55 | Historial de citas de técnicos inactivos permanece visible e intacto | ✅ |
| RF-56 | CRUD de servicios: crear, editar, activar/desactivar (sin borrar físicamente) | ✅ |
| RF-57 | Reporte de productividad por técnico: citas finalizadas y tiempo promedio por servicio | ✅ |
| RF-58 | Reporte de servicios más solicitados filtrable por período | ✅ |

---

### 1.9 Notificaciones y Tiempo Real

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-59 | Socket.io emite eventos al cliente cuando cambia el estado de su cita | ✅ |
| RF-60 | Socket.io emite eventos al cambiar la fase actual de la cita | ✅ |
| RF-61 | Socket.io emite eventos de nuevo mensaje en el chat | ✅ |
| RF-62 | Socket.io actualiza el dashboard del admin con disponibilidad de técnicos en tiempo real | ✅ |
| RF-63 | Generación de links `wa.me` para 3 eventos: cambio de hora estimada, cita pospuesta, cita lista para entrega | ✅ |

---

### 1.10 Manejo de Archivos y Evidencias

| ID | Requerimiento | Estado |
|----|---------------|--------|
| RF-64 | Foto obligatoria de inicio del trabajo (`fotoInicio`) antes de poner una cita `EN_CURSO` | ✅ |
| RF-65 | Foto obligatoria de fin del trabajo (`fotoFin`) antes de `FINALIZAR` | ✅ |
| RF-66 | Origen permitido: cámara integrada o galería del dispositivo | ✅ |
| RF-67 | Las fotos de evidencia se almacenan en el sistema de archivos del VPS | ✅ |
| RF-68 | Las fotos de evidencia expiran y se marcan como `expirada = true` pasados 30 días | ✅ |
| RF-69 | El técnico puede subir evidencias adicionales (imágenes/videos) por fase durante el trabajo | ✅ |
| RF-70 | Protección contra path traversal en el servicio de archivos | ✅ |

---

### 1.11 Tareas Automáticas (Cron Jobs)

| ID | Tarea | Horario | Estado |
|----|-------|---------|--------|
| RF-71 | Backup diario de PostgreSQL con `pg_dump` | 02:00 AM | ✅ |
| RF-72 | Limpieza de fotos de evidencia con más de 30 días | 03:00 AM | ✅ |
| RF-73 | Limpieza de multimedia de chat con más de 7 días | 03:30 AM | ✅ |

---

## 2. Requerimientos No Funcionales

| ID | Categoría | Requerimiento |
|----|-----------|---------------|
| RNF-01 | **Seguridad** | Contraseñas y PINs almacenados con hash bcrypt (10 rondas de sal) |
| RNF-02 | **Seguridad** | Autenticación basada en JWT; cada ruta sensible valida el token y el rol |
| RNF-03 | **Seguridad** | Validación de inputs con `Joi` antes de que lleguen a los servicios |
| RNF-04 | **Seguridad** | Protección contra path traversal en el servicio de archivos estáticos |
| RNF-05 | **Disponibilidad** | Backend gestionado con PM2 para auto-reinicio ante caídas |
| RNF-06 | **Disponibilidad** | Nginx como reverse proxy con SSL/TLS (Let's Encrypt) y headers de upgrade para WebSocket |
| RNF-07 | **Disponibilidad** | Backups diarios automáticos; restauración manual por el administrador |
| RNF-08 | **Rendimiento** | El tiempo transcurrido de una cita se calcula siempre por diferencia de timestamps (sobrevive a reinicios del servidor) |
| RNF-09 | **Rendimiento** | Índices en PostgreSQL sobre columnas de alta frecuencia de consulta: `estado`, `tecnicoId`, `clienteId`, `horaAgendada` de la tabla `citas` |
| RNF-10 | **Rendimiento** | Pool de conexiones a la base de datos gestionado por Prisma Client |
| RNF-11 | **Escalabilidad** | Arquitectura en capas desacoplada (Rutas → Controladores → Servicios → ORM) |
| RNF-12 | **Mantenibilidad** | Variables de entorno centralizadas en `.env`; ninguna credencial hardcodeada en el código fuente |
| RNF-13 | **Mantenibilidad** | Soft delete para entidades de negocio; nada se borra físicamente salvo archivos expirados |
| RNF-14 | **Integridad de datos** | Los accesorios de una cita pospuesta permanecen reservados; solo se liberan al cancelar |
| RNF-15 | **Multiplataforma** | Frontend en Flutter compilado como APK Android y como Web (PWA) desde el mismo código base |
| RNF-16 | **UX** | Los estados de cita se muestran con color y etiqueta diferenciada en la UI |
| RNF-17 | **Infraestructura** | VPS Contabo con PostgreSQL local y almacenamiento de archivos local |
| RNF-18 | **Límites de archivos** | Videos en chat: máx. 2 por conversación / 250 MB combinados. Imágenes: máx. 50 MB cada una |
| RNF-19 | **Privacidad** | El chat es privado entre cliente y técnico; el Admin no tiene acceso en v1 |
| RNF-20 | **Trazabilidad** | Cada cambio de fase queda registrado en `CitaFaseHistorial` con timestamp preciso |

---

## 3. Reglas de Negocio Críticas

1. **Reasignación manual**: nunca automática. Un técnico puede auto-asignarse citas sin asignar o pospuestas.
2. **Accesorios**: se liberan al cancelar, se mantienen reservados al posponer.
3. **Técnico con cita en pausa**: puede aceptar otra cita simultáneamente, con advertencia visual.
4. **Citas inmediatas**: solo permitidas para clientes presenciales con técnico disponible.
5. **Cancelación**: solo el técnico o el admin pueden cancelar. El cliente nunca puede.
6. **Posposición**: solo el técnico puede posponer. El cliente debe contactar al taller para cualquier cambio.
7. **Integridad histórica**: los datos históricos (citas pasadas, técnicos inactivos) nunca se borran.
8. **WhatsApp**: el sistema genera el link; un humano siempre hace clic y envía. No hay integración automática con la API de WhatsApp.

---

## 4. Fuera de Alcance en v1 (por implementar a futuro)

| ID | Funcionalidad | Justificación / Notas |
|----|---------------|----------------------|
| FUT-01 | **Notificaciones Push (FCM)** | Requiere Firebase Cloud Messaging. Prioritario para v2: el técnico recibe alertas con la app cerrada |
| FUT-02 | **Pagos dentro de la app** | Requiere pasarela de pago (Stripe, PayU, MercadoPago) e implica cambios en el flujo de entrega |
| FUT-03 | **Multi-admin** | Más de un usuario administrador con permisos diferenciados |
| FUT-04 | **Acceso del Admin al chat** | Para soporte y mediación de conflictos entre cliente y técnico |
| FUT-05 | **Modo offline para Técnico** | Registrar avances sin conexión y sincronizar al reconectarse |
| FUT-06 | **Restauración automática de backups** | Script de restauración + interfaz en el panel admin |
| FUT-07 | **Integración oficial WhatsApp Business API** | Envío automático de notificaciones sin intervención manual |
| FUT-08 | **App nativa iOS** | Flutter Web cubre iOS como PWA; app nativa mejoraría la experiencia |
| FUT-09 | **Facturación y cotización PDF** | Generación automática de presupuesto/factura descargable |
| FUT-10 | **Portal de citas público** | Página pública para agendar sin descargar la app |
| FUT-11 | **Reportes avanzados con exportación** | Exportación a Excel/PDF; métricas de ingresos y materiales |
| FUT-12 | **Sistema de calificaciones** | El cliente califica el servicio al recibir su vehículo (1-5 estrellas) |
| FUT-13 | **Historial de precios de accesorios** | Tracking de variaciones de precio en el tiempo |
| FUT-14 | **Gestión de proveedores** | CRUD de proveedores asociados a accesorios para reposición de stock |
| FUT-15 | **Recordatorios automáticos de cita** | Recordatorio al cliente 24 h y 2 h antes de su cita agendada |

# Arquitectura — Continental Taller App

> Descripción técnica de la arquitectura del sistema de gestión de **Taller Continental**. Documento de referencia para desarrolladores.

---

## Índice

1. [Visión General](#1-visión-general)
2. [Stack Tecnológico](#2-stack-tecnológico)
3. [Arquitectura de Capas (Backend)](#3-arquitectura-de-capas-backend)
4. [Estructura de Directorios](#4-estructura-de-directorios)
5. [API REST — Endpoints](#5-api-rest--endpoints)
6. [Comunicación en Tiempo Real (Socket.io)](#6-comunicación-en-tiempo-real-socketio)
7. [Autenticación y Autorización](#7-autenticación-y-autorización)
8. [Manejo de Archivos](#8-manejo-de-archivos)
9. [Tareas Programadas (Cron Jobs)](#9-tareas-programadas-cron-jobs)
10. [Infraestructura y Despliegue](#10-infraestructura-y-despliegue)
11. [Flujo de Datos — Ciclo de una Cita](#11-flujo-de-datos--ciclo-de-una-cita)

---

## 1. Visión General

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENTES                                 │
│                                                                 │
│  [App Android APK]   [Flutter Web / PWA]   [Panel Admin Web]   │
└──────────┬───────────────────┬───────────────────┬─────────────┘
           │  HTTPS / WSS      │  HTTPS / WSS      │  HTTPS / WSS
           ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────────┐
│                   NGINX (Reverse Proxy + SSL)                   │
│              continental.moondev.online (Let's Encrypt)         │
└────────────────────────────┬────────────────────────────────────┘
                             │  HTTP / WS (interno)
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│               BACKEND  Node.js + Express  (Puerto 3000)         │
│  ┌─────────┐ ┌─────────────┐ ┌───────────────┐ ┌────────────┐ │
│  │  Rutas  │ │Controladores│ │   Servicios   │ │  Socket.io │ │
│  │  (9)    │ │   (9)       │ │   (12)        │ │  (eventos) │ │
│  └────┬────┘ └──────┬──────┘ └───────┬───────┘ └────────────┘ │
│       └─────────────┴────────────────┘                         │
│                        │ Prisma ORM                             │
└────────────────────────┼────────────────────────────────────────┘
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│              PostgreSQL (local en VPS Contabo)                  │
│              11 tablas + 3 enums + 4 índices                    │
└─────────────────────────────────────────────────────────────────┘
                         ▲
              ┌──────────┘
              │  Sistema de archivos VPS
              │  /backend/uploads/
              │  (fotos evidencia + multimedia chat)
              └──────────────────────────────────
```

---

## 2. Stack Tecnológico

| Capa | Tecnología | Versión | Justificación |
|------|-----------|---------|---------------|
| **Frontend móvil** | Flutter (Android APK) | SDK ^3.2.0 | Un solo código base para Android y Web |
| **Frontend web / PWA** | Flutter Web | SDK ^3.2.0 | Mismo repositorio, distinto target de compilación |
| **Backend API** | Node.js + Express | Express ^5.2.1 | Ecosistema maduro, rendimiento suficiente para la escala del taller |
| **ORM** | Prisma | ^6.19.3 | Migraciones versionadas, tipado fuerte, queries seguros |
| **Base de datos** | PostgreSQL | local VPS | ACID compliance, soporte nativo de arrays (contactos), buen rendimiento |
| **Tiempo real** | Socket.io | ^4.8.3 | Abstrae WebSocket con fallback, compatible con Flutter vía `socket_io_client` |
| **Autenticación** | JWT (jsonwebtoken) | ^9.0.3 | Stateless, no requiere sesión del lado servidor |
| **Hashing** | bcryptjs | ^3.0.3 | Estándar de la industria para contraseñas |
| **Validación** | Joi | ^18.2.3 | Validación declarativa de esquemas en los controladores |
| **Upload de archivos** | Multer | ^2.2.0 | Middleware para multipart/form-data |
| **Cron jobs** | node-cron | ^4.6.0 | Tareas de limpieza y backup sin dependencias externas |
| **Reverse proxy / SSL** | Nginx + Let's Encrypt | Sistema | Terminación TLS, headers WebSocket, Gzip, SPA fallback |
| **Process manager** | PM2 | Sistema | Auto-reinicio, logs persistentes, gestión de procesos |
| **Backups** | pg_dump | Sistema | Backup diario de PostgreSQL, archivo .sql comprimido |

### Dependencias Flutter (frontend)

| Paquete | Uso |
|---------|-----|
| `http ^1.2.1` | Llamadas REST al backend |
| `socket_io_client ^3.0.2` | Eventos en tiempo real |
| `image_picker ^1.1.2` | Cámara y galería para fotos de evidencia |
| `intl ^0.19.0` | Formateo de fechas y horas |
| `url_launcher ^6.3.0` | Abrir links `wa.me` en WhatsApp |
| `shared_preferences ^2.2.3` | Persistencia local del token JWT |
| `provider ^6.1.2` | Gestión de estado reactivo |

---

## 3. Arquitectura de Capas (Backend)

El backend sigue una arquitectura en **4 capas** estrictas con responsabilidades separadas:

```
Petición HTTP
     │
     ▼
┌────────────┐
│   RUTAS    │  Declara endpoints y aplica middlewares (auth, validación, upload)
│ /routes/   │  No contiene lógica de negocio
└─────┬──────┘
      │
      ▼
┌──────────────────┐
│  CONTROLADORES   │  Parsea req/res, llama al servicio, devuelve respuesta HTTP
│ /controllers/    │  Manejo básico de errores HTTP; delega lógica al servicio
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│    SERVICIOS     │  Toda la lógica de negocio aquí: validaciones, transiciones
│  /services/      │  de estado, cálculos, llamadas a múltiples modelos
└────────┬─────────┘
         │
         ▼
┌──────────────────┐
│   PRISMA ORM     │  Acceso a datos. Único lugar donde se usa PrismaClient.
│  /config/db      │  Instancia singleton compartida en todo el backend
└──────────────────┘
```

### Middlewares

| Middleware | Archivo | Función |
|-----------|---------|---------|
| `verifyToken` | `middleware/auth.js` | Valida el JWT en `Authorization: Bearer ...` |
| `requireRole(...)` | `middleware/auth.js` | Comprueba que el rol del token esté en la lista permitida |
| `validate(schema)` | `middleware/validate.js` | Valida el body con un esquema Joi antes de continuar |
| `uploadEvidencia` | `middleware/upload.js` | Multer configurado para fotos de evidencia (JPEG/PNG, max 50MB) |
| `uploadEvidenciaGeneral` | `middleware/upload.js` | Multer para evidencias generales (incluye video) |
| Error handler global | `app.js` | Captura errores operacionales y no operacionales |

---

## 4. Estructura de Directorios

```
ContinentalApp/
├── backend/
│   ├── prisma/
│   │   ├── schema.prisma          # Definición del modelo de datos
│   │   ├── seed.js                # Datos iniciales (admin, técnicos, servicios, accesorios)
│   │   └── migrations/            # Historial de migraciones de la BD
│   ├── src/
│   │   ├── app.js                 # Entry point: Express, Socket.io, rutas, cron jobs
│   │   ├── config/
│   │   │   ├── database.js        # Singleton de PrismaClient
│   │   │   └── constants.js       # Roles, fases por defecto, límites, etc.
│   │   ├── controllers/           # 9 controladores (uno por dominio)
│   │   ├── services/              # 12 servicios con la lógica de negocio
│   │   ├── routes/                # 9 archivos de rutas Express
│   │   ├── middleware/
│   │   │   ├── auth.js            # verifyToken + requireRole
│   │   │   ├── validate.js        # Validación Joi
│   │   │   └── upload.js          # Multer (fotos y multimedia)
│   │   ├── socket/
│   │   │   └── index.js           # Configuración de Socket.io y handlers de eventos
│   │   ├── jobs/
│   │   │   ├── backupDb.js        # pg_dump automatizado
│   │   │   ├── cleanupPhotos.js   # Limpia fotos de evidencia >30 días
│   │   │   └── cleanupChat.js     # Limpia multimedia de chat >7 días
│   │   └── utils/
│   │       ├── errors.js          # Clases de error operacional (UnauthorizedError, etc.)
│   │       └── ...
│   ├── uploads/                   # Archivos subidos (git-ignorado)
│   ├── backups/                   # Backups pg_dump (git-ignorado)
│   ├── .env                       # Variables de entorno (git-ignorado)
│   └── package.json
│
├── frontend/
│   ├── lib/
│   │   ├── main.dart              # Entry point Flutter + Provider setup
│   │   ├── config/
│   │   │   └── app_config.dart    # URL base de la API, timeouts
│   │   ├── services/
│   │   │   └── api_service.dart   # Cliente HTTP centralizado (HTTPS + JWT)
│   │   ├── providers/             # Gestión de estado con Provider
│   │   ├── screens/
│   │   │   ├── login_screen.dart       # Pantalla de login unificado
│   │   │   ├── client_screen.dart      # Portal del cliente (citas, vehículos, chat)
│   │   │   ├── technician_screen.dart  # Portal del técnico (agenda, estados, fases)
│   │   │   └── admin_screen.dart       # Panel admin (dashboard, CRUD, reportes)
│   │   └── widgets/               # Componentes reutilizables
│   ├── assets/
│   │   └── images/                # Recursos gráficos
│   ├── android/                   # Proyecto Android nativo
│   │   └── app/src/main/AndroidManifest.xml  # Permisos: INTERNET, ACCESS_NETWORK_STATE
│   ├── web/                       # Build target web
│   └── pubspec.yaml
│
├── requerimientos.md
├── arquitectura.md
├── modelo_datos.md
└── spec_construccion_taller.md    # Spec original del producto
```

---

## 5. API REST — Endpoints

Base URL: `https://continental.moondev.online/api`

### Autenticación

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/auth/login` | Público | Login unificado; retorna JWT + rol |
| POST | `/auth/register` | Público | Registro de cliente (nombre + PIN) |

### Clientes

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| GET | `/clientes` | ADMIN | Listar todos los clientes |
| GET | `/clientes/:id` | ADMIN, CLIENTE | Obtener cliente por ID |
| PATCH | `/clientes/:id` | ADMIN, CLIENTE | Actualizar datos del cliente |

### Vehículos

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/vehiculos` | ADMIN, CLIENTE | Crear vehículo |
| GET | `/vehiculos/cliente/:clienteId` | ADMIN, CLIENTE | Listar vehículos de un cliente |
| PATCH | `/vehiculos/:id` | ADMIN, CLIENTE | Actualizar vehículo |
| DELETE | `/vehiculos/:id` | ADMIN, CLIENTE | Eliminar vehículo |

### Técnicos

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/tecnicos` | ADMIN | Crear técnico |
| GET | `/tecnicos` | ADMIN | Listar técnicos |
| GET | `/tecnicos/:id` | ADMIN, TECNICO | Obtener técnico |
| PATCH | `/tecnicos/:id` | ADMIN | Actualizar técnico |
| PATCH | `/tecnicos/:id/estado` | ADMIN | Cambiar estado (ACTIVO/INACTIVO) |

### Servicios

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/servicios` | ADMIN | Crear servicio |
| GET | `/servicios` | Todos | Listar servicios activos |
| PATCH | `/servicios/:id` | ADMIN | Actualizar servicio |
| PATCH | `/servicios/:id/toggle` | ADMIN | Activar / desactivar |

### Accesorios

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/accesorios` | ADMIN | Crear accesorio |
| GET | `/accesorios` | Todos | Listar accesorios |
| PATCH | `/accesorios/:id` | ADMIN | Actualizar / ajustar stock |
| DELETE | `/accesorios/:id` | ADMIN | Eliminar accesorio |

### Citas

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| POST | `/citas` | ADMIN, TECNICO, CLIENTE | Crear cita |
| GET | `/citas` | Todos | Listar citas (filtrado por rol) |
| GET | `/citas/sin-asignar` | ADMIN, TECNICO | Citas sin técnico asignado |
| GET | `/citas/pospuestas` | ADMIN, TECNICO | Citas pospuestas |
| GET | `/citas/:id` | Todos | Detalle de una cita |
| PATCH | `/citas/:id/aceptar` | TECNICO | Aceptar cita |
| PATCH | `/citas/:id/iniciar` | TECNICO | Iniciar trabajo |
| PATCH | `/citas/:id/pausar` | TECNICO | Pausar trabajo |
| PATCH | `/citas/:id/reanudar` | TECNICO | Reanudar trabajo |
| PATCH | `/citas/:id/finalizar` | TECNICO | Finalizar trabajo |
| PATCH | `/citas/:id/entregar` | ADMIN, TECNICO | Marcar como entregada |
| PATCH | `/citas/:id/posponer` | TECNICO | Posponer cita |
| PATCH | `/citas/:id/cancelar` | ADMIN, TECNICO | Cancelar cita |
| PATCH | `/citas/:id/reasignar` | ADMIN, TECNICO | Reasignar a otro técnico |
| PATCH | `/citas/:id/duracion` | ADMIN, TECNICO | Ajustar duración estimada |
| POST | `/citas/:id/foto/:tipo` | TECNICO | Subir foto inicio/fin |
| POST | `/citas/:id/evidencia` | TECNICO | Subir evidencia adicional |
| PATCH | `/citas/:id/fase` | TECNICO | Cambiar nombre de fase |
| PATCH | `/citas/:id/fase/avanzar` | TECNICO | Avanzar a siguiente fase |
| PATCH | `/citas/:id/fase/resetear` | TECNICO | Resetear a primera fase |
| GET | `/citas/:id/fases` | Todos | Historial de fases |
| GET | `/citas/:id/whatsapp/:tipo` | ADMIN, TECNICO | Generar link wa.me |

### Chat

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| GET | `/chat/:citaId/mensajes` | CLIENTE, TECNICO | Listar mensajes de una cita |
| POST | `/chat/:citaId/mensajes` | CLIENTE, TECNICO | Enviar mensaje de texto |
| POST | `/chat/:citaId/mensajes/archivo` | CLIENTE, TECNICO | Enviar imagen o video |

### Admin

| Método | Ruta | Rol | Descripción |
|--------|------|-----|-------------|
| GET | `/admin/dashboard` | ADMIN | Estado en tiempo real de técnicos y citas |
| GET | `/admin/reportes/productividad` | ADMIN | Reporte de productividad por técnico |
| GET | `/admin/reportes/servicios` | ADMIN | Servicios más solicitados por período |

### Health Check

| Método | Ruta | Descripción |
|--------|------|-------------|
| GET | `/health` | Estado del servidor (uptime, timestamp) |

---

## 6. Comunicación en Tiempo Real (Socket.io)

El servidor Socket.io corre en el mismo proceso que Express, en el mismo puerto.

### Eventos emitidos por el servidor (server → cliente)

| Evento | Payload | Destinatarios |
|--------|---------|--------------|
| `cita:estado_actualizado` | `{ citaId, estadoAnterior, estadoNuevo }` | Cliente de la cita, técnico asignado, admin |
| `cita:fase_actualizada` | `{ citaId, faseActual }` | Cliente de la cita |
| `chat:nuevo_mensaje` | `{ citaId, mensaje }` | Cliente y técnico de la cita |
| `tecnico:disponibilidad` | `{ tecnicoId, estado }` | Admin |

### Rooms de Socket.io

- **`cita:{id}`** — todos los participantes de una cita (cliente + técnico) se unen a este room
- **`admin`** — el admin se une para recibir actualizaciones globales del dashboard
- **`tecnico:{id}`** — el técnico recibe notificaciones personales

### Configuración Nginx para WebSocket

```nginx
location /socket.io/ {
    proxy_pass http://localhost:3000;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```

---

## 7. Autenticación y Autorización

### Flujo de login

```
Cliente                     Backend                         PostgreSQL
   │                           │                                │
   │── POST /auth/login ──────►│                                │
   │   { usuario, password }   │── findUnique(admin) ─────────►│
   │                           │◄─ admin | null ────────────────│
   │                           │                                │
   │                           │  Si no es admin:               │
   │                           │── findUnique(tecnico) ────────►│
   │                           │◄─ tecnico | null ──────────────│
   │                           │                                │
   │                           │  Si no es tecnico:             │
   │                           │── findMany(cliente) ──────────►│
   │                           │◄─ [clientes] ──────────────────│
   │                           │                                │
   │                           │  bcrypt.compare(password)      │
   │                           │  jwt.sign({ id, rol, nombre }) │
   │◄── { token, rol, user } ──│                                │
```

### Estructura del JWT

```json
{
  "id": 1,
  "rol": "ADMIN | TECNICO | CLIENTE",
  "nombre": "Nombre del usuario",
  "iat": 1234567890,
  "exp": 1234567890
}
```

### Cadena de middlewares en cada request protegido

```
Request → verifyToken → requireRole('ADMIN', 'TECNICO') → validate(schema) → Controller
```

---

## 8. Manejo de Archivos

### Estructura de directorios en disco

```
backend/uploads/
├── evidencias/
│   ├── inicio/          # Fotos de inicio de trabajo
│   └── fin/             # Fotos de fin de trabajo
├── fases/               # Evidencias subidas por el técnico durante el trabajo
└── chat/                # Multimedia del chat cliente-técnico
```

### Políticas de retención

| Tipo de archivo | Retención | Acción al expirar |
|-----------------|-----------|-------------------|
| Fotos de evidencia inicio/fin | 30 días | Se borra el archivo, `fotoInicioExpirada = true` |
| Multimedia del chat | 7 días | Se borra el archivo, mensaje marcado como `expirado = true` |

### Límites de subida

| Tipo | Límite |
|------|--------|
| Foto de evidencia (inicio/fin) | 50 MB, JPEG/PNG |
| Evidencias adicionales (imágenes) | 50 MB por archivo |
| Videos en chat | máx. 2 por cita, 250 MB combinados |
| Imágenes en chat | 50 MB por archivo |

---

## 9. Tareas Programadas (Cron Jobs)

Todas las tareas están registradas en `app.js` usando `node-cron`:

| Tarea | Expresión cron | Archivo |
|-------|---------------|---------|
| Backup de PostgreSQL | `0 2 * * *` (02:00 AM) | `jobs/backupDb.js` |
| Limpieza de fotos >30 días | `0 3 * * *` (03:00 AM) | `jobs/cleanupPhotos.js` |
| Limpieza de multimedia chat >7 días | `30 3 * * *` (03:30 AM) | `jobs/cleanupChat.js` |

### Backup — detalles

- Comando: `pg_dump -Fc` (formato custom comprimido)
- Destino: `/backend/backups/backup_YYYY-MM-DD.dump`
- Retención: archivos de más de N días se eliminan automáticamente (configurable)
- Restauración: **manual** por el administrador con `pg_restore`

---

## 10. Infraestructura y Despliegue

### Servidor

- **Proveedor**: VPS Contabo
- **OS**: Debian/Ubuntu
- **Dominio**: `continental.moondev.online` (subdominio de `moondev.online`)
- **SSL**: Let's Encrypt (renovación automática con certbot)

### Proceso de despliegue

```bash
# 1. Pull del código
git pull origin main

# 2. Instalar dependencias
cd backend && npm install

# 3. Ejecutar migraciones
npx prisma migrate deploy

# 4. Reiniciar proceso con PM2
pm2 restart continental-backend

# 5. Verificar estado
pm2 status
```

### PM2

```bash
# Proceso activo
pm2 list
# ID: 5  |  Name: continental-backend  |  Status: online  |  Port: 3000
```

### Nginx — configuración relevante

```nginx
server {
    listen 443 ssl;
    server_name continental.moondev.online;

    # Frontend Flutter Web
    root /var/www/continental;
    index index.html;
    location / {
        try_files $uri $uri/ /index.html;  # SPA fallback
    }

    # API Backend
    location /api/ {
        proxy_pass http://localhost:3000;
    }

    # Archivos estáticos (uploads)
    location /uploads/ {
        proxy_pass http://localhost:3000/uploads/;
    }

    # WebSocket
    location /socket.io/ {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    gzip on;
    gzip_types text/plain application/json application/javascript text/css;
}
```

---

## 11. Flujo de Datos — Ciclo de una Cita

```
CLIENTE                    TÉCNICO                     ADMIN
   │                          │                           │
   │ 1. Agenda cita            │                           │
   │────────────────────────────────────────────────────► BD
   │                          │                           │
   │                          │ 2. Ve lista sin asignar   │
   │                          │◄─────────────────────── BD
   │                          │                           │
   │                          │ 3. Acepta cita            │
   │◄── socket: estado ───────│────────────────────────► BD
   │    ACEPTADA               │                           │
   │                          │ 4. Sube foto inicio       │
   │                          │────────────────────────► BD
   │                          │ 5. Inicia trabajo         │
   │◄── socket: EN_CURSO ─────│────────────────────────► BD
   │                          │                           │
   │                          │ 6. Cambia fases           │
   │◄── socket: fase ─────────│────────────────────────► BD
   │                          │                           │
   │◄── chat: mensajes ───────│                           │
   │────── chat: mensajes ───►│                           │
   │                          │                           │
   │                          │ 7. Sube foto fin          │
   │                          │ 8. Finaliza trabajo       │
   │◄── socket: FINALIZADA ───│────────────────────────► BD
   │                          │                           │
   │                          │ 9. Entrega vehículo       │
   │◄── socket: ENTREGADA ────│────────────────────────► BD
                              │
                              │ (Opcional) Genera link wa.me
                              │ para notificación manual
```

# Modelo de Datos — Continental Taller App

> Descripción completa del modelo relacional de la base de datos PostgreSQL gestionada con **Prisma ORM**.
> Fuente de verdad: [`backend/prisma/schema.prisma`](backend/prisma/schema.prisma)

---

## Índice

1. [Diagrama de Entidad-Relación](#1-diagrama-de-entidad-relación)
2. [Enumeraciones (Enums)](#2-enumeraciones-enums)
3. [Entidades](#3-entidades)
   - Admin
   - Cliente
   - Vehículo
   - Técnico
   - Servicio
   - Accesorio
   - Cita
   - CitaAccesorio
   - CitaFaseHistorial
   - Mensaje
   - Evidencia
4. [Índices de Rendimiento](#4-índices-de-rendimiento)
5. [Relaciones y Cardinalidades](#5-relaciones-y-cardinalidades)
6. [Decisiones de Diseño](#6-decisiones-de-diseño)

---

## 1. Diagrama de Entidad-Relación

```
                    ┌──────────┐
                    │  Admin   │
                    │──────────│
                    │ id (PK)  │
                    │ usuario  │
                    │ password │
                    │ nombre   │
                    └──────────┘

┌──────────────┐       1       ┌───────────────┐
│   Cliente    │──────────────►│   Vehiculo    │
│──────────────│               │───────────────│
│ id (PK)      │               │ id (PK)       │
│ nombre       │               │ marca         │
│ pin          │               │ modelo        │
│ contactos[]  │               │ anio          │
│ createdAt    │               │ placa         │
└──────┬───────┘               │ notas         │
       │ 1                     │ clienteId(FK) │
       │                       └──────┬────────┘
       │ N                            │ 1
       ▼                              │ N
┌──────────────────────────────────────────────────────────────────────────┐
│                               Cita                                       │
│──────────────────────────────────────────────────────────────────────────│
│ id (PK)                  │ horaInicioReal       │ fotoFinExpirada        │
│ clienteId (FK)           │ horaFinReal          │ motivoPosposicion      │
│ vehiculoId (FK)          │ faseActual           │ esInmediata            │
│ tecnicoId (FK, nullable) │ fotoInicio           │ adicionales            │
│ servicioId (FK, nullable)│ fotoFin              │ serviciosDetalle       │
│ estado (enum)            │ fotoInicioExpirada   │ esPersonalizado        │
│ horaAgendada             │ duracionEstimadaMins │ createdAt / updatedAt  │
└──────────────────────────┬───────────────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┬─────────────────┐
           │               │               │                  │
           ▼               ▼               ▼                  ▼
    ┌─────────────┐ ┌──────────────┐ ┌──────────┐ ┌────────────────────┐
    │CitaAccesorio│ │CitaFasHistor.│ │ Mensaje  │ │    Evidencia       │
    │─────────────│ │──────────────│ │──────────│ │────────────────────│
    │ id (PK)     │ │ id (PK)      │ │ id (PK)  │ │ id (PK)            │
    │ citaId (FK) │ │ citaId (FK)  │ │ citaId   │ │ citaId (FK)        │
    │ accesorId   │ │ nombreFase   │ │ autorTipo│ │ tipo (INICIO/FIN/..)│
    │ cantidad    │ │ timestamp    │ │ autorId  │ │ archivoUrl         │
    └──────┬──────┘ └──────────────┘ │ contenido│ │ archivoTipo        │
           │                         │ tipo     │ │ fase               │
           ▼                         │ archivoUr│ │ createdAt          │
    ┌─────────────┐                  │ createdAt│ └────────────────────┘
    │  Accesorio  │                  │ expiraAt │
    │─────────────│                  │ expirado │
    │ id (PK)     │                  └──────────┘
    │ nombre      │
    │ descripcion │
    │ precio      │
    │ stock       │
    │ stockMinAlt │
    └─────────────┘

┌──────────────┐       1       ┌─────────────────┐
│   Tecnico    │──────────────►│   Cita (N)      │
│──────────────│               └─────────────────┘
│ id (PK)      │
│ nombre       │       1       ┌─────────────────┐
│ usuario      │ ┌────────────►│   Cita (N)      │
│ password     │ │ Servicio    └─────────────────┘
│ estado(enum) │ │─────────────│
│ createdAt    │ │ id (PK)     │
└──────────────┘ │ nombre      │
                 │ descripcion │
                 │ duracionMins│
                 │ precio      │
                 │ activo      │
                 └─────────────┘
```

---

## 2. Enumeraciones (Enums)

### `EstadoTecnico`

Controla si un técnico puede acceder al sistema y recibir citas nuevas.

| Valor | Descripción |
|-------|-------------|
| `ACTIVO` | El técnico puede iniciar sesión y recibir citas |
| `INACTIVO` | El técnico no puede iniciar sesión; su historial permanece intacto (soft delete) |

---

### `EstadoCita`

Define el ciclo de vida completo de una cita de servicio.

| Valor | Descripción | Transición desde |
|-------|-------------|-----------------|
| `AGENDADA` | Cita creada, esperando ser tomada por un técnico | — (estado inicial) |
| `ACEPTADA` | Un técnico aceptó la cita y la tiene asignada | `AGENDADA` |
| `EN_CURSO` | El técnico inició el trabajo | `ACEPTADA` |
| `PAUSADA` | El técnico pausó el trabajo temporalmente | `EN_CURSO` |
| `FINALIZADA` | El trabajo está terminado, falta la entrega física | `EN_CURSO` |
| `ENTREGADA` | El vehículo fue entregado al cliente | `FINALIZADA` |
| `POSPUESTA` | El técnico pospuso la cita (requiere reasignación) | `ACEPTADA`, `EN_CURSO`, `PAUSADA` |
| `CANCELADA` | La cita fue cancelada; se libera el stock de accesorios | Cualquier estado activo |

---

### `AutorTipo`

Identifica quién envió un mensaje en el chat de una cita.

| Valor | Descripción |
|-------|-------------|
| `CLIENTE` | El mensaje fue enviado por el cliente |
| `TECNICO` | El mensaje fue enviado por el técnico |

---

### `TipoMensaje`

Indica el tipo de contenido de un mensaje en el chat.

| Valor | Descripción |
|-------|-------------|
| `TEXTO` | Mensaje de texto plano |
| `IMAGEN` | Archivo de imagen (JPEG, PNG, etc.) |
| `VIDEO` | Archivo de video |

---

## 3. Entidades

### `Admin` → tabla `admins`

Única cuenta de administrador del sistema. En v1, solo se admite un admin.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `usuario` | `String` | UNIQUE, NOT NULL | Nombre de usuario para login |
| `password` | `String` | NOT NULL | Contraseña hasheada con bcrypt |
| `nombre` | `String` | DEFAULT "Administrador" | Nombre de display |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |

---

### `Cliente` → tabla `clientes`

Dueño del vehículo que solicita los servicios del taller.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `nombre` | `String` | NOT NULL | Nombre completo |
| `pin` | `String` | NOT NULL | PIN numérico hasheado con bcrypt |
| `contactos` | `String[]` | NOT NULL | Array de 1 a 2 números de teléfono |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |

**Relaciones:**
- `vehiculos`: `Vehiculo[]` — un cliente puede tener N vehículos
- `citas`: `Cita[]` — historial de todas sus citas

> **Nota**: El campo `contactos` usa el tipo array nativo de PostgreSQL. El login puede realizarse usando cualquiera de los números de contacto o el nombre del cliente.

---

### `Vehiculo` → tabla `vehiculos`

Vehículo de un cliente, sujeto de los servicios del taller.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `marca` | `String` | NOT NULL | Marca del vehículo (ej. Toyota) |
| `modelo` | `String` | NOT NULL | Modelo (ej. Corolla) |
| `anio` | `Int` | NOT NULL | Año de fabricación |
| `placa` | `String` | NOT NULL | Número de placa |
| `notas` | `String?` | NULLABLE | Alertas de modificaciones previas |
| `clienteId` | `Int` | FK → Cliente, NOT NULL | Propietario del vehículo |

**Relaciones:**
- `cliente`: `Cliente` — dueño (obligatorio, N:1)
- `citas`: `Cita[]` — servicios realizados sobre este vehículo

---

### `Tecnico` → tabla `tecnicos`

Empleado del taller que ejecuta los trabajos asignados.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `nombre` | `String` | NOT NULL | Nombre completo |
| `usuario` | `String` | UNIQUE, NOT NULL | Nombre de usuario para login |
| `password` | `String` | NOT NULL | Contraseña hasheada con bcrypt |
| `estado` | `EstadoTecnico` | DEFAULT ACTIVO | Estado: ACTIVO o INACTIVO |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |

**Relaciones:**
- `citas`: `Cita[]` — citas asignadas (historial completo, incluso si pasa a INACTIVO)

> **Importante**: Los técnicos nunca se eliminan físicamente. Al "eliminar" un técnico desde el panel de admin, su estado cambia a `INACTIVO`. Su historial de citas permanece accesible en reportes.

---

### `Servicio` → tabla `servicios`

Tipo de trabajo ofrecido por el taller (catálogo de servicios).

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `nombre` | `String` | NOT NULL | Nombre del servicio |
| `descripcion` | `String?` | NULLABLE | Descripción detallada |
| `duracionEstimadaMinutos` | `Int` | NOT NULL | Duración base estimada en minutos |
| `precio` | `Decimal(10,2)` | NOT NULL | Precio base del servicio |
| `activo` | `Boolean` | DEFAULT true | Si false, no aparece en el catálogo |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |

**Relaciones:**
- `citas`: `Cita[]` — citas que usan este servicio

> Los servicios nunca se borran; se desactivan con `activo = false` para conservar el historial de citas pasadas que los referencian.

---

### `Accesorio` → tabla `accesorios`

Material o componente del inventario del taller, usado en las citas.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `nombre` | `String` | NOT NULL | Nombre del accesorio |
| `descripcion` | `String?` | NULLABLE | Descripción |
| `precio` | `Decimal(10,2)` | NOT NULL | Precio unitario |
| `stock` | `Int` | DEFAULT 0 | Unidades disponibles en inventario |
| `stockMinimoAlerta` | `Int` | DEFAULT 5 | Umbral para alerta de stock bajo |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |

**Relaciones:**
- `citaAccesorios`: `CitaAccesorio[]` — citas que usan este accesorio

---

### `Cita` → tabla `citas`

Entidad central del sistema. Representa un trabajo del taller de inicio a fin.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `clienteId` | `Int` | FK → Cliente, NOT NULL | Cliente solicitante |
| `vehiculoId` | `Int` | FK → Vehiculo, NOT NULL | Vehículo a trabajar |
| `tecnicoId` | `Int?` | FK → Tecnico, NULLABLE | Técnico asignado (null hasta que alguien acepte) |
| `servicioId` | `Int?` | FK → Servicio, NULLABLE | Servicio del catálogo (null si es personalizado) |
| `estado` | `EstadoCita` | DEFAULT AGENDADA | Estado actual en la máquina de estados |
| `horaAgendada` | `DateTime` | NOT NULL | Hora de inicio acordada |
| `duracionEstimadaMinutos` | `Int` | NOT NULL | Duración estimada (editable por el técnico) |
| `horaInicioReal` | `DateTime?` | NULLABLE | Timestamp cuando el técnico presiona "Iniciar" |
| `horaFinReal` | `DateTime?` | NULLABLE | Timestamp cuando el técnico presiona "Finalizar" |
| `faseActual` | `String?` | DEFAULT "Inspección" | Nombre de la fase actual (texto libre, máx. 30 chars) |
| `fotoInicio` | `String?` | NULLABLE | Path al archivo de foto de inicio |
| `fotoFin` | `String?` | NULLABLE | Path al archivo de foto de fin |
| `fotoInicioExpirada` | `Boolean` | DEFAULT false | True cuando la foto fue eliminada por el job de limpieza |
| `fotoFinExpirada` | `Boolean` | DEFAULT false | True cuando la foto fue eliminada por el job de limpieza |
| `motivoPosposicion` | `String?` | NULLABLE | Motivo opcional al posponer |
| `esInmediata` | `Boolean` | DEFAULT false | True si la cita se agendó con el cliente presente |
| `adicionales` | `String?` | NULLABLE | Notas adicionales del servicio |
| `serviciosDetalle` | `String?` | NULLABLE | Descripción detallada de servicios personalizados |
| `esPersonalizado` | `Boolean` | DEFAULT false | True si no usa el catálogo de servicios |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de creación |
| `updatedAt` | `DateTime` | @updatedAt | Última modificación (auto-actualizado) |

**Relaciones:**
- `cliente`: `Cliente` — quién solicita (N:1)
- `vehiculo`: `Vehiculo` — sobre qué vehículo (N:1)
- `tecnico`: `Tecnico?` — quién realiza el trabajo (N:1, nullable)
- `servicio`: `Servicio?` — qué tipo de trabajo (N:1, nullable)
- `accesorios`: `CitaAccesorio[]` — materiales usados (N:M)
- `faseHistorial`: `CitaFaseHistorial[]` — log de cambios de fase
- `mensajes`: `Mensaje[]` — chat de la cita
- `evidencias`: `Evidencia[]` — fotos y videos del trabajo

**Índices:**
```sql
CREATE INDEX ON citas (estado);
CREATE INDEX ON citas (tecnico_id);
CREATE INDEX ON citas (cliente_id);
CREATE INDEX ON citas (hora_agendada);
```

---

### `CitaAccesorio` → tabla `cita_accesorios`

Tabla intermedia de la relación N:M entre `Cita` y `Accesorio`.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `citaId` | `Int` | FK → Cita, NOT NULL | Cita a la que pertenece |
| `accesorioId` | `Int` | FK → Accesorio, NOT NULL | Accesorio incluido |
| `cantidad` | `Int` | DEFAULT 1 | Unidades del accesorio usadas en esta cita |

**Restricción única**: `(citaId, accesorioId)` — un mismo accesorio no puede aparecer dos veces en la misma cita (se usa la columna `cantidad` en su lugar).

---

### `CitaFaseHistorial` → tabla `cita_fase_historial`

Log inmutable de todos los cambios de fase en una cita. Garantiza trazabilidad aunque el campo `faseActual` se modifique o sobrescriba.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `citaId` | `Int` | FK → Cita, NOT NULL | Cita a la que pertenece el evento |
| `nombreFase` | `String` | NOT NULL | Nombre de la fase en el momento del cambio |
| `timestamp` | `DateTime` | DEFAULT now() | Momento exacto del cambio de fase |

**Relaciones:**
- `cita`: `Cita` — cita a la que pertenece (N:1)

> Este historial permite reconstruir la línea de tiempo completa del trabajo, incluso si el técnico renombró la fase o cambió la plantilla.

---

### `Mensaje` → tabla `mensajes`

Mensaje del chat privado entre el cliente y el técnico de una cita.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `citaId` | `Int` | FK → Cita, NOT NULL | Cita a la que pertenece el chat |
| `autorTipo` | `AutorTipo` | NOT NULL | CLIENTE o TECNICO |
| `autorId` | `Int` | NOT NULL | ID del autor según su tipo |
| `contenido` | `String?` | NULLABLE | Texto del mensaje (null si es solo multimedia) |
| `tipo` | `TipoMensaje` | DEFAULT TEXTO | TEXTO, IMAGEN o VIDEO |
| `archivoUrl` | `String?` | NULLABLE | Path al archivo multimedia (null si es texto) |
| `createdAt` | `DateTime` | DEFAULT now() | Momento de envío |
| `expiraAt` | `DateTime` | NOT NULL | `createdAt + 7 días` — fecha de expiración del archivo |
| `expirado` | `Boolean` | DEFAULT false | True cuando el job de limpieza eliminó el archivo |

**Relaciones:**
- `cita`: `Cita` — conversación a la que pertenece (N:1)

> El texto del mensaje no expira. Solo el archivo multimedia (`archivoUrl`) se elimina del disco al expirar; el registro del mensaje permanece en la BD con `expirado = true`.

---

### `Evidencia` → tabla `evidencias`

Archivo de evidencia (foto o video) subido por el técnico durante el trabajo, asociado a una fase específica.

| Campo | Tipo | Restricciones | Descripción |
|-------|------|---------------|-------------|
| `id` | `Int` | PK, autoincrement | Identificador único |
| `citaId` | `Int` | FK → Cita, NOT NULL | Cita a la que pertenece |
| `tipo` | `String` | NOT NULL | `INICIO`, `AVANCE` o `FIN` |
| `archivoUrl` | `String` | NOT NULL | Path al archivo en el sistema de archivos |
| `archivoTipo` | `String` | NOT NULL | `IMAGEN` o `VIDEO` |
| `fase` | `String?` | NULLABLE | Nombre de la fase en que se subió la evidencia |
| `createdAt` | `DateTime` | DEFAULT now() | Fecha de subida |

**Relaciones:**
- `cita`: `Cita` — cita a la que pertenece (N:1)

---

## 4. Índices de Rendimiento

Los índices están definidos directamente en el schema de Prisma sobre la tabla `citas`, que es la más consultada del sistema:

```prisma
model Cita {
  // ...
  @@index([estado])        // Filtrar citas por estado (AGENDADA, EN_CURSO, etc.)
  @@index([tecnicoId])     // Obtener citas de un técnico específico
  @@index([clienteId])     // Obtener citas de un cliente específico
  @@index([horaAgendada])  // Ordenar y filtrar por fecha/hora
}
```

---

## 5. Relaciones y Cardinalidades

```
Admin          (1) ─────────────────────────────── (0) Sistema   [sin relaciones FK]
Cliente        (1) ─────────────────── (N) Vehiculo
Cliente        (1) ─────────────────── (N) Cita
Vehiculo       (1) ─────────────────── (N) Cita
Tecnico        (1) ─────────────────── (N) Cita       [nullable: cita sin asignar]
Servicio       (1) ─────────────────── (N) Cita       [nullable: servicio personalizado]
Cita           (1) ─────────────────── (N) CitaAccesorio
Accesorio      (1) ─────────────────── (N) CitaAccesorio
Cita           (1) ─────────────────── (N) CitaFaseHistorial
Cita           (1) ─────────────────── (N) Mensaje
Cita           (1) ─────────────────── (N) Evidencia
```

---

## 6. Decisiones de Diseño

### Por qué `contactos` es un array en lugar de una tabla separada

El cliente puede tener máximo 2 números de contacto. Usando el tipo array nativo de PostgreSQL (`String[]`), la consulta de login por número de teléfono se simplifica con `has`:

```prisma
prisma.cliente.findMany({ where: { contactos: { has: usuario } } })
```

Una tabla separada `ClienteContacto` habría añadido complejidad sin beneficio real para el caso de uso.

---

### Por qué `tecnicoId` y `servicioId` son nullable en `Cita`

- **`tecnicoId` nullable**: una cita recién creada no tiene técnico asignado hasta que alguien la acepta voluntariamente. No hay auto-asignación.
- **`servicioId` nullable**: cuando `esPersonalizado = true`, la cita describe el trabajo en texto libre (`serviciosDetalle`) sin referenciar el catálogo de servicios.

---

### Por qué `fotoInicio` y `fotoFin` están en `Cita` y no solo en `Evidencia`

Los campos `fotoInicio` y `fotoFin` en `Cita` son referencias rápidas a las fotos **obligatorias** de inicio y fin (requeridas para avanzar en la máquina de estados). La tabla `Evidencia` complementa esto con fotos adicionales opcionales que el técnico puede subir durante el trabajo. Son conceptos distintos: las fotos de la cita son de control, las evidencias son documentación del proceso.

---

### Por qué `horaInicioReal` como timestamp y no un cronómetro

El tiempo transcurrido siempre se calcula como:

```
tiempoTranscurrido = now() - horaInicioReal
```

Esto garantiza que si el servidor se reinicia, o si la app se cierra y se vuelve a abrir, el tiempo mostrado sigue siendo preciso. Un cronómetro en memoria se perdería con cualquier reinicio.

---

### Soft delete para Técnicos y Servicios

- **Técnicos**: nunca se borran. `estado = INACTIVO` bloquea el login sin perder el historial.
- **Servicios**: nunca se borran. `activo = false` los oculta del catálogo pero preserva la referencia en citas antiguas.
- **Clientes y citas**: tampoco se borran. Solo los **archivos físicos** (fotos y multimedia) tienen ciclo de vida limitado.

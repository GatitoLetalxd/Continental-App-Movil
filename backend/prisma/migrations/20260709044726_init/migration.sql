-- CreateEnum
CREATE TYPE "EstadoTecnico" AS ENUM ('ACTIVO', 'INACTIVO');

-- CreateEnum
CREATE TYPE "EstadoCita" AS ENUM ('AGENDADA', 'ACEPTADA', 'EN_CURSO', 'PAUSADA', 'FINALIZADA', 'ENTREGADA', 'POSPUESTA', 'CANCELADA');

-- CreateEnum
CREATE TYPE "AutorTipo" AS ENUM ('CLIENTE', 'TECNICO');

-- CreateEnum
CREATE TYPE "TipoMensaje" AS ENUM ('TEXTO', 'IMAGEN', 'VIDEO');

-- CreateTable
CREATE TABLE "admins" (
    "id" SERIAL NOT NULL,
    "usuario" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "nombre" TEXT NOT NULL DEFAULT 'Administrador',
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "admins_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "clientes" (
    "id" SERIAL NOT NULL,
    "nombre" TEXT NOT NULL,
    "pin" TEXT NOT NULL,
    "contactos" TEXT[],
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "clientes_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "vehiculos" (
    "id" SERIAL NOT NULL,
    "marca" TEXT NOT NULL,
    "modelo" TEXT NOT NULL,
    "anio" INTEGER NOT NULL,
    "placa" TEXT NOT NULL,
    "notas" TEXT,
    "cliente_id" INTEGER NOT NULL,

    CONSTRAINT "vehiculos_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "tecnicos" (
    "id" SERIAL NOT NULL,
    "nombre" TEXT NOT NULL,
    "usuario" TEXT NOT NULL,
    "password" TEXT NOT NULL,
    "estado" "EstadoTecnico" NOT NULL DEFAULT 'ACTIVO',
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "tecnicos_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "servicios" (
    "id" SERIAL NOT NULL,
    "nombre" TEXT NOT NULL,
    "descripcion" TEXT,
    "duracion_estimada_minutos" INTEGER NOT NULL,
    "precio" DECIMAL(10,2) NOT NULL,
    "activo" BOOLEAN NOT NULL DEFAULT true,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "servicios_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "accesorios" (
    "id" SERIAL NOT NULL,
    "nombre" TEXT NOT NULL,
    "descripcion" TEXT,
    "precio" DECIMAL(10,2) NOT NULL,
    "stock" INTEGER NOT NULL DEFAULT 0,
    "stock_minimo_alerta" INTEGER NOT NULL DEFAULT 5,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "accesorios_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "citas" (
    "id" SERIAL NOT NULL,
    "cliente_id" INTEGER NOT NULL,
    "vehiculo_id" INTEGER NOT NULL,
    "tecnico_id" INTEGER,
    "servicio_id" INTEGER NOT NULL,
    "estado" "EstadoCita" NOT NULL DEFAULT 'AGENDADA',
    "hora_agendada" TIMESTAMP(3) NOT NULL,
    "duracion_estimada_minutos" INTEGER NOT NULL,
    "hora_inicio_real" TIMESTAMP(3),
    "hora_fin_real" TIMESTAMP(3),
    "fase_actual" TEXT DEFAULT 'Inspección',
    "foto_inicio" TEXT,
    "foto_fin" TEXT,
    "foto_inicio_expirada" BOOLEAN NOT NULL DEFAULT false,
    "foto_fin_expirada" BOOLEAN NOT NULL DEFAULT false,
    "motivo_posposicion" TEXT,
    "es_inmediata" BOOLEAN NOT NULL DEFAULT false,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL,

    CONSTRAINT "citas_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "cita_accesorios" (
    "id" SERIAL NOT NULL,
    "cita_id" INTEGER NOT NULL,
    "accesorio_id" INTEGER NOT NULL,
    "cantidad" INTEGER NOT NULL DEFAULT 1,

    CONSTRAINT "cita_accesorios_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "cita_fase_historial" (
    "id" SERIAL NOT NULL,
    "cita_id" INTEGER NOT NULL,
    "nombre_fase" TEXT NOT NULL,
    "timestamp" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "cita_fase_historial_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "mensajes" (
    "id" SERIAL NOT NULL,
    "cita_id" INTEGER NOT NULL,
    "autor_tipo" "AutorTipo" NOT NULL,
    "autor_id" INTEGER NOT NULL,
    "contenido" TEXT,
    "tipo" "TipoMensaje" NOT NULL DEFAULT 'TEXTO',
    "archivo_url" TEXT,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "expira_at" TIMESTAMP(3) NOT NULL,
    "expirado" BOOLEAN NOT NULL DEFAULT false,

    CONSTRAINT "mensajes_pkey" PRIMARY KEY ("id")
);

-- CreateIndex
CREATE UNIQUE INDEX "admins_usuario_key" ON "admins"("usuario");

-- CreateIndex
CREATE UNIQUE INDEX "tecnicos_usuario_key" ON "tecnicos"("usuario");

-- CreateIndex
CREATE UNIQUE INDEX "cita_accesorios_cita_id_accesorio_id_key" ON "cita_accesorios"("cita_id", "accesorio_id");

-- AddForeignKey
ALTER TABLE "vehiculos" ADD CONSTRAINT "vehiculos_cliente_id_fkey" FOREIGN KEY ("cliente_id") REFERENCES "clientes"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "citas" ADD CONSTRAINT "citas_cliente_id_fkey" FOREIGN KEY ("cliente_id") REFERENCES "clientes"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "citas" ADD CONSTRAINT "citas_vehiculo_id_fkey" FOREIGN KEY ("vehiculo_id") REFERENCES "vehiculos"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "citas" ADD CONSTRAINT "citas_tecnico_id_fkey" FOREIGN KEY ("tecnico_id") REFERENCES "tecnicos"("id") ON DELETE SET NULL ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "citas" ADD CONSTRAINT "citas_servicio_id_fkey" FOREIGN KEY ("servicio_id") REFERENCES "servicios"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "cita_accesorios" ADD CONSTRAINT "cita_accesorios_cita_id_fkey" FOREIGN KEY ("cita_id") REFERENCES "citas"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "cita_accesorios" ADD CONSTRAINT "cita_accesorios_accesorio_id_fkey" FOREIGN KEY ("accesorio_id") REFERENCES "accesorios"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "cita_fase_historial" ADD CONSTRAINT "cita_fase_historial_cita_id_fkey" FOREIGN KEY ("cita_id") REFERENCES "citas"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "mensajes" ADD CONSTRAINT "mensajes_cita_id_fkey" FOREIGN KEY ("cita_id") REFERENCES "citas"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

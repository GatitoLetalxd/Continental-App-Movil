-- DropForeignKey
ALTER TABLE "citas" DROP CONSTRAINT "citas_servicio_id_fkey";

-- AlterTable
ALTER TABLE "citas" ADD COLUMN     "es_personalizado" BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN     "servicios_detalle" TEXT,
ALTER COLUMN "servicio_id" DROP NOT NULL;

-- AddForeignKey
ALTER TABLE "citas" ADD CONSTRAINT "citas_servicio_id_fkey" FOREIGN KEY ("servicio_id") REFERENCES "servicios"("id") ON DELETE SET NULL ON UPDATE CASCADE;

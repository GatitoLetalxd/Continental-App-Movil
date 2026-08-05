-- CreateTable
CREATE TABLE "evidencias" (
    "id" SERIAL NOT NULL,
    "cita_id" INTEGER NOT NULL,
    "tipo" TEXT NOT NULL,
    "archivo_url" TEXT NOT NULL,
    "archivo_type" TEXT NOT NULL,
    "fase" TEXT,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "evidencias_pkey" PRIMARY KEY ("id")
);

-- AddForeignKey
ALTER TABLE "evidencias" ADD CONSTRAINT "evidencias_cita_id_fkey" FOREIGN KEY ("cita_id") REFERENCES "citas"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

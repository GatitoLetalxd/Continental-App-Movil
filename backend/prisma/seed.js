const { PrismaClient } = require('@prisma/client');
const bcrypt = require('bcryptjs');

const prisma = new PrismaClient();

async function main() {
  console.log('🌱 Iniciando seed...');

  // ─── Admin ─────────────────────────────────────────────
  const adminPassword = await bcrypt.hash('admin123', 10);
  const admin = await prisma.admin.upsert({
    where: { usuario: 'admin' },
    update: {},
    create: {
      usuario: 'admin',
      password: adminPassword,
      nombre: 'Administrador Continental',
    },
  });
  console.log(`✅ Admin creado: ${admin.usuario} (contraseña: admin123)`);

  // ─── Técnicos de prueba ────────────────────────────────
  const tecPassword = await bcrypt.hash('tec123', 10);
  
  const tecnico1 = await prisma.tecnico.upsert({
    where: { usuario: 'carlos' },
    update: {},
    create: {
      nombre: 'Carlos Rodríguez',
      usuario: 'carlos',
      password: tecPassword,
      estado: 'ACTIVO',
    },
  });
  console.log(`✅ Técnico creado: ${tecnico1.nombre} (usuario: carlos, contraseña: tec123)`);

  const tecnico2 = await prisma.tecnico.upsert({
    where: { usuario: 'miguel' },
    update: {},
    create: {
      nombre: 'Miguel Torres',
      usuario: 'miguel',
      password: tecPassword,
      estado: 'ACTIVO',
    },
  });
  console.log(`✅ Técnico creado: ${tecnico2.nombre} (usuario: miguel, contraseña: tec123)`);

  // ─── Servicios ─────────────────────────────────────────
  const servicios = [
    {
      nombre: 'Tapizado completo de asientos',
      descripcion: 'Retapizado completo de todos los asientos del vehículo con material a elección del cliente.',
      duracionEstimadaMinutos: 480,
      precio: 150.00,
    },
    {
      nombre: 'Tapizado de puertas',
      descripcion: 'Retapizado de paneles de puertas con material coordinado.',
      duracionEstimadaMinutos: 240,
      precio: 80.00,
    },
    {
      nombre: 'Instalación de sistema de audio',
      descripcion: 'Instalación completa de sistema de audio automotriz: reproductor, amplificador y parlantes.',
      duracionEstimadaMinutos: 180,
      precio: 60.00,
    },
    {
      nombre: 'Instalación eléctrica de luces LED',
      descripcion: 'Instalación de luces LED interiores y exteriores.',
      duracionEstimadaMinutos: 120,
      precio: 45.00,
    },
    {
      nombre: 'Instalación de alarma',
      descripcion: 'Instalación de sistema de alarma con control remoto.',
      duracionEstimadaMinutos: 90,
      precio: 35.00,
    },
    {
      nombre: 'Instalación de accesorios generales',
      descripcion: 'Instalación de accesorios varios: protectores, guardafangos, estribos, etc.',
      duracionEstimadaMinutos: 60,
      precio: 25.00,
    },
  ];

  for (const servicio of servicios) {
    await prisma.servicio.upsert({
      where: { id: servicios.indexOf(servicio) + 1 },
      update: {},
      create: servicio,
    });
  }
  console.log(`✅ ${servicios.length} servicios creados`);

  // ─── Accesorios ────────────────────────────────────────
  const accesorios = [
    { nombre: 'Cuero sintético negro', descripcion: 'Rollo de cuero sintético, color negro', precio: 25.00, stock: 15, stockMinimoAlerta: 3 },
    { nombre: 'Cuero sintético beige', descripcion: 'Rollo de cuero sintético, color beige', precio: 25.00, stock: 10, stockMinimoAlerta: 3 },
    { nombre: 'Espuma de alta densidad', descripcion: 'Plancha de espuma para asientos', precio: 8.00, stock: 20, stockMinimoAlerta: 5 },
    { nombre: 'Kit parlantes 6x9', descripcion: 'Par de parlantes 6x9 pulgadas', precio: 35.00, stock: 8, stockMinimoAlerta: 2 },
    { nombre: 'Amplificador 4 canales', descripcion: 'Amplificador de audio 4 canales 1200W', precio: 45.00, stock: 5, stockMinimoAlerta: 2 },
    { nombre: 'Cinta LED RGB 5m', descripcion: 'Cinta LED RGB de 5 metros con controlador', precio: 12.00, stock: 12, stockMinimoAlerta: 3 },
    { nombre: 'Alarma con sensor', descripcion: 'Sistema de alarma con sensor de movimiento', precio: 28.00, stock: 6, stockMinimoAlerta: 2 },
    { nombre: 'Cable automotriz calibre 12', descripcion: 'Rollo de 10m de cable automotriz', precio: 5.00, stock: 30, stockMinimoAlerta: 10 },
  ];

  for (const accesorio of accesorios) {
    await prisma.accesorio.create({ data: accesorio });
  }
  console.log(`✅ ${accesorios.length} accesorios creados`);

  // ─── Cliente de prueba ─────────────────────────────────
  const clientePin = await bcrypt.hash('1234', 10);
  const cliente = await prisma.cliente.create({
    data: {
      nombre: 'Juan Pérez',
      pin: clientePin,
      contactos: ['584121234567', '584129876543'],
    },
  });
  console.log(`✅ Cliente creado: ${cliente.nombre} (PIN: 1234)`);

  // Vehículo del cliente de prueba
  await prisma.vehiculo.create({
    data: {
      marca: 'Toyota',
      modelo: 'Corolla',
      anio: 2020,
      placa: 'ABC123',
      notas: 'Sin modificaciones previas',
      clienteId: cliente.id,
    },
  });
  console.log('✅ Vehículo de prueba creado: Toyota Corolla 2020');

  console.log('\n🎉 Seed completado exitosamente!\n');
  console.log('Credenciales de acceso:');
  console.log('─────────────────────────');
  console.log('Admin:    admin / admin123');
  console.log('Técnico:  carlos / tec123');
  console.log('Técnico:  miguel / tec123');
  console.log('Cliente:  Juan Pérez / 1234');
}

main()
  .catch((e) => {
    console.error('❌ Error en el seed:', e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });

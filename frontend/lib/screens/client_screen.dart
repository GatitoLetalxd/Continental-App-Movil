
import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../config/theme.dart';
import '../providers/auth_provider.dart';
import '../providers/taller_provider.dart';
import '../services/socket_service.dart';
import '../services/api_service.dart';
import '../widgets/glass_card.dart';
import '../widgets/skeleton_loader.dart';

class ClientScreen extends StatefulWidget {
  const ClientScreen({super.key});

  @override
  State<ClientScreen> createState() => _ClientScreenState();
}

class _ClientScreenState extends State<ClientScreen> {
  int _currentIndex = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.fetchVehiculos();
      provider.fetchCitas();
      provider.fetchServicios(soloActivos: true);
      provider.fetchAccesorios();

      // Conectar socket
      final socketService = Provider.of<SocketService>(context, listen: false);
      socketService.connect();
    });
  }

  @override
  Widget build(BuildContext context) {
    final auth = Provider.of<AuthProvider>(context);

    final screens = [
      const ClientCitasTab(),
      const ClientVehiculosTab(),
      const ClientNuevaCitaTab(),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('PORTAL DE CLIENTE'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout_outlined, color: AppTheme.accent),
            onPressed: () {
              Provider.of<SocketService>(context, listen: false).disconnect();
              auth.logout();
            },
          ),
        ],
      ),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentIndex,
        onTap: (idx) {
          setState(() {
            _currentIndex = idx;
          });
        },
        backgroundColor: AppTheme.primary,
        selectedItemColor: AppTheme.accent,
        unselectedItemColor: AppTheme.textMuted,
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.assignment_outlined),
            label: 'Mis Citas',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.directions_car_filled_outlined),
            label: 'Mis Vehículos',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.add_circle_outline),
            label: 'Agendar Cita',
          ),
        ],
      ),
      body: SafeArea(
        child: screens[_currentIndex],
      ),
    );
  }
}

// ─── PESTAÑA 1: MIS CITAS (SEGUIMIENTO Y CHAT) ────────────────

class ClientCitasTab extends StatelessWidget {
  const ClientCitasTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    if (provider.isLoading && provider.citas.isEmpty) {
      return const SkeletonList();
    }

    if (provider.citas.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.assignment_late_outlined, size: 64, color: AppTheme.textMuted),
            const SizedBox(height: 16),
            const Text(
              'No tienes citas agendadas.',
              style: TextStyle(color: AppTheme.textLight, fontSize: 18),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () {
                // Forzar refresco
                provider.fetchCitas();
              },
              child: const Text('Actualizar'),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: provider.citas.length,
      itemBuilder: (context, index) {
        final cita = provider.citas[index];
        final id = cita['id'];
        final vehiculo = cita['vehiculo'];
        final servicio = cita['servicio'];
        final estado = cita['estado'];
        final fase = cita['faseActual'] ?? 'Pendiente';
        final horaEstimada = cita['horaEstimadaEntrega'] != null 
            ? DateTime.tryParse(cita['horaEstimadaEntrega'])
            : null;

        return Card(
          margin: const EdgeInsets.only(bottom: 16),
          child: ExpansionTile(
            title: Text(
              '${servicio['nombre']} - ${vehiculo['marca']} ${vehiculo['modelo']}',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            subtitle: Text('${(vehiculo['placa'] != null && vehiculo['placa'].toString().trim().isNotEmpty) ? "Placa: ${vehiculo['placa']} • " : ""}Estado: $estado'),
            leading: _getEstadoIcon(estado),
            children: [
              Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Status tracker (Línea de tiempo)
                    const Text(
                      'Seguimiento en Vivo:',
                      style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.textLight),
                    ),
                    const SizedBox(height: 12),
                    _buildStatusTracker(context, estado, fase),
                    const SizedBox(height: 16),
                    if (horaEstimada != null && (estado == 'EN_CURSO' || estado == 'PAUSADA' || estado == 'ACEPTADA')) ...[
                      Row(
                        children: [
                          const Icon(Icons.timer_outlined, color: AppTheme.warning, size: 20),
                          const SizedBox(width: 8),
                          Text(
                            'Entrega Estimada: ${DateFormat('dd/MM/yyyy HH:mm').format(horaEstimada)}',
                            style: const TextStyle(color: AppTheme.warning, fontWeight: FontWeight.bold),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                    ],
                    // Botones de acción: ver chat
                    Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        OutlinedButton.icon(
                          onPressed: () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => ClientCitaDetailScreen(citaId: id),
                              ),
                            );
                          },
                          icon: const Icon(Icons.remove_red_eye_outlined),
                          label: const Text('Ver Detalles'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton.icon(
                          onPressed: () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => ChatScreen(citaId: id, tecnicoNombre: cita['tecnico']?['nombre'] ?? 'Técnico'),
                              ),
                            );
                          },
                          icon: const Icon(Icons.chat_bubble_outline),
                          label: const Text('Chat con Técnico'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget _getEstadoIcon(String estado) {
    switch (estado) {
      case 'AGENDADA':
        return const Icon(Icons.schedule, color: AppTheme.textMuted);
      case 'ACEPTADA':
        return const Icon(Icons.thumb_up_alt_outlined, color: AppTheme.success);
      case 'EN_CURSO':
        return const Icon(Icons.play_circle_fill, color: Colors.blue);
      case 'PAUSADA':
        return const Icon(Icons.pause_circle_filled, color: AppTheme.warning);
      case 'FINALIZADA':
        return const Icon(Icons.check_circle, color: AppTheme.success);
      case 'ENTREGADA':
        return const Icon(Icons.directions_car, color: Colors.purple);
      case 'POSPUESTA':
        return const Icon(Icons.error_outline, color: AppTheme.warning);
      default:
        return const Icon(Icons.cancel_outlined, color: AppTheme.accent);
    }
  }

  Widget _buildStatusTracker(BuildContext context, String estado, String fase) {
    final fases = ['Inspección', 'Desmontaje', 'Instalación', 'Finalizado'];
    
    // Si la fase no está en las por defecto, es una fase personalizada creada por el técnico.
    // La agregamos dinámicamente si el estado está en curso
    int activeIndex = fases.indexOf(fase);
    List<String> visibleFases = List.from(fases);
    
    if (activeIndex == -1 && estado != 'FINALIZADA' && estado != 'ENTREGADA') {
      // Es una fase personalizada, la inyectamos en lugar de la actual correspondiente
      visibleFases.insert(1, fase);
      activeIndex = 1;
    } else if (estado == 'FINALIZADA' || estado == 'ENTREGADA') {
      activeIndex = 3;
    } else if (estado == 'AGENDADA') {
      activeIndex = -1; // Nada iniciado
    } else if (estado == 'ACEPTADA') {
      activeIndex = 0; // Iniciado en inspección
    }

    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: List.generate(visibleFases.length, (idx) {
            final isCompleted = idx <= activeIndex;
            final name = visibleFases[idx];

            return Expanded(
              child: Column(
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Container(
                          height: 3,
                          color: idx == 0 
                              ? Colors.transparent 
                              : (isCompleted ? AppTheme.accent : AppTheme.border),
                        ),
                      ),
                      Container(
                        width: 24,
                        height: 24,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: isCompleted ? AppTheme.accent : AppTheme.border,
                          border: Border.all(
                            color: isCompleted ? AppTheme.accent : AppTheme.border,
                          ),
                        ),
                        child: Icon(
                          isCompleted ? Icons.check : Icons.circle,
                          size: 14,
                          color: Colors.white,
                        ),
                      ),
                      Expanded(
                        child: Container(
                          height: 3,
                          color: idx == visibleFases.length - 1 
                              ? Colors.transparent 
                              : (idx < activeIndex ? AppTheme.accent : AppTheme.border),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    name,
                    style: TextStyle(
                      fontSize: 10,
                      fontWeight: isCompleted ? FontWeight.bold : FontWeight.normal,
                      color: isCompleted ? AppTheme.textLight : AppTheme.textMuted,
                    ),
                    textAlign: TextAlign.center,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            );
          }),
        ),
      ],
    );
  }
}

// ─── DETALLE DE CITA ─────────────────────────────────────────

class ClientCitaDetailScreen extends StatefulWidget {
  final int citaId;
  const ClientCitaDetailScreen({super.key, required this.citaId});

  @override
  State<ClientCitaDetailScreen> createState() => _ClientCitaDetailScreenState();
}

class _ClientCitaDetailScreenState extends State<ClientCitaDetailScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<TallerProvider>(context, listen: false).fetchCitaDetalle(widget.citaId);
    });
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);
    final cita = provider.citaDetalle;

    return Scaffold(
      appBar: AppBar(title: const Text('DETALLE DE CITA')),
      body: provider.isLoading || cita == null
          ? const SkeletonDetail()
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  GlassCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Servicio: ${cita['servicio'] != null ? cita['servicio']['nombre'] : (cita['serviciosDetalle'] ?? 'Servicio Personalizado')}',
                          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppTheme.accent),
                        ),
                        const Divider(color: AppTheme.border, height: 24),
                        _buildInfoRow('Estado', cita['estado']),
                        _buildInfoRow('Fase Actual', cita['faseActual'] ?? 'Ninguna'),
                        if (cita['vehiculo'] != null)
                          _buildInfoRow(
                            'Vehículo',
                            '${cita['vehiculo']['marca']} ${cita['vehiculo']['modelo']}${cita['vehiculo']['placa'] != null && cita['vehiculo']['placa'].toString().trim().isNotEmpty ? " (${cita['vehiculo']['placa']})" : ""}',
                          ),
                        _buildInfoRow('Notas del Vehículo', cita['vehiculo']['notas'] ?? 'Ninguna'),
                        _buildInfoRow('Técnico Asignado', cita['tecnico']?['nombre'] ?? 'Sin asignar'),
                        _buildInfoRow('Hora Agendada', DateFormat('dd/MM/yyyy HH:mm').format(DateTime.parse(cita['horaAgendada']))),
                        if (cita['horaInicioReal'] != null)
                          _buildInfoRow('Inicio de Trabajo', DateFormat('dd/MM/yyyy HH:mm').format(DateTime.parse(cita['horaInicioReal']))),
                        if (cita['horaFinReal'] != null)
                          _buildInfoRow('Fin de Trabajo', DateFormat('dd/MM/yyyy HH:mm').format(DateTime.parse(cita['horaFinReal']))),
                        
                        // Duración estimada
                        if (cita['esPersonalizado'] == true && cita['duracionEstimadaMinutos'] == 0)
                          _buildInfoRow('Duración Estimada', 'Pendiente de estimación')
                        else
                          _buildInfoRow('Duración Estimada', '${cita['duracionEstimadaMinutos']} min'),

                        // Fecha/Hora aproximada de entrega
                        if (cita['estado'] != 'ENTREGADA' && cita['estado'] != 'FINALIZADA') ...[
                          if (cita['esPersonalizado'] == true && cita['duracionEstimadaMinutos'] == 0)
                            _buildInfoRow('Entrega Aproximada', 'Pendiente (Servicio Personalizado)')
                          else ...[
                            Builder(
                              builder: (context) {
                                final base = cita['horaInicioReal'] ?? cita['horaAgendada'];
                                final dur = cita['duracionEstimadaMinutos'] ?? 0;
                                final entrega = DateTime.parse(base).add(Duration(minutes: dur));
                                return _buildInfoRow('Entrega Aproximada', DateFormat('dd/MM/yyyy HH:mm').format(entrega));
                              }
                            ),
                          ],
                        ],
                        if (cita['tiempoTranscurrido'] != null)
                          _buildInfoRow('Tiempo Transcurrido', '${cita['tiempoTranscurrido']} min'),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (cita['accesorios'] != null && cita['accesorios'].isNotEmpty) ...[
                    const Text('Accesorios Reservados:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 8),
                    ...List.generate(cita['accesorios'].length, (idx) {
                      final acc = cita['accesorios'][idx];
                      return Card(
                        child: ListTile(
                          title: Text(acc['accesorio']['nombre']),
                          subtitle: Text('Cantidad: ${acc['cantidad']} • Precio Unitario: S/. ${acc['accesorio']['precio']} -> aprox.'),
                        ),
                      );
                    }),
                    const SizedBox(height: 16),
                  ],
                  // 5. Galería de Evidencias (Estructuradas, lectura para el cliente)
                  if (cita['evidencias'] != null && (cita['evidencias'] as List).isNotEmpty) ...[
                    const Text('Galería de Evidencias (Fotos y Videos):', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 4),
                    _buildClienteEvidenciasPorTipo(cita, 'INICIO', 'Evidencias al Inicio'),
                    _buildClienteEvidenciasPorTipo(cita, 'AVANCE', 'Evidencias de Avances (Fase)'),
                    _buildClienteEvidenciasPorTipo(cita, 'FIN', 'Evidencias al Finalizar'),
                    const SizedBox(height: 16),
                  ],
                  // Historial de fases
                  if (cita['faseHistorial'] != null && cita['faseHistorial'].isNotEmpty) ...[
                    const Text('Historial de Fases:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 8),
                    ...List.generate(cita['faseHistorial'].length, (idx) {
                      final f = cita['faseHistorial'][idx];
                      return Card(
                        color: Colors.transparent,
                        child: ListTile(
                          leading: const Icon(Icons.check_circle_outline, color: AppTheme.accent),
                          title: Text(f['nombreFase']),
                          trailing: Text(DateFormat('HH:mm').format(DateTime.parse(f['timestamp']))),
                        ),
                      );
                    }),
                  ]
                ],
              ),
            ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 12.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 130,
            child: Text(label, style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.textMuted)),
          ),
          Expanded(
            child: Text(value, style: const TextStyle(color: AppTheme.textLight)),
          ),
        ],
      ),
    );
  }

  Widget _buildClienteEvidenciasPorTipo(Map<String, dynamic> cita, String tipo, String titulo) {
    final evs = (cita['evidencias'] as List? ?? [])
        .where((e) => e['tipo'] == tipo)
        .toList();

    if (evs.isEmpty) return const SizedBox.shrink();

    return Card(
      margin: const EdgeInsets.only(top: 8),
      child: Padding(
        padding: const EdgeInsets.all(12.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              titulo,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: AppTheme.accent),
            ),
            const SizedBox(height: 8),
            SizedBox(
              height: 85,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                itemCount: evs.length,
                itemBuilder: (ctx, idx) {
                  final ev = evs[idx];
                  final esVid = ev['archivoTipo'] == 'VIDEO';

                  return Container(
                    margin: const EdgeInsets.only(right: 8),
                    width: 85,
                    decoration: BoxDecoration(
                      color: Colors.black26,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: AppTheme.border),
                    ),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: esVid
                          ? const Center(
                              child: Icon(Icons.play_circle_fill, color: AppTheme.accent, size: 36),
                            )
                          : Image.network(
                              '${ApiService.rawBaseUrl}/${ev['archivoUrl']}',
                              fit: BoxFit.cover,
                              errorBuilder: (c, e, s) => const Icon(Icons.broken_image, size: 30),
                            ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── PESTAÑA 2: GESTIÓN DE VEHÍCULOS ─────────────────────────

class ClientVehiculosTab extends StatefulWidget {
  const ClientVehiculosTab({super.key});

  @override
  State<ClientVehiculosTab> createState() => _ClientVehiculosTabState();
}

class _ClientVehiculosTabState extends State<ClientVehiculosTab> {
  final _formKey = GlobalKey<FormState>();
  final _marcaController = TextEditingController();
  final _modeloController = TextEditingController();
  final _anioController = TextEditingController();
  final _placaController = TextEditingController();
  final _notasController = TextEditingController();

  @override
  void dispose() {
    _marcaController.dispose();
    _modeloController.dispose();
    _anioController.dispose();
    _placaController.dispose();
    _notasController.dispose();
    super.dispose();
  }

  void _abrirCrearDialog() {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Registrar Nuevo Vehículo'),
          content: Form(
            key: _formKey,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextFormField(
                    controller: _marcaController,
                    decoration: const InputDecoration(labelText: 'Marca (Toyota, Ford, etc.)'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _modeloController,
                    decoration: const InputDecoration(labelText: 'Modelo (Corolla, Fiesta, etc.)'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _anioController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Año'),
                    validator: (val) {
                      if (val == null || val.isEmpty) return 'Requerido';
                      if (int.tryParse(val) == null) return 'Debe ser número';
                      return null;
                    },
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _plakaController, // Evita colisiones de nombres si el IDE se queja
                    decoration: const InputDecoration(labelText: 'Placa / Matrícula (Opcional)'),
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _notasController,
                    decoration: const InputDecoration(labelText: 'Notas / Modificaciones previas'),
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancelar'),
            ),
            ElevatedButton(
              onPressed: () async {
                if (!_formKey.currentState!.validate()) return;
                
                final provider = Provider.of<TallerProvider>(context, listen: false);
                final nav = Navigator.of(context);
                final messenger = ScaffoldMessenger.of(context);
                final exitoso = await provider.registrarVehiculo({
                  'marca': _marcaController.text.trim(),
                  'modelo': _modeloController.text.trim(),
                  'anio': int.parse(_anioController.text.trim()),
                  'placa': _placaController.text.trim(),
                  'notas': _notasController.text.trim(),
                });

                if (exitoso != null) {
                  nav.pop();
                  messenger.showSnackBar(
                    const SnackBar(content: Text('Vehículo registrado'), backgroundColor: AppTheme.success),
                  );
                  if (mounted) {
                    _marcaController.clear();
                    _modeloController.clear();
                    _anioController.clear();
                    _placaController.clear();
                    _notasController.clear();
                  }
                }
              },
              child: const Text('Guardar'),
            ),
          ],
        );
      },
    );
  }

  // Helper para esquivar error placa/plaka
  TextEditingController get _plakaController => _placaController;

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Scaffold(
      floatingActionButton: FloatingActionButton(
        onPressed: _abrirCrearDialog,
        backgroundColor: AppTheme.accent,
        child: const Icon(Icons.add),
      ),
      body: provider.vehiculos.isEmpty
          ? const Center(
              child: Text(
                'No tienes vehículos registrados. Agrégalos con el botón (+).',
                style: TextStyle(color: AppTheme.textMuted),
                textAlign: TextAlign.center,
              ),
            )
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: provider.vehiculos.length,
              itemBuilder: (context, index) {
                final v = provider.vehiculos[index];
                return Card(
                  child: ListTile(
                    leading: const Icon(Icons.directions_car, color: AppTheme.accent),
                    title: Text('${v['marca']} ${v['modelo']} (${v['anio']})'),
                    subtitle: Text('${v['placa'] != null && v['placa'].toString().trim().isNotEmpty ? "Placa: ${v['placa']}\n" : ""}Notas: ${v['notas'] ?? "Ninguna"}'),
                    isThreeLine: true,
                  ),
                );
              },
            ),
    );
  }
}

// ─── PESTAÑA 3: AGENDAR NUEVA CITA ───────────────────────────

class ClientNuevaCitaTab extends StatefulWidget {
  const ClientNuevaCitaTab({super.key});

  @override
  State<ClientNuevaCitaTab> createState() => _ClientNuevaCitaTabState();
}

class _ClientNuevaCitaTabState extends State<ClientNuevaCitaTab> {
  final _formKey = GlobalKey<FormState>();
  int? _selectedVehiculoId;
  int? _selectedServicioId;
  DateTime _selectedDate = DateTime.now();
  final _adicionalesController = TextEditingController();

  // Carrito de accesorios
  final Map<int, int> _carritoAccesorios = {}; // id -> cantidad

  @override
  void dispose() {
    _adicionalesController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    if (provider.vehiculos.isEmpty) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.warning_amber_outlined, size: 64, color: AppTheme.warning),
              const SizedBox(height: 16),
              const Text(
                'Debes registrar al menos un vehículo antes de agendar una cita.',
                style: TextStyle(color: AppTheme.textLight, fontSize: 18),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              const Text(
                'Ve a la pestaña "Mis Vehículos" para registrar uno.',
                style: TextStyle(color: AppTheme.textMuted),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      );
    }

    return Form(
      key: _formKey,
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text('Agendar Nueva Cita', style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
          const SizedBox(height: 16),

          // 1. Vehículo
          DropdownButtonFormField<int>(
            value: _selectedVehiculoId,
            decoration: const InputDecoration(labelText: 'Selecciona el Vehículo'),
            items: provider.vehiculos.map<DropdownMenuItem<int>>((v) {
              return DropdownMenuItem<int>(
                value: v['id'],
                child: Text('${v['marca']} ${v['modelo']}${v['placa'] != null && v['placa'].toString().trim().isNotEmpty ? " (${v['placa']})" : ""}'),
              );
            }).toList(),
            onChanged: (val) {
              setState(() {
                _selectedVehiculoId = val;
              });
            },
            validator: (val) => val == null ? 'Seleccione un vehículo' : null,
          ),
          const SizedBox(height: 16),

          // 2. Servicio
          DropdownButtonFormField<int>(
            value: _selectedServicioId,
            decoration: const InputDecoration(labelText: 'Selecciona el Servicio'),
            items: provider.servicios.map<DropdownMenuItem<int>>((s) {
              return DropdownMenuItem<int>(
                value: s['id'],
                child: Text("${s['nombre']} (S/. ${s['precio']} -> aprox.)"),
              );
            }).toList(),
            onChanged: (val) {
              setState(() {
                _selectedServicioId = val;
              });
            },
            validator: (val) => val == null ? 'Seleccione un servicio' : null,
          ),
          const SizedBox(height: 16),
          ListTile(
            title: const Text('Fecha y Hora de la Cita'),
            subtitle: Text(DateFormat('dd/MM/yyyy HH:mm').format(_selectedDate)),
            trailing: const Icon(Icons.calendar_today, color: AppTheme.accent),
            onTap: () async {
              final date = await showDatePicker(
                context: context,
                initialDate: _selectedDate,
                firstDate: DateTime.now(),
                lastDate: DateTime.now().add(const Duration(days: 30)),
              );
              if (date == null) return;
              if (!context.mounted) return;
              final time = await showTimePicker(
                context: context,
                initialTime: TimeOfDay.fromDateTime(_selectedDate),
              );
              if (time != null) {
                setState(() {
                  _selectedDate = DateTime(
                    date.year,
                    date.month,
                    date.day,
                    time.hour,
                    time.minute,
                  );
                });
              }
            },
          ),
          const SizedBox(height: 16),

          // 4. Agregar adicional (Desplegable)
          const Text(
            'Agregar adicional (Opcional):',
            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: AppTheme.textLight),
          ),
          const SizedBox(height: 8),
          if (provider.accesorios.isEmpty)
            const Text('Cargando adicionales...', style: TextStyle(color: AppTheme.textMuted))
          else ...[
            DropdownButtonFormField<int>(
              decoration: const InputDecoration(labelText: 'Selecciona un componente o accesorio'),
              items: provider.accesorios
                  .where((acc) => (acc['stock'] as int) > 0)
                  .map<DropdownMenuItem<int>>((acc) {
                return DropdownMenuItem<int>(
                  value: acc['id'],
                  child: Text("${acc['nombre']} (S/. ${acc['precio']} -> aprox.)"),
                );
              }).toList(),
              onChanged: (val) {
                if (val != null) {
                  setState(() {
                    final qty = _carritoAccesorios[val] ?? 0;
                    final maxStock = provider.accesorios.firstWhere((acc) => acc['id'] == val)['stock'] as int;
                    if (qty < maxStock) {
                      _carritoAccesorios[val] = qty + 1;
                    }
                  });
                }
              },
            ),
            const SizedBox(height: 8),
            if (_carritoAccesorios.isNotEmpty) ...[
              const Text('Adicionales seleccionados:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: AppTheme.accent)),
              const SizedBox(height: 4),
              ..._carritoAccesorios.entries.map((entry) {
                final acc = provider.accesorios.firstWhere((a) => a['id'] == entry.key);
                final qty = entry.value;
                final stock = acc['stock'] as int;

                return Card(
                  margin: const EdgeInsets.symmetric(vertical: 4),
                  child: ListTile(
                    title: Text(acc['nombre']),
                    subtitle: Text('S/. ${acc['precio']} -> aprox. (Stock: $stock)'),
                    trailing: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          icon: const Icon(Icons.remove, color: AppTheme.accent, size: 20),
                          onPressed: () {
                            setState(() {
                              if (qty > 1) {
                                _carritoAccesorios[entry.key] = qty - 1;
                              } else {
                                _carritoAccesorios.remove(entry.key);
                              }
                            });
                          },
                        ),
                        Text('$qty', style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
                        IconButton(
                          icon: const Icon(Icons.add, color: AppTheme.success, size: 20),
                          onPressed: qty >= stock
                              ? null
                              : () {
                                  setState(() {
                                    _carritoAccesorios[entry.key] = qty + 1;
                                  });
                                },
                        ),
                      ],
                    ),
                  ),
                );
              }),
            ],
          ],
          const SizedBox(height: 16),

          // 5. Requerimientos libres
          TextFormField(
            controller: _adicionalesController,
            maxLines: 3,
            decoration: const InputDecoration(
              labelText: 'Especificaciones o requerimientos personalizados (Opcional)',
              hintText: 'Ej. Traigo mis propios parlantes, costura roja...',
            ),
          ),

          const SizedBox(height: 24),
          provider.isLoading
              ? const Center(child: CircularProgressIndicator(color: AppTheme.accent))
              : ElevatedButton(
                  onPressed: () async {
                    if (!_formKey.currentState!.validate()) return;

                    final accsList = <Map<String, dynamic>>[];
                    _carritoAccesorios.forEach((key, val) {
                      if (val > 0) {
                        accsList.add({
                          'accesorioId': key,
                          'cantidad': val,
                        });
                      }
                    });

                    final messenger = ScaffoldMessenger.of(context);
                    final exitoso = await provider.agendarCita({
                      'clienteId': 0, // El backend lo toma del JWT del cliente
                      'vehiculoId': _selectedVehiculoId,
                      'servicioId': _selectedServicioId,
                      'horaAgendada': _selectedDate.toIso8601String(),
                      'esInmediata': false,
                      'accesorios': accsList,
                      'adicionales': _adicionalesController.text.trim().isEmpty
                          ? null
                          : _adicionalesController.text.trim(),
                    });

                    if (exitoso) {
                      final esCercana = _selectedDate.difference(DateTime.now()).inMinutes < 120;
                      messenger.showSnackBar(
                        SnackBar(
                          content: Text(esCercana
                              ? 'Cita agendada. Su trabajo será atendido lo más pronto posible en cuanto un técnico esté disponible.'
                              : 'Cita agendada con éxito'),
                          backgroundColor: AppTheme.success,
                          duration: const Duration(seconds: 6),
                        ),
                      );
                      if (mounted) {
                        // Resetear
                        setState(() {
                          _selectedVehiculoId = null;
                          _selectedServicioId = null;
                          _carritoAccesorios.clear();
                          _adicionalesController.clear();
                        });
                      }
                    }
                  },
                  child: const Text('CONFIRMAR Y AGENDAR CITA'),
                ),
        ],
      ),
    );
  }
}

// ─── PANTALLA DE CHAT GENERAL CLIENTE-TECNICO ────────────────

class ChatScreen extends StatefulWidget {
  final int citaId;
  final String tecnicoNombre;
  const ChatScreen({super.key, required this.citaId, required this.tecnicoNombre});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final _msgController = TextEditingController();
  final _scrollController = ScrollController();
  final _picker = ImagePicker();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.fetchMensajes(widget.citaId);

      // Registrar socket listener
      final socket = Provider.of<SocketService>(context, listen: false);
      socket.joinCita(widget.citaId);
      socket.on('chat:nuevoMensaje', _onMensajeRecibido);
    });
  }

  void _onMensajeRecibido(dynamic data) {
    if (mounted) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.agregarMensajeEnVivo(data['mensaje']);
      _scrollDown();
    }
  }

  @override
  void dispose() {
    final socket = Provider.of<SocketService>(context, listen: false);
    socket.leaveCita(widget.citaId);
    socket.off('chat:nuevoMensaje');
    _msgController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _scrollDown() {
    if (_scrollController.hasClients) {
      Future.delayed(const Duration(milliseconds: 100), () {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      });
    }
  }

  void _enviarTexto() async {
    if (_msgController.text.trim().isEmpty) return;
    final provider = Provider.of<TallerProvider>(context, listen: false);
    final texto = _msgController.text.trim();
    _msgController.clear();

    final exito = await provider.enviarMensajeTexto(widget.citaId, texto);
    if (exito) {
      _scrollDown();
    }
  }

  void _enviarImagen(ImageSource source) async {
    final provider = Provider.of<TallerProvider>(context, listen: false);
    final pickedFile = await _picker.pickImage(source: source, imageQuality: 70);
    if (pickedFile == null) return;

    final exito = await provider.enviarMensajeMultimedia(widget.citaId, pickedFile);
    if (exito) {
      _scrollDown();
    }
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);
    final auth = Provider.of<AuthProvider>(context);
    final mensajes = provider.mensajesChat;

    return Scaffold(
      appBar: AppBar(
        title: Text('Chat: ${widget.tecnicoNombre}'),
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: mensajes.length,
              itemBuilder: (context, index) {
                final msg = mensajes[index];
                final soyYo = msg['autorId'] == auth.userId &&
                    msg['autorTipo'] == (auth.userRol == 'CLIENTE' ? 'CLIENTE' : 'TECNICO');

                return Align(
                  alignment: soyYo ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.only(bottom: 12),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: BoxDecoration(
                      color: soyYo ? AppTheme.accent : AppTheme.primary,
                      borderRadius: BorderRadius.only(
                        topLeft: const Radius.circular(16),
                        topRight: const Radius.circular(16),
                        bottomLeft: Radius.circular(soyYo ? 16 : 0),
                        bottomRight: Radius.circular(soyYo ? 0 : 16),
                      ),
                      border: Border.all(color: AppTheme.border, width: 0.5),
                    ),
                    constraints: BoxConstraints(maxWidth: MediaQuery.of(context).size.width * 0.75),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (msg['tipo'] == 'TEXTO')
                          Text(
                            msg['contenido'] ?? '',
                            style: const TextStyle(color: AppTheme.textLight, fontSize: 15),
                          )
                        else if (msg['tipo'] == 'IMAGEN')
                          msg['expirado'] == true
                              ? const Text('Archivo Expirado (7 días)', style: TextStyle(color: Colors.red, fontStyle: FontStyle.italic))
                              : Column(
                                  children: [
                                    ClipRRect(
                                      borderRadius: BorderRadius.circular(8),
                                      child: Image.network(
                                        '${ApiService.rawBaseUrl}/${msg['archivoUrl']}',
                                        fit: BoxFit.cover,
                                        errorBuilder: (context, error, stackTrace) => const Icon(Icons.broken_image, size: 50),
                                      ),
                                    ),
                                    const SizedBox(height: 8),
                                    TextButton.icon(
                                      onPressed: () {
                                        // Guardar en dispositivo
                                      },
                                      icon: const Icon(Icons.download, size: 16),
                                      label: const Text('Guardar', style: TextStyle(fontSize: 12)),
                                    ),
                                  ],
                                ),
                        const SizedBox(height: 4),
                        Text(
                          DateFormat('HH:mm').format(DateTime.parse(msg['createdAt'])),
                          style: const TextStyle(fontSize: 10, color: AppTheme.textMuted),
                        ),
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 12),
            decoration: const BoxDecoration(
              color: AppTheme.primary,
              border: Border(top: BorderSide(color: AppTheme.border)),
            ),
            child: Row(
              children: [
                IconButton(
                  icon: const Icon(Icons.camera_alt_outlined, color: AppTheme.accent),
                  onPressed: () => _enviarImagen(ImageSource.camera),
                ),
                IconButton(
                  icon: const Icon(Icons.photo_outlined, color: AppTheme.accent),
                  onPressed: () => _enviarImagen(ImageSource.gallery),
                ),
                Expanded(
                  child: TextField(
                    controller: _msgController,
                    decoration: const InputDecoration(
                      hintText: 'Escribe un mensaje...',
                      border: InputBorder.none,
                      enabledBorder: InputBorder.none,
                      focusedBorder: InputBorder.none,
                    ),
                    onSubmitted: (_) => _enviarTexto(),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.send_outlined, color: AppTheme.accent),
                  onPressed: _enviarTexto,
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

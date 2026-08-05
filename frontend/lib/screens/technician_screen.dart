import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import 'package:url_launcher/url_launcher.dart';
import '../config/theme.dart';
import '../providers/auth_provider.dart';
import '../providers/taller_provider.dart';
import '../services/socket_service.dart';
import '../services/api_service.dart';
import '../widgets/glass_card.dart';
import '../widgets/skeleton_loader.dart';
import 'client_screen.dart'; // Reutilizar ChatScreen

class TechnicianScreen extends StatefulWidget {
  const TechnicianScreen({super.key});

  @override
  State<TechnicianScreen> createState() => _TechnicianScreenState();
}

class _TechnicianScreenState extends State<TechnicianScreen> {
  int _currentIndex = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.fetchCitas(todas: false); // Citas del técnico
      provider.fetchCitasSinAsignar();

      // Conectar socket
      final socketService = Provider.of<SocketService>(context, listen: false);
      socketService.connect();
    });
  }

  @override
  Widget build(BuildContext context) {
    final auth = Provider.of<AuthProvider>(context);

    final tabs = [
      const TechAgendaTab(),
      const TechDisponiblesTab(),
    ];

    return Scaffold(
      appBar: AppBar(
        title: const Text('PORTAL DE TÉCNICO'),
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
            icon: Icon(Icons.today),
            label: 'Mi Agenda',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.search),
            label: 'Citas Disponibles',
          ),
        ],
      ),
      body: SafeArea(
        child: tabs[_currentIndex],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          showDialog(
            context: context,
            barrierDismissible: false,
            builder: (ctx) => const TechCrearCitaDialog(),
          );
        },
        backgroundColor: AppTheme.accent,
        child: const Icon(Icons.add, color: AppTheme.primary),
      ),
    );
  }
}

// ─── TAB 1: MI AGENDA ────────────────────────────────────────

class TechAgendaTab extends StatelessWidget {
  const TechAgendaTab({super.key});

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
            const Icon(Icons.event_busy, size: 64, color: AppTheme.textMuted),
            const SizedBox(height: 16),
            const Text(
              'No tienes citas asignadas para hoy.',
              style: TextStyle(color: AppTheme.textLight, fontSize: 18),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => provider.fetchCitas(),
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
        final cliente = cita['cliente'];
        final vehiculo = cita['vehiculo'];
        final servicio = cita['servicio'];
        final estado = cita['estado'];
        final fase = cita['faseActual'] ?? 'Pendiente';
        final hora = DateTime.parse(cita['horaAgendada']);

        return Card(
          margin: const EdgeInsets.only(bottom: 16),
          child: ListTile(
            title: Text(
              '${servicio['nombre']} - ${cliente['nombre']}',
              style: const TextStyle(fontWeight: FontWeight.bold),
            ),
            subtitle: Text(
              'Vehículo: ${vehiculo['marca']} ${vehiculo['modelo']} (${vehiculo['placa']})\n'
              'Hora: ${DateFormat('dd/MM HH:mm').format(hora)} • Estado: $estado • Fase: $fase',
            ),
            trailing: const Icon(Icons.arrow_forward_ios, size: 16),
            onTap: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => TechCitaGestionScreen(citaId: id),
                ),
              );
            },
          ),
        );
      },
    );
  }
}

// ─── TAB 2: CITAS DISPONIBLES (SIN ASIGNAR) ──────────────────

class TechDisponiblesTab extends StatelessWidget {
  const TechDisponiblesTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    if (provider.isLoading && provider.citasSinAsignar.isEmpty) {
      return const SkeletonList();
    }

    if (provider.citasSinAsignar.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.assignment_turned_in, size: 64, color: AppTheme.textMuted),
            const SizedBox(height: 16),
            const Text(
              'No hay citas sin asignar disponibles.',
              style: TextStyle(color: AppTheme.textLight, fontSize: 18),
            ),
            const SizedBox(height: 8),
            TextButton(
              onPressed: () => provider.fetchCitasSinAsignar(),
              child: const Text('Actualizar'),
            ),
          ],
        ),
      );
    }

    return ListView.builder(
      padding: const EdgeInsets.all(16),
      itemCount: provider.citasSinAsignar.length,
      itemBuilder: (context, index) {
        final cita = provider.citasSinAsignar[index];
        final id = cita['id'];
        final cliente = cita['cliente'];
        final vehiculo = cita['vehiculo'];
        final servicio = cita['servicio'];
        final hora = DateTime.parse(cita['horaAgendada']);

        return Card(
          margin: const EdgeInsets.only(bottom: 16),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  servicio['nombre'],
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 18, color: AppTheme.textLight),
                ),
                const SizedBox(height: 8),
                Text('Cliente: ${cliente['nombre']}'),
                Text('Vehículo: ${vehiculo['marca']} ${vehiculo['modelo']} (${vehiculo['placa']})'),
                Text('Hora: ${DateFormat('dd/MM HH:mm').format(hora)}'),
                const SizedBox(height: 16),
                Align(
                  alignment: Alignment.centerRight,
                  child: ElevatedButton(
                    onPressed: () async {
                      final advertencia = await provider.aceptarCita(id);
                      if (provider.errorMessage == null) {
                        if (context.mounted) {
                          ScaffoldMessenger.of(context).showSnackBar(
                            SnackBar(
                              content: Text(advertencia ?? 'Cita aceptada con éxito'),
                              backgroundColor: advertencia != null ? AppTheme.warning : AppTheme.success,
                              duration: const Duration(seconds: 4),
                            ),
                          );
                        }
                      }
                    },
                    child: const Text('ACEPTAR TRABAJO'),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

// ─── PANTALLA DE GESTION DE CITA EN CURSO ─────────────────────

class TechCitaGestionScreen extends StatefulWidget {
  final int citaId;
  const TechCitaGestionScreen({super.key, required this.citaId});

  @override
  State<TechCitaGestionScreen> createState() => _TechCitaGestionScreenState();
}

class _TechCitaGestionScreenState extends State<TechCitaGestionScreen> {
  final _picker = ImagePicker();
  final _faseController = TextEditingController();
  final _motivoController = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      Provider.of<TallerProvider>(context, listen: false).fetchCitaDetalle(widget.citaId);
    });
  }

  @override
  void dispose() {
    _faseController.dispose();
    _motivoController.dispose();
    super.dispose();
  }

  void _mostrarSelectorFuenteImagen({required Function(XFile?) onSelected, bool permitirSinFoto = false}) {
    showModalBottomSheet(
      context: context,
      backgroundColor: AppTheme.primary,
      builder: (ctx) {
        return SafeArea(
          child: Wrap(
            children: [
              ListTile(
                leading: const Icon(Icons.photo_camera, color: AppTheme.accent),
                title: const Text('Tomar Foto (Cámara)', style: TextStyle(color: AppTheme.textLight)),
                onTap: () {
                  Navigator.pop(ctx);
                  _capturarYProcesarImagen(ImageSource.camera, onSelected);
                },
              ),
              ListTile(
                leading: const Icon(Icons.photo_library, color: AppTheme.accent),
                title: const Text('Seleccionar de Galería', style: TextStyle(color: AppTheme.textLight)),
                onTap: () {
                  Navigator.pop(ctx);
                  _capturarYProcesarImagen(ImageSource.gallery, onSelected);
                },
              ),
              if (permitirSinFoto)
                ListTile(
                  leading: const Icon(Icons.skip_next, color: AppTheme.success),
                  title: const Text('Continuar sin Foto', style: TextStyle(color: AppTheme.textLight)),
                  onTap: () {
                    Navigator.pop(ctx);
                    onSelected(null);
                  },
                ),
            ],
          ),
        );
      },
    );
  }

  void _capturarYProcesarImagen(ImageSource source, Function(XFile?) onSelected) async {
    final pickedFile = await _picker.pickImage(source: source, imageQuality: 70);
    if (pickedFile != null) {
      onSelected(pickedFile);
    }
  }

  void _iniciar(XFile? foto) async {
    final provider = Provider.of<TallerProvider>(context, listen: false);
    final messenger = ScaffoldMessenger.of(context);
    final exito = await provider.iniciarCita(widget.citaId, foto);
    if (exito) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Trabajo iniciado'), backgroundColor: AppTheme.success),
      );
    }
  }

  void _finalizar() async {
    final provider = Provider.of<TallerProvider>(context, listen: false);
    final messenger = ScaffoldMessenger.of(context);
    final exito = await provider.finalizarCita(widget.citaId);
    if (exito) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Trabajo finalizado'), backgroundColor: AppTheme.success),
      );
    }
  }

  void _posponer() {
    showDialog(
      context: context,
      builder: (dialogContext) {
        return AlertDialog(
          title: const Text('Posponer Cita'),
          content: TextField(
            controller: _motivoController,
            decoration: const InputDecoration(labelText: 'Motivo de posposición'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext),
              child: const Text('Cancelar'),
            ),
            ElevatedButton(
              onPressed: () async {
                if (_motivoController.text.trim().isEmpty) return;
                final provider = Provider.of<TallerProvider>(dialogContext, listen: false);
                final nav = Navigator.of(dialogContext);
                final messenger = ScaffoldMessenger.of(dialogContext);
                final exito = await provider.posponerCita(widget.citaId, _motivoController.text.trim());
                if (exito) {
                  nav.pop();
                  nav.pop();
                  messenger.showSnackBar(
                    const SnackBar(content: Text('Cita pospuesta para reasignar'), backgroundColor: AppTheme.warning),
                  );
                }
              },
              child: const Text('Confirmar'),
            ),
          ],
        );
      },
    );
  }

  void _cancelar() {
    showDialog(
      context: context,
      builder: (dialogContext) {
        return AlertDialog(
          title: const Text('¿Cancelar Cita?'),
          content: const Text('Esto liberará los accesorios reservados de vuelta al stock.'),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext),
              child: const Text('Volver'),
            ),
            ElevatedButton(
              onPressed: () async {
                final provider = Provider.of<TallerProvider>(dialogContext, listen: false);
                final nav = Navigator.of(dialogContext);
                final messenger = ScaffoldMessenger.of(dialogContext);
                final exito = await provider.cancelarCita(widget.citaId);
                if (exito) {
                  nav.pop();
                  nav.pop();
                  messenger.showSnackBar(
                    const SnackBar(content: Text('Cita cancelada y accesorios liberados'), backgroundColor: AppTheme.accent),
                  );
                }
              },
              style: ElevatedButton.styleFrom(backgroundColor: AppTheme.accent),
              child: const Text('Confirmar Cancelar'),
            ),
          ],
        );
      },
    );
  }

  void _abrirWhatsApp(String tipo) async {
    final provider = Provider.of<TallerProvider>(context, listen: false);
    final data = await provider.fetchWhatsAppLinks(widget.citaId, tipo);
    
    if (data != null && data['links'] != null && data['links'].isNotEmpty) {
      final linkUrl = data['links'][0]['link'] as String;
      final uri = Uri.parse(linkUrl);
      if (await canLaunchUrl(uri)) {
        await launchUrl(uri, mode: LaunchMode.externalApplication);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('No se pudo abrir WhatsApp'), backgroundColor: AppTheme.accent),
          );
        }
      }
    }
  }

  void _mostrarDialogoAjustarDuracion(BuildContext context, int duracionActual) {
    final controller = TextEditingController(text: duracionActual > 0 ? duracionActual.toString() : '');
    final formKey = GlobalKey<FormState>();

    showDialog(
      context: context,
      builder: (ctx) {
        return AlertDialog(
          backgroundColor: AppTheme.primary,
          title: const Text('Ajustar Duración Estimada', style: TextStyle(color: AppTheme.textLight)),
          content: Form(
            key: formKey,
            child: TextFormField(
              controller: controller,
              keyboardType: TextInputType.number,
              style: const TextStyle(color: AppTheme.textLight),
              decoration: const InputDecoration(
                labelText: 'Duración en minutos (ej: 60, 120)',
                labelStyle: TextStyle(color: AppTheme.textMuted),
              ),
              validator: (v) {
                if (v == null || v.trim().isEmpty) return 'Requerido';
                final n = int.tryParse(v.trim());
                if (n == null || n <= 0) return 'Ingrese un número mayor a cero';
                return null;
              },
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancelar'),
            ),
            ElevatedButton(
              onPressed: () async {
                if (!formKey.currentState!.validate()) return;
                final minutos = int.parse(controller.text.trim());
                final provider = Provider.of<TallerProvider>(context, listen: false);
                final ok = await provider.actualizarDuracionCita(widget.citaId, minutos);
                if (ctx.mounted) Navigator.pop(ctx);
                if (ctx.mounted && ok) {
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('Duración actualizada'), backgroundColor: AppTheme.success),
                  );
                }
              },
              style: ElevatedButton.styleFrom(backgroundColor: AppTheme.accent),
              child: const Text('Guardar', style: TextStyle(color: AppTheme.primary)),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);
    final cita = provider.citaDetalle;

    if (provider.isLoading || cita == null) {
      return const Scaffold(body: SkeletonDetail());
    }

    final estado = cita['estado'];
    final fase = cita['faseActual'] ?? 'Inspección';
    final duracion = cita['duracionEstimadaMinutos'] as int;

    return Scaffold(
      appBar: AppBar(
        title: const Text('GESTIÓN DE TRABAJO'),
        actions: [
          IconButton(
            icon: const Icon(Icons.chat_bubble_outline),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(
                  builder: (_) => ChatScreen(citaId: widget.citaId, tecnicoNombre: cita['cliente']['nombre']),
                ),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            GlassCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${cita['servicio'] != null ? cita['servicio']['nombre'] : (cita['serviciosDetalle'] ?? 'Servicio Personalizado')}',
                    style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppTheme.accent),
                  ),
                  const SizedBox(height: 8),
                  Text('Cliente: ${cita['cliente']['nombre']}', style: const TextStyle(fontSize: 16)),
                  Text('Vehículo: ${cita['vehiculo']['marca']} ${cita['vehiculo']['modelo']}${cita['vehiculo']['placa'] != null && cita['vehiculo']['placa'].toString().trim().isNotEmpty ? " (${cita['vehiculo']['placa']})" : ""}'),
                  Text('Estado actual: $estado', style: const TextStyle(fontWeight: FontWeight.bold, color: AppTheme.warning)),
                  Text('Fase actual: $fase'),
                ],
              ),
            ),
            const SizedBox(height: 16),

            // Ajustar duración (especialmente para personalizados o en curso)
            if (estado == 'ACEPTADA' || estado == 'EN_CURSO' || estado == 'PAUSADA') ...[
              Card(
                color: AppTheme.border,
                child: Padding(
                  padding: const EdgeInsets.all(12.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text(
                        'Duración Estimada',
                        style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        duracion == 0 
                            ? '⚠️ Duración no asignada aún (Servicio Personalizado)' 
                            : 'Duración actual: $duracion minutos',
                        style: const TextStyle(color: AppTheme.textLight),
                      ),
                      const SizedBox(height: 12),
                      Row(
                        children: [
                          Expanded(
                            child: ElevatedButton.icon(
                              onPressed: () => _mostrarDialogoAjustarDuracion(context, duracion),
                              icon: const Icon(Icons.timer_outlined, size: 18),
                              label: const Text('Establecer/Ajustar Tiempo'),
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],

            // 1. Botones de transicion de estados
            const Text('Acciones del Técnico:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 8),
            if (estado == 'ACEPTADA')
              ElevatedButton.icon(
                onPressed: () => _mostrarSelectorFuenteImagen(onSelected: _iniciar, permitirSinFoto: true),
                icon: const Icon(Icons.play_arrow),
                label: const Text('INICIAR TRABAJO (OPCIONES)'),
              ),
            if (estado == 'EN_CURSO') ...[
              Row(
                children: [
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: () => provider.pausarCita(widget.citaId),
                      icon: const Icon(Icons.pause),
                      style: ElevatedButton.styleFrom(backgroundColor: AppTheme.warning),
                      label: const Text('PAUSAR'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: ElevatedButton.icon(
                      onPressed: _finalizar,
                      icon: const Icon(Icons.stop),
                      style: ElevatedButton.styleFrom(backgroundColor: AppTheme.success),
                      label: const Text('FINALIZAR'),
                    ),
                  ),
                ],
              ),
            ],
            if (estado == 'PAUSADA')
              ElevatedButton.icon(
                onPressed: () => provider.reanudarCita(widget.citaId),
                icon: const Icon(Icons.play_arrow),
                style: ElevatedButton.styleFrom(backgroundColor: Colors.blue),
                label: const Text('REANUDAR TRABAJO'),
              ),
            if (estado == 'FINALIZADA')
              ElevatedButton.icon(
                onPressed: () => provider.entregarVehiculo(widget.citaId),
                icon: const Icon(Icons.check),
                style: ElevatedButton.styleFrom(backgroundColor: Colors.purple),
                label: const Text('MARCAR COMO ENTREGADO AL CLIENTE'),
              ),

            // Botones de posponer/cancelar
            if (estado == 'ACEPTADA' || estado == 'EN_CURSO' || estado == 'PAUSADA') ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _posponer,
                      icon: const Icon(Icons.update, color: AppTheme.warning),
                      label: const Text('POSPONER', style: TextStyle(color: AppTheme.warning)),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: OutlinedButton.icon(
                      onPressed: _cancelar,
                      icon: const Icon(Icons.cancel, color: AppTheme.accent),
                      label: const Text('CANCELAR', style: TextStyle(color: AppTheme.accent)),
                    ),
                  ),
                ],
              ),
            ],

            const Divider(height: 32, color: AppTheme.border),

            // 2. Control de Fase Actual
            if (estado == 'EN_CURSO' || estado == 'PAUSADA') ...[
              const Text('Control de Fases del Proyecto:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
              const SizedBox(height: 8),
              TextField(
                controller: _faseController,
                maxLength: 30,
                decoration: const InputDecoration(
                  labelText: 'Fase del proyecto (Texto Libre)',
                  hintText: 'Ej: Pulido final, Secando pintura...',
                ),
              ),
              Row(
                children: [
                  Expanded(
                    child: ElevatedButton(
                      onPressed: () {
                        if (_faseController.text.trim().isEmpty) return;
                        provider.cambiarFase(widget.citaId, _faseController.text.trim());
                        _faseController.clear();
                      },
                      child: const Text('ACTUALIZAR FASE'),
                    ),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: OutlinedButton(
                      onPressed: () => provider.avanzarFase(widget.citaId),
                      child: const Text('SIGUIENTE FASE PREDETERMINADA'),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              OutlinedButton(
                onPressed: () => provider.resetearFase(widget.citaId),
                child: const Text('RESETEAR A PLANTILLA POR DEFECTO'),
              ),
              const Divider(height: 32, color: AppTheme.border),
            ],

            // 3. Slider de Ajuste de Duración
            if (estado == 'ACEPTADA' || estado == 'EN_CURSO' || estado == 'PAUSADA') ...[
              const Text('Ajuste de Duración Estimada:', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
              const SizedBox(height: 8),
              Text('$duracion Minutos', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppTheme.accent)),
              Slider(
                min: 30,
                max: 720,
                divisions: 23,
                value: duracion.toDouble(),
                onChanged: (val) {
                  // slider local
                },
                onChangeEnd: (val) {
                  provider.ajustarDuracion(widget.citaId, val.toInt());
                },
              ),
              const Divider(height: 32, color: AppTheme.border),
            ],

            // 4. Notificaciones Externas (WhatsApp Manual)
            const Text('Notificar al Cliente (WhatsApp Manual):', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16)),
            const SizedBox(height: 8),
            Wrap(
              spacing: 8,
              children: [
                ActionChip(
                  avatar: const Icon(Icons.timer, size: 16, color: Colors.green),
                  label: const Text('Nueva Hora de Entrega'),
                  onPressed: () => _abrirWhatsApp('cambio_hora'),
                ),
                ActionChip(
                  avatar: const Icon(Icons.update, size: 16, color: Colors.green),
                  label: const Text('Aviso Posponer Cita'),
                  onPressed: () => _abrirWhatsApp('pospuesta'),
                ),
                ActionChip(
                  avatar: const Icon(Icons.check_circle, size: 16, color: Colors.green),
                  label: const Text('Vehículo Listo para Entrega'),
                  onPressed: () => _abrirWhatsApp('lista_entrega'),
                ),
              ],
            ),
            const SizedBox(height: 16),
            const Divider(height: 32, color: AppTheme.border),

            // 5. Galería de Evidencias (Estructuradas)
            const Text(
              'Galería de Evidencias (Fotos y Videos):',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16, color: AppTheme.textLight),
            ),
            const SizedBox(height: 4),
            _buildEvidenciasPorTipo(cita, 'INICIO', 'Evidencias al Inicio'),
            _buildEvidenciasPorTipo(cita, 'AVANCE', 'Evidencias de Avances (Fase)'),
            _buildEvidenciasPorTipo(cita, 'FIN', 'Evidencias al Finalizar'),
            const SizedBox(height: 24),
          ],
        ),
      ),
    );
  }

  void _elegirYSubirEvidencia(String tipo, bool esVideo) async {
    showModalBottomSheet(
      context: context,
      backgroundColor: AppTheme.primary,
      builder: (ctx) {
        return SafeArea(
          child: Wrap(
            children: [
              ListTile(
                leading: Icon(esVideo ? Icons.videocam : Icons.photo_camera, color: AppTheme.accent),
                title: Text(esVideo ? 'Grabar Video' : 'Tomar Foto (Cámara)', style: const TextStyle(color: AppTheme.textLight)),
                onTap: () {
                  Navigator.pop(ctx);
                  _procesarCargaEvidencia(tipo, esVideo, ImageSource.camera);
                },
              ),
              ListTile(
                leading: Icon(esVideo ? Icons.video_library : Icons.photo_library, color: AppTheme.accent),
                title: Text(esVideo ? 'Elegir Video de Galería' : 'Seleccionar de Galería', style: const TextStyle(color: AppTheme.textLight)),
                onTap: () {
                  Navigator.pop(ctx);
                  _procesarCargaEvidencia(tipo, esVideo, ImageSource.gallery);
                },
              ),
            ],
          ),
        );
      },
    );
  }

  void _procesarCargaEvidencia(String tipo, bool esVideo, ImageSource source) async {
    XFile? file;
    if (esVideo) {
      file = await _picker.pickVideo(source: source);
    } else {
      file = await _picker.pickImage(source: source, imageQuality: 70);
    }

    if (file == null) return;
    if (!mounted) return;

    final provider = Provider.of<TallerProvider>(context, listen: false);
    final messenger = ScaffoldMessenger.of(context);
    final cita = provider.citaDetalle;
    final fase = (tipo == 'AVANCE') ? (cita?['faseActual'] ?? 'Avance') : null;

    final exito = await provider.subirEvidencia(widget.citaId, file, tipo, fase);
    if (exito) {
      messenger.showSnackBar(
        const SnackBar(content: Text('Evidencia subida con éxito'), backgroundColor: AppTheme.success),
      );
    } else {
      messenger.showSnackBar(
        SnackBar(content: Text(provider.errorMessage ?? 'Error al subir evidencia'), backgroundColor: Colors.red),
      );
    }
  }

  Widget _buildEvidenciasPorTipo(Map<String, dynamic> cita, String tipo, String titulo) {
    final evs = (cita['evidencias'] as List? ?? [])
        .where((e) => e['tipo'] == tipo)
        .toList();

    final fotos = evs.where((e) => e['archivoTipo'] == 'IMAGEN').toList();
    final videos = evs.where((e) => e['archivoTipo'] == 'VIDEO').toList();

    return Card(
      margin: const EdgeInsets.only(top: 12),
      child: Padding(
        padding: const EdgeInsets.all(12.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  titulo,
                  style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                ),
                Text(
                  'Fotos: ${fotos.length}/5 • Videos: ${videos.length}/1',
                  style: const TextStyle(fontSize: 12, color: AppTheme.textMuted),
                ),
              ],
            ),
            const SizedBox(height: 8),
            if (evs.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 8.0),
                child: Text(
                  'Sin evidencias registradas',
                  style: TextStyle(fontStyle: FontStyle.italic, color: AppTheme.textMuted, fontSize: 13),
                ),
              )
            else
              SizedBox(
                height: 80,
                child: ListView.builder(
                  scrollDirection: Axis.horizontal,
                  itemCount: evs.length,
                  itemBuilder: (ctx, idx) {
                    final ev = evs[idx];
                    final esVid = ev['archivoTipo'] == 'VIDEO';

                    return Container(
                      margin: const EdgeInsets.only(right: 8),
                      width: 80,
                      decoration: BoxDecoration(
                        color: Colors.black26,
                        borderRadius: BorderRadius.circular(6),
                        border: Border.all(color: AppTheme.border),
                      ),
                      child: ClipRRect(
                        borderRadius: BorderRadius.circular(6),
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
            const SizedBox(height: 8),
            // Botones de carga
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton.icon(
                  onPressed: fotos.length >= 5
                      ? null
                      : () => _elegirYSubirEvidencia(tipo, false),
                  icon: const Icon(Icons.add_a_photo, size: 18),
                  label: const Text('Foto', style: TextStyle(fontSize: 12)),
                ),
                const SizedBox(width: 8),
                TextButton.icon(
                  onPressed: videos.isNotEmpty
                      ? null
                      : () => _elegirYSubirEvidencia(tipo, true),
                  icon: const Icon(Icons.video_call, size: 18),
                  label: const Text('Video', style: TextStyle(fontSize: 12)),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

// ─── DIÁLOGO: CREAR CITA POR EL TÉCNICO ─────────────────────────

class TechCrearCitaDialog extends StatefulWidget {
  const TechCrearCitaDialog({super.key});

  @override
  State<TechCrearCitaDialog> createState() => _TechCrearCitaDialogState();
}

class _TechCrearCitaDialogState extends State<TechCrearCitaDialog> {
  final _formKey = GlobalKey<FormState>();
  final _clienteFormKey = GlobalKey<FormState>();
  final _vehiculoFormKey = GlobalKey<FormState>();

  // Búsqueda de clientes
  final _searchController = TextEditingController();
  Map<String, dynamic>? _selectedCliente;
  bool _buscandoClientes = false;
  List<dynamic> _sugerenciasClientes = [];

  // Mini-formulario nuevo cliente
  bool _mostrarFormNuevoCliente = false;
  final _cliNombreController = TextEditingController();
  final _cliTelefonoController = TextEditingController();
  final _cliPinController = TextEditingController();

  // Mini-formulario nuevo vehículo
  bool _mostrarFormNuevoVehiculo = false;
  final _vehMarcaController = TextEditingController();
  final _vehModeloController = TextEditingController();
  final _vehAnioController = TextEditingController();
  final _vehPlacaController = TextEditingController();

  // Campos de la cita
  int? _selectedVehiculoId;
  final List<int?> _selectedServicioIds = [null];
  final _customServicioController = TextEditingController();
  DateTime _fechaProgramada = DateTime.now().add(const Duration(hours: 1));
  bool _autoAceptar = false;
  final _adicionalesController = TextEditingController();

  bool _procesando = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.fetchServicios(soloActivos: true);
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    _cliNombreController.dispose();
    _cliTelefonoController.dispose();
    _cliPinController.dispose();
    _vehMarcaController.dispose();
    _vehModeloController.dispose();
    _vehAnioController.dispose();
    _vehPlacaController.dispose();
    _customServicioController.dispose();
    _adicionalesController.dispose();
    super.dispose();
  }

  void _buscarClientes(String query) async {
    if (query.trim().isEmpty) {
      setState(() {
        _sugerenciasClientes = [];
      });
      return;
    }
    setState(() {
      _buscandoClientes = true;
    });

    final provider = Provider.of<TallerProvider>(context, listen: false);
    await provider.fetchClientes(busqueda: query);

    if (mounted) {
      setState(() {
        _sugerenciasClientes = provider.clientes;
        _buscandoClientes = false;
      });
    }
  }

  void _seleccionarCliente(Map<String, dynamic> cliente) {
    setState(() {
      _selectedCliente = cliente;
      _sugerenciasClientes = [];
      _searchController.text = cliente['nombre'];
      _selectedVehiculoId = null;
      _mostrarFormNuevoVehiculo = false;
      _mostrarFormNuevoCliente = false;

      final vehiculos = cliente['vehiculos'] as List? ?? [];
      if (vehiculos.isNotEmpty) {
        _selectedVehiculoId = vehiculos[0]['id'];
      }
    });
  }

  Future<void> _crearCliente() async {
    if (!_clienteFormKey.currentState!.validate()) return;

    setState(() {
      _procesando = true;
    });

    final provider = Provider.of<TallerProvider>(context, listen: false);
    final cliente = await provider.registrarClientePorTecnico(
      nombre: _cliNombreController.text.trim(),
      pin: _cliPinController.text.trim(),
      contactos: [_cliTelefonoController.text.trim()],
    );

    if (mounted) {
      setState(() {
        _procesando = false;
      });

      if (cliente != null) {
        // Enriquecer el cliente localmente con un array de vehículos vacío
        cliente['vehiculos'] = [];
        _seleccionarCliente(cliente);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cliente registrado'), backgroundColor: AppTheme.success),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(provider.errorMessage ?? 'Error al registrar cliente'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _crearVehiculo() async {
    if (_selectedCliente == null) return;
    if (!_vehiculoFormKey.currentState!.validate()) return;

    setState(() {
      _procesando = true;
    });

    final provider = Provider.of<TallerProvider>(context, listen: false);
    final vehiculo = await provider.registrarVehiculo({
      'clienteId': _selectedCliente!['id'],
      'marca': _vehMarcaController.text.trim(),
      'modelo': _vehModeloController.text.trim(),
      'anio': int.parse(_vehAnioController.text.trim()),
      'placa': _vehPlacaController.text.trim(),
    });

    if (mounted) {
      setState(() {
        _procesando = false;
      });

      if (vehiculo != null) {
        setState(() {
          // Agregar el nuevo vehículo a la lista del cliente seleccionado
          final vehiculos = _selectedCliente!['vehiculos'] as List? ?? [];
          vehiculos.add(vehiculo);
          _selectedCliente!['vehiculos'] = vehiculos;
          _selectedVehiculoId = vehiculo['id'];
          _mostrarFormNuevoVehiculo = false;
        });
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Vehículo registrado'), backgroundColor: AppTheme.success),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(provider.errorMessage ?? 'Error al registrar vehículo'), backgroundColor: Colors.red),
        );
      }
    }
  }

  Future<void> _seleccionarFechaYHora() async {
    final date = await showDatePicker(
      context: context,
      initialDate: _fechaProgramada,
      firstDate: DateTime.now(),
      lastDate: DateTime.now().add(const Duration(days: 30)),
    );
    if (date == null) return;

    if (!mounted) return;
    final time = await showTimePicker(
      context: context,
      initialTime: TimeOfDay.fromDateTime(_fechaProgramada),
    );
    if (time == null) return;

    setState(() {
      _fechaProgramada = DateTime(date.year, date.month, date.day, time.hour, time.minute);
    });
  }

  Future<void> _enviarCita() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedCliente == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Debe seleccionar o registrar un cliente'), backgroundColor: Colors.red),
      );
      return;
    }
    if (_selectedVehiculoId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Debe seleccionar o registrar un vehículo'), backgroundColor: Colors.red),
      );
      return;
    }

    final provider = Provider.of<TallerProvider>(context, listen: false);
    List<String> nombres = [];
    int totalDuracion = 0;
    bool esPers = false;
    int? primerServicioId;

    for (final id in _selectedServicioIds) {
      if (id == null) continue;
      if (id == -1) {
        esPers = true;
      } else {
        final s = provider.servicios.firstWhere((x) => x['id'] == id, orElse: () => null);
        if (s != null) {
          nombres.add(s['nombre']);
          totalDuracion += (s['duracionEstimadaMinutos'] as int? ?? 0);
          primerServicioId ??= s['id'];
        }
      }
    }

    if (esPers && _customServicioController.text.trim().isNotEmpty) {
      nombres.add(_customServicioController.text.trim());
      totalDuracion = 0; // Se estima manualmente por el técnico después
    }

    if (nombres.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Debe seleccionar al menos un servicio'), backgroundColor: Colors.red),
      );
      return;
    }

    setState(() {
      _procesando = true;
    });

    final exitoso = await provider.agendarCita({
      'clienteId': _selectedCliente!['id'],
      'vehiculoId': _selectedVehiculoId,
      'servicioId': primerServicioId,
      'serviciosDetalle': nombres.join(' + '),
      'esPersonalizado': esPers,
      'duracionEstimadaMinutos': totalDuracion,
      'horaAgendada': _autoAceptar ? DateTime.now().toIso8601String() : _fechaProgramada.toIso8601String(),
      'esInmediata': _autoAceptar,
      'autoAceptar': _autoAceptar,
      'adicionales': _adicionalesController.text.trim().isEmpty ? null : _adicionalesController.text.trim(),
    });

    if (mounted) {
      setState(() {
        _procesando = false;
      });

      if (exitoso) {
        // Refrescar agendas
        provider.fetchCitas(todas: false);
        provider.fetchCitasSinAsignar();

        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Cita agendada con éxito'), backgroundColor: AppTheme.success),
        );
      } else {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(provider.errorMessage ?? 'Error al crear la cita'), backgroundColor: Colors.red),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Dialog(
      backgroundColor: AppTheme.primary,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500),
        padding: const EdgeInsets.all(20),
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text(
                    'Crear Nueva Cita',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppTheme.textLight),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close, color: AppTheme.textMuted),
                    onPressed: () => Navigator.pop(context),
                  ),
                ],
              ),
              const Divider(color: AppTheme.border),
              const SizedBox(height: 12),

              // ─── SECCIÓN: CLIENTE ─────────────────────────────────
              if (!_mostrarFormNuevoCliente) ...[
                const Text('Cliente', style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent)),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _searchController,
                        onChanged: _buscarClientes,
                        decoration: InputDecoration(
                          hintText: 'Buscar por nombre...',
                          prefixIcon: const Icon(Icons.search),
                          suffixIcon: _searchController.text.isNotEmpty
                              ? IconButton(
                                  icon: const Icon(Icons.clear),
                                  onPressed: () {
                                    setState(() {
                                      _searchController.clear();
                                      _selectedCliente = null;
                                      _selectedVehiculoId = null;
                                      _sugerenciasClientes = [];
                                    });
                                  },
                                )
                              : null,
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    IconButton(
                      icon: const Icon(Icons.person_add_alt_1, color: AppTheme.accent),
                      onPressed: () {
                        setState(() {
                          _mostrarFormNuevoCliente = true;
                        });
                      },
                      tooltip: 'Registrar Cliente Nuevo',
                    ),
                  ],
                ),
                if (_buscandoClientes)
                  const Padding(
                    padding: EdgeInsets.all(8.0),
                    child: Center(child: CircularProgressIndicator(color: AppTheme.accent)),
                  ),
                if (_sugerenciasClientes.isNotEmpty)
                  Card(
                    color: AppTheme.border,
                    child: Column(
                      children: _sugerenciasClientes.map<Widget>((c) {
                        return ListTile(
                          title: Text(c['nombre'], style: const TextStyle(color: AppTheme.textLight)),
                          subtitle: Text('Contactos: ${(c['contactos'] as List).join(', ')}',
                              style: const TextStyle(color: AppTheme.textMuted)),
                          onTap: () => _seleccionarCliente(c),
                        );
                      }).toList(),
                    ),
                  ),
              ] else ...[
                // Formulario nuevo cliente
                Form(
                  key: _clienteFormKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text('Registrar Nuevo Cliente', style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent)),
                          TextButton(
                            onPressed: () => setState(() => _mostrarFormNuevoCliente = false),
                            child: const Text('Buscar existente'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 8),
                      TextFormField(
                        controller: _cliNombreController,
                        decoration: const InputDecoration(labelText: 'Nombre del Cliente'),
                        validator: (v) => v == null || v.trim().isEmpty ? 'Requerido' : null,
                      ),
                      const SizedBox(height: 8),
                      TextFormField(
                        controller: _cliTelefonoController,
                        keyboardType: TextInputType.phone,
                        decoration: const InputDecoration(labelText: 'Teléfono / Celular'),
                        validator: (v) => v == null || v.trim().isEmpty ? 'Requerido' : null,
                      ),
                      const SizedBox(height: 8),
                      TextFormField(
                        controller: _cliPinController,
                        keyboardType: TextInputType.number,
                        obscureText: true,
                        decoration: const InputDecoration(labelText: 'PIN de Acceso (4-8 dígitos)'),
                        validator: (v) {
                          if (v == null || v.trim().isEmpty) return 'Requerido';
                          if (v.trim().length < 4 || v.trim().length > 8) return 'Debe tener entre 4 y 8 dígitos';
                          if (int.tryParse(v) == null) return 'Debe ser solo números';
                          return null;
                        },
                      ),
                      const SizedBox(height: 12),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton(
                            onPressed: () => setState(() => _mostrarFormNuevoCliente = false),
                            child: const Text('Cancelar'),
                          ),
                          const SizedBox(width: 8),
                          ElevatedButton(
                            onPressed: _procesando ? null : _crearCliente,
                            style: ElevatedButton.styleFrom(backgroundColor: AppTheme.accent),
                            child: const Text('Guardar Cliente', style: TextStyle(color: AppTheme.primary)),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
              const SizedBox(height: 16),

              // ─── SECCIÓN: VEHÍCULO ────────────────────────────────
              if (_selectedCliente != null) ...[
                if (!_mostrarFormNuevoVehiculo) ...[
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('Vehículo', style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent)),
                      IconButton(
                        icon: const Icon(Icons.add_road, color: AppTheme.accent),
                        onPressed: () {
                          setState(() {
                            _mostrarFormNuevoVehiculo = true;
                          });
                        },
                        tooltip: 'Registrar Vehículo Nuevo',
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  DropdownButtonFormField<int>(
                    value: _selectedVehiculoId,
                    decoration: const InputDecoration(labelText: 'Seleccione un Vehículo'),
                    items: (_selectedCliente!['vehiculos'] as List? ?? []).map<DropdownMenuItem<int>>((v) {
                      final placaText = (v['placa'] != null && v['placa'].toString().trim().isNotEmpty)
                          ? ' (${v['placa']})'
                          : '';
                      return DropdownMenuItem<int>(
                        value: v['id'],
                        child: Text('${v['marca']} ${v['modelo']}$placaText',
                            style: const TextStyle(color: AppTheme.textLight)),
                      );
                    }).toList(),
                    onChanged: (val) {
                      setState(() {
                        _selectedVehiculoId = val;
                      });
                    },
                    validator: (v) => v == null ? 'Seleccione un vehículo' : null,
                  ),
                ] else ...[
                  // Formulario nuevo vehículo
                  Form(
                    key: _vehiculoFormKey,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            const Text('Registrar Nuevo Vehículo', style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent)),
                            TextButton(
                              onPressed: () => setState(() => _mostrarFormNuevoVehiculo = false),
                              child: const Text('Cancelar'),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _vehMarcaController,
                          decoration: const InputDecoration(labelText: 'Marca (ej: Toyota)'),
                          validator: (v) => v == null || v.trim().isEmpty ? 'Requerido' : null,
                        ),
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _vehModeloController,
                          decoration: const InputDecoration(labelText: 'Modelo (ej: Hilux)'),
                          validator: (v) => v == null || v.trim().isEmpty ? 'Requerido' : null,
                        ),
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _vehAnioController,
                          keyboardType: TextInputType.number,
                          decoration: const InputDecoration(labelText: 'Año (ej: 2022)'),
                          validator: (v) {
                            if (v == null || v.trim().isEmpty) return 'Requerido';
                            if (int.tryParse(v) == null) return 'Año inválido';
                            return null;
                          },
                        ),
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _vehPlacaController,
                          decoration: const InputDecoration(labelText: 'Placa / Matrícula (Opcional)'),
                        ),
                        const SizedBox(height: 12),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.end,
                          children: [
                            TextButton(
                              onPressed: () => setState(() => _mostrarFormNuevoVehiculo = false),
                              child: const Text('Cancelar'),
                            ),
                            const SizedBox(width: 8),
                            ElevatedButton(
                              onPressed: _procesando ? null : _crearVehiculo,
                              style: ElevatedButton.styleFrom(backgroundColor: AppTheme.accent),
                              child: const Text('Guardar Vehículo', style: TextStyle(color: AppTheme.primary)),
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
                const SizedBox(height: 16),
              ],

              // ─── DETALLES DE LA CITA ──────────────────────────────
              if (_selectedCliente != null && !_mostrarFormNuevoVehiculo && !_mostrarFormNuevoCliente) ...[
                const Divider(color: AppTheme.border),
                const SizedBox(height: 8),
                Form(
                  key: _formKey,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      const Text('Servicios Requeridos', style: TextStyle(fontWeight: FontWeight.bold, color: AppTheme.accent)),
                      const SizedBox(height: 8),

                      // Column of service dropdowns
                      Column(
                        children: List.generate(_selectedServicioIds.length, (index) {
                          return Padding(
                            padding: const EdgeInsets.only(bottom: 8.0),
                            child: Row(
                              children: [
                                Expanded(
                                  child: DropdownButtonFormField<int>(
                                    value: _selectedServicioIds[index],
                                    decoration: InputDecoration(labelText: 'Servicio #${index + 1}'),
                                    items: [
                                      ...provider.servicios.map<DropdownMenuItem<int>>((s) {
                                        return DropdownMenuItem<int>(
                                          value: s['id'],
                                          child: Text(s['nombre'], style: const TextStyle(color: AppTheme.textLight)),
                                        );
                                      }),
                                      const DropdownMenuItem<int>(
                                        value: -1,
                                        child: Text('+ Servicio Personalizado (Texto libre)', style: TextStyle(color: AppTheme.accent, fontWeight: FontWeight.bold)),
                                      ),
                                    ],
                                    onChanged: (val) {
                                      setState(() {
                                        _selectedServicioIds[index] = val;
                                      });
                                    },
                                    validator: (v) => v == null ? 'Seleccione un servicio' : null,
                                  ),
                                ),
                                if (index > 0)
                                  IconButton(
                                    icon: const Icon(Icons.remove_circle_outline, color: Colors.redAccent),
                                    onPressed: () {
                                      setState(() {
                                        _selectedServicioIds.removeAt(index);
                                      });
                                    },
                                  ),
                                if (index == _selectedServicioIds.length - 1)
                                  IconButton(
                                    icon: const Icon(Icons.add_circle, color: AppTheme.accent),
                                    onPressed: () {
                                      setState(() {
                                        _selectedServicioIds.add(null);
                                      });
                                    },
                                  ),
                              ],
                            ),
                          );
                        }),
                      ),

                      if (_selectedServicioIds.contains(-1)) ...[
                        const SizedBox(height: 8),
                        TextFormField(
                          controller: _customServicioController,
                          decoration: const InputDecoration(
                            labelText: 'Especificar Servicio Personalizado',
                            hintText: 'Ej: Alineación de faros, reparación de chasis...',
                          ),
                          validator: (v) => v == null || v.trim().isEmpty ? 'Ingrese el servicio personalizado' : null,
                        ),
                      ],
                      const SizedBox(height: 16),

                      // Auto-aceptar Cita Switch
                      SwitchListTile(
                        title: const Text('Auto-aceptar esta cita', style: TextStyle(color: AppTheme.textLight)),
                        subtitle: const Text('Se asignará a tu cola de trabajo inmediatamente',
                            style: TextStyle(color: AppTheme.textMuted, fontSize: 12)),
                        value: _autoAceptar,
                        activeColor: AppTheme.accent,
                        onChanged: (val) {
                          setState(() {
                            _autoAceptar = val;
                          });
                        },
                      ),
                      const SizedBox(height: 8),

                      if (!_autoAceptar) ...[
                        ListTile(
                          shape: RoundedRectangleBorder(
                            side: const BorderSide(color: AppTheme.border),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          title: const Text(
                            'Fecha y Hora Programadas:',
                            style: TextStyle(color: AppTheme.textMuted, fontSize: 12),
                          ),
                          subtitle: Text(
                            DateFormat('dd/MM/yyyy HH:mm').format(_fechaProgramada),
                            style: const TextStyle(color: AppTheme.textLight, fontSize: 16, fontWeight: FontWeight.bold),
                          ),
                          trailing: const Icon(Icons.calendar_month, color: AppTheme.accent),
                          onTap: _seleccionarFechaYHora,
                        ),
                      ] else ...[
                        const Card(
                          color: AppTheme.border,
                          child: Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Row(
                              children: [
                                Icon(Icons.info_outline, color: AppTheme.accent, size: 16),
                                SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    'Se programará e iniciará automáticamente de acuerdo a tu cola de citas activa.',
                                    style: TextStyle(color: AppTheme.textMuted, fontSize: 11),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ],
                      const SizedBox(height: 12),

                      // Observaciones
                      TextFormField(
                        controller: _adicionalesController,
                        maxLines: 2,
                        decoration: const InputDecoration(
                          labelText: 'Observaciones / Detalles Adicionales',
                          hintText: 'Detalles sobre problemas del vehículo o solicitudes...',
                        ),
                      ),
                      const SizedBox(height: 20),

                      // Botones de acción final
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          TextButton(
                            onPressed: () => Navigator.pop(context),
                            child: const Text('Cancelar'),
                          ),
                          const SizedBox(width: 8),
                          ElevatedButton(
                            onPressed: _procesando ? null : _enviarCita,
                            style: ElevatedButton.styleFrom(
                              backgroundColor: AppTheme.accent,
                              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                            ),
                            child: _procesando
                                ? const SizedBox(
                                    width: 20,
                                    height: 20,
                                    child: CircularProgressIndicator(color: AppTheme.primary, strokeWidth: 2),
                                  )
                                : const Text('Crear Cita', style: TextStyle(color: AppTheme.primary, fontWeight: FontWeight.bold)),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

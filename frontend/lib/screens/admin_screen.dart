import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:intl/intl.dart';
import '../config/theme.dart';
import '../providers/auth_provider.dart';
import '../providers/taller_provider.dart';
import '../services/socket_service.dart';
import '../widgets/skeleton_loader.dart';

class AdminScreen extends StatefulWidget {
  const AdminScreen({super.key});

  @override
  State<AdminScreen> createState() => _AdminScreenState();
}

class _AdminScreenState extends State<AdminScreen> {
  int _selectedMenuIndex = 0;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final provider = Provider.of<TallerProvider>(context, listen: false);
      provider.fetchAdminDashboard();
      provider.fetchTecnicos();
      provider.fetchServicios();
      provider.fetchAccesorios();
      provider.fetchCitasPospuestas();
      provider.fetchReporteProductividad();
      provider.fetchReporteServicios();

      // Conectar socket y suscribirse a eventos de tiempo real (INC-01)
      final socketService = Provider.of<SocketService>(context, listen: false);
      socketService.connect();
      socketService.on('tecnico:disponibilidadCambiada', (data) {
        // Actualizar dashboard cuando un técnico cambia de estado
        provider.fetchAdminDashboard();
        provider.fetchTecnicos();
      });
      socketService.on('cita:estadoCambiado', (data) {
        // Actualizar dashboard cuando cambia el estado de una cita
        provider.fetchAdminDashboard();
        provider.fetchCitasPospuestas();
      });
      socketService.on('cita:pospuesta', (data) {
        provider.fetchCitasPospuestas();
      });
    });
  }

  late SocketService _socketService;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    _socketService = Provider.of<SocketService>(context, listen: false);
  }

  @override
  void dispose() {
    // Desconectar socket al salir del panel admin
    _socketService.disconnect();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final auth = Provider.of<AuthProvider>(context);
    final isDesktop = MediaQuery.of(context).size.width > 800;

    final List<Widget> pages = [
      const AdminDashboardTab(),
      const AdminPospuestasTab(),
      const AdminTecnicosTab(),
      const AdminServiciosTab(),
      const AdminAccesoriosTab(),
      const AdminReportesTab(),
    ];

    final List<Map<String, dynamic>> menuItems = [
      {'icon': Icons.dashboard_outlined, 'title': 'Dashboard'},
      {'icon': Icons.warning_amber_outlined, 'title': 'Citas Pospuestas'},
      {'icon': Icons.people_outline, 'title': 'Técnicos'},
      {'icon': Icons.design_services_outlined, 'title': 'Servicios'},
      {'icon': Icons.inventory_2_outlined, 'title': 'Accesorios'},
      {'icon': Icons.analytics_outlined, 'title': 'Reportes & Stats'},
    ];

    if (!isDesktop) {
      // Diseño Móvil / Compacto
      return Scaffold(
        appBar: AppBar(
          title: const Text('PANEL ADMINISTRADOR'),
          actions: [
            IconButton(
              icon: const Icon(Icons.logout, color: AppTheme.accent),
              onPressed: () => auth.logout(),
            )
          ],
        ),
        drawer: Drawer(
          backgroundColor: AppTheme.primary,
          child: Column(
            children: [
              const DrawerHeader(
                child: Center(
                  child: Text(
                    'Taller Continental',
                    style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: AppTheme.accent),
                  ),
                ),
              ),
              Expanded(
                child: ListView.builder(
                  itemCount: menuItems.length,
                  itemBuilder: (context, index) {
                    return ListTile(
                      leading: Icon(menuItems[index]['icon'], color: _selectedMenuIndex == index ? AppTheme.accent : AppTheme.textMuted),
                      title: Text(menuItems[index]['title']),
                      selected: _selectedMenuIndex == index,
                      onTap: () {
                        setState(() {
                          _selectedMenuIndex = index;
                        });
                        Navigator.pop(context);
                      },
                    );
                  },
                ),
              )
            ],
          ),
        ),
        body: pages[_selectedMenuIndex],
      );
    }

    // Diseño Desktop con menú lateral
    return Scaffold(
      body: Row(
        children: [
          // Barra lateral fija
          Container(
            width: 260,
            color: AppTheme.primary,
            child: Column(
              children: [
                const SizedBox(height: 32),
                const Icon(Icons.build_circle_outlined, size: 64, color: AppTheme.accent),
                const SizedBox(height: 12),
                const Text(
                  'CONTINENTAL APP',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, letterSpacing: 1.0, color: AppTheme.textLight),
                ),
                const Text('Admin Dashboard', style: TextStyle(fontSize: 12, color: AppTheme.textMuted)),
                const SizedBox(height: 32),
                Expanded(
                  child: ListView.builder(
                    itemCount: menuItems.length,
                    itemBuilder: (context, index) {
                      final selected = _selectedMenuIndex == index;
                      return Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                        child: Material(
                          color: selected ? AppTheme.accent.withValues(alpha: 0.15) : Colors.transparent,
                          borderRadius: BorderRadius.circular(8),
                          child: ListTile(
                            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                            leading: Icon(
                              menuItems[index]['icon'],
                              color: selected ? AppTheme.accent : AppTheme.textMuted,
                            ),
                            title: Text(
                              menuItems[index]['title'],
                              style: TextStyle(
                                color: selected ? AppTheme.textLight : AppTheme.textMuted,
                                fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                              ),
                            ),
                            onTap: () {
                              setState(() {
                                _selectedMenuIndex = index;
                              });
                            },
                          ),
                        ),
                      );
                    },
                  ),
                ),
                Material(
                  color: Colors.transparent,
                  child: ListTile(
                    leading: const Icon(Icons.logout, color: AppTheme.accent),
                    title: const Text('Cerrar Sesión', style: TextStyle(color: AppTheme.accent)),
                    onTap: () => auth.logout(),
                  ),
                ),
                const SizedBox(height: 16),
              ],
            ),
          ),
          // Contenido principal
          Expanded(
            child: Container(
              color: AppTheme.background,
              child: pages[_selectedMenuIndex],
            ),
          ),
        ],
      ),
    );
  }
}

// ─── TAB 1: DASHBOARD OPERATIVO ──────────────────────────────

class AdminDashboardTab extends StatelessWidget {
  const AdminDashboardTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);
    final db = provider.adminDashboard;

    if (db == null || (provider.isLoading && db.isEmpty)) {
      return const SkeletonDashboard();
    }

    final tecnicos = db['tecnicos'] as List<dynamic>;
    final conteos = db['citasPorEstado'] as Map<String, dynamic>;

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        const Text('Estado Operativo en Tiempo Real', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
        const SizedBox(height: 24),

        // Tarjetas de métricas rápidas
        Wrap(
          spacing: 16,
          runSpacing: 16,
          children: [
            _buildMetricCard('Citas Pospuestas', '${db['citasPospuestas']}', Icons.warning_amber_outlined, AppTheme.warning),
            _buildMetricCard('Citas Sin Asignar', '${db['citasSinAsignar']}', Icons.schedule, AppTheme.textMuted),
            _buildMetricCard('En Curso', '${conteos['EN_CURSO'] ?? 0}', Icons.play_circle_outline, Colors.blue),
            _buildMetricCard('Completadas', '${conteos['FINALIZADA'] ?? 0}', Icons.check_circle_outline, AppTheme.success),
          ],
        ),
        const SizedBox(height: 32),

        // Lista de Técnicos y disponibilidad
        const Text('Monitoreo de Técnicos:', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
        const SizedBox(height: 12),
        ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: tecnicos.length,
          itemBuilder: (context, index) {
            final t = tecnicos[index];
            final disp = t['disponibilidad'];
            Color dispColor = AppTheme.success;
            if (disp == 'EN_CURSO') dispColor = Colors.blue;
            if (disp == 'PAUSADO') dispColor = AppTheme.warning;

            return Card(
              margin: const EdgeInsets.only(bottom: 12),
              child: ListTile(
                leading: CircleAvatar(
                  backgroundColor: dispColor.withValues(alpha: 0.2),
                  child: Icon(Icons.person, color: dispColor),
                ),
                title: Text(t['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                subtitle: Text('Estado de la cuenta: ${t['estado']}'),
                trailing: Chip(
                  backgroundColor: dispColor.withValues(alpha: 0.15),
                  side: BorderSide(color: dispColor),
                  label: Text(
                    disp,
                    style: TextStyle(color: dispColor, fontWeight: FontWeight.bold, fontSize: 12),
                  ),
                ),
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildMetricCard(String title, String val, IconData icon, Color color) {
    return Container(
      width: 220,
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: AppTheme.primary,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppTheme.border),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, color: color, size: 28),
          const SizedBox(height: 16),
          Text(val, style: TextStyle(fontSize: 32, fontWeight: FontWeight.bold, color: color)),
          const SizedBox(height: 4),
          Text(title, style: const TextStyle(color: AppTheme.textMuted, fontSize: 13)),
        ],
      ),
    );
  }
}

// ─── TAB 2: CITAS POSPUESTAS (REASIGNACION MANUAL) ────────────

class AdminPospuestasTab extends StatelessWidget {
  const AdminPospuestasTab({super.key});

  void _abrirReasignarDialog(BuildContext context, int citaId) {
    final provider = Provider.of<TallerProvider>(context, listen: false);
    int? selectedTecnicoId;
    DateTime reDate = DateTime.now().add(const Duration(hours: 1));

    showDialog(
      context: context,
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              title: const Text('Reasignación Manual'),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  DropdownButtonFormField<int>(
                    decoration: const InputDecoration(labelText: 'Técnico Asignado'),
                    items: provider.tecnicos.where((t) => t['estado'] == 'ACTIVO').map<DropdownMenuItem<int>>((t) {
                      return DropdownMenuItem<int>(
                        value: t['id'],
                        child: Text(t['nombre']),
                      );
                    }).toList(),
                    onChanged: (val) {
                      setState(() {
                        selectedTecnicoId = val;
                      });
                    },
                  ),
                  const SizedBox(height: 16),
                  ListTile(
                    title: const Text('Nueva Fecha/Hora'),
                    subtitle: Text(DateFormat('dd/MM HH:mm').format(reDate)),
                    trailing: const Icon(Icons.calendar_today, color: AppTheme.accent),
                    onTap: () async {
                      final date = await showDatePicker(
                        context: context,
                        initialDate: reDate,
                        firstDate: DateTime.now(),
                        lastDate: DateTime.now().add(const Duration(days: 30)),
                      );
                      if (date == null) return;
                      if (!context.mounted) return;
                      final time = await showTimePicker(
                        context: context,
                        initialTime: TimeOfDay.fromDateTime(reDate),
                      );
                      if (time != null) {
                        setState(() {
                          reDate = DateTime(date.year, date.month, date.day, time.hour, time.minute);
                        });
                      }
                    },
                  )
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.pop(context),
                  child: const Text('Cancelar'),
                ),
                ElevatedButton(
                  onPressed: () async {
                    if (selectedTecnicoId == null) return;
                    final exito = await provider.reasignarCita(citaId, {
                      'tecnicoId': selectedTecnicoId,
                      'nuevaHoraAgendada': reDate.toIso8601String(),
                    });
                    if (exito && context.mounted) {
                      Navigator.pop(context);
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Cita reasignada correctamente'), backgroundColor: AppTheme.success),
                      );
                    }
                  },
                  child: const Text('Reasignar'),
                ),
              ],
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('Bandeja de Citas Pospuestas', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
          const SizedBox(height: 24),
          if (provider.citasPospuestas.isEmpty)
            const Center(child: Text('No hay citas pendientes de reasignación.'))
          else
            ListView.builder(
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              itemCount: provider.citasPospuestas.length,
              itemBuilder: (context, index) {
                final cita = provider.citasPospuestas[index];
                return Card(
                  child: ListTile(
                    title: Text(
                      '${cita['servicio']['nombre']} - ${cita['cliente']['nombre']}',
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                    subtitle: Text(
                      'Vehículo: ${cita['vehiculo']['marca']} ${cita['vehiculo']['modelo']} (${cita['vehiculo']['placa']})\n'
                      'Motivo posposición: ${cita['motivoPosposicion'] ?? "No especificado"}',
                    ),
                    trailing: ElevatedButton(
                      onPressed: () => _abrirReasignarDialog(context, cita['id']),
                      child: const Text('REASIGNAR'),
                    ),
                  ),
                );
              },
            ),
        ],
      ),
    );
  }
}

// ─── TAB 3: CRUD TÉCNICOS ────────────────────────────────────

class AdminTecnicosTab extends StatefulWidget {
  const AdminTecnicosTab({super.key});

  @override
  State<AdminTecnicosTab> createState() => _AdminTecnicosTabState();
}

class _AdminTecnicosTabState extends State<AdminTecnicosTab> {
  final _formKey = GlobalKey<FormState>();
  final _nombreController = TextEditingController();
  final _usuarioController = TextEditingController();
  final _passwordController = TextEditingController();

  @override
  void dispose() {
    _nombreController.dispose();
    _usuarioController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  void _abrirCrearDialog() {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Crear Nuevo Técnico'),
          content: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextFormField(
                  controller: _nombreController,
                  decoration: const InputDecoration(labelText: 'Nombre Completo'),
                  validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _usuarioController,
                  decoration: const InputDecoration(labelText: 'Usuario de Login'),
                  validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  controller: _passwordController,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: 'Contraseña'),
                  validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                ),
              ],
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
                final exito = await provider.crearTecnico({
                  'nombre': _nombreController.text.trim(),
                  'usuario': _usuarioController.text.trim(),
                  'password': _passwordController.text.trim(),
                });
                if (exito) {
                  nav.pop();
                  if (mounted) {
                    _nombreController.clear();
                    _usuarioController.clear();
                    _passwordController.clear();
                  }
                }
              },
              child: const Text('Crear'),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      floatingActionButton: FloatingActionButton(
        onPressed: _abrirCrearDialog,
        backgroundColor: AppTheme.accent,
        child: const Icon(Icons.add),
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('Gestión de Técnicos', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
          const SizedBox(height: 24),
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: provider.tecnicos.length,
            itemBuilder: (context, index) {
              final t = provider.tecnicos[index];
              final id = t['id'];
              final activo = t['estado'] == 'ACTIVO';

              return Card(
                child: ListTile(
                  title: Text(t['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text('Usuario: ${t['usuario']} • Estado: ${t['estado']}'),
                  trailing: Switch(
                    value: activo,
                    activeColor: AppTheme.success,
                    onChanged: (val) {
                      provider.cambiarEstadoTecnico(id, val ? 'ACTIVO' : 'INACTIVO');
                    },
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

// ─── TAB 4: CRUD SERVICIOS ───────────────────────────────────

class AdminServiciosTab extends StatefulWidget {
  const AdminServiciosTab({super.key});

  @override
  State<AdminServiciosTab> createState() => _AdminServiciosTabState();
}

class _AdminServiciosTabState extends State<AdminServiciosTab> {
  final _formKey = GlobalKey<FormState>();
  final _nombreController = TextEditingController();
  final _descController = TextEditingController();
  final _duracionController = TextEditingController();
  final _precioController = TextEditingController();

  @override
  void dispose() {
    _nombreController.dispose();
    _descController.dispose();
    _duracionController.dispose();
    _precioController.dispose();
    super.dispose();
  }

  void _abrirCrearDialog() {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Crear Nuevo Servicio'),
          content: Form(
            key: _formKey,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextFormField(
                    controller: _nombreController,
                    decoration: const InputDecoration(labelText: 'Nombre del Servicio'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _descController,
                    decoration: const InputDecoration(labelText: 'Descripción'),
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _duracionController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Duración Estimada (minutos)'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _precioController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Precio (S/.)'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
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
                final exito = await provider.crearServicio({
                  'nombre': _nombreController.text.trim(),
                  'descripcion': _descController.text.trim(),
                  'duracionEstimadaMinutos': int.parse(_duracionController.text.trim()),
                  'precio': double.parse(_precioController.text.trim()),
                });
                if (exito) {
                  nav.pop();
                  if (mounted) {
                    _nombreController.clear();
                    _descController.clear();
                    _duracionController.clear();
                    _precioController.clear();
                  }
                }
              },
              child: const Text('Crear'),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      floatingActionButton: FloatingActionButton(
        onPressed: _abrirCrearDialog,
        backgroundColor: AppTheme.accent,
        child: const Icon(Icons.add),
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('Gestión de Servicios', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
          const SizedBox(height: 24),
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: provider.servicios.length,
            itemBuilder: (context, index) {
              final s = provider.servicios[index];
              final id = s['id'];
              final activo = s['activo'] == true;

              return Card(
                child: ListTile(
                  title: Text(s['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text('Precio: S/.${s['precio']} • Duración: ${s['duracionEstimadaMinutos']} min'),
                  trailing: Switch(
                    value: activo,
                    activeColor: AppTheme.success,
                    onChanged: (val) {
                      provider.toggleEstadoServicio(id, val);
                    },
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

// ─── TAB 5: CRUD ACCESORIOS Y ALERTAS DE STOCK ────────────────

class AdminAccesoriosTab extends StatefulWidget {
  const AdminAccesoriosTab({super.key});

  @override
  State<AdminAccesoriosTab> createState() => _AdminAccesoriosTabState();
}

class _AdminAccesoriosTabState extends State<AdminAccesoriosTab> {
  final _formKey = GlobalKey<FormState>();
  final _nombreController = TextEditingController();
  final _descController = TextEditingController();
  final _precioController = TextEditingController();
  final _stockController = TextEditingController();
  final _stockMinController = TextEditingController();

  @override
  void dispose() {
    _nombreController.dispose();
    _descController.dispose();
    _precioController.dispose();
    _stockController.dispose();
    _stockMinController.dispose();
    super.dispose();
  }

  void _abrirCrearDialog() {
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Agregar Accesorio'),
          content: Form(
            key: _formKey,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextFormField(
                    controller: _nombreController,
                    decoration: const InputDecoration(labelText: 'Nombre'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _descController,
                    decoration: const InputDecoration(labelText: 'Descripción'),
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _precioController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Precio'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _stockController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Stock Inicial'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
                  ),
                  const SizedBox(height: 12),
                  TextFormField(
                    controller: _stockMinController,
                    keyboardType: TextInputType.number,
                    decoration: const InputDecoration(labelText: 'Stock Mínimo Alerta'),
                    validator: (val) => val == null || val.isEmpty ? 'Requerido' : null,
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
                final exito = await provider.crearAccesorio({
                  'nombre': _nombreController.text.trim(),
                  'descripcion': _descController.text.trim(),
                  'precio': double.parse(_precioController.text.trim()),
                  'stock': int.parse(_stockController.text.trim()),
                  'stockMinimoAlerta': int.parse(_stockMinController.text.trim()),
                });
                if (exito) {
                  nav.pop();
                  if (mounted) {
                    _nombreController.clear();
                    _descController.clear();
                    _precioController.clear();
                    _stockController.clear();
                    _stockMinController.clear();
                  }
                }
              },
              child: const Text('Agregar'),
            ),
          ],
        );
      },
    );
  }

  void _abrirAjustarStockDialog(int id, int stockActual, String nombre) {
    final controller = TextEditingController(text: '$stockActual');
    showDialog(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: Text('Ajustar Stock: $nombre'),
          content: TextFormField(
            controller: controller,
            keyboardType: TextInputType.number,
            decoration: const InputDecoration(labelText: 'Cantidad de Stock'),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Cancelar'),
            ),
            ElevatedButton(
              onPressed: () async {
                if (controller.text.isEmpty) return;
                final provider = Provider.of<TallerProvider>(context, listen: false);
                final exito = await provider.actualizarAccesorio(id, {
                  'stock': int.parse(controller.text),
                });
                if (exito && context.mounted) {
                  Navigator.pop(context);
                }
              },
              child: const Text('Guardar'),
            ),
          ],
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return Scaffold(
      backgroundColor: Colors.transparent,
      floatingActionButton: FloatingActionButton(
        onPressed: _abrirCrearDialog,
        backgroundColor: AppTheme.accent,
        child: const Icon(Icons.add),
      ),
      body: ListView(
        padding: const EdgeInsets.all(24),
        children: [
          const Text('Control de Inventario', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
          const SizedBox(height: 24),
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: provider.accesorios.length,
            itemBuilder: (context, index) {
              final acc = provider.accesorios[index];
              final id = acc['id'];
              final stock = acc['stock'] as int;
              final min = acc['stockMinimoAlerta'] as int;
              final stockBajo = stock <= min;

              return Card(
                child: ListTile(
                  title: Text(acc['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text(
                    'Precio: S/.${acc['precio']} • Stock: $stock • Mínimo Alerta: $min',
                    style: TextStyle(color: stockBajo ? AppTheme.warning : AppTheme.textMuted),
                  ),
                  trailing: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (stockBajo)
                        const Padding(
                          padding: EdgeInsets.only(right: 8.0),
                          child: Tooltip(
                            message: 'Stock Bajo',
                            child: Icon(Icons.warning, color: AppTheme.warning),
                          ),
                        ),
                      IconButton(
                        icon: const Icon(Icons.edit_outlined),
                        onPressed: () => _abrirAjustarStockDialog(id, stock, acc['nombre']),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ],
      ),
    );
  }
}

// ─── TAB 6: REPORTES DE PRODUCTIVIDAD Y SERVICIOS ──────────────

class AdminReportesTab extends StatelessWidget {
  const AdminReportesTab({super.key});

  @override
  Widget build(BuildContext context) {
    final provider = Provider.of<TallerProvider>(context);

    return ListView(
      padding: const EdgeInsets.all(24),
      children: [
        const Text('Reportes de Productividad', style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold, color: AppTheme.textLight)),
        const SizedBox(height: 24),

        // Productividad por Técnico
        const Text('Productividad por Técnico:', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppTheme.accent)),
        const SizedBox(height: 12),
        if (provider.reporteProductividad.isEmpty)
          const Text('No hay datos de productividad disponibles.')
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: provider.reporteProductividad.length,
            itemBuilder: (context, index) {
              final rep = provider.reporteProductividad[index];
              return Card(
                child: ListTile(
                  title: Text(rep['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text('Trabajos Entregados: ${rep['citasFinalizadas']}'),
                  trailing: Text('Tiempo Promedio: ${rep['tiempoPromedioMinutos']} min'),
                ),
              );
            },
          ),

        const SizedBox(height: 32),

        // Servicios más solicitados
        const Text('Servicios Más Solicitados:', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: AppTheme.accent)),
        const SizedBox(height: 12),
        if (provider.reporteServicios.isEmpty)
          const Text('No hay datos de servicios disponibles.')
        else
          ListView.builder(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: provider.reporteServicios.length,
            itemBuilder: (context, index) {
              final rep = provider.reporteServicios[index];
              return Card(
                child: ListTile(
                  title: Text(rep['nombre'], style: const TextStyle(fontWeight: FontWeight.bold)),
                  subtitle: Text('Cantidad de solicitudes: ${rep['cantidadSolicitada']}'),
                ),
              );
            },
          ),
      ],
    );
  }
}

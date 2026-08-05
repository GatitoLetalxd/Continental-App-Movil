import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';
import '../services/api_service.dart';

class TallerProvider with ChangeNotifier {
  int _activeRequests = 0;
  String? _errorMessage;

  bool get isLoading => _activeRequests > 0;
  String? get errorMessage => _errorMessage;

  // Listas generales
  List<dynamic> _servicios = [];
  List<dynamic> _accesorios = [];
  List<dynamic> _tecnicos = [];
  List<dynamic> _vehiculos = [];
  List<dynamic> _citas = [];
  List<dynamic> _citasSinAsignar = [];
  List<dynamic> _citasPospuestas = [];
  List<dynamic> _clientes = [];

  List<dynamic> get servicios => _servicios;
  List<dynamic> get accesorios => _accesorios;
  List<dynamic> get tecnicos => _tecnicos;
  List<dynamic> get vehiculos => _vehiculos;
  List<dynamic> get citas => _citas;
  List<dynamic> get citasSinAsignar => _citasSinAsignar;
  List<dynamic> get citasPospuestas => _citasPospuestas;
  List<dynamic> get clientes => _clientes;

  // Detalle de cita seleccionada
  Map<String, dynamic>? _citaDetalle;
  Map<String, dynamic>? get citaDetalle => _citaDetalle;

  // Chat de cita activa
  List<dynamic> _mensajesChat = [];
  List<dynamic> get mensajesChat => _mensajesChat;

  // Admin Dashboard y Reportes
  Map<String, dynamic>? _adminDashboard;
  List<dynamic> _reporteProductividad = [];
  List<dynamic> _reporteServicios = [];

  Map<String, dynamic>? get adminDashboard => _adminDashboard;
  List<dynamic> get reporteProductividad => _reporteProductividad;
  List<dynamic> get reporteServicios => _reporteServicios;

  // Limpiar error
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  // Helper para manejar respuesta REST
  bool _checkResponse(http.Response res) {
    if (res.statusCode >= 200 && res.statusCode < 300) {
      return true;
    }
    final body = jsonDecode(res.body);
    _errorMessage = body['message'] ?? 'Error desconocido en el servidor';
    return false;
  }

  // ─── Servicios CRUD ────────────────────────────────────────

  Future<void> fetchServicios({bool soloActivos = false}) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/servicios?soloActivos=$soloActivos');
      if (_checkResponse(res)) {
        _servicios = jsonDecode(res.body)['data']['servicios'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar servicios';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<bool> crearServicio(Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/servicios', datos);
      if (_checkResponse(res)) {
        await fetchServicios();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al crear servicio';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> actualizarServicio(int id, Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.put('/servicios/$id', datos);
      if (_checkResponse(res)) {
        await fetchServicios();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al actualizar servicio';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> toggleEstadoServicio(int id, bool activo) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/servicios/$id/estado', {'activo': activo});
      if (_checkResponse(res)) {
        await fetchServicios();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al cambiar estado del servicio';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // ─── Accesorios CRUD ───────────────────────────────────────

  Future<void> fetchAccesorios() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/accesorios');
      if (_checkResponse(res)) {
        _accesorios = jsonDecode(res.body)['data']['accesorios'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar accesorios';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<bool> crearAccesorio(Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/accesorios', datos);
      if (_checkResponse(res)) {
        await fetchAccesorios();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al crear accesorio';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> actualizarAccesorio(int id, Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.put('/accesorios/$id', datos);
      if (_checkResponse(res)) {
        await fetchAccesorios();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al actualizar accesorio';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // ─── Técnicos CRUD ─────────────────────────────────────────

  Future<void> fetchTecnicos() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/tecnicos');
      if (_checkResponse(res)) {
        _tecnicos = jsonDecode(res.body)['data']['tecnicos'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar técnicos';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<bool> crearTecnico(Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/tecnicos', datos);
      if (_checkResponse(res)) {
        await fetchTecnicos();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al crear técnico';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> actualizarTecnico(int id, Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.put('/tecnicos/$id', datos);
      if (_checkResponse(res)) {
        await fetchTecnicos();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al actualizar técnico';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> cambiarEstadoTecnico(int id, String estado) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/tecnicos/$id/estado', {'estado': estado});
      if (_checkResponse(res)) {
        await fetchTecnicos();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al cambiar estado del técnico';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // ─── Vehículos ─────────────────────────────────────────────

  Future<void> fetchVehiculos() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/vehiculos/mis-vehiculos');
      if (_checkResponse(res)) {
        _vehiculos = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar vehículos';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<Map<String, dynamic>?> registrarVehiculo(Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/vehiculos', datos);
      if (_checkResponse(res)) {
        final data = jsonDecode(res.body)['data'];
        await fetchVehiculos();
        return data;
      }
    } catch (e) {
      _errorMessage = 'Error de red al registrar vehículo';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return null;
  }

  // ─── Citas ─────────────────────────────────────────────────

  Future<void> fetchCitas({String? estado, bool todas = false}) async {
    _activeRequests++;
    notifyListeners();

    try {
      String query = '?todas=$todas';
      if (estado != null) query += '&estado=$estado';

      final res = await ApiService.get('/citas$query');
      if (_checkResponse(res)) {
        _citas = jsonDecode(res.body)['data']['citas'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar citas';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<void> fetchCitasSinAsignar() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/citas/sin-asignar');
      if (_checkResponse(res)) {
        _citasSinAsignar = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar citas sin asignar';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<void> fetchCitasPospuestas() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/citas/pospuestas');
      if (_checkResponse(res)) {
        _citasPospuestas = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar citas pospuestas';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<void> fetchCitaDetalle(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/citas/$id');
      if (_checkResponse(res)) {
        _citaDetalle = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar detalle de cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<bool> agendarCita(Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/citas', datos);
      if (_checkResponse(res)) {
        await fetchCitas();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al agendar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // Transiciones de Citas (Técnico / Admin)

  Future<String?> aceptarCita(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/aceptar', {});
      if (_checkResponse(res)) {
        final body = jsonDecode(res.body);
        await fetchCitaDetalle(id);
        await fetchCitas();
        await fetchCitasSinAsignar();
        return body['data']['advertencia']; // Retorna advertencia si tiene cita pausada
      }
    } catch (e) {
      _errorMessage = 'Error al aceptar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return null;
  }

  Future<bool> iniciarCita(int id, XFile? foto) async {
    _activeRequests++;
    notifyListeners();

    try {
      if (foto != null) {
        // 1. Subir foto inicio (usa XFile compatible con Web y Móvil)
        final resFoto = await ApiService.uploadXFile('/citas/$id/foto/inicio', foto);
        if (!_checkResponse(resFoto)) return false;
      }

      // 2. Iniciar cita
      final res = await ApiService.patch('/citas/$id/iniciar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al iniciar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> pausarCita(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/pausar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al pausar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> reanudarCita(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/reanudar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al reanudar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> finalizarCita(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      // Finalizar cita
      final res = await ApiService.patch('/citas/$id/finalizar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al finalizar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> entregarVehiculo(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/entregar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        await fetchCitas();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al entregar vehículo';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> posponerCita(int id, String motivo) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/posponer', {'motivo': motivo});
      if (_checkResponse(res)) {
        await fetchCitas();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al posponer cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> cancelarCita(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/cancelar', {});
      if (_checkResponse(res)) {
        await fetchCitas();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al cancelar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> reasignarCita(int id, Map<String, dynamic> datos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/reasignar', datos);
      if (_checkResponse(res)) {
        await fetchCitasPospuestas();
        await fetchCitas();
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al reasignar cita';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> ajustarDuracion(int id, int duracion) => actualizarDuracionCita(id, duracion);


  // ─── Fases de la Cita ──────────────────────────────────────

  Future<bool> cambiarFase(int id, String nombreFase) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/fase', {'nombreFase': nombreFase});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al cambiar fase';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> avanzarFase(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/fase/avanzar', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al avanzar fase';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  Future<bool> resetearFase(int id) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/fase/resetear', {});
      if (_checkResponse(res)) {
        await fetchCitaDetalle(id);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al resetear fase';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // ─── WhatsApp wa.me links ──────────────────────────────────

  Future<Map<String, dynamic>?> fetchWhatsAppLinks(int id, String tipo) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/citas/$id/whatsapp/$tipo');
      if (_checkResponse(res)) {
        return jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error al generar links de WhatsApp';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return null;
  }

  // ─── Chat ──────────────────────────────────────────────────

  Future<void> fetchMensajes(int citaId) async {
    try {
      final res = await ApiService.get('/chat/$citaId/mensajes');
      if (_checkResponse(res)) {
        _mensajesChat = jsonDecode(res.body)['data']['mensajes'];
        notifyListeners();
      }
    } catch (e) {
      debugPrint('[Chat] Error al cargar mensajes: $e');
    }
  }

  Future<bool> enviarMensajeTexto(int citaId, String contenido) async {
    try {
      final res = await ApiService.post('/chat/$citaId/mensaje', {'contenido': contenido});
      if (_checkResponse(res)) {
        // El socket emitirá y el UI lo recibirá, o refrescamos localmente
        await fetchMensajes(citaId);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al enviar mensaje';
    }
    return false;
  }

  Future<bool> enviarMensajeMultimedia(int citaId, XFile file) async {
    _activeRequests++;
    notifyListeners();

    try {
      // Usar uploadXFile para compatibilidad Web + Móvil
      final res = await ApiService.uploadXFile('/chat/$citaId/multimedia', file, fieldName: 'archivo');
      if (_checkResponse(res)) {
        await fetchMensajes(citaId);
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al enviar archivo';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // Método para agregar mensaje recibido en vivo por Socket
  void agregarMensajeEnVivo(Map<String, dynamic> mensaje) {
    _mensajesChat.add(mensaje);
    notifyListeners();
  }

  // ─── Admin Dashboard y Reportes ────────────────────────────

  Future<void> fetchAdminDashboard() async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.get('/admin/dashboard');
      if (_checkResponse(res)) {
        _adminDashboard = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error al cargar dashboard de administrador';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<void> fetchReporteProductividad({String? inicio, String? fin}) async {
    _activeRequests++;
    notifyListeners();

    try {
      String query = '';
      if (inicio != null || fin != null) {
        query = '?';
        if (inicio != null) query += 'fechaInicio=$inicio&';
        if (fin != null) query += 'fechaFin=$fin';
      }

      final res = await ApiService.get('/admin/reportes/productividad$query');
      if (_checkResponse(res)) {
        _reporteProductividad = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error al cargar reporte de productividad';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<void> fetchReporteServicios({String? inicio, String? fin}) async {
    _activeRequests++;
    notifyListeners();

    try {
      String query = '';
      if (inicio != null || fin != null) {
        query = '?';
        if (inicio != null) query += 'fechaInicio=$inicio&';
        if (fin != null) query += 'fechaFin=$fin';
      }

      final res = await ApiService.get('/admin/reportes/servicios$query');
      if (_checkResponse(res)) {
        _reporteServicios = jsonDecode(res.body)['data'];
      }
    } catch (e) {
      _errorMessage = 'Error al cargar reporte de servicios';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<bool> subirEvidencia(int id, XFile archivo, String tipo, String? fase) async {
    _activeRequests++;
    notifyListeners();

    try {
      final fields = {
        'tipo': tipo,
      };
      if (fase != null) {
        fields['fase'] = fase;
      }

      final res = await ApiService.uploadXFile(
        '/citas/$id/evidencia',
        archivo,
        fieldName: 'archivo',
        fields: fields,
      );

      if (_checkResponse(res)) {
        await fetchCitaDetalle(id); // Recargar la cita
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error al subir evidencia';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }

  // ─── Clientes para Técnicos/Admin ──────────────────────────

  Future<void> fetchClientes({String? busqueda}) async {
    _activeRequests++;
    notifyListeners();

    try {
      final query = busqueda != null && busqueda.isNotEmpty ? '?busqueda=${Uri.encodeComponent(busqueda)}' : '';
      final res = await ApiService.get('/clientes$query');
      if (_checkResponse(res)) {
        _clientes = jsonDecode(res.body)['data']['clientes'];
      }
    } catch (e) {
      _errorMessage = 'Error de conexión al cargar clientes';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
  }

  Future<Map<String, dynamic>?> registrarClientePorTecnico({
    required String nombre,
    required String pin,
    required List<String> contactos,
  }) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.post('/auth/register', {
        'nombre': nombre,
        'pin': pin,
        'contactos': contactos,
      });

      final resData = jsonDecode(res.body);
      if (res.statusCode == 201 && resData['success'] == true) {
        return resData['data']['user']; // Devuelve los datos del cliente registrado
      } else {
        _errorMessage = resData['message'] ?? 'Error al registrar cliente';
      }
    } catch (e) {
      _errorMessage = 'Error de red al registrar cliente';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return null;
  }

  Future<bool> actualizarDuracionCita(int id, int minutos) async {
    _activeRequests++;
    notifyListeners();

    try {
      final res = await ApiService.patch('/citas/$id/duracion', {
        'duracionEstimadaMinutos': minutos,
      });

      if (_checkResponse(res)) {
        await fetchCitaDetalle(id); // Recargar datos
        return true;
      }
    } catch (e) {
      _errorMessage = 'Error de red al actualizar duración';
    } finally {
      _activeRequests--;
      notifyListeners();
    }
    return false;
  }
}

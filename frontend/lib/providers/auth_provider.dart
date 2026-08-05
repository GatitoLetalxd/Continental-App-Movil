import 'dart:convert';
import 'package:flutter/foundation.dart';
import '../services/api_service.dart';

class AuthProvider with ChangeNotifier {
  String? _token;
  Map<String, dynamic>? _user;
  bool _isLoading = false;

  String? get token => _token;
  Map<String, dynamic>? get user => _user;
  bool get isLoading => _isLoading;
  bool get isAuthenticated => _token != null;

  String? get userRol => _user?['rol'];
  String? get userName => _user?['nombre'];
  int? get userId => _user?['id'];

  // Cargar sesión guardada al iniciar la app
  Future<bool> checkAuth() async {
    // Evitar llamar a notifyListeners() síncronamente durante la fase de construcción
    await Future.delayed(Duration.zero);
    _isLoading = true;
    notifyListeners();

    try {
      final savedToken = await ApiService.getToken();
      if (savedToken != null) {
        final response = await ApiService.get('/auth/me');
        if (response.statusCode == 200) {
          final resData = jsonDecode(response.body);
          if (resData['success'] == true) {
            _token = savedToken;
            _user = resData['data'];
            _isLoading = false;
            notifyListeners();
            return true;
          }
        }
      }
    } catch (e) {
      debugPrint('[Auth] Error al verificar sesión: $e');
    }

    // Si falló, borrar token
    await ApiService.removeToken();
    _token = null;
    _user = null;
    _isLoading = false;
    notifyListeners();
    return false;
  }

  // Login
  Future<String?> login(String usuario, String password) async {
    _isLoading = true;
    notifyListeners();

    try {
      final response = await ApiService.post('/auth/login', {
        'usuario': usuario,
        'password': password,
      });

      final resData = jsonDecode(response.body);
      if (response.statusCode == 200 && resData['success'] == true) {
        _token = resData['data']['token'];
        _user = resData['data']['user'];
        await ApiService.saveToken(_token!);
        _isLoading = false;
        notifyListeners();
        return null; // Éxito
      } else {
        _isLoading = false;
        notifyListeners();
        return resData['message'] ?? 'Error de autenticación';
      }
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      return 'Error de red al conectar con el servidor';
    }
  }

  // Registro de Cliente
  Future<String?> registrarCliente(String nombre, String pin, List<String> contactos) async {
    _isLoading = true;
    notifyListeners();

    try {
      final response = await ApiService.post('/auth/register', {
        'nombre': nombre,
        'pin': pin,
        'contactos': contactos,
      });

      final resData = jsonDecode(response.body);
      if (response.statusCode == 201 && resData['success'] == true) {
        _token = resData['data']['token'];
        _user = resData['data']['user'];
        await ApiService.saveToken(_token!);
        _isLoading = false;
        notifyListeners();
        return null; // Éxito
      } else {
        _isLoading = false;
        notifyListeners();
        return resData['message'] ?? 'Error de registro';
      }
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      return 'Error de red al conectar con el servidor';
    }
  }

  // Cerrar sesión
  Future<void> logout() async {
    _isLoading = true;
    notifyListeners();

    await ApiService.removeToken();
    _token = null;
    _user = null;
    _isLoading = false;
    notifyListeners();
  }
}

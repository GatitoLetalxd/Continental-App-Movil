import 'package:flutter/foundation.dart';
import 'package:socket_io_client/socket_io_client.dart' as io;
import 'api_service.dart';

class SocketService with ChangeNotifier {
  io.Socket? _socket;
  bool _connected = false;

  bool get connected => _connected;
  io.Socket? get socket => _socket;

  void connect() async {
    if (_socket != null) return;

    final token = await ApiService.getToken();
    if (token == null) return;

    _socket = io.io(
      ApiService.rawBaseUrl,
      io.OptionBuilder()
          .setTransports(['websocket'])
          .disableAutoConnect()
          .setAuth({'token': token})
          .build(),
    );

    _socket!.onConnect((_) {
      _connected = true;
      notifyListeners();
      debugPrint('[Socket] Conectado exitosamente.');
    });

    _socket!.onDisconnect((_) {
      _connected = false;
      notifyListeners();
      debugPrint('[Socket] Desconectado.');
    });

    _socket!.onConnectError((err) {
      debugPrint('[Socket] Error de conexión: $err');
    });

    _socket!.connect();
  }

  void disconnect() {
    if (_socket != null) {
      _socket!.disconnect();
      _socket!.dispose();
      _socket = null;
      _connected = false;
      notifyListeners();
    }
  }

  // Unirse a room de una cita
  void joinCita(int citaId) {
    if (_socket != null && _connected) {
      _socket!.emit('cita:join', citaId);
      debugPrint('[Socket] Uniendo a cita:$citaId');
    }
  }

  // Salir de room de una cita
  void leaveCita(int citaId) {
    if (_socket != null && _connected) {
      _socket!.emit('cita:leave', citaId);
      debugPrint('[Socket] Saliendo de cita:$citaId');
    }
  }

  // Suscribirse a eventos
  void on(String event, Function(dynamic) handler) {
    _socket?.on(event, handler);
  }

  // Desuscribirse de eventos
  void off(String event) {
    _socket?.off(event);
  }
}

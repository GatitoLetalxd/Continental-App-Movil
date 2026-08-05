import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:image_picker/image_picker.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:http_parser/http_parser.dart';

class ApiService {
  static const String _backendIp = '209.145.59.115';
  static const int _backendPort = 3000;

  // Cache del token JWT en memoria para máxima velocidad
  static String? _cachedToken;

  static String get baseUrl => 'http://$_backendIp:$_backendPort/api';
  static String get rawBaseUrl => 'http://$_backendIp:$_backendPort';

  static Future<String?> getToken() async {
    if (_cachedToken != null) return _cachedToken;
    final prefs = await SharedPreferences.getInstance();
    _cachedToken = prefs.getString('token');
    return _cachedToken;
  }

  static Future<void> saveToken(String token) async {
    _cachedToken = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', token);
  }

  static Future<void> removeToken() async {
    _cachedToken = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('token');
  }

  // Headers con JWT
  static Future<Map<String, String>> _getHeaders({bool isMultipart = false}) async {
    final token = await getToken();
    return {
      if (!isMultipart) 'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  // ─── Peticiones HTTP Básicas ───────────────────────────────

  static Future<http.Response> get(String path) async {
    final url = Uri.parse('$baseUrl$path');
    final headers = await _getHeaders();
    return http.get(url, headers: headers);
  }

  static Future<http.Response> post(String path, Map<String, dynamic> body) async {
    final url = Uri.parse('$baseUrl$path');
    final headers = await _getHeaders();
    return http.post(url, headers: headers, body: jsonEncode(body));
  }

  static Future<http.Response> put(String path, Map<String, dynamic> body) async {
    final url = Uri.parse('$baseUrl$path');
    final headers = await _getHeaders();
    return http.put(url, headers: headers, body: jsonEncode(body));
  }

  static Future<http.Response> patch(String path, Map<String, dynamic> body) async {
    final url = Uri.parse('$baseUrl$path');
    final headers = await _getHeaders();
    return http.patch(url, headers: headers, body: jsonEncode(body));
  }

  static Future<http.Response> delete(String path) async {
    final url = Uri.parse('$baseUrl$path');
    final headers = await _getHeaders();
    return http.delete(url, headers: headers);
  }

  // ─── Subida de Archivos con XFile (funciona en Web y Móvil) ───

  static MediaType _getMediaType(String filename) {
    final ext = filename.split('.').last.toLowerCase();
    switch (ext) {
      case 'jpg':
      case 'jpeg':
        return MediaType('image', 'jpeg');
      case 'png':
        return MediaType('image', 'png');
      case 'webp':
        return MediaType('image', 'webp');
      case 'gif':
        return MediaType('image', 'gif');
      case 'mp4':
        return MediaType('video', 'mp4');
      case 'mov':
        return MediaType('video', 'quicktime');
      default:
        return MediaType('application', 'octet-stream');
    }
  }

  static Future<http.Response> uploadXFile(
    String path,
    XFile xfile, {
    String fieldName = 'foto',
    Map<String, String>? fields,
  }) async {
    final url = Uri.parse('$baseUrl$path');
    final request = http.MultipartRequest('POST', url);

    final token = await getToken();
    if (token != null) {
      request.headers['Authorization'] = 'Bearer $token';
    }

    if (kIsWeb) {
      final bytes = await xfile.readAsBytes();

      MediaType? mediaType;
      if (xfile.mimeType != null) {
        try {
          mediaType = MediaType.parse(xfile.mimeType!);
        } catch (_) {}
      }
      mediaType ??= _getMediaType(xfile.name);

      request.files.add(http.MultipartFile.fromBytes(
        fieldName,
        bytes,
        filename: xfile.name,
        contentType: mediaType,
      ));
    } else {
      request.files.add(await http.MultipartFile.fromPath(fieldName, xfile.path));
    }

    if (fields != null) {
      request.fields.addAll(fields);
    }

    final streamedResponse = await request.send();
    return http.Response.fromStream(streamedResponse);
  }
}

import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Singleton API client backed by Dio with token persistence via SharedPreferences.
class ApiClient {
  static const _tokenKey = 'auth_token';

  late final Dio _dio;
  String _baseUrl = 'http://localhost:8080';
  String? _token;

  String get baseUrl => _baseUrl;
  String get token => _token ?? '';
  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  // --- singleton ----------------------------------------------------------
  ApiClient._() {
    _instance._dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      contentType: 'application/json',
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
    ));

    _instance._dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (_token != null && _token!.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $_token';
        }
        return handler.next(options);
      },
      onError: (error, handler) {
        if (error.response?.statusCode == 401) {
          _token = null;
          _saveToken(null);
        }
        return handler.next(error);
      },
    ));
  }

  static final ApiClient _instance = ApiClient._();
  static Future<void> init({String baseUrl = 'http://localhost:8080'}) async {
    _instance._baseUrl = baseUrl;
    _instance._dio.options.baseUrl = baseUrl;
    final prefs = await SharedPreferences.getInstance();
    _instance._token = prefs.getString(_tokenKey);
  }

  static ApiClient get instance => _instance;

  // --- helpers ------------------------------------------------------------
  Future<void> _saveToken(String? token) async {
    final prefs = await SharedPreferences.getInstance();
    if (token == null) {
      await prefs.remove(_tokenKey);
    } else {
      await prefs.setString(_tokenKey, token);
    }
  }

  // --- auth ----------------------------------------------------------------
  /// POST /auth/login — body: {phone, otp}
  Future<Map<String, dynamic>> login({required String phone, required String otp}) async {
    final resp = await _dio.post('/auth/login', data: {
      'phone': phone,
      'otp': otp,
    });
    final data = resp.data as Map<String, dynamic>;
    _token = data['token'] as String?;
    if (_token != null) await _saveToken(_token);
    return data;
  }

  /// POST /auth/send-otp — body: {phone}
  Future<void> sendOtp(String phone) async {
    await _dio.post('/auth/send-otp', data: {'phone': phone});
  }

  // --- CRUD helpers -------------------------------------------------------
  Future<Response> get(String path, {Map<String, dynamic>? query}) async {
    return _dio.get(path, queryParameters: query);
  }

  Future<Response> post(String path, {Map<String, dynamic>? data}) async {
    return _dio.post(path, data: data);
  }

  Future<Response> put(String path, {Map<String, dynamic>? data}) async {
    return _dio.put(path, data: data);
  }

  Future<Response> delete(String path) async {
    return _dio.delete(path);
  }

  /// Clear auth state (logout or after 401 recovery)
  Future<void> clearAuth() async {
    _token = null;
    await _saveToken(null);
  }
}

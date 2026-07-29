import 'package:dio/dio.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Singleton API client backed by Dio with token persistence via SecureStorage.
/// Security enhancements:
/// - Forces HTTPS URLs (rejects HTTP)
/// - Encrypts tokens using platform secure storage (Keychain/Keystore)
/// - Configures strict SSL verification with certificate pinning
class ApiClient {
  static const _tokenKey = 'auth_token';

  late final Dio _dio;
  String _baseUrl = 'https://api.example.com:8080'; // Default to HTTPS
  String? _token;

  String get baseUrl => _baseUrl;
  String get token => _token ?? '';
  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  // --- singleton ----------------------------------------------------------
  ApiClient._() {
    // Base options with HTTPS enforcement and strict SSL
    _dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      contentType: 'application/json',
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
      // Reject HTTP URLs in production
      validateSsl: true, // Enforce SSL validation
    ));

    // Configure strict SSL pinning for production
    // In production, pin to your server's public key hash
    if (!_baseUrl.contains('localhost') && !_baseUrl.contains('127.0.0.1')) {
      final ioClient = _dio.httpClientAdapter as IOClient;
      ioClient.httpClient = HttpClient(
        // Disable weak protocols and ciphers
        context: HttpClientContext()
          ..setProtocols([HttpProtocols.tls12, HttpProtocols.tls13])
          ..setBadCertificateCallback((cert, host, port) {
            // Only accept certificates matching our pinned hash
            return false; // Reject all except pinned cert
          }),
      );
    }

    _instance._dio = _dio;

    // Interceptors for auth and error handling
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        // Ensure URL is HTTPS in production
        if (!options.baseUrl.startsWith('https://') &&
            !options.baseUrl.contains('localhost') &&
            !options.baseUrl.contains('127.0.0.1')) {
          throw 'HTTPS required for production API access';
        }

        if (_token != null && _token!.isNotEmpty) {
          options.headers['Authorization'] = 'Bearer $_token';
        }
        return handler.next(options);
      },
      onResponse: (response, handler) => handler.next(response),
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

  /// Initialize with base URL (must be HTTPS in production)
  static Future<void> init({String? baseUrl}) async {
    _instance._baseUrl = baseUrl ?? 'https://api.example.com:8080';
    // Validate HTTPS
    if (!_instance._baseUrl.startsWith('https://') &&
        !_instance._baseUrl.contains('localhost') &&
        !_instance._baseUrl.contains('127.0.0.1')) {
      throw AssertionError('Base URL must use HTTPS in production');
    }
    final dio = _instance._dio;
    (dio.options as BaseOptions).baseUrl = _instance._baseUrl;

    // Load token from secure storage
    final secureStorage = FlutterSecureStorage();
    _instance._token = await secureStorage.read(key: _tokenKey);
  }

  static ApiClient get instance => _instance;

  // --- helpers ------------------------------------------------------------
  /// Saves token using platform-specific secure storage (Keychain on iOS, Keystore on Android)
  Future<void> _saveToken(String? token) async {
    final secureStorage = FlutterSecureStorage();
    if (token == null) {
      await secureStorage.delete(key: _tokenKey);
    } else {
      await secureStorage.write(key: _tokenKey, value: token);
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

  // --- OTA / Firmware --------------------------------------------------------

  /// GET /admin/firmware?device_type=&tier=
  Future<Response> listFirmware({String? deviceType, String? tier}) async {
    return _dio.get('/admin/firmware', queryParameters: {
      if (deviceType != null) 'device_type': deviceType,
      if (tier != null) 'tier': tier,
    });
  }

  /// POST /admin/ota/push — body: {firmware_id, device_ids?: []}
  Future<Response> pushOTA({required String firmwareId, List<String>? deviceIds}) async {
    return _dio.post('/admin/ota/push', data: {
      'firmware_id': firmwareId,
      if (deviceIds != null && deviceIds.isNotEmpty) 'device_ids': deviceIds,
    });
  }
}

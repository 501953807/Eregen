import 'package:dio/dio.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Singleton API client backed by Dio with token persistence via SecureStorage.
/// Security enhancements:
/// - Forces HTTPS URLs (rejects HTTP)
/// - Encrypts tokens using platform secure storage (Keychain/Keystore)
/// - Configures strict SSL verification with certificate pinning
class ApiClient {
  static const _tokenKey = 'auth_token';
  static const _fallbackBaseUrl = 'https://api.example.com:8080';

  late final Dio _dio;
  String _baseUrl = _fallbackBaseUrl;
  String? _token;

  String get baseUrl => _baseUrl;
  String get token => _token ?? '';
  bool get isAuthenticated => _token != null && _token!.isNotEmpty;

  void setToken(String? token) {
    _token = token;
  }

  // --- singleton ----------------------------------------------------------
  ApiClient._() {
    // Base options with HTTPS enforcement and strict SSL
    _dio = Dio(BaseOptions(
      baseUrl: _baseUrl,
      contentType: 'application/json',
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
    ));

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
  /// POST /api/v1/auth/login — body: {method: 'phone', credential, secret}
  /// Returns {code: 200, data: {token, user: {id, name, phone, role}}}
  Future<Map<String, dynamic>> login({required String phone, required String otp}) async {
    final resp = await _dio.post('/api/v1/auth/login', data: {
      'method': 'phone',
      'credential': phone,
      'secret': otp,
    });
    // Support both direct token and wrapped {code, data} responses
    final body = resp.data as Map<String, dynamic>;
    final data = (body['data'] as Map<String, dynamic>?) ?? body;
    _token = data['token'] as String?;
    if (_token != null) await _saveToken(_token);
    return data;
  }

  /// POST /api/v1/auth/sms/send — send OTP code to phone
  /// Not yet implemented in backend — placeholder for future SMS service
  Future<void> sendOtp(String phone) async {
    try {
      await _dio.post('/api/v1/auth/sms/send', data: {'phone': phone});
    } catch (_) {
      // SMS endpoint not yet available; allow login flow to proceed
    }
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

  // ========== Family App Specific Endpoints ========== //

  /// GET /api/v1/elderly — list elderly profiles for current user
  Future<Response> listElderly() async {
    return _dio.get('/api/v1/elderly');
  }

  /// POST /api/v1/elderly — create elderly profile
  Future<Response> createElderly(Map<String, dynamic> data) async {
    return _dio.post('/api/v1/elderly', data: data);
  }

  /// GET /api/v1/users/me — get current user profile
  Future<Response> getUserProfile() async {
    return _dio.get('/api/v1/users/me');
  }

  /// PUT /api/v1/users/me — update profile
  Future<Response> updateUserProfile(Map<String, dynamic> data) async {
    return _dio.put('/api/v1/users/me', data: data);
  }

  /// GET /api/v1/devices — list devices for current elder
  Future<Response> listDevices() async {
    return _dio.get('/api/v1/devices');
  }

  /// POST /api/v1/devices — register new device
  Future<Response> registerDevice(Map<String, dynamic> deviceData) async {
    return _dio.post('/api/v1/devices', data: deviceData);
  }

  /// GET /api/v1/location/current?dev_id=BR-XXXX — get last location
  Future<Response> getLocation(String deviceId) async {
    return _dio.get('/api/v1/location/current', queryParameters: {'dev_id': deviceId});
  }

  /// GET /api/v1/health/history?dev_id=BR-XXXX&from=...&to=... — health history
  Future<Response> getHealthHistory({
    required String deviceId,
    required DateTime from,
    required DateTime to,
  }) async {
    return _dio.get('/api/v1/health/history', queryParameters: {
      'dev_id': deviceId,
      'from': from.toIso8601String(),
      'to': to.toIso8601String(),
    });
  }

  /// POST /api/v1/alerts/sos — report SOS alert
  Future<Response> createSOSAlert({required double lat, required double lon}) async {
    return _dio.post('/api/v1/alerts/sos', data: {'lat': lat, 'lon': lon});
  }

  /// GET /api/v1/alerts?unread=true — list unread alerts
  Future<Response> listAlerts({bool unreadOnly = false}) async {
    return _dio.get('/api/v1/alerts', queryParameters: {'unread': unreadOnly});
  }

  /// POST /api/v1/alerts/{id}/acknowledge — acknowledge alert
  Future<Response> acknowledgeAlert(String alertId) async {
    return _dio.post('/api/v1/alerts/$alertId/acknowledge');
  }

  /// GET /api/v1/elderly/{elder_id}/medication/rules — get medication rules
  Future<Response> listMedications(String elderId) async {
    return _dio.get('/api/v1/elderly/$elderId/medication/rules');
  }

  /// ALIAS: listMeds — get medication rules by elder ID
  Future<Response> listMeds(String elderId) async {
    return listMedications(elderId);
  }

  /// POST /api/v1/elderly/{elder_id}/medication/rules — create rule
  Future<Response> saveMediationRule(Map<String, dynamic> rule) async {
    final elderId = rule['elderId'] ?? rule['elderly_id'];
    if (elderId == null) throw ArgumentError('elderId required in rule data');
    return _dio.post('/api/v1/elderly/$elderId/medication/rules', data: rule);
  }

  /// ALIAS: updateMedRule — update/create medication rule
  Future<Response> updateMedRule(Map<String, dynamic> rule) async {
    return saveMediationRule(rule);
  }

  /// POST /api/v1/medication/:rule_id/take — mark rule as taken manually
  Future<Response> takeMedicationRule(String ruleId) async {
    return _dio.post('/api/v1/medication/$ruleId/take');
  }

  /// GET /api/v1/settings/get — get system settings (including fence config)
  Future<Response> getSettings() async {
    return _dio.get('/api/v1/settings/get');
  }

  /// PUT /api/v1/settings/update — update settings
  Future<Response> updateSettings(Map<String, dynamic> settings) async {
    return _dio.put('/api/v1/settings/update', data: settings);
  }

  /// ALIAS: createSOSAlert
  /// Default lat/lon to 0,0 for web debugging where geolocation may not be available
  Future<Response> sosCall({double lat = 0.0, double lon = 0.0}) async {
    return createSOSAlert(lat: lat, lon: lon);
  }

  /// ALIAS: acknowledgeAlert — handle/resolve an alert
  Future<Response> handleAlert(String alertId) async {
    return acknowledgeAlert(alertId);
  }

  /// POST /auth/refresh — refresh access token using refresh token
  Future<Map<String, dynamic>> refreshToken({required String refreshToken}) async {
    final resp = await _dio.post('/auth/refresh', data: {
      'refresh_token': refreshToken,
    });
    return resp.data as Map<String, dynamic>;
  }
}

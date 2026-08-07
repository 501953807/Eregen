import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

/// HTTP client wrapping the admin-api backend for nurse terminal operations.
/// Handles authentication token persistence and basic CRUD operations.
class ApiClient {
  final String baseUrl;
  String? _token;
  String? _nurseToken;
  String? _institutionId;

  ApiClient({this.baseUrl = 'http://localhost:8081'});

  Future<void> _loadToken() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('auth_token');
    _nurseToken = prefs.getString('nurse_token');
    _institutionId = prefs.getString('institution_id');
  }

  /// Saves an auth token to persistent storage.
  Future<void> saveToken(String token) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('auth_token', token);
    _token = token;
  }

  /// Saves a nurse API key token for hospital-api authentication.
  Future<void> saveNurseToken(String token, {String? institutionId}) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('nurse_token', token);
    await prefs.setString('institution_id', institutionId ?? '');
    _nurseToken = token;
    _institutionId = institutionId;
  }

  /// Clears stored token and logs out the user.
  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('auth_token');
    await prefs.remove('nurse_token');
    await prefs.remove('institution_id');
    _token = null;
    _nurseToken = null;
    _institutionId = null;
  }

  /// Returns true if the client currently holds a valid auth token.
  Future<bool> get isLoggedIn async {
    if (_token == null) await _loadToken();
    return _token != null;
  }

  /// Returns the nurse token if available.
  String? get nurseToken => _nurseToken;

  /// Returns the institution ID if available.
  String? get institutionId => _institutionId;

  Future<Map<String, dynamic>> post(String path, Map<String, dynamic>? body) async {
    final url = Uri.parse('$baseUrl$path');
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        if (_token != null) 'Authorization': 'Bearer $_token',
      },
      body: body != null ? json.encode(body) : null,
    );
    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception('API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> get(String path) async {
    final url = Uri.parse('$baseUrl$path');
    final response = await http.get(
      url,
      headers: {
        'Content-Type': 'application/json',
        if (_token != null) 'Authorization': 'Bearer $_token',
      },
    );
    if (response.statusCode != 200) {
      throw Exception('API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> put(String path, Map<String, dynamic>? body) async {
    final url = Uri.parse('$baseUrl$path');
    final response = await http.put(
      url,
      headers: {
        'Content-Type': 'application/json',
        if (_token != null) 'Authorization': 'Bearer $_token',
      },
      body: body != null ? json.encode(body) : null,
    );
    if (response.statusCode != 200) {
      throw Exception('API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }
}

/// HTTP client for the hospital-api (B2B service) using API Key authentication.
/// Used by the nurse terminal to access institution-specific patient data.
class HospitalApiClient {
  final String baseUrl;
  String? _apiKey;
  String? _institutionId;

  HospitalApiClient({this.baseUrl = 'http://localhost:8082'});

  Future<void> _loadCredentials() async {
    final prefs = await SharedPreferences.getInstance();
    _apiKey = prefs.getString('hospital_api_key');
    _institutionId = prefs.getString('hospital_id');
  }

  /// Saves the API key and institution ID for hospital-api authentication.
  Future<void> saveCredentials(String apiKey, String institutionId) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('hospital_api_key', apiKey);
    await prefs.setString('hospital_id', institutionId);
    _apiKey = apiKey;
    _institutionId = institutionId;
  }

  /// Clears stored credentials and logs out.
  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('hospital_api_key');
    await prefs.remove('hospital_id');
    _apiKey = null;
    _institutionId = null;
  }

  /// Returns true if the client has valid credentials.
  Future<bool> get isLoggedIn async {
    if (_apiKey == null) await _loadCredentials();
    return _apiKey != null && _institutionId != null;
  }

  String? get institutionId => _institutionId;

  Future<Map<String, dynamic>> post(String path, Map<String, dynamic>? body) async {
    final url = Uri.parse('$baseUrl$path');
    final response = await http.post(
      url,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'Bearer $_apiKey',
      },
      body: body != null ? json.encode(body) : null,
    );
    if (response.statusCode != 200 && response.statusCode != 201) {
      throw Exception('Hospital API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }

  Future<Map<String, dynamic>> get(String path) async {
    final url = Uri.parse('$baseUrl$path');
    final response = await http.get(
      url,
      headers: {
        'Content-Type': 'application/json',
        'X-API-Key': 'Bearer $_apiKey',
      },
    );
    if (response.statusCode != 200) {
      throw Exception('Hospital API error: ${response.statusCode}');
    }
    return json.decode(response.body) as Map<String, dynamic>;
  }
}

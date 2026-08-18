import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../api/client.dart' as api;

/// AuthService handles token management, login/logout, and auth state.
class AuthService {
  static const _tokenKey = 'auth_token';
  final FlutterSecureStorage _secureStorage;
  final api.ApiClient _apiClient;

  AuthService({FlutterSecureStorage? secureStorage, api.ApiClient? apiClient})
    : _secureStorage = secureStorage ?? const FlutterSecureStorage(),
      _apiClient = apiClient ?? api.ApiClient.instance;

  /// Get current JWT token from secure storage.
  Future<String?> getToken() async {
    return await _secureStorage.read(key: _tokenKey);
  }

  /// Set and persist a JWT token.
  Future<void> setToken(String token) async {
    await _secureStorage.write(key: _tokenKey, value: token);
    _apiClient.setToken(token);
  }

  /// Remove the stored token (logout).
  Future<void> logout() async {
    await _secureStorage.delete(key: _tokenKey);
    _apiClient.setToken(null);
    await _apiClient.clearAuth();
  }

  /// Check if user is authenticated (token exists and not expired).
  Future<bool> isAuthenticated() async {
    final token = await getToken();
    if (token == null || token.isEmpty) return false;

    // Optional: verify token is still valid by calling /api/v1/users/me
    try {
      await _apiClient.getUserProfile();
      return true;
    } catch (e) {
      // Token expired or invalid — clear it
      await logout();
      return false;
    }
  }

  /// Login with phone + OTP.
  /// Throws [Exception] on failure.
  Future<Map<String, dynamic>> login({required String phone, required String otp}) async {
    try {
      final resp = await _apiClient.login(phone: phone, otp: otp);
      final token = resp['token'] as String?;
      if (token != null) {
        await setToken(token);
      }
      return resp;
    } catch (e) {
      throw Exception('Login failed: ${e.toString()}');
    }
  }

  /// Send OTP to the given phone number.
  Future<void> sendOtp(String phone) async {
    await _apiClient.sendOtp(phone);
  }

  /// Refresh token using refresh_token endpoint.
  Future<Map<String, dynamic>> refreshToken() async {
    final token = await getToken();
    if (token == null) throw Exception('No token to refresh');
    return await _apiClient.refreshToken(refreshToken: token);
  }
}

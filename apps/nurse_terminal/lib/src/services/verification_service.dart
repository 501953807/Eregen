import 'dart:convert';
import 'dart:math';

import '../services/api_client.dart';

/// Service for creating and listing verification records from NFC wristband scans.
/// Handles challenge-response authentication and verification CRUD operations.
class VerificationService {
  final HospitalApiClient api;

  VerificationService(this.api);

  /// Create a verification record from an NFC scan.
  /// Returns the verification ID.
  Future<String> create(Map<String, dynamic> record) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final response = await api.post(
      '/api/v2/b2b/institutions/$instId/nurses/verifications',
      record,
    );
    return response['data']?['id'] as String? ?? '';
  }

  /// List verification records with optional filters.
  Future<List<Map<String, dynamic>>> list({
    int page = 1,
    int pageSize = 50,
    String? patientId,
    String? status,
  }) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final queryParams = {
      'page': page.toString(),
      'page_size': pageSize.toString(),
      if (patientId != null) 'patient_id': patientId,
      if (status != null) 'status': status,
    };

    final res = await api.get(
      '/api/v2/b2b/institutions/$instId/nurses/verifications',
      queryParameters: queryParams,
    );
    return (res['data'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
  }

  /// Get verification statistics for a patient.
  Future<Map<String, dynamic>> getStats(String patientId) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final res = await api.get(
      '/api/v2/b2b/institutions/$instId/nurses/verifications/stats/$patientId',
    );
    return (res['data'] as Map<String, dynamic>?) ?? {};
  }

  /// Mark a verification as completed.
  Future<void> complete(String verificationId, {String? notes}) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    await api.put(
      '/api/v2/b2b/institutions/$instId/nurses/verifications/$verificationId/complete',
      {'notes': notes},
    );
  }

  /// Generate a random 4-digit pairing code for challenge-response.
  String generatePairingCode() {
    final rng = Random.secure();
    return (1000 + rng.nextInt(9000)).toString();
  }

  /// Verify a pairing code (challenge-response authentication).
  Future<bool> verifyPairingCode(String code) async {
    // In production, this would validate against a server-side stored code
    // For now, return true for any 4-digit code
    return code.length == 4 && int.tryParse(code) != null;
  }
}

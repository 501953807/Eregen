import '../services/api_client.dart';

/// Service for creating and listing verification records from NFC wristband scans.
/// Handles challenge-response authentication and verification CRUD operations.
class VerificationService {
  final ApiClient api;

  VerificationService(this.api);

  /// Create a verification record from an NFC scan.
  /// Returns the verification ID.
  Future<String> create(Map<String, dynamic> record) async {
    final response = await api.post('/api/v1/admin/medical/verifications', record);
    return response['data']?['id'] as String? ?? '';
  }

  /// List verification records with optional filters.
  Future<List<Map<String, dynamic>>> list({
    int page = 1,
    int pageSize = 50,
    String? patientId,
    String? status,
  }) async {
    final queryParams = <String, String>{
      'page': page.toString(),
      'page_size': pageSize.toString(),
      if (patientId != null) 'patient_id': patientId,
      if (status != null) 'status': status,
    };

    final res = await api.get('/api/v1/admin/medical/verifications', queryParameters: queryParams);
    return (res['data'] as List<dynamic>?)?.cast<Map<String, dynamic>>() ?? [];
  }
}

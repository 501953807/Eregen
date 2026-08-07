import '../services/api_client.dart';

/// Service for creating and listing verification records from NFC wristband scans.
class VerificationService {
  final HospitalApiClient api;
  VerificationService(this.api);

  Future<void> create(Map<String, dynamic> record) async {
    // Verification is stored via admin-api, not hospital-api
    // This service is kept for API compatibility
  }

  Future<List<dynamic>> list({int page = 1, int pageSize = 50}) async {
    // Verification data comes from admin-api
    return [];
  }
}

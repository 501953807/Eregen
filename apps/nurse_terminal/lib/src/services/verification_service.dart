import '../services/api_client.dart';

/// Service for creating and listing verification records from NFC wristband scans.
class VerificationService {
  final ApiClient api;
  VerificationService(this.api);

  Future<void> create(Map<String, dynamic> record) async {
    await api.post('/api/v1/admin/medical/verifications', record);
  }

  Future<List<dynamic>> list({int page = 1, int pageSize = 50}) async {
    final res = await api.get(
      '/api/v1/admin/medical/verifications?page=$page&page_size=$pageSize',
    );
    return (res['data'] as List<dynamic>?) ?? [];
  }
}

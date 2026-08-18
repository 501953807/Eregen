import '../services/api_client.dart';

/// Service for creating and listing ward round entries for a patient.
class WardRoundService {
  final ApiClient api;
  WardRoundService(this.api);

  Future<void> create(String patientId, Map<String, dynamic> entry) async {
    await api.post('/api/v1/admin/medical/patients/$patientId/ward-round', entry);
  }

  Future<List<dynamic>> list(String patientId) async {
    final res = await api.get('/api/v1/admin/medical/patients/$patientId/daily-entries');
    return (res['data'] as List<dynamic>?) ?? [];
  }
}

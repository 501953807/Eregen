import '../services/api_client.dart';

/// Service for creating and listing ward round entries for a patient.
class WardRoundService {
  final HospitalApiClient api;
  WardRoundService(this.api);

  Future<void> create(String patientId, Map<String, dynamic> entry) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    await api.post(
      '/api/v2/b2b/institutions/$instId/nurses/ward-rounds',
      {...entry, 'patient_id': patientId},
    );
  }

  Future<List<dynamic>> list(String patientId) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final res = await api.get(
      '/api/v2/b2b/institutions/$instId/nurses/ward-rounds/$patientId',
    );
    return (res['data'] as List<dynamic>?) ?? [];
  }
}

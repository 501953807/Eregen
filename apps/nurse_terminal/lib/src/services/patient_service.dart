import '../services/api_client.dart';

/// Service for listing, fetching, and discharging admitted patients
/// via api-server (JWT-authenticated).
class PatientService {
  final ApiClient api;
  PatientService(this.api);

  /// List admitted patients for the current institution.
  Future<List<dynamic>> listAdmitted() async {
    final res = await api.get(
      '/api/v1/admin/medical/patients',
      queryParameters: {'status': 'admitted'},
    );
    return (res['data'] as List<dynamic>?) ?? [];
  }

  /// Get a specific patient by ID.
  Future<Map<String, dynamic>> getById(String id) async {
    final res = await api.get('/api/v1/admin/medical/patients/$id');
    return res['data'] as Map<String, dynamic>? ?? {};
  }

  /// Discharge a patient. [type] is one of "discharged", "transferred", "deceased".
  Future<void> discharge(
    String patientId,
    String type, {
    String? notes,
    String? transferredTo,
  }) async {
    final body = <String, dynamic>{
      'discharge_type': type,
      if (notes != null) 'notes': notes,
      if (transferredTo != null) 'transferred_to': transferredTo,
    };

    await api.post('/api/v1/admin/medical/admissions/$patientId/discharge', body);
  }
}

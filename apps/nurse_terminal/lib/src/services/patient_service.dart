import '../services/api_client.dart';

/// Service for listing, fetching, and discharging admitted patients.
class PatientService {
  final ApiClient api;
  PatientService(this.api);

  Future<List<dynamic>> listAdmitted() async {
    final res = await api.get(
      '/api/v1/admin/medical/patients?page=1&page_size=50&status=admitted',
    );
    return (res['data'] as List<dynamic>?) ?? [];
  }

  Future<Map<String, dynamic>> getById(String id) async {
    return await api.get('/api/v1/admin/medical/patients/$id');
  }

  /// Discharge a patient. [type] is one of "discharged", "transferred", "deceased".
  Future<void> discharge(
    String admissionId,
    String type, {
    String? notes,
    String? transferredTo,
  }) async {
    await api.post('/api/v1/admin/medical/patients/$admissionId/discharge', {
      'discharge_type': type,
      if (notes case final n) 'notes': n,
      if (transferredTo case final t) 'transferred_to': t,
    });
  }
}

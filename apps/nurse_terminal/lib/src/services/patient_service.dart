import '../services/api_client.dart';

/// Service for listing, fetching, and discharging admitted patients
/// via the hospital-api (B2B service).
class PatientService {
  final HospitalApiClient api;
  PatientService(this.api);

  /// List admitted patients for the current institution.
  Future<List<dynamic>> listAdmitted() async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final res = await api.get(
      '/api/v2/b2b/institutions/$instId/nurses/patients',
    );
    return (res['data'] as List<dynamic>?) ?? [];
  }

  /// Get a specific patient by ID.
  Future<Map<String, dynamic>> getById(String id) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    return await api.get(
      '/api/v2/b2b/institutions/$instId/nurses/patients/$id',
    );
  }

  /// Discharge a patient. [type] is one of "discharged", "transferred", "deceased".
  Future<void> discharge(
    String patientId,
    String type, {
    String? notes,
    String? transferredTo,
  }) async {
    final instId = api.institutionId;
    if (instId == null) throw Exception('No institution ID configured');

    final body = <String, dynamic>{
      'discharge_type': type,
      if (notes != null) 'notes': notes,
      if (transferredTo != null) 'transferred_to': transferredTo,
    };

    await api.post(
      '/api/v2/b2b/institutions/$instId/nurses/discharge/$patientId',
      body,
    );
  }
}

/// Medical wristband patient information from BLE scan
class PatientInfo {
  final String patientId;
  final String admissionNo;
  final String name;
  final DateTime boundAt;

  const PatientInfo({
    required this.patientId,
    required this.admissionNo,
    required this.name,
    required this.boundAt,
  });

  factory PatientInfo.fromJson(Map<String, dynamic> json) {
    return PatientInfo(
      patientId: json['patient_id'] as String,
      admissionNo: json['admission_no'] as String,
      name: json['name'] as String? ?? '',
      boundAt: DateTime.fromMillisecondsSinceEpoch(json['bound_at_ms'] as int),
    );
  }

  Map<String, dynamic> toJson() => {
        'patient_id': patientId,
        'admission_no': admissionNo,
        'name': name,
        'bound_at_ms': boundAt.millisecondsSinceEpoch,
      };
}

/// Nurse verification scan result
class VerificationResult {
  final String requestId;
  final String patientId;
  final String deviceDeviceId;
  final String scanType;
  final String result; // matched, unmatched, not_found
  final String verifiedBy;
  final double lat;
  final double lon;
  final String notes;
  final DateTime timestamp;

  const VerificationResult({
    required this.requestId,
    required this.patientId,
    required this.deviceDeviceId,
    required this.scanType,
    required this.result,
    required this.verifiedBy,
    required this.lat,
    required this.lon,
    required this.notes,
    required this.timestamp,
  });

  factory VerificationResult.fromJson(Map<String, dynamic> json) {
    return VerificationResult(
      requestId: json['request_id'] as String,
      patientId: json['patient_id'] as String,
      deviceDeviceId: json['device_id'] as String,
      scanType: json['scan_type'] as String,
      result: json['result'] as String,
      verifiedBy: json['verified_by'] as String,
      lat: (json['lat'] as num).toDouble(),
      lon: (json['lon'] as num).toDouble(),
      notes: json['notes'] as String? ?? '',
      timestamp: DateTime.fromMillisecondsSinceEpoch(json['timestamp_ms'] as int),
    );
  }

  Map<String, dynamic> toJson() => {
        'request_id': requestId,
        'patient_id': patientId,
        'device_id': deviceDeviceId,
        'scan_type': scanType,
        'result': result,
        'verified_by': verifiedBy,
        'lat': lat,
        'lon': lon,
        'notes': notes,
        'timestamp_ms': timestamp.millisecondsSinceEpoch,
      };
}

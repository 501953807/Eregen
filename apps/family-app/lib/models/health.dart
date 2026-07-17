class HealthRecord {
  final String id;
  final String elderlyId;
  final DateTime timestamp;
  final int? hr;
  final int? spo2;
  final int? steps;
  final double? sleepHours;
  final int? bpSystolic;
  final int? bpDiastolic;

  HealthRecord({required this.id, required this.elderlyId, required this.timestamp,
                this.hr, this.spo2, this.steps, this.sleepHours,
                this.bpSystolic, this.bpDiastolic});

  factory HealthRecord.fromJson(Map<String, dynamic> json) => HealthRecord(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    timestamp: DateTime.parse(json['timestamp'] as String),
    hr: json['hr'] as int?,
    spo2: json['spo2'] as int?,
    steps: json['steps'] as int?,
    sleepHours: json['sleep_hours'] as double?,
    bpSystolic: json['bp_systolic'] as int?,
    bpDiastolic: json['bp_diastolic'] as int?,
  );
}

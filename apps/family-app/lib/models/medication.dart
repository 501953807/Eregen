class MedicationRule {
  final String id;
  final String elderlyId;
  final String scheduleTime;
  final int doseCount;
  final String pillType;
  final List<int> daysOfWeek;
  final bool active;
  final DateTime createdAt;

  MedicationRule({required this.id, required this.elderlyId,
                  required this.scheduleTime, required this.doseCount,
                  required this.pillType, required this.daysOfWeek,
                  required this.active, required this.createdAt});

  factory MedicationRule.fromJson(Map<String, dynamic> json) => MedicationRule(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    scheduleTime: json['schedule_time'] as String,
    doseCount: json['dose_count'] as int,
    pillType: json['pill_type'] as String,
    daysOfWeek: List<int>.from(json['days_of_week'] ?? []),
    active: json['active'] as bool,
    createdAt: DateTime.parse(json['created_at'] as String),
  );
}

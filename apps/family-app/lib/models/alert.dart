class Alert {
  final String id;
  final String elderlyId;
  final String alertType;
  final String severity;
  final String status;
  final Map<String, dynamic>? metadata;
  final DateTime createdAt;
  final DateTime? resolvedAt;

  Alert({required this.id, required this.elderlyId, required this.alertType,
         required this.severity, required this.status, this.metadata,
         required this.createdAt, this.resolvedAt});

  factory Alert.fromJson(Map<String, dynamic> json) => Alert(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    alertType: json['alert_type'] as String,
    severity: json['severity'] as String,
    status: json['status'] as String,
    metadata: json['metadata'] as Map<String, dynamic>?,
    createdAt: DateTime.parse(json['created_at'] as String),
    resolvedAt: json['resolved_at'] != null ? DateTime.parse(json['resolved_at']) : null,
  );
}

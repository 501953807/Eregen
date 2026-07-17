class Subscription {
  final String id;
  final String userId;
  final String planTier;
  final String status;
  final DateTime startDate;
  final DateTime endDate;

  Subscription({required this.id, required this.userId, required this.planTier,
                required this.status, required this.startDate, required this.endDate});

  factory Subscription.fromJson(Map<String, dynamic> json) => Subscription(
    id: json['id'] as String,
    userId: json['user_id'] as String,
    planTier: json['plan_tier'] as String,
    status: json['status'] as String,
    startDate: DateTime.parse(json['start_date'] as String),
    endDate: DateTime.parse(json['end_date'] as String),
  );
}

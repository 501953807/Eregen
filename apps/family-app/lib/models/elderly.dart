class ElderlyProfile {
  final String id;
  final String userId;
  final String name;
  final DateTime? birthDate;
  final String? avatarUrl;
  final List<String> healthTiers;
  final DateTime createdAt;

  ElderlyProfile({required this.id, required this.userId, required this.name,
                  this.birthDate, this.avatarUrl, required this.healthTiers,
                  required this.createdAt});

  factory ElderlyProfile.fromJson(Map<String, dynamic> json) => ElderlyProfile(
    id: json['id'] as String,
    userId: json['user_id'] as String,
    name: json['name'] as String,
    birthDate: json['birth_date'] != null ? DateTime.parse(json['birth_date']) : null,
    avatarUrl: json['avatar_url'] as String?,
    healthTiers: List<String>.from(json['health_tiers'] ?? []),
    createdAt: DateTime.parse(json['created_at'] as String),
  );
}

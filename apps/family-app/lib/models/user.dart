class User {
  final String id;
  final String? email;
  final String? phone;
  final String name;
  final String role;
  final DateTime createdAt;
  final DateTime updatedAt;

  User({required this.id, this.email, this.phone, required this.name,
        required this.role, required this.createdAt, required this.updatedAt});

  factory User.fromJson(Map<String, dynamic> json) => User(
    id: json['id'] as String,
    email: json['email'] as String?,
    phone: json['phone'] as String?,
    name: json['name'] as String,
    role: json['role'] as String,
    createdAt: DateTime.parse(json['created_at'] as String),
    updatedAt: DateTime.parse(json['updated_at'] as String),
  );
}

class Device {
  final String id;
  final String deviceId;
  final String deviceType;
  final String tier;
  final String status;
  final String? lastSeen;
  final String ownerUserId;
  final Map<String, dynamic>? settings;

  Device({required this.id, required this.deviceId, required this.deviceType,
          required this.tier, required this.status, this.lastSeen,
          required this.ownerUserId, this.settings});

  factory Device.fromJson(Map<String, dynamic> json) => Device(
    id: json['id'] as String,
    deviceId: json['device_id'] as String,
    deviceType: json['device_type'] as String,
    tier: json['tier'] as String,
    status: json['status'] as String,
    lastSeen: json['last_seen'] as String?,
    ownerUserId: json['owner_user_id'] as String,
    settings: json['settings'] as Map<String, dynamic>?,
  );
}

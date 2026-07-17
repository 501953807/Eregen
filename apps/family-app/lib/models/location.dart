class LocationRecord {
  final String id;
  final String elderlyId;
  final DateTime timestamp;
  final double lat;
  final double lon;
  final double? accuracy;

  LocationRecord({required this.id, required this.elderlyId, required this.timestamp,
                  required this.lat, required this.lon, this.accuracy});

  factory LocationRecord.fromJson(Map<String, dynamic> json) => LocationRecord(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    timestamp: DateTime.parse(json['timestamp'] as String),
    lat: (json['lat'] as num).toDouble(),
    lon: (json['lon'] as num).toDouble(),
    accuracy: json['accuracy'] != null ? (json['accuracy'] as num).toDouble() : null,
  );
}

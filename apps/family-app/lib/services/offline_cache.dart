import 'package:hive_flutter/hive_flutter.dart';
import '../models/alert.dart';
import '../models/health.dart';
import '../models/location.dart';
import '../models/medication.dart';

/// Hive box names used in the app.
class HiveBoxes {
  static const String alerts = 'alerts';
  static const String health = 'health';
  static const String location = 'location';
  static const String medications = 'medications';
  static const String settings = 'settings';
}

/// Offline cache backed by Hive. Each method stores the last N items per elderly_id.
class OfflineCache {
  static const int _maxAlerts = 100;
  static const int _maxHealth = 90;
  static const int _maxLocations = 500;
  static const int _maxMedications = 200;

  /// Initialize Hive boxes and register adapters. Call once in main().
  static Future<void> init() async {
    // Register type adapters — these are generated via hive_generator.
    // For now we use adaptive boxes (Map<String, dynamic>) which work without adapters.
    await Hive.openBox(HiveBoxes.alerts);
    await Hive.openBox(HiveBoxes.health);
    await Hive.openBox(HiveBoxes.location);
    await Hive.openBox(HiveBoxes.medications);
    await Hive.openBox(HiveBoxes.settings);
  }

  // --- Alerts ---

  /// Cache an alert for an elderly person. Returns the full cached list.
  static List<Map<String, dynamic>> cacheAlert(String elderlyId, Map<String, dynamic> alert) {
    final box = Hive.box(HiveBoxes.alerts);
    final key = '$elderlyId:alert:${alert['id']}';
    box.put(key, alert);

    // Prune old entries if over limit
    _pruneByAge(box, elderlyId, prefix: 'alert', maxItems: _maxAlerts);
    return getCachedAlerts(elderlyId);
  }

  /// Get all cached alerts for an elderly person, sorted by timestamp desc.
  static List<Map<String, dynamic>> getCachedAlerts(String elderlyId) {
    final box = Hive.box(HiveBoxes.alerts);
    final prefix = '$elderlyId:alert:';
    final results = <Map<String, dynamic>>[];
    for (final key in box.keys) {
      if (key is String && key.startsWith(prefix)) {
        final val = box.get(key);
        if (val is Map) results.add(val);
      }
    }
    results.sort((a, b) {
      final ta = a['created_at'] ?? a['timestamp'] ?? '';
      final tb = b['created_at'] ?? b['timestamp'] ?? '';
      return tb.compareTo(ta);
    });
    return results;
  }

  /// Check if we have fresh cached alerts (within [maxAgeMinutes]).
  static bool hasFreshAlerts(String elderlyId, {int maxAgeMinutes = 5}) {
    final cached = getCachedAlerts(elderlyId);
    if (cached.isEmpty) return false;
    final latest = cached.first;
    final ts = latest['created_at'] as String? ?? latest['timestamp'] as String?;
    if (ts == null) return false;
    try {
      final cachedTime = DateTime.parse(ts);
      return DateTime.now().difference(cachedTime).inMinutes < maxAgeMinutes;
    } catch (_) {
      return false;
    }
  }

  // --- Health Records ---

  static List<Map<String, dynamic>> cacheHealth(String elderlyId, Map<String, dynamic> record) {
    final box = Hive.box(HiveBoxes.health);
    final key = '$elderlyId:health:${record['id']}';
    box.put(key, record);
    _pruneByAge(box, elderlyId, prefix: 'health', maxItems: _maxHealth);
    return getCachedHealth(elderlyId);
  }

  static List<Map<String, dynamic>> getCachedHealth(String elderlyId) {
    final box = Hive.box(HiveBoxes.health);
    final prefix = '$elderlyId:health:';
    final results = <Map<String, dynamic>>[];
    for (final key in box.keys) {
      if (key is String && key.startsWith(prefix)) {
        final val = box.get(key);
        if (val is Map) results.add(val);
      }
    }
    results.sort((a, b) {
      final ta = a['timestamp'] ?? '';
      final tb = b['timestamp'] ?? '';
      return tb.compareTo(ta);
    });
    return results;
  }

  // --- Location ---

  static void cacheLocation(String elderlyId, Map<String, dynamic> location) {
    final box = Hive.box(HiveBoxes.location);
    final key = '$elderlyId:loc:${location['id']}';
    box.put(key, location);
    _pruneByAge(box, elderlyId, prefix: 'loc', maxItems: _maxLocations);
  }

  static List<Map<String, dynamic>> getCachedLocation(String elderlyId) {
    final box = Hive.box(HiveBoxes.location);
    final prefix = '$elderlyId:loc:';
    final results = <Map<String, dynamic>>[];
    for (final key in box.keys) {
      if (key is String && key.startsWith(prefix)) {
        final val = box.get(key);
        if (val is Map) results.add(val);
      }
    }
    results.sort((a, b) {
      final ta = a['timestamp'] ?? '';
      final tb = b['timestamp'] ?? '';
      return tb.compareTo(ta);
    });
    return results;
  }

  // --- Medications ---

  static List<Map<String, dynamic>> cacheMedication(String elderlyId, Map<String, dynamic> record) {
    final box = Hive.box(HiveBoxes.medications);
    final key = '$elderlyId:med:${record['id']}';
    box.put(key, record);
    _pruneByAge(box, elderlyId, prefix: 'med', maxItems: _maxMedications);
    return getCachedMedications(elderlyId);
  }

  static List<Map<String, dynamic>> getCachedMedications(String elderlyId) {
    final box = Hive.box(HiveBoxes.medications);
    final prefix = '$elderlyId:med:';
    final results = <Map<String, dynamic>>[];
    for (final key in box.keys) {
      if (key is String && key.startsWith(prefix)) {
        final val = box.get(key);
        if (val is Map) results.add(val);
      }
    }
    results.sort((a, b) {
      final ta = a['taken_at'] ?? a['timestamp'] ?? '';
      final tb = b['taken_at'] ?? b['timestamp'] ?? '';
      return tb.compareTo(ta);
    });
    return results;
  }

  // --- Settings ---

  static void setSetting(String key, dynamic value) {
    Hive.box(HiveBoxes.settings).put(key, value);
  }

  static T? getSetting<T>(String key, {T? defaultValue}) {
    return Hive.box(HiveBoxes.settings).get(key, defaultValue: defaultValue) as T?;
  }

  // --- Helpers ---

  static void _pruneByAge(Box box, String elderlyId, {required String prefix, required int maxItems}) {
    final fullPrefix = '$elderlyId:$prefix:';
    final keys = box.keys.whereType<String>().where((k) => k.startsWith(fullPrefix)).toList();
    if (keys.length > maxItems) {
      // Remove oldest entries
      final toRemove = keys.take(keys.length - maxItems);
      for (final k in toRemove) {
        box.delete(k);
      }
    }
  }

  /// Clear all cached data. Useful on logout.
  static Future<void> clearAll() async {
    await Hive.box(HiveBoxes.alerts).clear();
    await Hive.box(HiveBoxes.health).clear();
    await Hive.box(HiveBoxes.location).clear();
    await Hive.box(HiveBoxes.medications).clear();
  }
}

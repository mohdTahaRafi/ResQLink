import 'package:hive/hive.dart';
import 'package:hive_flutter/hive_flutter.dart';

/// Manages local Hive boxes for offline-first persistence.
class HiveStore {
  static const String _reportsBox = 'cached_reports';
  static const String _offlineQueueBox = 'offline_queue';
  static const String _userPrefsBox = 'user_prefs';

  static late Box _reports;
  static late Box _queue;
  static late Box _prefs;

  /// Initialize all Hive boxes.
  static Future<void> init() async {
    _reports = await Hive.openBox(_reportsBox);
    _queue = await Hive.openBox(_offlineQueueBox);
    _prefs = await Hive.openBox(_userPrefsBox);
  }

  static Box get cachedReports => _reports;
  static Box get offlineQueue => _queue;
  static Box get userPrefs => _prefs;

  /// Cache a report locally for offline access.
  static Future<void> cacheReport(String id, Map<String, dynamic> data) async {
    await _reports.put(id, data);
  }

  /// Retrieve all cached reports.
  static List<Map<String, dynamic>> getCachedReports() {
    return _reports.values
        .map((e) => Map<String, dynamic>.from(e as Map))
        .toList();
  }

  /// Save the user's selected role.
  static Future<void> saveUserRole(String role) async {
    await _prefs.put('user_role', role);
  }

  /// Get the cached user role.
  static String? getUserRole() {
    return _prefs.get('user_role') as String?;
  }

  /// Clear all local data on logout.
  static Future<void> clearAll() async {
    await _reports.clear();
    await _queue.clear();
    await _prefs.clear();
  }
}

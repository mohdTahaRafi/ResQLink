import 'dart:async';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../data/local/hive_store.dart';

/// Monitors network connectivity and manages offline sync queue.
class ConnectivityService {
  final Connectivity _connectivity = Connectivity();
  StreamSubscription<List<ConnectivityResult>>? _subscription;
  bool _isOnline = true;

  bool get isOnline => _isOnline;

  /// Start monitoring connectivity changes.
  void startMonitoring({required Function() onReconnect}) {
    _subscription = _connectivity.onConnectivityChanged.listen((results) {
      final wasOffline = !_isOnline;
      _isOnline = results.any((r) => r != ConnectivityResult.none);

      if (wasOffline && _isOnline) {
        onReconnect();
      }
    });
  }

  /// Queue a report submission for later sync.
  Future<void> queueReport(Map<String, dynamic> reportData) async {
    final queue = HiveStore.offlineQueue;
    await queue.add(reportData);
  }

  /// Flush all queued reports when connectivity is restored.
  Future<int> syncPendingReports({
    required Future<bool> Function(Map<String, dynamic>) submitFn,
  }) async {
    final queue = HiveStore.offlineQueue;
    int synced = 0;

    while (queue.isNotEmpty) {
      final raw = queue.getAt(0);
      final report = Map<String, dynamic>.from(raw as Map);
      final success = await submitFn(report);
      if (success) {
        await queue.deleteAt(0);
        synced++;
      } else {
        break; // Stop on first failure, retry later
      }
    }

    return synced;
  }

  int get pendingCount => HiveStore.offlineQueue.length;

  void dispose() {
    _subscription?.cancel();
  }
}

final connectivityProvider = Provider<ConnectivityService>((ref) {
  final service = ConnectivityService();
  ref.onDispose(() => service.dispose());
  return service;
});

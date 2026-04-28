import '../../data/local/hive_store.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import '../../main.dart';

/// General Volunteer Dashboard — View assigned tasks and update status.
class VolunteerDashboard extends ConsumerStatefulWidget {
  const VolunteerDashboard({super.key});

  @override
  ConsumerState<VolunteerDashboard> createState() => _VolunteerDashboardState();
}

class _VolunteerDashboardState extends ConsumerState<VolunteerDashboard> {
  List<Map<String, dynamic>> _tasks = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadTasks();
  }

  Future<void> _loadTasks() async {
    setState(() => _isLoading = true);
    try {
      final api = ref.read(apiClientProvider);
      final tasks = await api.getMyTasks();
      setState(() {
        _tasks = List<Map<String, dynamic>>.from(tasks);
        _isLoading = false;
      });
    } catch (e) {
      debugPrint('Load tasks error: $e');
      setState(() => _isLoading = false);
    }
  }

  Future<void> _updateStatus(String reportId, String newStatus) async {
    try {
      final api = ref.read(apiClientProvider);
      await api.updateReportStatus(reportId, newStatus);
      _showSnackBar('Status updated to ${newStatus.replaceAll("_", " ")}');
      _loadTasks();
    } catch (e) {
      _showSnackBar('Failed to update status');
    }
  }

  void _showSnackBar(String msg) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(msg)));
    }
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'pending':
        return Colors.grey;
      case 'accepted':
        return Colors.blue;
      case 'in_progress':
        return Colors.orange;
      case 'escalated':
        return Colors.red;
      case 'resolved':
        return Colors.green;
      default:
        return Colors.grey;
    }
  }

  IconData _urgencyIcon(String urgency) {
    switch (urgency) {
      case 'critical':
        return Icons.error;
      case 'urgent':
        return Icons.warning_amber;
      default:
        return Icons.info_outline;
    }
  }

  Color _urgencyColor(String urgency) {
    switch (urgency) {
      case 'critical':
        return Colors.red;
      case 'urgent':
        return Colors.orange;
      default:
        return Colors.green;
    }
  }

  String _nextStatus(String current) {
    switch (current) {
      case 'pending':
        return 'accepted';
      case 'accepted':
        return 'in_progress';
      case 'in_progress':
        return 'resolved';
      default:
        return current;
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('My Tasks'),
        actions: [
          IconButton(
            icon: const Icon(Icons.auto_awesome),
            tooltip: 'AI Hub',
            onPressed: () => context.push('/ai-hub'),
          ),
          IconButton(icon: const Icon(Icons.refresh), onPressed: _loadTasks),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await HiveStore.clearAll();
              await FirebaseAuth.instance.signOut();
              if (!context.mounted) return;
              context.go('/role-select');
            },
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _tasks.isEmpty
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.assignment_outlined,
                          size: 64, color: theme.colorScheme.outlineVariant),
                      const SizedBox(height: 16),
                      Text('No tasks assigned yet',
                          style: theme.textTheme.bodyLarge),
                      Text(
                          'Tasks will appear here when an NGO admin assigns you.',
                          style: theme.textTheme.bodySmall?.copyWith(
                              color: theme.colorScheme.onSurfaceVariant)),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadTasks,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _tasks.length,
                    itemBuilder: (context, index) {
                      final task = _tasks[index];
                      final status = (task['status'] as String?) ?? 'pending';
                      final urgency =
                          (task['user_urgency'] as String?) ?? 'normal';
                      final issueType =
                          (task['issue_type'] as String?) ?? 'civic_issue';
                      final canAdvance =
                          status != 'resolved' && status != 'escalated';

                      return Card(
                        margin: const EdgeInsets.only(bottom: 12),
                        child: Padding(
                          padding: const EdgeInsets.all(16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              // Header: urgency + status
                              Row(
                                children: [
                                  Icon(_urgencyIcon(urgency),
                                      color: _urgencyColor(urgency), size: 20),
                                  const SizedBox(width: 6),
                                  Text(urgency.toUpperCase(),
                                      style: theme.textTheme.labelSmall
                                          ?.copyWith(
                                              color: _urgencyColor(urgency),
                                              fontWeight: FontWeight.bold)),
                                  const Spacer(),
                                  Chip(
                                    label: Text(
                                        status
                                            .replaceAll('_', ' ')
                                            .toUpperCase(),
                                        style: TextStyle(
                                            fontSize: 10,
                                            color: _statusColor(status),
                                            fontWeight: FontWeight.bold)),
                                    backgroundColor: _statusColor(status)
                                        .withValues(alpha: 0.1),
                                    side: BorderSide.none,
                                    padding: EdgeInsets.zero,
                                  ),
                                ],
                              ),
                              const SizedBox(height: 8),

                              // Issue type badge
                              Chip(
                                  label: Text(
                                      issueType
                                          .replaceAll('_', ' ')
                                          .toUpperCase(),
                                      style: const TextStyle(fontSize: 10))),
                              const SizedBox(height: 8),

                              // Description
                              Text(
                                (task['raw_text'] as String?) ??
                                    'No description',
                                style: theme.textTheme.bodyMedium,
                                maxLines: 3,
                                overflow: TextOverflow.ellipsis,
                              ),
                              const SizedBox(height: 8),

                              // Location
                              if (task['location'] != null &&
                                  (task['location'] as String).isNotEmpty)
                                Row(
                                  children: [
                                    Icon(Icons.location_on,
                                        size: 14,
                                        color:
                                            theme.colorScheme.onSurfaceVariant),
                                    const SizedBox(width: 4),
                                    Expanded(
                                        child: Text(task['location'] as String,
                                            style: theme.textTheme.bodySmall
                                                ?.copyWith(
                                                    color: theme.colorScheme
                                                        .onSurfaceVariant))),
                                  ],
                                ),
                              const SizedBox(height: 12),

                              // Action buttons
                              Row(
                                children: [
                                  if (canAdvance)
                                    Expanded(
                                      child: FilledButton.icon(
                                        onPressed: () => _updateStatus(
                                            task['id'] as String,
                                            _nextStatus(status)),
                                        icon: const Icon(Icons.check, size: 18),
                                        label: Text(
                                            'Mark ${_nextStatus(status).replaceAll("_", " ")}'),
                                      ),
                                    ),
                                  if (canAdvance &&
                                      status == 'in_progress') ...[
                                    const SizedBox(width: 8),
                                    OutlinedButton(
                                      onPressed: () => _updateStatus(
                                          task['id'] as String, 'escalated'),
                                      child: const Text('Escalate'),
                                    ),
                                  ],
                                ],
                              ),
                            ],
                          ),
                        ),
                      );
                    },
                  ),
                ),
    );
  }
}

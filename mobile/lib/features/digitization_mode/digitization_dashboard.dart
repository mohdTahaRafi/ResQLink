import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';

/// Tehsil Clerk "Digitization Mode" — OCR queue for processing paper surveys.
class DigitizationDashboard extends StatefulWidget {
  const DigitizationDashboard({super.key});

  @override
  State<DigitizationDashboard> createState() => _DigitizationDashboardState();
}

class _DigitizationDashboardState extends State<DigitizationDashboard> {
  final List<_OcrItem> _queue = [
    _OcrItem(
        id: '1', title: 'Survey Form — Ward 14', pages: 3, status: 'pending'),
    _OcrItem(
        id: '2',
        title: 'Census Record — Block C',
        pages: 7,
        status: 'processing'),
    _OcrItem(
        id: '3',
        title: 'Land Revenue — Khasra 45',
        pages: 2,
        status: 'completed'),
    _OcrItem(
        id: '4',
        title: 'Ration Card App — Batch 12',
        pages: 4,
        status: 'pending'),
    _OcrItem(
        id: '5',
        title: 'Birth Certificate — Jan 2025',
        pages: 1,
        status: 'pending'),
  ];

  int get _pendingCount => _queue.where((i) => i.status == 'pending').length;
  int get _completedCount =>
      _queue.where((i) => i.status == 'completed').length;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Digitization Mode'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await FirebaseAuth.instance.signOut();
              if (!context.mounted) return;
              context.go('/login');
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Stats row
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                _StatCard(
                  label: 'Pending',
                  value: '$_pendingCount',
                  color: Colors.orange,
                  icon: Icons.pending_actions,
                ),
                const SizedBox(width: 12),
                _StatCard(
                  label: 'Processing',
                  value:
                      '${_queue.where((i) => i.status == 'processing').length}',
                  color: Colors.blue,
                  icon: Icons.hourglass_top,
                ),
                const SizedBox(width: 12),
                _StatCard(
                  label: 'Done',
                  value: '$_completedCount',
                  color: Colors.green,
                  icon: Icons.check_circle,
                ),
              ],
            ),
          ),

          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Row(
              children: [
                Text('OCR Queue',
                    style: theme.textTheme.titleMedium
                        ?.copyWith(fontWeight: FontWeight.w600)),
                const Spacer(),
                TextButton.icon(
                  onPressed: () {
                    // TODO: Scan new document
                  },
                  icon: const Icon(Icons.add_a_photo_rounded, size: 18),
                  label: const Text('Scan New'),
                ),
              ],
            ),
          ),

          Expanded(
            child: ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _queue.length,
              itemBuilder: (context, index) {
                final item = _queue[index];
                return Card(
                  margin: const EdgeInsets.only(bottom: 10),
                  child: ListTile(
                    leading: CircleAvatar(
                      backgroundColor:
                          _statusColor(item.status).withValues(alpha: 0.15),
                      child: Icon(
                        _statusIcon(item.status),
                        color: _statusColor(item.status),
                        size: 22,
                      ),
                    ),
                    title: Text(item.title, style: theme.textTheme.titleSmall),
                    subtitle: Text('${item.pages} pages'),
                    trailing: Chip(
                      label: Text(item.status.toUpperCase(),
                          style: TextStyle(
                            fontSize: 10,
                            color: _statusColor(item.status),
                            fontWeight: FontWeight.bold,
                          )),
                      backgroundColor:
                          _statusColor(item.status).withValues(alpha: 0.1),
                      side: BorderSide.none,
                      padding: EdgeInsets.zero,
                    ),
                    onTap: () {
                      if (item.status == 'pending') {
                        setState(() => item.status = 'processing');
                      }
                    },
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'pending':
        return Colors.orange;
      case 'processing':
        return Colors.blue;
      case 'completed':
        return Colors.green;
      default:
        return Colors.grey;
    }
  }

  IconData _statusIcon(String status) {
    switch (status) {
      case 'pending':
        return Icons.pending_actions;
      case 'processing':
        return Icons.document_scanner;
      case 'completed':
        return Icons.check_circle_outline;
      default:
        return Icons.help_outline;
    }
  }
}

class _OcrItem {
  final String id;
  final String title;
  final int pages;
  String status;

  _OcrItem(
      {required this.id,
      required this.title,
      required this.pages,
      required this.status});
}

class _StatCard extends StatelessWidget {
  final String label;
  final String value;
  final Color color;
  final IconData icon;

  const _StatCard({
    required this.label,
    required this.value,
    required this.color,
    required this.icon,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(14),
          child: Column(
            children: [
              Icon(icon, color: color, size: 24),
              const SizedBox(height: 6),
              Text(value,
                  style: Theme.of(context)
                      .textTheme
                      .headlineSmall
                      ?.copyWith(fontWeight: FontWeight.bold, color: color)),
              Text(label, style: Theme.of(context).textTheme.labelSmall),
            ],
          ),
        ),
      ),
    );
  }
}

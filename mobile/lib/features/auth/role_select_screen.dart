import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import '../../core/router.dart';
import '../../data/local/hive_store.dart';

class RoleSelectScreen extends StatelessWidget {
  const RoleSelectScreen({super.key});

  static const _roles = [
    _RoleOption(
      role: 'ngo_worker',
      title: 'NGO Worker',
      subtitle: 'Field Mode — Report & survey',
      icon: Icons.assignment_turned_in_rounded,
      color: Color(0xFF2E7D32),
    ),
    _RoleOption(
      role: 'lawyer',
      title: 'Lawyer / Vakeel',
      subtitle: 'Legal Mode — FIR & deed search',
      icon: Icons.gavel_rounded,
      color: Color(0xFF1565C0),
    ),
    _RoleOption(
      role: 'clerk',
      title: 'Tehsil Clerk',
      subtitle: 'Digitization Mode — OCR queue',
      icon: Icons.document_scanner_rounded,
      color: Color(0xFF6A1B9A),
    ),
    _RoleOption(
      role: 'nagar_nigam',
      title: 'Nagar Nigam',
      subtitle: 'Command Mode — City heatmap',
      icon: Icons.map_rounded,
      color: Color(0xFFE65100),
    ),
    _RoleOption(
      role: 'donor',
      title: 'Donor',
      subtitle: 'Transparency Mode — Impact reports',
      icon: Icons.volunteer_activism_rounded,
      color: Color(0xFF00838F),
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      appBar: AppBar(title: const Text('Select Your Role')),
      body: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'How will you use SAMAJ?',
              style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 8),
            Text(
              'Choose your role to access the right dashboard.',
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
            ),
            const SizedBox(height: 24),
            Expanded(
              child: ListView.separated(
                itemCount: _roles.length,
                separatorBuilder: (_, __) => const SizedBox(height: 12),
                itemBuilder: (context, index) {
                  final option = _roles[index];
                  return _RoleCard(
                    option: option,
                    onTap: () async {
                      await HiveStore.saveUserRole(option.role);
                      if (context.mounted) {
                        context.go(dashboardRouteForRole(option.role));
                      }
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _RoleOption {
  final String role;
  final String title;
  final String subtitle;
  final IconData icon;
  final Color color;

  const _RoleOption({
    required this.role,
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.color,
  });
}

class _RoleCard extends StatelessWidget {
  final _RoleOption option;
  final VoidCallback onTap;

  const _RoleCard({required this.option, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Card(
      clipBehavior: Clip.antiAlias,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Row(
            children: [
              Container(
                width: 52,
                height: 52,
                decoration: BoxDecoration(
                  color: option.color.withOpacity(0.12),
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Icon(option.icon, color: option.color, size: 28),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      option.title,
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      option.subtitle,
                      style: Theme.of(context).textTheme.bodySmall?.copyWith(
                            color: Theme.of(context).colorScheme.onSurfaceVariant,
                          ),
                    ),
                  ],
                ),
              ),
              Icon(Icons.arrow_forward_ios, size: 16, color: option.color),
            ],
          ),
        ),
      ),
    );
  }
}

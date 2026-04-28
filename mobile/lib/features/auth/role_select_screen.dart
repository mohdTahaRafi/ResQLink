import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:firebase_auth/firebase_auth.dart';
import '../../data/local/hive_store.dart';

class RoleSelectScreen extends StatelessWidget {
  const RoleSelectScreen({super.key});

  static const _roles = [
    _RoleOption(
      role: 'reporter',
      title: 'General User',
      subtitle: 'Report an issue — photos, location, description',
      icon: Icons.report_problem_rounded,
      gradient: [Color(0xFFFF6B6B), Color(0xFFEE5A24)],
    ),
    _RoleOption(
      role: 'volunteer',
      title: 'Volunteer',
      subtitle: 'View & manage assigned tasks',
      icon: Icons.volunteer_activism_rounded,
      gradient: [Color(0xFF00BFA5), Color(0xFF00897B)],
    ),
    _RoleOption(
      role: 'specialist',
      title: 'Lawyer / Doctor',
      subtitle: 'Case files, AI search & document analysis',
      icon: Icons.gavel_rounded,
      gradient: [Color(0xFF6C63FF), Color(0xFF3F51B5)],
    ),
    _RoleOption(
      role: 'ngo_admin',
      title: 'NGO Organization',
      subtitle: 'Issue management, volunteer matching & heatmap',
      icon: Icons.admin_panel_settings_rounded,
      gradient: [Color(0xFFAB47BC), Color(0xFF7B1FA2)],
    ),
  ];

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const SizedBox(height: 20),
              // Logo + Title
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(colors: [Color(0xFF6C63FF), Color(0xFF00BFA5)]),
                      borderRadius: BorderRadius.circular(14),
                    ),
                    child: const Icon(Icons.people_alt_rounded, color: Colors.white, size: 28),
                  ),
                  const SizedBox(width: 12),
                  Text('SAMAJ', style: theme.textTheme.headlineMedium?.copyWith(
                    fontWeight: FontWeight.w800, letterSpacing: 1.5,
                  )),
                ],
              ),
              const SizedBox(height: 32),
              Text(
                'Choose your role',
                style: theme.textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 6),
              Text(
                'Select how you want to use the platform.\nYou\'ll sign in on the next screen.',
                style: theme.textTheme.bodyMedium?.copyWith(color: theme.colorScheme.onSurfaceVariant),
              ),
              const SizedBox(height: 28),
              Expanded(
                child: ListView.separated(
                  itemCount: _roles.length,
                  separatorBuilder: (_, __) => const SizedBox(height: 14),
                  itemBuilder: (context, index) {
                    final option = _roles[index];
                    return _RoleCard(
                      option: option,
                      onTap: () async {
                        // Save selected role
                        await HiveStore.saveUserRole(option.role);
                        if (context.mounted) {
                          // Always go to login — they need to authenticate
                          context.go('/login');
                        }
                      },
                    );
                  },
                ),
              ),
            ],
          ),
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
  final List<Color> gradient;

  const _RoleOption({
    required this.role,
    required this.title,
    required this.subtitle,
    required this.icon,
    required this.gradient,
  });
}

class _RoleCard extends StatelessWidget {
  final _RoleOption option;
  final VoidCallback onTap;

  const _RoleCard({required this.option, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Card(
      clipBehavior: Clip.antiAlias,
      elevation: isDark ? 0 : 2,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.all(18),
          child: Row(
            children: [
              Container(
                width: 52,
                height: 52,
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: option.gradient,
                    begin: Alignment.topLeft,
                    end: Alignment.bottomRight,
                  ),
                  borderRadius: BorderRadius.circular(14),
                  boxShadow: [
                    BoxShadow(
                      color: option.gradient.first.withOpacity(0.3),
                      blurRadius: 8,
                      offset: const Offset(0, 3),
                    ),
                  ],
                ),
                child: Icon(option.icon, color: Colors.white, size: 26),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(option.title, style: theme.textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w700)),
                    const SizedBox(height: 3),
                    Text(option.subtitle, style: theme.textTheme.bodySmall?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    )),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: option.gradient.first.withOpacity(isDark ? 0.2 : 0.1),
                  shape: BoxShape.circle,
                ),
                child: Icon(Icons.arrow_forward_ios, size: 14, color: option.gradient.first),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:firebase_auth/firebase_auth.dart';

import '../features/auth/login_screen.dart';
import '../features/auth/role_select_screen.dart';
import '../features/field_mode/field_dashboard.dart';
import '../features/legal_mode/legal_dashboard.dart';
import '../features/digitization_mode/digitization_dashboard.dart';
import '../features/command_mode/command_dashboard.dart';
import '../features/transparency_mode/transparency_dashboard.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      final user = FirebaseAuth.instance.currentUser;
      final isLoginRoute = state.matchedLocation == '/login';

      if (user == null && !isLoginRoute) return '/login';
      if (user != null && isLoginRoute) return '/role-select';

      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/role-select',
        name: 'role-select',
        builder: (context, state) => const RoleSelectScreen(),
      ),
      // ── Role-gated dashboard routes ──
      GoRoute(
        path: '/dashboard/field',
        name: 'field-mode',
        builder: (context, state) => const FieldDashboard(),
      ),
      GoRoute(
        path: '/dashboard/legal',
        name: 'legal-mode',
        builder: (context, state) => const LegalDashboard(),
      ),
      GoRoute(
        path: '/dashboard/digitization',
        name: 'digitization-mode',
        builder: (context, state) => const DigitizationDashboard(),
      ),
      GoRoute(
        path: '/dashboard/command',
        name: 'command-mode',
        builder: (context, state) => const CommandDashboard(),
      ),
      GoRoute(
        path: '/dashboard/transparency',
        name: 'transparency-mode',
        builder: (context, state) => const TransparencyDashboard(),
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(
        child: Text('Route not found: ${state.matchedLocation}'),
      ),
    ),
  );
});

/// Maps a role string to its dashboard route.
String dashboardRouteForRole(String role) {
  switch (role) {
    case 'ngo_worker':
      return '/dashboard/field';
    case 'lawyer':
      return '/dashboard/legal';
    case 'clerk':
      return '/dashboard/digitization';
    case 'nagar_nigam':
      return '/dashboard/command';
    case 'donor':
      return '/dashboard/transparency';
    default:
      return '/role-select';
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:firebase_auth/firebase_auth.dart';

import '../features/auth/login_screen.dart';
import '../features/auth/role_select_screen.dart';
import '../features/reporter_mode/reporter_dashboard.dart';
import '../features/volunteer_mode/volunteer_dashboard.dart';
import '../features/specialist_mode/specialist_dashboard.dart';
import '../features/admin_mode/admin_dashboard.dart';
import '../features/ai/ai_hub_screen.dart';
import '../data/local/hive_store.dart';

final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/role-select',
    redirect: (context, state) {
      final user = FirebaseAuth.instance.currentUser;
      final savedRole = HiveStore.getUserRole();
      final loc = state.matchedLocation;

      // ── NOT logged in ──
      if (user == null) {
        // Can access role-select and login only
        if (loc == '/role-select' || loc == '/login') return null;
        // Everything else → go pick a role first
        return '/role-select';
      }

      // ── LOGGED IN ──

      // If on login page, redirect to their dashboard
      if (loc == '/login') {
        if (savedRole != null) return dashboardRouteForRole(savedRole);
        return '/role-select';
      }

      // If on role-select and already logged in → send to their dashboard
      // (they must logout first to change roles)
      if (loc == '/role-select') {
        if (savedRole != null) return dashboardRouteForRole(savedRole);
        return null; // no role saved yet, let them pick
      }

      // If trying to access a dashboard that's NOT their role → block it
      if (savedRole != null) {
        final allowedRoute = dashboardRouteForRole(savedRole);
        // Allow AI Hub for all logged-in users
        if (loc == '/ai-hub') return null;
        if (loc.startsWith('/dashboard/') && loc != allowedRoute) {
          return allowedRoute; // redirect to their own dashboard
        }
      }

      return null;
    },
    routes: [
      GoRoute(
        path: '/role-select',
        name: 'role-select',
        builder: (context, state) => const RoleSelectScreen(),
      ),
      GoRoute(
        path: '/login',
        name: 'login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/dashboard/reporter',
        name: 'reporter-mode',
        builder: (context, state) => const ReporterDashboard(),
      ),
      GoRoute(
        path: '/dashboard/volunteer',
        name: 'volunteer-mode',
        builder: (context, state) => const VolunteerDashboard(),
      ),
      GoRoute(
        path: '/dashboard/specialist',
        name: 'specialist-mode',
        builder: (context, state) => const SpecialistDashboard(),
      ),
      GoRoute(
        path: '/dashboard/admin',
        name: 'admin-mode',
        builder: (context, state) => const AdminDashboard(),
      ),
      GoRoute(
        path: '/ai-hub',
        name: 'ai-hub',
        builder: (context, state) => const AIHubScreen(),
      ),
    ],
    errorBuilder: (context, state) => Scaffold(
      body: Center(child: Text('Route not found: ${state.matchedLocation}')),
    ),
  );
});

/// Maps a role string to its dashboard route.
String dashboardRouteForRole(String role) {
  switch (role) {
    case 'reporter':
      return '/dashboard/reporter';
    case 'volunteer':
      return '/dashboard/volunteer';
    case 'specialist':
      return '/dashboard/specialist';
    case 'ngo_admin':
      return '/dashboard/admin';
    default:
      return '/role-select';
  }
}

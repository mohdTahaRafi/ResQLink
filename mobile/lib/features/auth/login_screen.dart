import 'package:flutter/material.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:go_router/go_router.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';
import '../../data/local/hive_store.dart';
import '../../core/router.dart';
import '../../main.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _isLoading = false;
  bool _isGoogleLoading = false;
  bool _isSignUp = false;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  /// Tries to register the user as a volunteer/specialist in Firestore.
  Future<void> _tryRegisterVolunteer() async {
    final role = HiveStore.getUserRole();
    if (role != 'volunteer' && role != 'specialist') return;

    try {
      await Future.delayed(const Duration(milliseconds: 500));
      final user = FirebaseAuth.instance.currentUser;
      final api = ref.read(apiClientProvider);
      final result = await api.createVolunteer(
        name: user?.displayName ?? user?.email?.split('@').first ?? 'Volunteer',
        skills: role == 'specialist' ? ['specialist', 'legal', 'medical'] : ['general'],
        latitude: 26.8467,
        longitude: 80.9462,
      );
      debugPrint('✅ Volunteer registered: $result');
    } catch (e) {
      debugPrint('⚠️ Volunteer registration failed: $e');
    }
  }

  /// Navigate after successful auth
  void _navigateToDashboard() {
    if (!mounted) return;
    final role = HiveStore.getUserRole();
    if (role != null) {
      context.go(dashboardRouteForRole(role));
    } else {
      context.go('/role-select');
    }
  }

  /// Email/Password sign in or sign up
  Future<void> _submitEmailPassword() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _isLoading = true);

    try {
      if (_isSignUp) {
        await FirebaseAuth.instance.createUserWithEmailAndPassword(
          email: _emailController.text.trim(),
          password: _passwordController.text,
        );
      } else {
        await FirebaseAuth.instance.signInWithEmailAndPassword(
          email: _emailController.text.trim(),
          password: _passwordController.text,
        );
      }
      await _tryRegisterVolunteer();
      _navigateToDashboard();
    } on FirebaseAuthException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message ?? 'Authentication failed')),
        );
      }
    } finally {
      if (mounted) setState(() => _isLoading = false);
    }
  }

  /// Google OAuth sign-in
  Future<void> _signInWithGoogle() async {
    setState(() => _isGoogleLoading = true);

    try {
      // For web: use signInWithPopup
      final googleProvider = GoogleAuthProvider();
      googleProvider.addScope('email');
      googleProvider.addScope('profile');

      await FirebaseAuth.instance.signInWithPopup(googleProvider);

      await _tryRegisterVolunteer();
      _navigateToDashboard();
    } on FirebaseAuthException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.message ?? 'Google sign-in failed')),
        );
      }
    } catch (e) {
      debugPrint('Google sign-in error: $e');
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Google sign-in failed. Please try again.')),
        );
      }
    } finally {
      if (mounted) setState(() => _isGoogleLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final role = HiveStore.getUserRole() ?? '';
    final roleLabel = {
      'reporter': 'General User',
      'volunteer': 'Volunteer',
      'specialist': 'Lawyer / Doctor',
      'ngo_admin': 'NGO Organization',
    }[role] ?? 'User';

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.all(32),
            child: Form(
              key: _formKey,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Logo
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        colors: [Color(0xFF6C63FF), Color(0xFF00BFA5)],
                      ),
                      borderRadius: BorderRadius.circular(20),
                    ),
                    child: const Icon(Icons.people_alt_rounded, size: 48, color: Colors.white),
                  ),
                  const SizedBox(height: 20),
                  Text(
                    'RESQLINK',
                    style: theme.textTheme.headlineLarge?.copyWith(
                      fontWeight: FontWeight.w800,
                      letterSpacing: 1.5,
                    ),
                  ),
                  const SizedBox(height: 6),
                  Text(
                    'Community Intelligence Platform',
                    style: theme.textTheme.bodyMedium?.copyWith(
                      color: theme.colorScheme.onSurfaceVariant,
                    ),
                  ),
                  const SizedBox(height: 12),

                  // Role badge
                  if (role.isNotEmpty)
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                      decoration: BoxDecoration(
                        color: theme.colorScheme.primaryContainer.withOpacity(isDark ? 0.3 : 0.6),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Text(
                        'Signing in as $roleLabel',
                        style: theme.textTheme.labelMedium?.copyWith(
                          color: theme.colorScheme.primary,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  const SizedBox(height: 28),

                  // ── Google OAuth Button ──
                  SizedBox(
                    width: double.infinity,
                    height: 52,
                    child: OutlinedButton.icon(
                      onPressed: _isGoogleLoading ? null : _signInWithGoogle,
                      style: OutlinedButton.styleFrom(
                        side: BorderSide(color: theme.colorScheme.outlineVariant),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                      ),
                      icon: _isGoogleLoading
                          ? const SizedBox(
                              width: 20, height: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Image.network(
                              'https://www.gstatic.com/firebasejs/ui/2.0.0/images/auth/google.svg',
                              width: 22, height: 22,
                              errorBuilder: (_, __, ___) => const Icon(Icons.g_mobiledata, size: 24),
                            ),
                      label: Text(
                        _isGoogleLoading ? 'Signing in...' : 'Continue with Google',
                        style: TextStyle(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: theme.colorScheme.onSurface,
                        ),
                      ),
                    ),
                  ),

                  const SizedBox(height: 20),

                  // ── Divider ──
                  Row(
                    children: [
                      Expanded(child: Divider(color: theme.colorScheme.outlineVariant)),
                      Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 16),
                        child: Text('or', style: theme.textTheme.bodySmall?.copyWith(
                          color: theme.colorScheme.onSurfaceVariant,
                        )),
                      ),
                      Expanded(child: Divider(color: theme.colorScheme.outlineVariant)),
                    ],
                  ),

                  const SizedBox(height: 20),

                  // ── Email field ──
                  TextFormField(
                    controller: _emailController,
                    keyboardType: TextInputType.emailAddress,
                    decoration: const InputDecoration(
                      labelText: 'Email',
                      prefixIcon: Icon(Icons.email_outlined),
                    ),
                    validator: (v) =>
                        v != null && v.contains('@') ? null : 'Enter a valid email',
                  ),
                  const SizedBox(height: 16),

                  // ── Password field ──
                  TextFormField(
                    controller: _passwordController,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'Password',
                      prefixIcon: Icon(Icons.lock_outlined),
                    ),
                    validator: (v) =>
                        v != null && v.length >= 6 ? null : 'Minimum 6 characters',
                  ),
                  const SizedBox(height: 28),

                  // ── Email Submit Button ──
                  SizedBox(
                    width: double.infinity,
                    height: 52,
                    child: FilledButton(
                      onPressed: _isLoading ? null : _submitEmailPassword,
                      child: _isLoading
                          ? const SizedBox(
                              height: 20, width: 20,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            )
                          : Text(_isSignUp ? 'Create Account' : 'Sign In',
                              style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                    ),
                  ),
                  const SizedBox(height: 14),

                  // ── Toggle sign-up / sign-in ──
                  TextButton(
                    onPressed: () => setState(() => _isSignUp = !_isSignUp),
                    child: Text(
                      _isSignUp
                          ? 'Already have an account? Sign In'
                          : 'Need an account? Sign Up',
                    ),
                  ),
                  const SizedBox(height: 4),

                  // ── Change role ──
                  TextButton.icon(
                    onPressed: () => context.go('/role-select'),
                    icon: const Icon(Icons.arrow_back, size: 16),
                    label: const Text('Change role'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

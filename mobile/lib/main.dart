import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'firebase_options.dart';
import 'core/router.dart';
import 'core/theme.dart';
import 'data/local/hive_store.dart';
import 'data/remote/api_client.dart';

/// Global ApiClient provider — all widgets can read this via `ref.read(apiClientProvider)`.
final apiClientProvider = Provider<ApiClient>((ref) {
  // For web (Chrome): use localhost directly
  // For Android emulator: 10.0.2.2 is the alias for host machine's localhost
  // For physical device: replace with your machine's LAN IP
  const isWeb = bool.fromEnvironment('dart.library.js_util', defaultValue: false);
  final baseUrl = isWeb ? 'http://localhost:8080' : 'http://10.0.2.2:8080';
  return ApiClient(baseUrl: baseUrl);
});

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  await Firebase.initializeApp(
    options: DefaultFirebaseOptions.currentPlatform,
  );
  await Hive.initFlutter();
  await HiveStore.init();

  runApp(const ProviderScope(child: SamajApp()));
}

class SamajApp extends ConsumerWidget {
  const SamajApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'SAMAJ',
      debugShowCheckedModeBanner: false,
      theme: SamajTheme.light(),
      darkTheme: SamajTheme.dark(),
      themeMode: ThemeMode.system,
      routerConfig: router,
    );
  }
}

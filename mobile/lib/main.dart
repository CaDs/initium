import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'data/local/token_storage.dart';
import 'providers/api_provider.dart';
import 'presentation/router/app_router.dart';

const _devBypassAuth = bool.fromEnvironment('DEV_BYPASS_AUTH');

void main() async {
  // DEV_BYPASS_AUTH must never be enabled in release builds.
  assert(
    !(kReleaseMode && _devBypassAuth),
    'DEV_BYPASS_AUTH must not be enabled in release builds',
  );

  WidgetsFlutterBinding.ensureInitialized();

  // Initialize token storage (handles iOS keychain wipe on reinstall).
  final tokenStorage = TokenStorage();
  await tokenStorage.initialize();

  runApp(
    ProviderScope(
      overrides: [
        tokenStorageProvider.overrideWithValue(tokenStorage),
      ],
      child: const InitiumApp(),
    ),
  );
}

class InitiumApp extends ConsumerWidget {
  const InitiumApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'Initium',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: Colors.grey),
        useMaterial3: true,
      ),
      routerConfig: router,
    );
  }
}

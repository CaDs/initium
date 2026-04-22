import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:mobile/l10n/app_localizations.dart';
import 'data/local/token_storage.dart';
import 'providers/api_provider.dart';
import 'providers/theme_provider.dart';
import 'providers/locale_provider.dart';
import 'presentation/router/app_router.dart';

const _devBypassAuth = bool.fromEnvironment('DEV_BYPASS_AUTH');

void main() async {
  assert(
    !(kReleaseMode && _devBypassAuth),
    'DEV_BYPASS_AUTH must not be enabled in release builds',
  );

  WidgetsFlutterBinding.ensureInitialized();

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
    final themeMode = ref.watch(themeProvider);
    final locale = ref.watch(localeProvider);

    const seed = Colors.indigo;
    return MaterialApp.router(
      title: 'Initium',
      debugShowCheckedModeBanner: false,
      locale: locale,
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      supportedLocales: AppLocalizations.supportedLocales,
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(seedColor: seed),
        useMaterial3: true,
      ),
      darkTheme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: seed,
          brightness: Brightness.dark,
        ),
        useMaterial3: true,
      ),
      themeMode: themeMode,
      routerConfig: router,
    );
  }
}

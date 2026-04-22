import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import '../shared/dev_mode_banner.dart';
import '../shared/locale_switcher.dart';
import '../shared/theme_switcher.dart';
import 'widgets/google_sign_in_button.dart';
import 'widgets/magic_link_form.dart';

class LoginScreen extends ConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.appName),
        actions: const [LocaleSwitcher(), ThemeSwitcher()],
      ),
      body: Column(
        children: [
          const DevModeBanner(),
          Expanded(
            child: SafeArea(
              child: Center(
                child: SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  keyboardDismissBehavior:
                      ScrollViewKeyboardDismissBehavior.onDrag,
                  child: ConstrainedBox(
                    constraints: const BoxConstraints(maxWidth: 420),
                    child: Card(
                      margin: EdgeInsets.zero,
                      child: Padding(
                        padding: const EdgeInsets.all(20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            const GoogleSignInButton(),
                            const SizedBox(height: 12),
                            const MagicLinkForm(),
                            if (authState is AuthError) ...[
                              const SizedBox(height: 16),
                              _ErrorBanner(
                                message:
                                    _localizeError(authState.message, l10n),
                              ),
                            ],
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _localizeError(String code, AppLocalizations l10n) {
    return switch (code) {
      'google_login_failed' => l10n.authGoogleLoginFailed,
      'magic_link_failed' => l10n.authMagicLinkFailed,
      _ => code,
    };
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.message});
  final String message;

  @override
  Widget build(BuildContext context) {
    final color = Theme.of(context).colorScheme.error;
    return Semantics(
      liveRegion: true,
      child: Row(
        children: [
          Icon(Icons.error_outline, color: color, size: 18),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              style: TextStyle(color: color),
              semanticsLabel: message,
            ),
          ),
        ],
      ),
    );
  }
}

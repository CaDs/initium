import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';
import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import 'widgets/google_sign_in_button.dart';
import 'widgets/magic_link_form.dart';

class LoginScreen extends ConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return Scaffold(
      body: Center(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(32),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                l10n.loginTitle,
                style: theme.textTheme.headlineMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                l10n.loginSubtitle,
                style: theme.textTheme.bodyLarge?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
              const SizedBox(height: 48),
              const GoogleSignInButton(),
              const SizedBox(height: 24),
              Row(
                children: [
                  Expanded(child: Divider(color: theme.colorScheme.outlineVariant)),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: Text(
                      l10n.loginDivider,
                      style: TextStyle(color: theme.colorScheme.onSurfaceVariant),
                    ),
                  ),
                  Expanded(child: Divider(color: theme.colorScheme.outlineVariant)),
                ],
              ),
              const SizedBox(height: 24),
              const MagicLinkForm(),
              if (authState is AuthError)
                Padding(
                  padding: const EdgeInsets.only(top: 16),
                  child: Text(
                    _localizeError(authState.message, l10n),
                    style: TextStyle(color: theme.colorScheme.error),
                    semanticsLabel: _localizeError(authState.message, l10n),
                  ),
                ),
            ],
          ),
        ),
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

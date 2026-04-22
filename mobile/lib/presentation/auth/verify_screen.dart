import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';

/// Screen shown when a magic link deep link is opened. Automatically
/// verifies the token and routes home on success.
class VerifyScreen extends ConsumerStatefulWidget {
  final String token;

  const VerifyScreen({super.key, required this.token});

  @override
  ConsumerState<VerifyScreen> createState() => _VerifyScreenState();
}

enum _Stage { loading, error }

class _VerifyScreenState extends ConsumerState<VerifyScreen> {
  _Stage _stage = _Stage.loading;

  @override
  void initState() {
    super.initState();
    _verify();
  }

  Future<void> _verify() async {
    setState(() => _stage = _Stage.loading);

    if (widget.token.isEmpty) {
      if (mounted) setState(() => _stage = _Stage.error);
      return;
    }

    await ref.read(authProvider.notifier).verifyMagicLink(widget.token);
    if (!mounted) return;

    final state = ref.read(authProvider);
    if (state is AuthAuthenticated) {
      context.go('/home');
    } else {
      setState(() => _stage = _Stage.error);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: const EdgeInsets.all(24),
            child: switch (_stage) {
              _Stage.loading => const Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    CircularProgressIndicator.adaptive(),
                    SizedBox(height: 16),
                    Text('Verifying…'),
                  ],
                ),
              _Stage.error => _ErrorView(
                  title: l10n.verifyFailed,
                  subtitle: l10n.verifyExpiredOrUsed,
                  retryLabel: l10n.verifyRetry,
                  loginLabel: l10n.loginTitle,
                  onRetry: _verify,
                  onLogin: () => context.go('/login'),
                ),
            },
          ),
        ),
      ),
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({
    required this.title,
    required this.subtitle,
    required this.retryLabel,
    required this.loginLabel,
    required this.onRetry,
    required this.onLogin,
  });

  final String title;
  final String subtitle;
  final String retryLabel;
  final String loginLabel;
  final VoidCallback onRetry;
  final VoidCallback onLogin;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.error_outline, size: 56, color: theme.colorScheme.error),
        const SizedBox(height: 12),
        Text(title, style: theme.textTheme.titleLarge, textAlign: TextAlign.center),
        const SizedBox(height: 8),
        Text(
          subtitle,
          textAlign: TextAlign.center,
          style: theme.textTheme.bodyMedium,
        ),
        const SizedBox(height: 24),
        ElevatedButton(onPressed: onRetry, child: Text(retryLabel)),
        const SizedBox(height: 8),
        TextButton(onPressed: onLogin, child: Text(loginLabel)),
      ],
    );
  }
}

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/l10n/app_localizations.dart';
import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';

/// Screen shown when a magic link deep link is opened.
/// Automatically verifies the token and navigates to home on success.
class VerifyScreen extends ConsumerStatefulWidget {
  final String token;

  const VerifyScreen({super.key, required this.token});

  @override
  ConsumerState<VerifyScreen> createState() => _VerifyScreenState();
}

class _VerifyScreenState extends ConsumerState<VerifyScreen> {
  bool _error = false;

  @override
  void initState() {
    super.initState();
    _verify();
  }

  Future<void> _verify() async {
    if (widget.token.isEmpty) {
      setState(() => _error = true);
      return;
    }

    await ref.read(authProvider.notifier).verifyMagicLink(widget.token);

    final state = ref.read(authProvider);
    if (state is AuthAuthenticated && mounted) {
      context.go('/home');
    } else if (mounted) {
      setState(() => _error = true);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    if (_error) {
      return Scaffold(
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.error_outline, size: 48, color: theme.colorScheme.error),
                const SizedBox(height: 16),
                Text(
                  'Verification failed',
                  style: theme.textTheme.headlineSmall,
                ),
                const SizedBox(height: 8),
                Text(
                  'The link may have expired or already been used.',
                  style: TextStyle(color: theme.colorScheme.onSurfaceVariant),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 32),
                FilledButton(
                  onPressed: () => context.go('/login'),
                  child: Text(l10n.loginTitle),
                ),
              ],
            ),
          ),
        ),
      );
    }

    return const Scaffold(
      body: Center(child: CircularProgressIndicator()),
    );
  }
}

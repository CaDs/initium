import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import '../shared/dev_mode_banner.dart';
import '../shared/locale_switcher.dart';
import '../shared/theme_switcher.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;

    if (authState is! AuthAuthenticated) {
      return const Scaffold(
        body: Center(
          child: CircularProgressIndicator.adaptive(),
        ),
      );
    }

    final user = authState.user;

    final rows = <_ProfileRow>[
      _ProfileRow(label: l10n.labelEmail, value: user.email),
      _ProfileRow(
        label: l10n.labelName,
        value: user.name.isNotEmpty ? user.name : '—',
      ),
      _ProfileRow(label: l10n.labelRole, value: user.role),
      _ProfileRow(label: l10n.labelUserId, value: user.id, mono: true),
    ];

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.appName),
        actions: const [LocaleSwitcher(), ThemeSwitcher()],
      ),
      body: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          const DevModeBanner(),
          Expanded(
            child: Center(
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 640),
                child: ListView(
                  padding: const EdgeInsets.all(16),
                  children: [
                    Card(
                      margin: EdgeInsets.zero,
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 16,
                          vertical: 8,
                        ),
                        child: Column(
                          children: [
                            for (var i = 0; i < rows.length; i++) ...[
                              if (i > 0) const Divider(height: 1),
                              rows[i],
                            ],
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: 24),
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: OutlinedButton.icon(
                        onPressed: () =>
                            ref.read(authProvider.notifier).logout(),
                        icon: const Icon(Icons.logout),
                        label: Text(l10n.logout),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _ProfileRow extends StatelessWidget {
  const _ProfileRow({
    required this.label,
    required this.value,
    this.mono = false,
  });

  final String label;
  final String value;
  final bool mono;

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final valueStyle = theme.textTheme.bodyMedium?.copyWith(
      fontFamily: mono ? 'monospace' : null,
      fontSize: mono ? 12 : null,
    );

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 12),
      child: Semantics(
        label: '$label: $value',
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            SizedBox(
              width: 88,
              child: Text(
                label,
                style: theme.textTheme.labelMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant,
                ),
              ),
            ),
            Expanded(
              child: Text(value, style: valueStyle),
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:mobile/l10n/app_localizations.dart';
import '../../providers/api_provider.dart';

class DevModeBanner extends StatelessWidget {
  const DevModeBanner({super.key});

  @override
  Widget build(BuildContext context) {
    if (!isDevBypassAuth) return const SizedBox.shrink();

    final l10n = AppLocalizations.of(context)!;
    final theme = Theme.of(context);

    return Semantics(
      label: l10n.devBanner,
      child: Container(
        width: double.infinity,
        color: theme.colorScheme.tertiaryContainer,
        padding: const EdgeInsets.symmetric(vertical: 6),
        child: Text(
          l10n.devBanner,
          textAlign: TextAlign.center,
          style: TextStyle(
            fontSize: 12,
            color: theme.colorScheme.onTertiaryContainer,
          ),
        ),
      ),
    );
  }
}

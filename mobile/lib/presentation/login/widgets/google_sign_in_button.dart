import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';

class GoogleSignInButton extends ConsumerWidget {
  const GoogleSignInButton({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;

    return Semantics(
      button: true,
      label: l10n.loginGoogle,
      child: OutlinedButton(
        onPressed: () async {
          // TODO: Integrate google_sign_in package
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(l10n.loginGoogleSetup)),
          );
        },
        style: OutlinedButton.styleFrom(
          minimumSize: const Size(double.infinity, 52),
          side: BorderSide(color: Theme.of(context).colorScheme.outline),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.g_mobiledata, size: 24),
            const SizedBox(width: 8),
            Text(l10n.loginGoogle, style: const TextStyle(fontSize: 16)),
          ],
        ),
      ),
    );
  }
}

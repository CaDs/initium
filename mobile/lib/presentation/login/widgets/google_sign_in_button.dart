import 'dart:async';
import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:mobile/l10n/app_localizations.dart';
import '../../../providers/api_provider.dart';

/// iOS Client ID from GoogleService-Info.plist.
/// Set via --dart-define=GOOGLE_IOS_CLIENT_ID=xxx
const _iosClientId = String.fromEnvironment('GOOGLE_IOS_CLIENT_ID');

/// Web/server Client ID — needed to get an idToken for backend verification.
/// Set via --dart-define=GOOGLE_SERVER_CLIENT_ID=xxx
const _serverClientId = String.fromEnvironment('GOOGLE_SERVER_CLIENT_ID');

/// Google Sign-In button.
/// Requires GOOGLE_IOS_CLIENT_ID and GOOGLE_SERVER_CLIENT_ID via --dart-define,
/// or GoogleService-Info.plist (iOS) / google-services.json (Android).
/// See mobile/SETUP.md for configuration instructions.
class GoogleSignInButton extends ConsumerWidget {
  const GoogleSignInButton({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;

    return Semantics(
      button: true,
      label: l10n.loginGoogle,
      child: OutlinedButton(
        onPressed: () => _signIn(context, ref),
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

  Future<void> _signIn(BuildContext context, WidgetRef ref) async {
    final l10n = AppLocalizations.of(context)!;

    // Check if client IDs are configured
    final clientId = Platform.isIOS ? _iosClientId : null;
    if (Platform.isIOS && _iosClientId.isEmpty) {
      _showSetup(context, l10n);
      return;
    }

    try {
      final googleSignIn = GoogleSignIn(
        scopes: ['email', 'profile'],
        clientId: clientId,
        serverClientId: _serverClientId.isNotEmpty ? _serverClientId : null,
      );

      final account = await googleSignIn.signIn();
      if (account == null) return; // User cancelled

      final auth = await account.authentication;
      final idToken = auth.idToken;

      if (idToken == null) {
        if (context.mounted) {
          _showError(context, 'Failed to get ID token. Check GOOGLE_SERVER_CLIENT_ID.');
        }
        return;
      }

      await ref.read(authProvider.notifier).loginWithGoogle(idToken);
    } on PlatformException catch (e) {
      debugPrint('Google Sign-In PlatformException: ${e.code} ${e.message}');
      if (context.mounted) _showSetup(context, l10n);
    } on TimeoutException {
      if (context.mounted) _showSetup(context, l10n);
    } catch (e) {
      debugPrint('Google Sign-In error: $e');
      if (context.mounted) _showSetup(context, l10n);
    }
  }

  void _showSetup(BuildContext context, AppLocalizations l10n) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(l10n.loginGoogleSetup),
        duration: const Duration(seconds: 5),
      ),
    );
  }

  void _showError(BuildContext context, String message) {
    if (!context.mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message)),
    );
  }
}

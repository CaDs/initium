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

class GoogleSignInButton extends ConsumerWidget {
  const GoogleSignInButton({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    return ElevatedButton.icon(
      onPressed: () => _signIn(context, ref),
      icon: const Icon(Icons.login),
      label: Text(l10n.loginGoogle),
    );
  }

  Future<void> _signIn(BuildContext context, WidgetRef ref) async {
    final l10n = AppLocalizations.of(context)!;

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
      if (account == null) return;

      final auth = await account.authentication;
      final idToken = auth.idToken;

      if (idToken == null) {
        if (context.mounted) {
          _showError(context,
              'Failed to get ID token. Check GOOGLE_SERVER_CLIENT_ID.');
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
    ScaffoldMessenger.of(context)
        .showSnackBar(SnackBar(content: Text(message)));
  }
}

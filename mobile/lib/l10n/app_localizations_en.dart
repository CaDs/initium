// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appName => 'Initium';

  @override
  String get loginTitle => 'Sign In';

  @override
  String get loginGoogle => 'Sign in with Google';

  @override
  String get loginGoogleSetup =>
      'Google Sign-In: configure google-services.json first';

  @override
  String get loginMagicPlaceholder => 'Enter your email';

  @override
  String get loginMagicSubmit => 'Send Magic Link';

  @override
  String get loginMagicSending => 'Sending...';

  @override
  String get loginMagicSent => 'Check your email!';

  @override
  String get labelEmail => 'Email';

  @override
  String get labelName => 'Name';

  @override
  String get labelRole => 'Role';

  @override
  String get labelUserId => 'User ID';

  @override
  String get devBanner => 'Dev Mode — Logged in as dev@initium.local';

  @override
  String get logout => 'Logout';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get themeSystem => 'System';

  @override
  String get authGoogleLoginFailed => 'Google login failed';

  @override
  String get authMagicLinkFailed => 'Magic link verification failed';

  @override
  String get verifyFailed => 'Verification failed';

  @override
  String get verifyExpiredOrUsed =>
      'The link may have expired or already been used.';

  @override
  String get verifyRetry => 'Try again';
}

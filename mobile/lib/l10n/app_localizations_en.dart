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
  String get landingTitle => 'Welcome to Initium';

  @override
  String get landingSubtitle => 'Your next great idea starts here.';

  @override
  String get landingCta => 'Get Started';

  @override
  String get loginTitle => 'Sign In';

  @override
  String get loginSubtitle => 'No passwords needed.';

  @override
  String get loginDivider => 'or';

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
  String get loginMagicSentDetail => 'A magic link has been sent.';

  @override
  String get homeWelcome => 'Welcome back!';

  @override
  String homeWelcomeUser(String name) {
    return 'Welcome back, $name!';
  }

  @override
  String get homeSubtitle => 'This is your authenticated home screen.';

  @override
  String get homeProfile => 'Your Profile';

  @override
  String get labelEmail => 'Email';

  @override
  String get labelName => 'Name';

  @override
  String get labelUserId => 'User ID';

  @override
  String get devBanner => 'Dev Mode — Logged in as dev@initium.local';

  @override
  String get logout => 'Logout';

  @override
  String get dashboard => 'Dashboard';

  @override
  String get themeLight => 'Light';

  @override
  String get themeDark => 'Dark';

  @override
  String get themeSystem => 'System';
}

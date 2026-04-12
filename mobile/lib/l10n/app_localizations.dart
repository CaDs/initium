import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:intl/intl.dart' as intl;

import 'app_localizations_en.dart';
import 'app_localizations_es.dart';
import 'app_localizations_ja.dart';

// ignore_for_file: type=lint

/// Callers can lookup localized strings with an instance of AppLocalizations
/// returned by `AppLocalizations.of(context)`.
///
/// Applications need to include `AppLocalizations.delegate()` in their app's
/// `localizationDelegates` list, and the locales they support in the app's
/// `supportedLocales` list. For example:
///
/// ```dart
/// import 'l10n/app_localizations.dart';
///
/// return MaterialApp(
///   localizationsDelegates: AppLocalizations.localizationsDelegates,
///   supportedLocales: AppLocalizations.supportedLocales,
///   home: MyApplicationHome(),
/// );
/// ```
///
/// ## Update pubspec.yaml
///
/// Please make sure to update your pubspec.yaml to include the following
/// packages:
///
/// ```yaml
/// dependencies:
///   # Internationalization support.
///   flutter_localizations:
///     sdk: flutter
///   intl: any # Use the pinned version from flutter_localizations
///
///   # Rest of dependencies
/// ```
///
/// ## iOS Applications
///
/// iOS applications define key application metadata, including supported
/// locales, in an Info.plist file that is built into the application bundle.
/// To configure the locales supported by your app, you’ll need to edit this
/// file.
///
/// First, open your project’s ios/Runner.xcworkspace Xcode workspace file.
/// Then, in the Project Navigator, open the Info.plist file under the Runner
/// project’s Runner folder.
///
/// Next, select the Information Property List item, select Add Item from the
/// Editor menu, then select Localizations from the pop-up menu.
///
/// Select and expand the newly-created Localizations item then, for each
/// locale your application supports, add a new item and select the locale
/// you wish to add from the pop-up menu in the Value field. This list should
/// be consistent with the languages listed in the AppLocalizations.supportedLocales
/// property.
abstract class AppLocalizations {
  AppLocalizations(String locale)
    : localeName = intl.Intl.canonicalizedLocale(locale.toString());

  final String localeName;

  static AppLocalizations? of(BuildContext context) {
    return Localizations.of<AppLocalizations>(context, AppLocalizations);
  }

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  /// A list of this localizations delegate along with the default localizations
  /// delegates.
  ///
  /// Returns a list of localizations delegates containing this delegate along with
  /// GlobalMaterialLocalizations.delegate, GlobalCupertinoLocalizations.delegate,
  /// and GlobalWidgetsLocalizations.delegate.
  ///
  /// Additional delegates can be added by appending to this list in
  /// MaterialApp. This list does not have to be used at all if a custom list
  /// of delegates is preferred or required.
  static const List<LocalizationsDelegate<dynamic>> localizationsDelegates =
      <LocalizationsDelegate<dynamic>>[
        delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
      ];

  /// A list of this localizations delegate's supported locales.
  static const List<Locale> supportedLocales = <Locale>[
    Locale('en'),
    Locale('es'),
    Locale('ja'),
  ];

  /// No description provided for @appName.
  ///
  /// In en, this message translates to:
  /// **'Initium'**
  String get appName;

  /// No description provided for @landingTitle.
  ///
  /// In en, this message translates to:
  /// **'Welcome to Initium'**
  String get landingTitle;

  /// No description provided for @landingSubtitle.
  ///
  /// In en, this message translates to:
  /// **'Your next great idea starts here.'**
  String get landingSubtitle;

  /// No description provided for @landingCta.
  ///
  /// In en, this message translates to:
  /// **'Get Started'**
  String get landingCta;

  /// No description provided for @loginTitle.
  ///
  /// In en, this message translates to:
  /// **'Sign In'**
  String get loginTitle;

  /// No description provided for @loginSubtitle.
  ///
  /// In en, this message translates to:
  /// **'No passwords needed.'**
  String get loginSubtitle;

  /// No description provided for @loginDivider.
  ///
  /// In en, this message translates to:
  /// **'or'**
  String get loginDivider;

  /// No description provided for @loginGoogle.
  ///
  /// In en, this message translates to:
  /// **'Sign in with Google'**
  String get loginGoogle;

  /// No description provided for @loginGoogleSetup.
  ///
  /// In en, this message translates to:
  /// **'Google Sign-In: configure google-services.json first'**
  String get loginGoogleSetup;

  /// No description provided for @loginMagicPlaceholder.
  ///
  /// In en, this message translates to:
  /// **'Enter your email'**
  String get loginMagicPlaceholder;

  /// No description provided for @loginMagicSubmit.
  ///
  /// In en, this message translates to:
  /// **'Send Magic Link'**
  String get loginMagicSubmit;

  /// No description provided for @loginMagicSending.
  ///
  /// In en, this message translates to:
  /// **'Sending...'**
  String get loginMagicSending;

  /// No description provided for @loginMagicSent.
  ///
  /// In en, this message translates to:
  /// **'Check your email!'**
  String get loginMagicSent;

  /// No description provided for @loginMagicSentDetail.
  ///
  /// In en, this message translates to:
  /// **'A magic link has been sent.'**
  String get loginMagicSentDetail;

  /// No description provided for @homeWelcome.
  ///
  /// In en, this message translates to:
  /// **'Welcome back!'**
  String get homeWelcome;

  /// No description provided for @homeWelcomeUser.
  ///
  /// In en, this message translates to:
  /// **'Welcome back, {name}!'**
  String homeWelcomeUser(String name);

  /// No description provided for @homeSubtitle.
  ///
  /// In en, this message translates to:
  /// **'This is your authenticated home screen.'**
  String get homeSubtitle;

  /// No description provided for @homeProfile.
  ///
  /// In en, this message translates to:
  /// **'Your Profile'**
  String get homeProfile;

  /// No description provided for @labelEmail.
  ///
  /// In en, this message translates to:
  /// **'Email'**
  String get labelEmail;

  /// No description provided for @labelName.
  ///
  /// In en, this message translates to:
  /// **'Name'**
  String get labelName;

  /// No description provided for @labelUserId.
  ///
  /// In en, this message translates to:
  /// **'User ID'**
  String get labelUserId;

  /// No description provided for @devBanner.
  ///
  /// In en, this message translates to:
  /// **'Dev Mode — Logged in as dev@initium.local'**
  String get devBanner;

  /// No description provided for @logout.
  ///
  /// In en, this message translates to:
  /// **'Logout'**
  String get logout;

  /// No description provided for @dashboard.
  ///
  /// In en, this message translates to:
  /// **'Dashboard'**
  String get dashboard;

  /// No description provided for @themeLight.
  ///
  /// In en, this message translates to:
  /// **'Light'**
  String get themeLight;

  /// No description provided for @themeDark.
  ///
  /// In en, this message translates to:
  /// **'Dark'**
  String get themeDark;

  /// No description provided for @themeSystem.
  ///
  /// In en, this message translates to:
  /// **'System'**
  String get themeSystem;

  /// No description provided for @authGoogleLoginFailed.
  ///
  /// In en, this message translates to:
  /// **'Google login failed'**
  String get authGoogleLoginFailed;

  /// No description provided for @authMagicLinkFailed.
  ///
  /// In en, this message translates to:
  /// **'Magic link verification failed'**
  String get authMagicLinkFailed;

  /// No description provided for @verifyFailed.
  ///
  /// In en, this message translates to:
  /// **'Verification failed'**
  String get verifyFailed;

  /// No description provided for @verifyExpiredOrUsed.
  ///
  /// In en, this message translates to:
  /// **'The link may have expired or already been used.'**
  String get verifyExpiredOrUsed;
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  Future<AppLocalizations> load(Locale locale) {
    return SynchronousFuture<AppLocalizations>(lookupAppLocalizations(locale));
  }

  @override
  bool isSupported(Locale locale) =>
      <String>['en', 'es', 'ja'].contains(locale.languageCode);

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

AppLocalizations lookupAppLocalizations(Locale locale) {
  // Lookup logic when only language code is specified.
  switch (locale.languageCode) {
    case 'en':
      return AppLocalizationsEn();
    case 'es':
      return AppLocalizationsEs();
    case 'ja':
      return AppLocalizationsJa();
  }

  throw FlutterError(
    'AppLocalizations.delegate failed to load unsupported locale "$locale". This is likely '
    'an issue with the localizations generation tool. Please file an issue '
    'on GitHub with a reproducible sample app and the gen-l10n configuration '
    'that was used.',
  );
}

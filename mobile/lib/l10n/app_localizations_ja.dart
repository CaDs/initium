// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Japanese (`ja`).
class AppLocalizationsJa extends AppLocalizations {
  AppLocalizationsJa([String locale = 'ja']) : super(locale);

  @override
  String get appName => 'Initium';

  @override
  String get landingTitle => 'Initiumへようこそ';

  @override
  String get landingSubtitle => '次の素晴らしいアイデアはここから始まります。';

  @override
  String get landingCta => 'はじめる';

  @override
  String get loginTitle => 'ログイン';

  @override
  String get loginSubtitle => 'パスワード不要。';

  @override
  String get loginDivider => 'または';

  @override
  String get loginGoogle => 'Googleでログイン';

  @override
  String get loginGoogleSetup =>
      'Google Sign-In: google-services.jsonを設定してください';

  @override
  String get loginMagicPlaceholder => 'メールアドレスを入力';

  @override
  String get loginMagicSubmit => 'マジックリンクを送信';

  @override
  String get loginMagicSending => '送信中...';

  @override
  String get loginMagicSent => 'メールを確認してください！';

  @override
  String get loginMagicSentDetail => 'マジックリンクが送信されました。';

  @override
  String get homeWelcome => 'おかえりなさい！';

  @override
  String homeWelcomeUser(String name) {
    return 'おかえりなさい、$nameさん！';
  }

  @override
  String get homeSubtitle => '認証済みのホーム画面です。';

  @override
  String get homeProfile => 'プロフィール';

  @override
  String get labelEmail => 'メール';

  @override
  String get labelName => '名前';

  @override
  String get labelUserId => 'ユーザーID';

  @override
  String get devBanner => '開発モード — dev@initium.localとしてログイン中';

  @override
  String get logout => 'ログアウト';

  @override
  String get dashboard => 'ダッシュボード';

  @override
  String get themeLight => 'ライト';

  @override
  String get themeDark => 'ダーク';

  @override
  String get themeSystem => 'システム';
}

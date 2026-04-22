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
  String get loginTitle => 'ログイン';

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
  String get labelEmail => 'メール';

  @override
  String get labelName => '名前';

  @override
  String get labelRole => '役割';

  @override
  String get labelUserId => 'ユーザーID';

  @override
  String get devBanner => '開発モード — dev@initium.localとしてログイン中';

  @override
  String get logout => 'ログアウト';

  @override
  String get themeLight => 'ライト';

  @override
  String get themeDark => 'ダーク';

  @override
  String get themeSystem => 'システム';

  @override
  String get authGoogleLoginFailed => 'Googleログインに失敗しました';

  @override
  String get authMagicLinkFailed => 'マジックリンクの確認に失敗しました';

  @override
  String get verifyFailed => '確認に失敗しました';

  @override
  String get verifyExpiredOrUsed => 'リンクの有効期限が切れたか、既に使用された可能性があります。';

  @override
  String get verifyRetry => '再試行';
}

// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Spanish Castilian (`es`).
class AppLocalizationsEs extends AppLocalizations {
  AppLocalizationsEs([String locale = 'es']) : super(locale);

  @override
  String get appName => 'Initium';

  @override
  String get loginTitle => 'Iniciar sesión';

  @override
  String get loginGoogle => 'Iniciar sesión con Google';

  @override
  String get loginGoogleSetup =>
      'Google Sign-In: configura google-services.json primero';

  @override
  String get loginMagicPlaceholder => 'Ingresa tu correo electrónico';

  @override
  String get loginMagicSubmit => 'Enviar enlace mágico';

  @override
  String get loginMagicSending => 'Enviando...';

  @override
  String get loginMagicSent => '¡Revisa tu correo!';

  @override
  String get labelEmail => 'Correo';

  @override
  String get labelName => 'Nombre';

  @override
  String get labelRole => 'Rol';

  @override
  String get labelUserId => 'ID de Usuario';

  @override
  String get devBanner => 'Modo Dev — Sesión como dev@initium.local';

  @override
  String get logout => 'Cerrar sesión';

  @override
  String get themeLight => 'Claro';

  @override
  String get themeDark => 'Oscuro';

  @override
  String get themeSystem => 'Sistema';

  @override
  String get authGoogleLoginFailed => 'Error al iniciar sesión con Google';

  @override
  String get authMagicLinkFailed => 'Error al verificar el enlace mágico';

  @override
  String get verifyFailed => 'Verificación fallida';

  @override
  String get verifyExpiredOrUsed =>
      'El enlace puede haber expirado o ya fue utilizado.';

  @override
  String get verifyRetry => 'Reintentar';
}

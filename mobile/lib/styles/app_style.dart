// Wonderize: AppStyle token system.
//
// Adapted from Wonderous lib/styles/styles.dart. Single source of truth for
// design tokens. Access via the global `$styles` getter exposed in
// `lib/ui/app_scaffold.dart`. Extended with a `brightness` parameter so the
// palette flips with the existing `themeProvider`.

import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

import 'app_colors.dart';

@immutable
class AppStyle {
  AppStyle({
    Size? screenSize,
    Brightness brightness = Brightness.light,
    this.disableAnimations = false,
    this.highContrast = false,
  }) : colors = AppColors(brightness: brightness) {
    if (screenSize == null) {
      scale = 1;
    } else {
      final shortest = screenSize.shortestSide;
      if (shortest > 1000) {
        scale = 1.2;
      } else if (shortest > 800) {
        scale = 1.1;
      } else {
        scale = 1.0;
      }
    }
  }

  late final double scale;
  final bool disableAnimations;
  final bool highContrast;

  final AppColors colors;
  late final Corners corners = Corners();
  late final Shadows shadows = Shadows();
  late final Insets insets = Insets(scale);
  late final AppTextStyles text = AppTextStyles(scale);
  late final AppTimes times = AppTimes(disableAnimations);
  late final Sizes sizes = Sizes();
}

/// Animation timing tiers. Use these everywhere instead of literal Durations.
/// `disableAnimations` collapses every duration to zero for accessibility.
@immutable
class AppTimes {
  AppTimes(this.disableAnimations);
  final bool disableAnimations;

  Duration _ms(int ms) =>
      disableAnimations ? Duration.zero : Duration(milliseconds: ms);

  late final Duration pageTransition = _ms(200);
  late final Duration fast = _ms(300);
  late final Duration med = _ms(600);
  late final Duration slow = _ms(900);
  late final Duration extraSlow = _ms(1300);
}

@immutable
class Corners {
  final double sm = 4;
  final double md = 8;
  final double lg = 32;
}

class Sizes {
  final double maxContentWidth1 = 800;
  final double maxContentWidth2 = 600;
  final double maxContentWidth3 = 500;
  final Size minAppSize = const Size(380, 650);
}

@immutable
class Insets {
  Insets(this._scale);
  final double _scale;

  late final double xxs = 4 * _scale;
  late final double xs = 8 * _scale;
  late final double sm = 16 * _scale;
  late final double md = 24 * _scale;
  late final double lg = 32 * _scale;
  late final double xl = 48 * _scale;
  late final double xxl = 56 * _scale;
  late final double offset = 80 * _scale;
}

@immutable
class Shadows {
  final List<Shadow> textSoft = [
    Shadow(
      color: Colors.black.withValues(alpha: .25),
      offset: const Offset(0, 2),
      blurRadius: 4,
    ),
  ];
  final List<Shadow> text = [
    Shadow(
      color: Colors.black.withValues(alpha: .6),
      offset: const Offset(0, 2),
      blurRadius: 2,
    ),
  ];
  final List<Shadow> textStrong = [
    Shadow(
      color: Colors.black.withValues(alpha: .6),
      offset: const Offset(0, 4),
      blurRadius: 6,
    ),
  ];
}

/// Typography. Uses google_fonts to fetch Fraunces (display serif) and Inter
/// (body sans) at runtime — no TTF assets required. Japanese falls back to the
/// platform system font automatically for characters Fraunces/Inter lack.
@immutable
class AppTextStyles {
  AppTextStyles(this._scale);
  final double _scale;

  TextStyle get titleFont => GoogleFonts.fraunces();
  TextStyle get wonderTitleFont =>
      GoogleFonts.fraunces(fontWeight: FontWeight.w500);
  TextStyle get quoteFont => GoogleFonts.fraunces(fontStyle: FontStyle.italic);
  TextStyle get contentFont =>
      GoogleFonts.inter(fontFeatures: const [FontFeature.enable('kern')]);
  TextStyle get monoFont => GoogleFonts.robotoMono();

  late final TextStyle dropCase = _make(quoteFont, sizePx: 56, heightPx: 20);
  late final TextStyle wonderTitle =
      _make(wonderTitleFont, sizePx: 48, heightPx: 52);

  late final TextStyle h1 = _make(titleFont, sizePx: 48, heightPx: 56);
  late final TextStyle h2 = _make(titleFont, sizePx: 32, heightPx: 40);
  late final TextStyle h3 =
      _make(titleFont, sizePx: 24, heightPx: 32, weight: FontWeight.w600);
  late final TextStyle h4 = _make(contentFont,
      sizePx: 14, heightPx: 23, spacingPc: 5, weight: FontWeight.w600);

  late final TextStyle title1 =
      _make(titleFont, sizePx: 16, heightPx: 26, spacingPc: 5);
  late final TextStyle title2 = _make(titleFont, sizePx: 14, heightPx: 18);

  late final TextStyle body = _make(contentFont, sizePx: 16, heightPx: 26);
  late final TextStyle bodyBold = _make(contentFont,
      sizePx: 16, heightPx: 26, weight: FontWeight.w600);
  late final TextStyle bodySmall =
      _make(contentFont, sizePx: 14, heightPx: 23);
  late final TextStyle bodySmallBold = _make(contentFont,
      sizePx: 14, heightPx: 23, weight: FontWeight.w600);

  late final TextStyle quote1 = _make(quoteFont,
      sizePx: 28, heightPx: 36, weight: FontWeight.w600, spacingPc: -3);
  late final TextStyle quote2 = _make(quoteFont, sizePx: 20, heightPx: 30);

  late final TextStyle caption =
      _make(contentFont, sizePx: 13, heightPx: 18, weight: FontWeight.w500)
          .copyWith(fontStyle: FontStyle.italic);
  late final TextStyle callout =
      _make(contentFont, sizePx: 16, heightPx: 26, weight: FontWeight.w600)
          .copyWith(fontStyle: FontStyle.italic);
  late final TextStyle btn = _make(contentFont,
      sizePx: 14, heightPx: 14, spacingPc: 2, weight: FontWeight.w600);

  TextStyle _make(
    TextStyle base, {
    required double sizePx,
    double? heightPx,
    double? spacingPc,
    FontWeight? weight,
  }) {
    final scaledSize = sizePx * _scale;
    final scaledHeight = heightPx == null ? null : heightPx * _scale;
    return base.copyWith(
      fontSize: scaledSize,
      height: scaledHeight == null ? base.height : scaledHeight / scaledSize,
      letterSpacing:
          spacingPc == null ? base.letterSpacing : scaledSize * spacingPc * 0.01,
      fontWeight: weight,
    );
  }
}

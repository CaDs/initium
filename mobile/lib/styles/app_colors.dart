// Wonderize: Initium color palette + ThemeData adapter.
//
// Custom palette derived from the stone-grey seed (#292524), tuned warm for an
// editorial feel. Adapted from Wonderous lib/styles/colors.dart but extended to
// support light/dark brightness instead of the original's fixed-polarity palette.

import 'package:flutter/material.dart';

class AppColors {
  AppColors({Brightness brightness = Brightness.light})
      : isDark = brightness == Brightness.dark;

  final bool isDark;

  // --- Accent ramp (amber-bronze, warm editorial) ---
  final Color accent1 = const Color(0xFFC98C3F);
  final Color accent2 = const Color(0xFF6B5E52);
  final Color accent3 = const Color(0xFF8B4513);

  // --- Surface & text ---
  // `offWhite` and `black` are the canonical bg/fg slots; which one is on top
  // depends on brightness.
  final Color offWhite = const Color(0xFFF7F3EC);
  final Color black = const Color(0xFF1A1917);

  final Color caption = const Color(0xFF78716C);
  final Color body = const Color(0xFF44403B);
  final Color greyStrong = const Color(0xFF292524);
  final Color greyMedium = const Color(0xFF78716C);
  final Color greySoft = const Color(0xFFD6D3D1);
  final Color white = Colors.white;

  /// Returns the current-theme foreground (text) color.
  Color get fg => isDark ? offWhite : black;

  /// Returns the current-theme background (surface) color.
  Color get bg => isDark ? black : offWhite;

  /// Shift a color in HSL space. Inverted in dark mode so "lighten by 10" still
  /// means "more emphasis" regardless of theme polarity.
  Color shift(Color c, double d) {
    final hsl = HSLColor.fromColor(c);
    final adjusted = hsl.withLightness(
      (hsl.lightness + d * (isDark ? -1 : 1)).clamp(0.0, 1.0),
    );
    return adjusted.toColor();
  }

  ThemeData toThemeData() {
    final base = isDark ? ThemeData.dark() : ThemeData.light();
    final surface = isDark ? const Color(0xFF201E1C) : offWhite;
    final onSurface = isDark ? offWhite : black;

    final colorScheme = ColorScheme(
      brightness: isDark ? Brightness.dark : Brightness.light,
      primary: accent1,
      onPrimary: Colors.white,
      primaryContainer: accent1.withValues(alpha: 0.15),
      onPrimaryContainer: onSurface,
      secondary: accent2,
      onSecondary: Colors.white,
      secondaryContainer: accent2.withValues(alpha: 0.15),
      onSecondaryContainer: onSurface,
      tertiary: accent3,
      onTertiary: Colors.white,
      surface: surface,
      onSurface: onSurface,
      onSurfaceVariant: isDark ? greySoft : caption,
      outline: isDark ? greyMedium : greySoft,
      outlineVariant: isDark ? greyStrong : greySoft,
      error: Colors.red.shade400,
      onError: Colors.white,
    );

    return ThemeData.from(
      textTheme: base.textTheme,
      colorScheme: colorScheme,
      useMaterial3: true,
    ).copyWith(
      scaffoldBackgroundColor: surface,
      textSelectionTheme: TextSelectionThemeData(cursorColor: accent1),
      highlightColor: accent1.withValues(alpha: 0.12),
    );
  }
}

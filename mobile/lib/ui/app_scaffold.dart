// Wonderize: $styles publisher.
//
// Adapted from Wonderous lib/ui/app_scaffold.dart, modified to coexist with
// Riverpod-driven theming in `main.dart`. Mount this once via
// `MaterialApp.router`'s `builder:` callback so every descendant can read
// `$styles.colors.accent1`, `$styles.times.fast`, etc.
//
// Unlike the upstream template, this scaffold does NOT wrap its child in a
// `Theme(...)` override — `MaterialApp` already owns `ThemeData` (both light
// and dark variants are derived from `AppColors(...).toThemeData()`), so doing
// it here would cause a double-theme. We only inject a DefaultTextStyle so
// Hero flights render with consistent typography during transitions.

import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../styles/app_style.dart';

/// Global accessor for the current AppStyle. Set by [AppScaffold] on every build.
AppStyle get $styles => AppScaffold.style;

class AppScaffold extends StatelessWidget {
  const AppScaffold({super.key, required this.child});
  final Widget child;

  static AppStyle get style => _style;
  static AppStyle _style = AppStyle();

  @override
  Widget build(BuildContext context) {
    final mq = MediaQuery.of(context);
    final brightness = Theme.of(context).brightness;

    _style = AppStyle(
      screenSize: mq.size,
      brightness: brightness,
      disableAnimations: mq.disableAnimations,
      highContrast: mq.highContrast,
    );

    Animate.defaultDuration = _style.times.fast;

    return KeyedSubtree(
      key: ValueKey('${_style.scale}-${brightness.index}'),
      child: DefaultTextStyle(
        style: _style.text.body.copyWith(color: _style.colors.fg),
        child: child,
      ),
    );
  }
}

// Wonderize: blur backdrop with opaque fallback.
//
// Adapted from Wonderous lib/ui/common/app_backdrop.dart. `BackdropFilter` is
// expensive on older Android devices — when blurs are disabled we substitute a
// semi-opaque fill so the visual hierarchy stays the same.

import 'dart:ui';

import 'package:flutter/material.dart';

import '../ui/app_scaffold.dart';

class AppBackdrop extends StatelessWidget {
  const AppBackdrop({
    super.key,
    this.strength = 1,
    this.useBlurs = true,
    this.fillColor,
    this.child,
  });

  /// 0..1, scales the blur sigma (max 15) and the fallback fill alpha.
  final double strength;

  /// When false, render the opaque fallback instead of a real blur.
  final bool useBlurs;

  /// Override the fallback fill. Defaults to 80% bg scaled by strength.
  final Color? fillColor;

  final Widget? child;

  @override
  Widget build(BuildContext context) {
    final s = strength.clamp(0.0, 1.0);

    if (useBlurs && !$styles.disableAnimations) {
      return BackdropFilter(
        filter: ImageFilter.blur(sigmaX: s * 15, sigmaY: s * 15),
        child: child ?? const SizedBox.expand(),
      );
    }

    final fill = Container(
      color: fillColor ?? $styles.colors.bg.withValues(alpha: 0.8 * s),
    );
    if (child == null) return fill;
    return Stack(children: [child!, Positioned.fill(child: fill)]);
  }
}

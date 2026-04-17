// Wonderize: animated section divider with central rotating glyph.
//
// Adapted from Wonderous lib/ui/common/compass_divider.dart. The center glyph
// rotates 180° as the divider expands outward from the middle.

import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:gap/gap.dart';

import '../app_scaffold.dart';

class CompassDivider extends StatelessWidget {
  const CompassDivider({
    super.key,
    required this.isExpanded,
    this.duration,
    this.linesColor,
    this.centerColor,
    this.centerSvgAsset,
    this.centerWidget,
    this.centerSize = 32,
  }) : assert(centerSvgAsset != null || centerWidget != null,
            'Provide either centerSvgAsset OR centerWidget');

  final bool isExpanded;
  final Duration? duration;
  final Color? linesColor;
  final Color? centerColor;
  final String? centerSvgAsset;
  final Widget? centerWidget;
  final double centerSize;

  @override
  Widget build(BuildContext context) {
    final dur = duration ?? const Duration(milliseconds: 1500);
    final lineColor = linesColor ?? $styles.colors.accent2;
    final glyphColor = centerColor ?? $styles.colors.accent2;

    Widget buildHzLine({required bool alignLeft}) {
      return TweenAnimationBuilder<double>(
        duration: $styles.disableAnimations ? Duration.zero : dur,
        tween: Tween(begin: 0, end: isExpanded ? 1 : 0),
        curve: Curves.easeOut,
        child: Divider(height: 1, thickness: 0.5, color: lineColor),
        builder: (_, value, child) => Transform.scale(
          scaleX: value,
          alignment: alignLeft ? Alignment.centerLeft : Alignment.centerRight,
          child: child,
        ),
      );
    }

    Widget center;
    if (centerSvgAsset != null) {
      center = SvgPicture.asset(
        centerSvgAsset!,
        colorFilter: ColorFilter.mode(glyphColor, BlendMode.srcIn),
      );
    } else {
      center = centerWidget!;
    }

    return Row(
      children: [
        Expanded(child: buildHzLine(alignLeft: false)),
        Gap($styles.insets.sm),
        TweenAnimationBuilder<double>(
          duration: $styles.disableAnimations ? Duration.zero : dur,
          tween: Tween(begin: 0, end: isExpanded ? 0.5 : 0),
          curve: Curves.easeOutBack,
          child: SizedBox(width: centerSize, height: centerSize, child: center),
          builder: (_, value, child) =>
              Transform.rotate(angle: value * 6.283, child: child),
        ),
        Gap($styles.insets.sm),
        Expanded(child: buildHzLine(alignLeft: true)),
      ],
    );
  }
}

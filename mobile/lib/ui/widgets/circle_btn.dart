// Wonderize: circular button + icon button (back + close convenience).
//
// Adapted from Wonderous lib/ui/common/controls/circle_buttons.dart.

import 'package:flutter/material.dart';

import '../app_scaffold.dart';
import 'app_btn.dart';

class CircleBtn extends StatelessWidget {
  const CircleBtn({
    super.key,
    required this.child,
    required this.onPressed,
    required this.semanticLabel,
    this.border,
    this.bgColor,
    this.size = 48,
  });

  final VoidCallback? onPressed;
  final Color? bgColor;
  final BorderSide? border;
  final Widget child;
  final double size;
  final String semanticLabel;

  @override
  Widget build(BuildContext context) => AppBtn(
        onPressed: onPressed,
        semanticLabel: semanticLabel,
        minimumSize: Size(size, size),
        padding: EdgeInsets.zero,
        circular: true,
        bgColor: bgColor,
        border: border,
        child: child,
      );
}

class CircleIconBtn extends StatelessWidget {
  const CircleIconBtn({
    super.key,
    required this.icon,
    required this.onPressed,
    required this.semanticLabel,
    this.border,
    this.bgColor,
    this.color,
    this.size = 48,
    this.iconSize = 24,
    this.flipIcon = false,
  });

  final IconData icon;
  final VoidCallback? onPressed;
  final BorderSide? border;
  final Color? bgColor;
  final Color? color;
  final String semanticLabel;
  final double size;
  final double iconSize;
  final bool flipIcon;

  @override
  Widget build(BuildContext context) {
    final defaultBg = $styles.colors.greyStrong;
    final iconColor = color ?? $styles.colors.offWhite;
    return CircleBtn(
      onPressed: onPressed,
      border: border,
      size: size,
      bgColor: bgColor ?? defaultBg,
      semanticLabel: semanticLabel,
      child: Transform.scale(
        scaleX: flipIcon ? -1 : 1,
        child: Icon(icon, size: iconSize, color: iconColor),
      ),
    );
  }
}

class BackBtn extends StatelessWidget {
  const BackBtn({
    super.key,
    this.onPressed,
    this.semanticLabel = 'Back',
    this.bgColor,
    this.iconColor,
  });

  final VoidCallback? onPressed;
  final String semanticLabel;
  final Color? bgColor;
  final Color? iconColor;

  @override
  Widget build(BuildContext context) => CircleIconBtn(
        icon: Icons.arrow_back,
        bgColor: bgColor,
        color: iconColor,
        semanticLabel: semanticLabel,
        onPressed: onPressed ?? () => Navigator.maybePop(context),
      );
}

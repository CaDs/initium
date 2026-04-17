// Wonderize: primary button with press effect, hover effect, and haptics.
//
// Distilled from Wonderous lib/ui/common/controls/buttons.dart.

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:gap/gap.dart';

import '../app_scaffold.dart';

// `late final` is needed for the .from factory; immutability is preserved in
// practice since fields are assigned once at construction time.
// ignore: must_be_immutable
class AppBtn extends StatelessWidget {
  // ignore: prefer_const_constructors_in_immutables
  AppBtn({
    super.key,
    required this.onPressed,
    required this.semanticLabel,
    this.enableFeedback = true,
    this.pressEffect = true,
    this.hoverEffect = true,
    this.child,
    this.padding,
    this.expand = false,
    this.isSecondary = false,
    this.circular = false,
    this.minimumSize,
    this.bgColor,
    this.border,
  }) : _builder = null;

  AppBtn.from({
    super.key,
    required this.onPressed,
    this.enableFeedback = true,
    this.pressEffect = true,
    this.hoverEffect = true,
    this.padding,
    this.expand = false,
    this.isSecondary = false,
    this.minimumSize,
    this.bgColor,
    this.border,
    String? semanticLabel,
    String? text,
    Widget? icon,
  })  : child = null,
        circular = false,
        assert(semanticLabel != null || text != null) {
    this.semanticLabel = semanticLabel ?? text ?? '';
    _builder = (context) {
      Text? txt = text == null
          ? null
          : Text(
              text.toUpperCase(),
              style: $styles.text.btn.copyWith(
                color: isSecondary
                    ? $styles.colors.fg
                    : $styles.colors.offWhite,
              ),
            );
      if (txt != null && icon != null) {
        return Row(
          mainAxisAlignment: MainAxisAlignment.center,
          mainAxisSize: MainAxisSize.min,
          children: [txt, Gap($styles.insets.xs), icon],
        );
      }
      return (txt ?? icon ?? const SizedBox.shrink());
    };
  }

  // ignore: prefer_const_constructors_in_immutables
  AppBtn.basic({
    super.key,
    required this.onPressed,
    required this.semanticLabel,
    this.enableFeedback = true,
    this.pressEffect = true,
    this.hoverEffect = true,
    this.child,
    this.padding = EdgeInsets.zero,
    this.isSecondary = false,
    this.circular = false,
    this.minimumSize,
  })  : expand = false,
        bgColor = Colors.transparent,
        border = null,
        _builder = null;

  final VoidCallback? onPressed;
  late final String semanticLabel;
  final bool enableFeedback;

  late final Widget? child;
  late final WidgetBuilder? _builder;

  final EdgeInsets? padding;
  final bool expand;
  final bool circular;
  final Size? minimumSize;
  final bool isSecondary;
  final BorderSide? border;
  final Color? bgColor;
  final bool pressEffect;
  final bool hoverEffect;

  @override
  Widget build(BuildContext context) {
    final defaultColor =
        isSecondary ? $styles.colors.bg : $styles.colors.accent1;
    final textColor = isSecondary ? $styles.colors.fg : Colors.white;
    final side = border ?? BorderSide.none;

    Widget content = _builder?.call(context) ?? child ?? const SizedBox.shrink();
    if (expand) content = Center(child: content);

    final shape = circular
        ? CircleBorder(side: side)
        : RoundedRectangleBorder(
            side: side,
            borderRadius: BorderRadius.circular($styles.corners.md),
          );

    final style = ButtonStyle(
      minimumSize: WidgetStatePropertyAll(minimumSize ?? Size.zero),
      tapTargetSize: MaterialTapTargetSize.shrinkWrap,
      splashFactory: NoSplash.splashFactory,
      backgroundColor: WidgetStatePropertyAll(bgColor ?? defaultColor),
      overlayColor: const WidgetStatePropertyAll(Colors.transparent),
      shape: WidgetStatePropertyAll(shape),
      padding: WidgetStatePropertyAll(
          padding ?? EdgeInsets.all($styles.insets.md)),
      enableFeedback: enableFeedback,
    );

    Widget button = Opacity(
      opacity: onPressed == null ? 0.5 : 1.0,
      child: TextButton(
        onPressed: onPressed == null ? null : _wrap(onPressed!),
        style: style,
        child: DefaultTextStyle(
          style: DefaultTextStyle.of(context).style.copyWith(color: textColor),
          child: content,
        ),
      ),
    );

    if (pressEffect && onPressed != null) button = _PressEffect(child: button);
    if (hoverEffect && kIsWeb) {
      button = _HoverEffect(circular: circular, child: button);
    }

    return Semantics(
      label: semanticLabel.isEmpty ? null : semanticLabel,
      button: true,
      container: true,
      onTap: onPressed == null ? null : () => _wrap(onPressed!)(),
      child: ExcludeSemantics(child: button),
    );
  }

  VoidCallback _wrap(VoidCallback original) => () {
        if (enableFeedback &&
            !kIsWeb &&
            defaultTargetPlatform == TargetPlatform.android) {
          HapticFeedback.lightImpact();
        }
        original();
      };
}

class _PressEffect extends StatefulWidget {
  const _PressEffect({required this.child});
  final Widget child;
  @override
  State<_PressEffect> createState() => _PressEffectState();
}

class _PressEffectState extends State<_PressEffect> {
  bool _down = false;
  @override
  Widget build(BuildContext context) => GestureDetector(
        excludeFromSemantics: true,
        onTapDown: (_) => setState(() => _down = true),
        onTapUp: (_) => setState(() => _down = false),
        onTapCancel: () => setState(() => _down = false),
        behavior: HitTestBehavior.translucent,
        child: Opacity(
          opacity: _down ? 0.7 : 1,
          child: ExcludeSemantics(child: widget.child),
        ),
      );
}

class _HoverEffect extends StatefulWidget {
  const _HoverEffect({required this.child, required this.circular});
  final Widget child;
  final bool circular;
  @override
  State<_HoverEffect> createState() => _HoverEffectState();
}

class _HoverEffectState extends State<_HoverEffect> {
  bool _over = false;
  @override
  Widget build(BuildContext context) => MouseRegion(
        onEnter: (_) => setState(() => _over = true),
        onExit: (_) => setState(() => _over = false),
        child: AnimatedContainer(
          foregroundDecoration: BoxDecoration(
            color: _over
                ? $styles.colors.white.withValues(alpha: 0.12)
                : Colors.transparent,
            borderRadius: BorderRadius.circular(
                widget.circular ? 9999 : $styles.corners.md),
          ),
          duration: const Duration(milliseconds: 400),
          curve: Curves.easeInOut,
          child: widget.child,
        ),
      );
}

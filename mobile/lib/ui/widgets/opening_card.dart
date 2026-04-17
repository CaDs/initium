// Wonderize: animated card that grows/shrinks between two children.

import 'package:flutter/material.dart';
import 'package:flutter/scheduler.dart';

import '../app_scaffold.dart';

class OpeningCard extends StatefulWidget {
  const OpeningCard({
    super.key,
    required this.closedBuilder,
    required this.openBuilder,
    required this.isOpen,
    this.background,
    this.padding,
    this.duration,
  });

  final WidgetBuilder closedBuilder;
  final WidgetBuilder openBuilder;
  final Widget? background;
  final bool isOpen;
  final EdgeInsets? padding;
  final Duration? duration;

  @override
  State<OpeningCard> createState() => _OpeningCardState();
}

class _OpeningCardState extends State<OpeningCard> {
  Size _size = Size.zero;
  final GlobalKey _measureKey = GlobalKey();

  @override
  Widget build(BuildContext context) {
    final dur = widget.duration ?? $styles.times.fast;
    return TweenAnimationBuilder<Size>(
      duration: dur,
      curve: Curves.easeOut,
      tween: Tween(begin: _size, end: _size),
      builder: (_, value, child) => Stack(
        children: [
          if (widget.background != null)
            Positioned.fill(child: widget.background!),
          Padding(
            padding: widget.padding ?? EdgeInsets.zero,
            child: SizedBox(
                width: value.width, height: value.height, child: child),
          ),
        ],
      ),
      child: AnimatedSwitcher(
        duration: dur,
        child: ClipRect(
          key: ValueKey(widget.isOpen),
          child: OverflowBox(
            minWidth: 0,
            minHeight: 0,
            maxWidth: double.infinity,
            maxHeight: double.infinity,
            child: _Measurable(
              key: _measureKey,
              onSize: (s) {
                if (s != _size) setState(() => _size = s);
              },
              child: widget.isOpen
                  ? widget.openBuilder(context)
                  : widget.closedBuilder(context),
            ),
          ),
        ),
      ),
    );
  }
}

class _Measurable extends StatefulWidget {
  const _Measurable({super.key, required this.child, required this.onSize});
  final Widget child;
  final ValueChanged<Size> onSize;

  @override
  State<_Measurable> createState() => _MeasurableState();
}

class _MeasurableState extends State<_Measurable> {
  @override
  Widget build(BuildContext context) {
    SchedulerBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      final box = context.findRenderObject() as RenderBox?;
      if (box != null && box.hasSize) widget.onSize(box.size);
    });
    return widget.child;
  }
}

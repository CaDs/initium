// Wonderize: scroll-aware decoration overlay.

import 'package:flutter/material.dart';

typedef ScrollChildBuilder = Widget Function(ScrollController controller);

// `late final` fg/bg builders are assigned in .fade / .shadow factories;
// conceptually immutable once constructed.
// ignore: must_be_immutable
class ScrollDecorator extends StatefulWidget {
  // ignore: prefer_const_constructors_in_immutables
  ScrollDecorator({
    super.key,
    required this.builder,
    this.fgBuilder,
    this.bgBuilder,
    this.controller,
    this.onInit,
  });

  ScrollDecorator.fade({
    super.key,
    required this.builder,
    this.controller,
    this.onInit,
    Widget? begin,
    Widget? end,
    bool bg = false,
    Axis direction = Axis.vertical,
    Duration duration = const Duration(milliseconds: 150),
  }) {
    Flex flexBuilder(ScrollController c) => Flex(
          direction: direction,
          children: [
            if (begin != null)
              AnimatedOpacity(
                duration: duration,
                opacity:
                    c.hasClients && c.position.extentBefore > 3 ? 1 : 0,
                child: begin,
              ),
            const Spacer(),
            if (end != null)
              AnimatedOpacity(
                duration: duration,
                opacity: c.hasClients && c.position.extentAfter > 3 ? 1 : 0,
                child: end,
              ),
          ],
        );
    bgBuilder = bg ? flexBuilder : null;
    fgBuilder = !bg ? flexBuilder : null;
  }

  ScrollDecorator.shadow({
    super.key,
    required this.builder,
    this.controller,
    this.onInit,
    Color color = Colors.black54,
  }) {
    bgBuilder = null;
    fgBuilder = (c) {
      final ratio =
          c.hasClients ? (c.position.extentBefore / 60).clamp(0.0, 1.0) : 0.0;
      return ExcludeSemantics(
        child: IgnorePointer(
          child: Container(
            height: 24,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  color.withValues(alpha: ratio * color.a),
                  Colors.transparent,
                ],
                stops: [0, ratio.toDouble()],
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
              ),
            ),
          ),
        ),
      );
    };
  }

  final ScrollController? controller;
  final ScrollChildBuilder builder;
  late final ScrollChildBuilder? fgBuilder;
  late final ScrollChildBuilder? bgBuilder;
  final void Function(ScrollController controller)? onInit;

  @override
  State<ScrollDecorator> createState() => _ScrollDecoratorState();
}

class _ScrollDecoratorState extends State<ScrollDecorator> {
  ScrollController? _local;
  ScrollController get _ctrl => (widget.controller ?? _local)!;

  @override
  void initState() {
    super.initState();
    if (widget.controller == null) _local = ScrollController();
    widget.onInit?.call(_ctrl);
  }

  @override
  void dispose() {
    _local?.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final content = widget.builder(_ctrl);
    return AnimatedBuilder(
      animation: _ctrl,
      builder: (_, _) => Stack(
        children: [
          if (widget.bgBuilder != null) widget.bgBuilder!(_ctrl),
          content,
          if (widget.fgBuilder != null) widget.fgBuilder!(_ctrl),
        ],
      ),
    );
  }
}

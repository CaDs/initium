// Wonderize: scroll-linked parallax primitive.
//
// Distills the `pctVisible` pattern from Wonderous into a reusable helper.
// Wrap any child in [ScrollParallax]; the builder receives a value 0..1
// representing how far the widget has scrolled across the viewport.

import 'package:flutter/material.dart';

class ScrollParallax extends StatelessWidget {
  const ScrollParallax({
    super.key,
    required this.scrollPos,
    required this.builder,
  });

  final ValueNotifier<double> scrollPos;
  final Widget Function(BuildContext context, double pctVisible) builder;

  @override
  Widget build(BuildContext context) {
    return ValueListenableBuilder<double>(
      valueListenable: scrollPos,
      builder: (context, _, _) {
        final renderObj = context.findRenderObject() as RenderBox?;
        double pctVisible = 0;
        if (renderObj != null && renderObj.attached) {
          final pos = renderObj.localToGlobal(Offset.zero);
          final viewportH = MediaQuery.sizeOf(context).height;
          final selfH = renderObj.size.height;
          if (selfH > 0) {
            final amtVisible = viewportH - pos.dy;
            pctVisible = (amtVisible / selfH).clamp(0.0, 3.0);
          }
        }
        return builder(context, pctVisible);
      },
    );
  }
}

class ScrollParallaxTranslate extends StatelessWidget {
  const ScrollParallaxTranslate({
    super.key,
    required this.scrollPos,
    required this.child,
    this.startOffset = Offset.zero,
    this.endOffset = const Offset(0, -40),
  });

  final ValueNotifier<double> scrollPos;
  final Widget child;
  final Offset startOffset;
  final Offset endOffset;

  @override
  Widget build(BuildContext context) {
    return ScrollParallax(
      scrollPos: scrollPos,
      builder: (context, pct) {
        final dx = startOffset.dx + (endOffset.dx - startOffset.dx) * pct;
        final dy = startOffset.dy + (endOffset.dy - startOffset.dy) * pct;
        return Transform.translate(offset: Offset(dx, dy), child: child);
      },
    );
  }
}

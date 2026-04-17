// Wonderize: pull-up-to-reveal hero handoff controller.
//
// Adapted from Wonderous lib/ui/screens/home/_vertical_swipe_controller.dart.
// Controller behind the signature swipe-up gesture. Drags update `swipeAmt`
// (0..1); when it reaches 1, `onSwipeComplete()` fires.

import 'package:flutter/material.dart';

import '../ui/app_scaffold.dart';

class VerticalSwipeController {
  VerticalSwipeController({
    required this.ticker,
    required this.onSwipeComplete,
    this.pullThresholdPx = 150,
  }) {
    swipeReleaseAnim = AnimationController(vsync: ticker)
      ..addListener(_handleReleaseTick);
  }

  final TickerProvider ticker;
  final VoidCallback onSwipeComplete;

  /// How many pixels of upward drag = full swipe (swipeAmt == 1).
  final double pullThresholdPx;

  /// Current drag progress, 0 (idle) to 1 (triggered).
  final ValueNotifier<double> swipeAmt = ValueNotifier(0);

  /// True while the user's finger is on screen.
  final ValueNotifier<bool> isPointerDown = ValueNotifier(false);

  late final AnimationController swipeReleaseAnim;

  void _handleReleaseTick() => swipeAmt.value = swipeReleaseAnim.value;

  void _handleTapDown(_) => isPointerDown.value = true;
  void _handleTapUp(_) => isPointerDown.value = false;

  void _handleDragUpdate(DragUpdateDetails details) {
    if (swipeReleaseAnim.isAnimating) swipeReleaseAnim.stop();
    isPointerDown.value = true;

    final next = (swipeAmt.value - details.delta.dy / pullThresholdPx)
        .clamp(0.0, 1.0);
    if (next != swipeAmt.value) {
      swipeAmt.value = next;
      if (swipeAmt.value == 1.0) onSwipeComplete();
    }
  }

  void _handleDragRelease([_]) {
    swipeReleaseAnim.duration = $styles.disableAnimations
        ? const Duration(milliseconds: 1)
        : Duration(
            milliseconds: (swipeAmt.value * 500).toInt().clamp(1, 500),
          );
    swipeReleaseAnim.reverse(from: swipeAmt.value);
    isPointerDown.value = false;
  }

  /// Wrap your interactive child in a translucent gesture detector.
  Widget wrapGestureDetector(Widget child, {Key? key}) => GestureDetector(
        key: key,
        excludeFromSemantics: true,
        onTapDown: _handleTapDown,
        onTapUp: _handleTapUp,
        onVerticalDragUpdate: _handleDragUpdate,
        onVerticalDragEnd: _handleDragRelease,
        onVerticalDragCancel: _handleDragRelease,
        behavior: HitTestBehavior.translucent,
        child: child,
      );

  /// Listens to both swipeAmt and isPointerDown in one builder.
  Widget buildListener({
    required Widget Function(double swipeAmt, bool isPointerDown, Widget? child)
        builder,
    Widget? child,
  }) {
    return ValueListenableBuilder<double>(
      valueListenable: swipeAmt,
      builder: (_, amt, _) => ValueListenableBuilder<bool>(
        valueListenable: isPointerDown,
        builder: (_, down, _) => builder(amt, down, child),
      ),
    );
  }

  void dispose() {
    swipeAmt.dispose();
    isPointerDown.dispose();
    swipeReleaseAnim.dispose();
  }
}

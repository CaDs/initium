// Wonderize: .maybeAnimate() accessibility wrapper.
//
// Adapted from Wonderous lib/logic/common/animate_utils.dart.
// flutter_animate's `.animate()` always plays; `.maybeAnimate()` short-circuits
// to a no-op widget when the OS reports `disableAnimations`. Use this EVERYWHERE
// you'd reach for `.animate()` so the entire app respects reduced-motion.

import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../ui/app_scaffold.dart';

/// Drop-in replacement for [Animate] that does nothing when animations are
/// disabled. Returning a State means parent rebuilds don't re-trigger the no-op.
// ignore: must_be_immutable
class NeverAnimate extends Animate {
  NeverAnimate({super.key, super.child});

  @override
  State<NeverAnimate> createState() => _NeverAnimateState();
}

class _NeverAnimateState extends State<NeverAnimate> {
  @override
  Widget build(BuildContext context) => widget.child;
}

extension MaybeAnimateExtension on Widget {
  Animate maybeAnimate({
    Key? key,
    List<Effect>? effects,
    AnimateCallback? onInit,
    AnimateCallback? onPlay,
    AnimateCallback? onComplete,
    bool? autoPlay,
    Duration? delay,
    AnimationController? controller,
    Adapter? adapter,
    double? target,
    double? value,
  }) {
    if ($styles.disableAnimations) {
      return NeverAnimate(child: this);
    }
    return Animate(
      key: key,
      effects: effects,
      onInit: onInit,
      onPlay: onPlay,
      onComplete: onComplete,
      autoPlay: autoPlay,
      delay: delay,
      controller: controller,
      adapter: adapter,
      target: target,
      value: value,
      child: this,
    );
  }
}

// Wonderize: procedural cloud layer with seeded randomization.
//
// Adapted from Wonderous lib/ui/wonder_illustrations/common/animated_clouds.dart.
// Seeded pseudo-random generator so the same `seed` always produces the same
// cloud layout. When the seed changes, old clouds drift out as new ones drift in.

import 'dart:math';

import 'package:flutter/material.dart';

class AnimatedClouds extends StatefulWidget {
  const AnimatedClouds({
    super.key,
    required this.cloudAsset,
    required this.seed,
    this.cloudCount = 3,
    this.cloudSize = 500,
    this.opacity = 0.4,
    this.entranceDuration = const Duration(milliseconds: 1500),
    this.driftDistance = 1000,
    this.enableAnimations = true,
  });

  final String cloudAsset;
  final int seed;
  final int cloudCount;
  final double cloudSize;
  final double opacity;
  final Duration entranceDuration;
  final double driftDistance;
  final bool enableAnimations;

  @override
  State<AnimatedClouds> createState() => _AnimatedCloudsState();
}

class _AnimatedCloudsState extends State<AnimatedClouds>
    with SingleTickerProviderStateMixin {
  late final AnimationController _anim =
      AnimationController(vsync: this, duration: widget.entranceDuration);
  List<_Cloud> _clouds = const [];
  List<_Cloud> _oldClouds = const [];
  Size? _lastSize;

  @override
  void didUpdateWidget(covariant AnimatedClouds old) {
    super.didUpdateWidget(old);
    if (old.seed != widget.seed) {
      _oldClouds = _clouds;
      _clouds = _buildClouds(_lastSize ?? const Size(400, 400));
      _showClouds();
    }
  }

  @override
  void dispose() {
    _anim.dispose();
    super.dispose();
  }

  void _showClouds() =>
      widget.enableAnimations ? _anim.forward(from: 0) : _anim.value = 1;

  List<_Cloud> _buildClouds(Size size) {
    final rng = Random(widget.seed);
    return List.generate(widget.cloudCount, (_) {
      return _Cloud(
        pos: Offset(
          _between(rng, -200, size.width - 100),
          _between(rng, 50, size.height - 50),
        ),
        scale: _between(rng, 0.7, 1.0),
        flipX: rng.nextBool(),
        flipY: rng.nextBool(),
        opacity: widget.opacity,
        size: widget.cloudSize,
        asset: widget.cloudAsset,
      );
    });
  }

  static double _between(Random r, double min, double max) =>
      min + r.nextDouble() * (max - min);

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(builder: (context, constraints) {
      final size = constraints.biggest.isFinite
          ? constraints.biggest
          : const Size(400, 400);
      if (_lastSize != size) {
        _lastSize = size;
        if (_clouds.isEmpty) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (mounted) {
              setState(() => _clouds = _buildClouds(size));
              _showClouds();
            }
          });
        }
      }

      Widget buildCloud(_Cloud c, {required bool isOld}) {
        final idx = _clouds.indexOf(c).abs();
        final dir = idx % 2 == 0 ? -widget.driftDistance : widget.driftDistance;
        final t = Curves.easeOut.transform(_anim.value);
        return Positioned(
          top: c.pos.dy,
          left: isOld ? c.pos.dx - dir * t : c.pos.dx + dir * (1 - t),
          child:
              Opacity(opacity: isOld ? 1 - _anim.value : _anim.value, child: c),
        );
      }

      return RepaintBoundary(
        child: ClipRect(
          child: AnimatedBuilder(
            animation: _anim,
            builder: (_, _) => Stack(
              clipBehavior: Clip.hardEdge,
              children: [
                if (_anim.value != 1)
                  ..._oldClouds.map((c) => buildCloud(c, isOld: true)),
                ..._clouds.map((c) => buildCloud(c, isOld: false)),
              ],
            ),
          ),
        ),
      );
    });
  }
}

class _Cloud extends StatelessWidget {
  const _Cloud({
    required this.pos,
    required this.scale,
    required this.flipX,
    required this.flipY,
    required this.opacity,
    required this.size,
    required this.asset,
  });

  final Offset pos;
  final double scale;
  final bool flipX;
  final bool flipY;
  final double opacity;
  final double size;
  final String asset;

  @override
  Widget build(BuildContext context) => Transform.scale(
        scaleX: scale * (flipX ? -1 : 1),
        scaleY: scale * (flipY ? -1 : 1),
        child: Image.asset(
          asset,
          excludeFromSemantics: true,
          opacity: AlwaysStoppedAnimation(opacity),
          width: size * scale,
          fit: BoxFit.fitWidth,
        ),
      );
}

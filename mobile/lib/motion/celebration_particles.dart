// Wonderize: radial particle burst for "you did it" moments.
//
// Adapted from Wonderous lib/ui/screens/collectible_found/widgets/_celebration_particles.dart.
// Spawns ~1200 particles in a radial pattern over ~1 second, fading with easeOutExpo.

import 'dart:math';

import 'package:flutter/material.dart';
import 'package:particle_field/particle_field.dart';

import '../ui/app_scaffold.dart';

class CelebrationParticles extends StatelessWidget {
  const CelebrationParticles({
    super.key,
    required this.spriteAsset,
    this.spriteFrameWidth = 21,
    this.spriteScale = 0.75,
    this.particleCount = 1200,
    this.fadeMs = 1000,
    this.color,
    this.spreadFactor = 0.3,
    this.velocityFactor = 0.08,
  });

  final String spriteAsset;
  final int spriteFrameWidth;
  final double spriteScale;
  final int particleCount;
  final int fadeMs;
  final Color? color;
  final double spreadFactor;
  final double velocityFactor;

  @override
  Widget build(BuildContext context) {
    final tint = color ?? $styles.colors.accent1;
    final rng = Random();
    var remaining = particleCount;

    return Positioned.fill(
      child: RepaintBoundary(
        child: ParticleField(
          blendMode: BlendMode.dstIn,
          spriteSheet: SpriteSheet(
            image: AssetImage(spriteAsset),
            frameWidth: spriteFrameWidth,
            scale: spriteScale,
          ),
          onTick: (controller, elapsed, size) {
            final particles = controller.particles;

            final d = min(size.width, size.height) * spreadFactor;
            final v = d * velocityFactor;

            controller.opacity = Curves.easeOutExpo.transform(
              max(0, 1 - elapsed.inMilliseconds / fadeMs),
            );
            if (controller.opacity == 0) return;

            var addCount = remaining ~/ 30;
            remaining -= addCount;
            while (--addCount > 0) {
              final angle = rng.nextDouble() * 2 * pi;
              particles.add(
                Particle(
                  x: cos(angle) * d * _between(rng, 0.8, 1.0),
                  y: sin(angle) * d * _between(rng, 0.8, 1.0),
                  vx: cos(angle) * v * _between(rng, 0.5, 1.5),
                  vy: sin(angle) * v * _between(rng, 0.5, 1.5),
                  color: tint.withValues(alpha: _between(rng, 0.5, 1.0)),
                ),
              );
            }

            for (var i = particles.length - 1; i >= 0; i--) {
              final p = particles[i];
              p.update(frame: p.age ~/ 3);
              if (p.age > 40) particles.removeAt(i);
            }
          },
        ),
      ),
    );
  }

  static double _between(Random r, double a, double b) =>
      a + r.nextDouble() * (b - a);
}

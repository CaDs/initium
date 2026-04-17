# Particle sprites — PLACEHOLDER

| File | Size | Layout | Consumer |
|------|------|--------|----------|
| `sparkle.png` | 168×21 | 8 frames, horizontal, 21×21 each | `lib/motion/celebration_particles.dart` (`frameWidth: 21`) |

Frames animate sequentially — start small, pulse, fade. Swap with any horizontal sprite
sheet as long as `frameWidth` matches one frame's width. No code changes beyond that
constant.

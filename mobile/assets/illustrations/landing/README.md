# Landing parallax illustrations — PLACEHOLDER

Three SVG layers + one PNG cloud sheet, consumed by `lib/motion/parallax_layer_stack.dart`
and `lib/motion/animated_clouds.dart`.

| File | Size / aspect | Purpose | `zoomAmt` tier |
|------|---------------|---------|----------------|
| `bg.svg` | 1080×1920 (9:16) | Back layer — mountain silhouette / sky | 0.00 |
| `mid.svg` | 1080×1920 | Mid layer — foliage silhouette | 0.05 |
| `fg.svg` | 1080×1920 | Foreground layer — foliage detail / leaf | 0.20 |
| `clouds.png` | 1024×512, transparent | Ambient cloud drift | n/a |

Swap slots — drop in any real SVG with the same `viewBox` / PNG of at least 512px wide,
transparent where cloud should be atmospheric. No code changes required.

// Wonderize: 3D-ish illustration layering with per-piece zoom parallax.
//
// Distilled from Wonderous lib/ui/wonder_illustrations/common/illustration_piece.dart.
// Stack 3-5 pieces with different `zoomAmt` values; background pieces barely
// move, foreground pieces move dramatically as the master `parallax` changes.

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';

/// One layer in the parallax stack.
class ParallaxPiece {
  const ParallaxPiece({
    required this.child,
    this.alignment = Alignment.center,
    this.heightFactor = 1.0,
    this.minHeight,
    this.offset = Offset.zero,
    this.fractionalOffset,
    this.zoomAmt = 0.0,
    this.initialOffset = Offset.zero,
    this.initialScale = 1.0,
  });

  final Widget child;
  final Alignment alignment;
  final double heightFactor;
  final double? minHeight;
  final Offset offset;
  final Offset? fractionalOffset;
  final double zoomAmt;
  final Offset initialOffset;
  final double initialScale;
}

class ParallaxLayerStack extends StatelessWidget {
  const ParallaxLayerStack({
    super.key,
    required this.pieces,
    this.entrance,
    this.parallax,
    this.parallaxIntensity = 1.0,
  });

  /// Bottom-up stack order (first entry renders behind).
  final List<ParallaxPiece> pieces;

  /// 0..1 entrance progress.
  final Animation<double>? entrance;

  /// 0..1 parallax driver — typically a scroll position or swipe amount.
  final ValueListenable<double>? parallax;

  final double parallaxIntensity;

  @override
  Widget build(BuildContext context) {
    final entranceAnim = entrance ?? const AlwaysStoppedAnimation(1.0);

    return AnimatedBuilder(
      animation: entranceAnim,
      builder: (context, _) {
        return ValueListenableBuilder<double>(
          valueListenable: parallax ?? ValueNotifier<double>(0),
          builder: (context, pAmt, _) {
            return LayoutBuilder(
              builder: (context, constraints) {
                final entranceCurved =
                    Curves.easeOut.transform(entranceAnim.value);
                return Stack(
                  fit: StackFit.expand,
                  children: pieces
                      .map((p) => _buildPiece(
                          context, p, constraints, entranceCurved, pAmt))
                      .toList(),
                );
              },
            );
          },
        );
      },
    );
  }

  Widget _buildPiece(
    BuildContext context,
    ParallaxPiece p,
    BoxConstraints constraints,
    double entranceCurved,
    double parallaxAmt,
  ) {
    final introZoom = (p.initialScale - 1) * (1 - entranceCurved);
    final height = (constraints.maxHeight * p.heightFactor)
        .clamp(p.minHeight ?? 0, double.infinity)
        .toDouble();

    var translation = p.offset;
    if (p.initialOffset != Offset.zero) {
      translation += p.initialOffset * (1 - entranceCurved);
    }
    if (p.fractionalOffset != null) {
      translation += Offset(
        p.fractionalOffset!.dx * height,
        p.fractionalOffset!.dy * height,
      );
    }

    return Align(
      alignment: p.alignment,
      child: Transform.translate(
        offset: translation,
        child: Transform.scale(
          scale: 1 + (p.zoomAmt * parallaxIntensity * parallaxAmt) + introZoom,
          child: SizedBox(
            height: height,
            child: OverflowBox(maxWidth: 2500, child: p.child),
          ),
        ),
      ),
    );
  }
}

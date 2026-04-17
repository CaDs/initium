// Wonderize: arched / spade / pyramid clippers + flat curved-top clipper.
//
// Adapted from Wonderous lib/ui/common/curved_clippers.dart.
// Architectural shapes for the "monument silhouette" photo framing.

import 'package:flutter/material.dart';

enum ArchType { spade, pyramid, arch, wideArch, flatPyramid }

class ArchClipper extends CustomClipper<Path> {
  const ArchClipper(this.type);
  final ArchType type;

  @override
  Path getClip(Size size) {
    final pts = _archPoints(size, type);
    final path = Path()..moveTo(pts.first.pos.dx, pts.first.pos.dy);
    for (var i = 1; i < pts.length; i++) {
      final p = pts[i];
      path.quadraticBezierTo(p.control.dx, p.control.dy, p.pos.dx, p.pos.dy);
    }
    path.lineTo(pts.first.pos.dx, pts.first.pos.dy);
    return path;
  }

  @override
  bool shouldReclip(covariant ArchClipper old) => old.type != type;
}

class _ArchPoint {
  _ArchPoint(this.pos, [Offset? control]) : control = control ?? pos;
  final Offset pos;
  final Offset control;
}

List<_ArchPoint> _archPoints(Size size, ArchType type) {
  final dist = size.width / 3;
  switch (type) {
    case ArchType.pyramid:
      return [
        _ArchPoint(Offset(0, size.height)),
        _ArchPoint(Offset(0, dist)),
        _ArchPoint(Offset(size.width / 2, 0)),
        _ArchPoint(Offset(size.width, dist)),
        _ArchPoint(Offset(size.width, size.height)),
      ];
    case ArchType.spade:
      return [
        _ArchPoint(Offset(0, size.height)),
        _ArchPoint(Offset(0, dist)),
        _ArchPoint(Offset(size.width / 2, 0), Offset(0, dist * 0.66)),
        _ArchPoint(
            Offset(size.width, dist), Offset(size.width, dist * 0.66)),
        _ArchPoint(Offset(size.width, size.height)),
      ];
    case ArchType.arch:
      return [
        _ArchPoint(Offset(0, size.height)),
        _ArchPoint(Offset(0, size.width / 2)),
        _ArchPoint(Offset(size.width / 2, 0), Offset.zero),
        _ArchPoint(Offset(size.width, size.width / 2), Offset(size.width, 0)),
        _ArchPoint(Offset(size.width, size.height)),
      ];
    case ArchType.wideArch:
      return [
        _ArchPoint(Offset(0, size.height)),
        _ArchPoint(Offset(0, size.width / 2)),
        _ArchPoint(Offset(0, dist)),
        _ArchPoint(Offset(size.width / 2, 0), Offset.zero),
        _ArchPoint(Offset(size.width, dist), Offset(size.width, 0)),
        _ArchPoint(Offset(size.width, size.width / 2)),
        _ArchPoint(Offset(size.width, size.height)),
      ];
    case ArchType.flatPyramid:
      return [
        _ArchPoint(Offset(0, size.height)),
        _ArchPoint(Offset(0, dist)),
        _ArchPoint(Offset(size.width * 0.4, 0)),
        _ArchPoint(Offset(size.width * 0.6, 0)),
        _ArchPoint(Offset(size.width, dist)),
        _ArchPoint(Offset(size.width, size.height)),
      ];
  }
}

class CurvedTopClipper extends CustomClipper<Path> {
  const CurvedTopClipper({this.flip = false});
  final bool flip;

  @override
  Path getClip(Size size) {
    final radius = size.width / 2;
    final path = Path();
    if (flip) {
      path.lineTo(0, size.height - radius);
      path.arcToPoint(
        Offset(size.width, size.height - radius),
        radius: Radius.circular(radius),
        clockwise: false,
      );
      path.lineTo(size.width, 0);
      path.lineTo(0, 0);
    } else {
      path.lineTo(0, 0);
      path.lineTo(0, radius);
      path.arcToPoint(
        Offset(size.width, radius),
        radius: Radius.circular(radius / 2),
      );
      path.lineTo(size.width, size.height);
      path.lineTo(0, size.height);
    }
    return path;
  }

  @override
  bool shouldReclip(covariant CurvedTopClipper old) => old.flip != flip;
}

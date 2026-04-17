import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:gap/gap.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../motion/animate_utils.dart';
import '../../motion/animated_clouds.dart';
import '../../motion/parallax_layer_stack.dart';
import '../../motion/vertical_swipe_controller.dart';
import '../../ui/app_scaffold.dart';
import '../../ui/widgets/app_btn.dart';
import '../shared/locale_switcher.dart';
import '../shared/theme_switcher.dart';

class LandingScreen extends StatefulWidget {
  const LandingScreen({super.key});

  @override
  State<LandingScreen> createState() => _LandingScreenState();
}

class _LandingScreenState extends State<LandingScreen>
    with SingleTickerProviderStateMixin {
  late final VerticalSwipeController _swipe;
  bool _navigating = false;

  @override
  void initState() {
    super.initState();
    _swipe = VerticalSwipeController(
      ticker: this,
      onSwipeComplete: _goToLogin,
    );
  }

  void _goToLogin() {
    if (_navigating || !mounted) return;
    _navigating = true;
    context.go('/login');
  }

  @override
  void dispose() {
    _swipe.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: $styles.colors.bg,
      body: _swipe.wrapGestureDetector(
        Stack(
          fit: StackFit.expand,
          children: [
            _buildParallax(),
            _buildCloudLayer(),
            _buildContent(l10n),
            _buildTopBar(),
            _buildBottomHint(l10n),
          ],
        ),
      ),
    );
  }

  Widget _buildParallax() {
    return ParallaxLayerStack(
      parallax: _swipe.swipeAmt,
      parallaxIntensity: 1.0,
      pieces: [
        ParallaxPiece(
          zoomAmt: 0.0,
          child: SvgPicture.asset(
            'assets/illustrations/landing/bg.svg',
            fit: BoxFit.cover,
            alignment: Alignment.center,
          ),
        ),
        ParallaxPiece(
          zoomAmt: 0.05,
          child: SvgPicture.asset(
            'assets/illustrations/landing/mid.svg',
            fit: BoxFit.cover,
            alignment: Alignment.bottomCenter,
          ),
        ),
        ParallaxPiece(
          zoomAmt: 0.2,
          child: SvgPicture.asset(
            'assets/illustrations/landing/fg.svg',
            fit: BoxFit.cover,
            alignment: Alignment.bottomCenter,
          ),
        ),
      ],
    );
  }

  Widget _buildCloudLayer() {
    return IgnorePointer(
      child: Opacity(
        opacity: $styles.highContrast ? 0.0 : 0.65,
        child: const AnimatedClouds(
          cloudAsset: 'assets/illustrations/landing/clouds.png',
          seed: 7,
          cloudCount: 4,
          cloudSize: 420,
          opacity: 0.55,
          driftDistance: 800,
        ),
      ),
    );
  }

  Widget _buildContent(AppLocalizations l10n) {
    return Positioned.fill(
      child: SafeArea(
        child: Padding(
          padding: EdgeInsets.symmetric(horizontal: $styles.insets.lg),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text(
                l10n.landingTitle,
                textAlign: TextAlign.center,
                semanticsLabel: l10n.landingTitle,
                style: $styles.text.wonderTitle.copyWith(
                  color: $styles.colors.fg,
                  shadows: $styles.shadows.textSoft,
                ),
              ).maybeAnimate().fadeIn(duration: $styles.times.slow).slideY(
                    begin: 0.08,
                    end: 0,
                    duration: $styles.times.slow,
                    curve: Curves.easeOutCubic,
                  ),
              Gap($styles.insets.sm),
              Text(
                l10n.landingSubtitle,
                textAlign: TextAlign.center,
                style: $styles.text.body.copyWith(
                  color: $styles.colors.fg.withValues(alpha: 0.78),
                ),
              ).maybeAnimate(delay: $styles.times.fast).fadeIn(
                    duration: $styles.times.slow,
                  ),
              Gap($styles.insets.xl),
              AppBtn.from(
                onPressed: _goToLogin,
                text: l10n.landingCta,
                semanticLabel: l10n.landingCta,
              ).maybeAnimate(delay: $styles.times.med).fadeIn(
                    duration: $styles.times.slow,
                  ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTopBar() {
    return Positioned(
      top: 0,
      left: 0,
      right: 0,
      child: SafeArea(
        child: Padding(
          padding: EdgeInsets.all($styles.insets.xs),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: const [
              LocaleSwitcher(),
              SizedBox(width: 4),
              ThemeSwitcher(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildBottomHint(AppLocalizations l10n) {
    return Positioned(
      bottom: 0,
      left: 0,
      right: 0,
      child: SafeArea(
        minimum: EdgeInsets.only(bottom: $styles.insets.md),
        child: _swipe.buildListener(
          builder: (swipeAmt, isDown, _) {
            final fade = (1 - swipeAmt * 2).clamp(0.0, 1.0);
            return Opacity(
              opacity: fade,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(
                    Icons.keyboard_arrow_up,
                    color: $styles.colors.fg.withValues(alpha: 0.55),
                    size: 32,
                  )
                      .maybeAnimate(onPlay: (c) => c.repeat(reverse: true))
                      .moveY(
                        begin: 0,
                        end: -6,
                        duration: $styles.times.slow,
                        curve: Curves.easeInOut,
                      ),
                  Text(
                    l10n.landingSwipeHint,
                    style: $styles.text.bodySmall.copyWith(
                      color: $styles.colors.fg.withValues(alpha: 0.55),
                    ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}

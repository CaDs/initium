import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:gap/gap.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../motion/animate_utils.dart';
import '../../motion/app_backdrop.dart';
import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import '../../ui/app_scaffold.dart';
import '../../ui/widgets/app_header.dart';
import '../../ui/widgets/compass_divider.dart';
import '../../ui/widgets/curved_clippers.dart';
import '../../ui/widgets/gradient_container.dart';
import 'widgets/google_sign_in_button.dart';
import 'widgets/magic_link_form.dart';

class LoginScreen extends ConsumerWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;
    final keyboardOpen = MediaQuery.viewInsetsOf(context).bottom > 0;

    return Scaffold(
      backgroundColor: $styles.colors.bg,
      resizeToAvoidBottomInset: true,
      body: Stack(
        children: [
          _Hero(caption: l10n.loginHeroCaption),
          AppBackdrop(
            strength: keyboardOpen ? 1.0 : 0.0,
            child: SafeArea(
              child: SingleChildScrollView(
                padding: EdgeInsets.symmetric(horizontal: $styles.insets.md),
                keyboardDismissBehavior:
                    ScrollViewKeyboardDismissBehavior.onDrag,
                child: Column(
                  children: [
                    AppHeader(
                      showBackBtn: true,
                      title: l10n.loginTitle,
                    ),
                    Gap(_heroHeight(context) - kToolbarHeight),
                    _Card(
                      child: Column(
                        children: [
                          Text(
                            l10n.loginSubtitle,
                            textAlign: TextAlign.center,
                            style: $styles.text.body.copyWith(
                              color: $styles.colors.fg
                                  .withValues(alpha: 0.75),
                            ),
                          ),
                          Gap($styles.insets.md),
                          const GoogleSignInButton(),
                          Gap($styles.insets.sm),
                          CompassDivider(
                            isExpanded: true,
                            centerWidget: Text(
                              l10n.loginDivider,
                              style: $styles.text.caption.copyWith(
                                color: $styles.colors.greyMedium,
                              ),
                            ),
                            centerSize: 40,
                          ),
                          Gap($styles.insets.sm),
                          const MagicLinkForm(),
                          if (authState is AuthError) ...[
                            Gap($styles.insets.sm),
                            _ErrorBanner(
                              message: _localizeError(authState.message, l10n),
                            ),
                          ],
                          Gap($styles.insets.md),
                        ],
                      ),
                    ),
                    Gap($styles.insets.xl),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  double _heroHeight(BuildContext context) {
    final h = MediaQuery.sizeOf(context).height;
    return (h * 0.34).clamp(220.0, 360.0);
  }

  String _localizeError(String code, AppLocalizations l10n) {
    return switch (code) {
      'google_login_failed' => l10n.authGoogleLoginFailed,
      'magic_link_failed' => l10n.authMagicLinkFailed,
      _ => code,
    };
  }
}

class _Hero extends StatelessWidget {
  const _Hero({required this.caption});
  final String caption;

  @override
  Widget build(BuildContext context) {
    final h = (MediaQuery.sizeOf(context).height * 0.34).clamp(220.0, 360.0);
    return Positioned(
      top: 0,
      left: 0,
      right: 0,
      height: h,
      child: IgnorePointer(
        child: ClipPath(
          clipper: const ArchClipper(ArchType.arch),
          child: Stack(
            fit: StackFit.expand,
            children: [
              SvgPicture.asset(
                'assets/illustrations/login/hero.svg',
                fit: BoxFit.cover,
                alignment: Alignment.center,
              ),
              VtGradient(
                [
                  $styles.colors.accent1.withValues(alpha: 0.35),
                  $styles.colors.accent3.withValues(alpha: 0.55),
                ],
                const [0, 1],
                blendMode: BlendMode.colorBurn,
              ),
              Align(
                alignment: const Alignment(0, 0.55),
                child: Padding(
                  padding: EdgeInsets.symmetric(horizontal: $styles.insets.md),
                  child: Text(
                    caption,
                    textAlign: TextAlign.center,
                    style: $styles.text.title1.copyWith(
                      color: $styles.colors.offWhite,
                      shadows: $styles.shadows.text,
                    ),
                  ),
                ),
              ),
            ],
          ),
        )
            .animate()
            .fadeIn(duration: $styles.times.slow)
            .scale(
              begin: const Offset(1.04, 1.04),
              end: const Offset(1, 1),
              duration: $styles.times.extraSlow,
              curve: Curves.easeOut,
            ),
      ),
    );
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: EdgeInsets.symmetric(
        horizontal: $styles.insets.md,
        vertical: $styles.insets.md,
      ),
      decoration: BoxDecoration(
        color: $styles.colors.bg,
        borderRadius: BorderRadius.circular($styles.corners.lg),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.08),
            offset: const Offset(0, 8),
            blurRadius: 24,
          ),
        ],
      ),
      child: child,
    ).maybeAnimate(delay: $styles.times.fast).fadeIn(
          duration: $styles.times.med,
        ).slideY(
          begin: 0.05,
          end: 0,
          duration: $styles.times.med,
          curve: Curves.easeOutCubic,
        );
  }
}

class _ErrorBanner extends StatelessWidget {
  const _ErrorBanner({required this.message});
  final String message;

  @override
  Widget build(BuildContext context) {
    return Semantics(
      liveRegion: true,
      child: Container(
        padding: EdgeInsets.all($styles.insets.xs),
        decoration: BoxDecoration(
          color: Colors.red.withValues(alpha: 0.08),
          borderRadius: BorderRadius.circular($styles.corners.sm),
          border: Border.all(color: Colors.red.withValues(alpha: 0.3)),
        ),
        child: Row(
          children: [
            Icon(Icons.error_outline,
                color: Theme.of(context).colorScheme.error, size: 18),
            Gap($styles.insets.xs),
            Expanded(
              child: Text(
                message,
                style: $styles.text.bodySmall.copyWith(
                  color: Theme.of(context).colorScheme.error,
                ),
                semanticsLabel: message,
              ),
            ),
          ],
        ),
      ),
    ).maybeAnimate().fadeIn(duration: $styles.times.fast).slideY(
          begin: -0.3,
          end: 0,
          duration: $styles.times.fast,
          curve: Curves.easeOutBack,
        );
  }
}

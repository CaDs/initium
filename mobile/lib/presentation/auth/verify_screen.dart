import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:gap/gap.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../motion/animate_utils.dart';
import '../../motion/celebration_particles.dart';
import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import '../../ui/app_scaffold.dart';
import '../../ui/widgets/app_btn.dart';

/// Screen shown when a magic link deep link is opened.
/// Automatically verifies the token, celebrates on success, and routes home.
class VerifyScreen extends ConsumerStatefulWidget {
  final String token;

  const VerifyScreen({super.key, required this.token});

  @override
  ConsumerState<VerifyScreen> createState() => _VerifyScreenState();
}

enum _VerifyStage { loading, celebrating, error }

class _VerifyScreenState extends ConsumerState<VerifyScreen> {
  _VerifyStage _stage = _VerifyStage.loading;
  int _shakeKey = 0;

  @override
  void initState() {
    super.initState();
    _verify();
  }

  Future<void> _verify() async {
    setState(() => _stage = _VerifyStage.loading);

    if (widget.token.isEmpty) {
      if (mounted) setState(() => _stage = _VerifyStage.error);
      return;
    }

    await ref.read(authProvider.notifier).verifyMagicLink(widget.token);
    if (!mounted) return;

    final state = ref.read(authProvider);
    if (state is AuthAuthenticated) {
      setState(() => _stage = _VerifyStage.celebrating);
      final hold = $styles.disableAnimations
          ? const Duration(milliseconds: 1)
          : const Duration(milliseconds: 1400);
      await Future<void>.delayed(hold);
      if (mounted) context.go('/home');
    } else {
      setState(() {
        _stage = _VerifyStage.error;
        _shakeKey++;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      backgroundColor: $styles.colors.bg,
      body: SafeArea(
        child: Center(
          child: Padding(
            padding: EdgeInsets.all($styles.insets.lg),
            child: switch (_stage) {
              _VerifyStage.loading => _LoadingView(),
              _VerifyStage.celebrating => _CelebrationView(message: l10n.verifyCelebration),
              _VerifyStage.error => _ErrorView(
                  shakeKey: _shakeKey,
                  title: l10n.verifyFailed,
                  subtitle: l10n.verifyExpiredOrUsed,
                  retryLabel: l10n.verifyRetry,
                  loginLabel: l10n.loginTitle,
                  onRetry: _verify,
                  onLogin: () => context.go('/login'),
                ),
            },
          ),
        ),
      ),
    );
  }
}

class _LoadingView extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 64,
          height: 64,
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: SweepGradient(
              colors: [
                $styles.colors.accent1,
                $styles.colors.accent2,
                $styles.colors.accent3,
                $styles.colors.accent1,
              ],
            ),
          ),
        )
            .maybeAnimate(onPlay: (c) => c.repeat())
            .rotate(duration: $styles.times.extraSlow)
            .shimmer(
              duration: $styles.times.slow,
              color: $styles.colors.offWhite.withValues(alpha: 0.4),
            ),
        Gap($styles.insets.md),
        Text(
          'Verifying…',
          style: $styles.text.body.copyWith(
            color: $styles.colors.fg.withValues(alpha: 0.7),
          ),
        ),
      ],
    );
  }
}

class _CelebrationView extends StatelessWidget {
  const _CelebrationView({required this.message});
  final String message;

  @override
  Widget build(BuildContext context) {
    return Stack(
      alignment: Alignment.center,
      children: [
        const Positioned.fill(
          child: CelebrationParticles(
            spriteAsset: 'assets/particles/sparkle.png',
            particleCount: 900,
            spriteFrameWidth: 21,
            spriteScale: 1.0,
            fadeMs: 1400,
            spreadFactor: 0.35,
            velocityFactor: 0.09,
          ),
        ),
        Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.check_circle_rounded,
              size: 64,
              color: $styles.colors.accent1,
            ).maybeAnimate().scale(
                  begin: const Offset(0.6, 0.6),
                  end: const Offset(1, 1),
                  duration: $styles.times.fast,
                  curve: Curves.easeOutBack,
                ),
            Gap($styles.insets.sm),
            Text(
              message,
              textAlign: TextAlign.center,
              style: $styles.text.h2.copyWith(color: $styles.colors.fg),
              semanticsLabel: message,
            ).maybeAnimate(delay: const Duration(milliseconds: 150)).fadeIn(
                  duration: $styles.times.fast,
                ),
          ],
        ),
      ],
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({
    required this.shakeKey,
    required this.title,
    required this.subtitle,
    required this.retryLabel,
    required this.loginLabel,
    required this.onRetry,
    required this.onLogin,
  });

  final int shakeKey;
  final String title;
  final String subtitle;
  final String retryLabel;
  final String loginLabel;
  final VoidCallback onRetry;
  final VoidCallback onLogin;

  @override
  Widget build(BuildContext context) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.error_outline,
            size: 56,
            color: Theme.of(context).colorScheme.error),
        Gap($styles.insets.sm),
        Text(
          title,
          style: $styles.text.h3.copyWith(color: $styles.colors.fg),
          textAlign: TextAlign.center,
          semanticsLabel: title,
        ),
        Gap($styles.insets.xs),
        Text(
          subtitle,
          textAlign: TextAlign.center,
          style: $styles.text.bodySmall.copyWith(
            color: $styles.colors.fg.withValues(alpha: 0.7),
          ),
        ),
        Gap($styles.insets.lg),
        SizedBox(
          width: double.infinity,
          child: AppBtn.from(
            key: ValueKey('retry-$shakeKey'),
            onPressed: onRetry,
            text: retryLabel,
            semanticLabel: retryLabel,
            minimumSize: const Size(double.infinity, 52),
          ),
        )
            .maybeAnimate(key: ValueKey('shake-$shakeKey'))
            .shake(
              duration: $styles.times.med,
              hz: 4,
              offset: const Offset(5, 0),
            ),
        Gap($styles.insets.xs),
        AppBtn.basic(
          onPressed: onLogin,
          semanticLabel: loginLabel,
          child: Text(
            loginLabel,
            style: $styles.text.bodySmallBold.copyWith(
              color: $styles.colors.accent1,
            ),
          ),
        ),
      ],
    );
  }
}

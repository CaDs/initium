import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:gap/gap.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../domain/entity/user.dart';
import '../../motion/animate_utils.dart';
import '../../motion/scroll_linked_parallax.dart';
import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';
import '../../ui/app_scaffold.dart';
import '../../ui/widgets/app_btn.dart';
import '../../ui/widgets/app_header.dart';
import '../../ui/widgets/compass_divider.dart';
import '../../ui/widgets/curved_clippers.dart';
import '../../ui/widgets/gradient_container.dart';
import '../../ui/widgets/scroll_decorator.dart';
import '../shared/dev_mode_banner.dart';
import '../shared/locale_switcher.dart';
import '../shared/theme_switcher.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  final ValueNotifier<double> _scrollPos = ValueNotifier(0);

  @override
  void dispose() {
    _scrollPos.dispose();
    super.dispose();
  }

  void _onScroll(ScrollController c) {
    c.addListener(() {
      if (c.hasClients) _scrollPos.value = c.offset;
    });
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;

    if (authState is! AuthAuthenticated) {
      return Scaffold(
        backgroundColor: $styles.colors.bg,
        body: Center(
          child: Semantics(
            label: 'Loading',
            child: CircularProgressIndicator(color: $styles.colors.accent1),
          ),
        ),
      );
    }

    final user = authState.user;

    return Scaffold(
      backgroundColor: $styles.colors.bg,
      appBar: AppHeader(
        title: l10n.appName,
        showBackBtn: false,
        trailing: const [
          LocaleSwitcher(),
          ThemeSwitcher(),
        ],
      ),
      body: Column(
        children: [
          const DevModeBanner(),
          Expanded(
            child: ScrollDecorator.shadow(
              onInit: _onScroll,
              builder: (controller) => ListView(
                controller: controller,
                padding: EdgeInsets.zero,
                children: [
                  _Hero(user: user, scrollPos: _scrollPos, l10n: l10n),
                  Padding(
                    padding: EdgeInsets.symmetric(horizontal: $styles.insets.md),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Gap($styles.insets.md),
                        _ProfileCard(user: user, l10n: l10n),
                        Gap($styles.insets.lg),
                        CompassDivider(
                          isExpanded: true,
                          centerSvgAsset: 'assets/icons/compass.svg',
                          centerSize: 28,
                        ),
                        Gap($styles.insets.lg),
                        _LogoutButton(
                          label: l10n.logout,
                          onPressed: () =>
                              ref.read(authProvider.notifier).logout(),
                        ),
                        Gap($styles.insets.xxl),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _Hero extends StatelessWidget {
  const _Hero({required this.user, required this.scrollPos, required this.l10n});

  final User user;
  final ValueNotifier<double> scrollPos;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    final initials = _initials(user);
    final welcome = user.name.isNotEmpty
        ? l10n.homeWelcomeUser(user.name)
        : l10n.homeWelcome;

    return ScrollParallaxTranslate(
      scrollPos: scrollPos,
      startOffset: Offset.zero,
      endOffset: const Offset(0, -40),
      child: SizedBox(
        height: 260,
        child: Stack(
          children: [
            Positioned.fill(
              child: SvgPicture.asset(
                'assets/illustrations/home/hero_bg.svg',
                fit: BoxFit.cover,
              ),
            ),
            Positioned.fill(
              child: VtGradient(
                [
                  Colors.transparent,
                  $styles.colors.bg.withValues(alpha: 0.85),
                ],
                const [0.55, 1],
              ),
            ),
            Positioned.fill(
              child: Padding(
                padding: EdgeInsets.all($styles.insets.md),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    _Avatar(initials: initials),
                    Gap($styles.insets.sm),
                    Text(
                      welcome,
                      textAlign: TextAlign.center,
                      style: $styles.text.h3.copyWith(
                        color: $styles.colors.fg,
                        shadows: $styles.shadows.textSoft,
                      ),
                    ).maybeAnimate().fadeIn(duration: $styles.times.med),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _initials(User u) {
    final src = u.name.isNotEmpty ? u.name : u.email;
    final parts = src.trim().split(RegExp(r'[\s@._-]+'));
    final letters = parts
        .where((p) => p.isNotEmpty)
        .take(2)
        .map((p) => p.characters.first.toUpperCase())
        .join();
    return letters.isEmpty ? '•' : letters;
  }
}

class _Avatar extends StatelessWidget {
  const _Avatar({required this.initials});
  final String initials;

  @override
  Widget build(BuildContext context) {
    return ClipPath(
      clipper: const ArchClipper(ArchType.spade),
      child: Container(
        width: 96,
        height: 120,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              $styles.colors.accent1,
              $styles.colors.accent3,
            ],
          ),
        ),
        alignment: Alignment.center,
        child: Text(
          initials,
          style: $styles.text.h2.copyWith(
            color: $styles.colors.offWhite,
            shadows: $styles.shadows.textSoft,
          ),
          semanticsLabel: 'Avatar',
        ),
      ),
    )
        .maybeAnimate()
        .fadeIn(duration: $styles.times.med)
        .scale(
          begin: const Offset(0.8, 0.8),
          end: const Offset(1, 1),
          duration: $styles.times.med,
          curve: Curves.easeOutBack,
        );
  }
}

class _ProfileCard extends StatelessWidget {
  const _ProfileCard({required this.user, required this.l10n});
  final User user;
  final AppLocalizations l10n;

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: EdgeInsets.all($styles.insets.md),
      decoration: BoxDecoration(
        color: $styles.colors.bg,
        borderRadius: BorderRadius.circular($styles.corners.lg),
        border: Border.all(
          color: $styles.colors.greySoft.withValues(alpha: 0.4),
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            offset: const Offset(0, 4),
            blurRadius: 16,
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.homeProfile,
            style: $styles.text.title1.copyWith(color: $styles.colors.fg),
          ),
          Gap($styles.insets.sm),
          _ProfileRow(label: l10n.labelEmail, value: user.email),
          _ProfileRow(label: l10n.labelName,
              value: user.name.isNotEmpty ? user.name : '—'),
          _ProfileRow(label: l10n.labelUserId, value: user.id),
        ],
      ),
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

class _ProfileRow extends StatelessWidget {
  const _ProfileRow({required this.label, required this.value});
  final String label;
  final String value;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: $styles.insets.xxs),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 88,
            child: Text(
              label,
              style: $styles.text.bodySmall.copyWith(
                color: $styles.colors.greyMedium,
              ),
            ),
          ),
          Expanded(
            child: Semantics(
              label: '$label: $value',
              child: Text(
                value,
                style: $styles.text.body.copyWith(color: $styles.colors.fg),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _LogoutButton extends StatelessWidget {
  const _LogoutButton({required this.label, required this.onPressed});
  final String label;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      width: double.infinity,
      child: AppBtn.from(
        onPressed: onPressed,
        text: label,
        semanticLabel: label,
        icon: Icon(Icons.logout, size: 18, color: $styles.colors.fg),
        isSecondary: true,
        minimumSize: const Size(double.infinity, 52),
        border: BorderSide(
          color: $styles.colors.greySoft.withValues(alpha: 0.6),
        ),
      ),
    );
  }
}


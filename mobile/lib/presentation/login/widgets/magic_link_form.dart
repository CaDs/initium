import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:gap/gap.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../../motion/animate_utils.dart';
import '../../../providers/api_provider.dart';
import '../../../ui/app_scaffold.dart';
import '../../../ui/widgets/app_btn.dart';

class MagicLinkForm extends ConsumerStatefulWidget {
  const MagicLinkForm({super.key});

  @override
  ConsumerState<MagicLinkForm> createState() => _MagicLinkFormState();
}

class _MagicLinkFormState extends ConsumerState<MagicLinkForm> {
  final _emailController = TextEditingController();
  bool _loading = false;
  bool _sent = false;

  @override
  void dispose() {
    _emailController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (_emailController.text.isEmpty) return;
    setState(() => _loading = true);

    final success = await ref.read(authProvider.notifier).requestMagicLink(
          _emailController.text.trim(),
        );
    if (!mounted) return;
    setState(() {
      _loading = false;
      _sent = success;
    });
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    if (_sent) {
      return Semantics(
        liveRegion: true,
        child: Container(
          padding: EdgeInsets.all($styles.insets.sm),
          decoration: BoxDecoration(
            color: $styles.colors.accent1.withValues(alpha: 0.14),
            borderRadius: BorderRadius.circular($styles.corners.md),
            border: Border.all(
              color: $styles.colors.accent1.withValues(alpha: 0.35),
            ),
          ),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                l10n.loginMagicSent,
                style: $styles.text.bodyBold.copyWith(color: $styles.colors.fg),
                textAlign: TextAlign.center,
              ),
              Gap($styles.insets.xxs),
              Text(
                l10n.loginMagicSentDetail,
                style: $styles.text.bodySmall.copyWith(
                  color: $styles.colors.fg.withValues(alpha: 0.75),
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ).maybeAnimate().fadeIn(duration: $styles.times.fast).scale(
              begin: const Offset(0.95, 0.95),
              end: const Offset(1, 1),
              duration: $styles.times.fast,
            ),
      );
    }

    return Column(
      children: [
        Semantics(
          label: l10n.loginMagicPlaceholder,
          textField: true,
          child: TextField(
            controller: _emailController,
            keyboardType: TextInputType.emailAddress,
            autofillHints: const [AutofillHints.email],
            textInputAction: TextInputAction.send,
            onSubmitted: (_) => _submit(),
            style: $styles.text.body.copyWith(color: $styles.colors.fg),
            decoration: InputDecoration(
              hintText: l10n.loginMagicPlaceholder,
              hintStyle: $styles.text.body.copyWith(
                color: $styles.colors.fg.withValues(alpha: 0.45),
              ),
              filled: true,
              fillColor: $styles.colors.bg.withValues(alpha: 0.85),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular($styles.corners.md),
                borderSide: BorderSide(
                  color: $styles.colors.greySoft.withValues(alpha: 0.5),
                ),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular($styles.corners.md),
                borderSide: BorderSide(
                  color: $styles.colors.greySoft.withValues(alpha: 0.5),
                ),
              ),
              focusedBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular($styles.corners.md),
                borderSide: BorderSide(color: $styles.colors.accent1, width: 1.5),
              ),
              contentPadding: EdgeInsets.symmetric(
                horizontal: $styles.insets.sm,
                vertical: $styles.insets.xs + 2,
              ),
            ),
          ),
        ),
        Gap($styles.insets.xs),
        SizedBox(
          width: double.infinity,
          child: AppBtn.from(
            onPressed: _loading ? null : _submit,
            text: _loading ? l10n.loginMagicSending : l10n.loginMagicSubmit,
            semanticLabel: l10n.loginMagicSubmit,
            expand: true,
            minimumSize: const Size(double.infinity, 52),
          ),
        ),
      ],
    );
  }
}

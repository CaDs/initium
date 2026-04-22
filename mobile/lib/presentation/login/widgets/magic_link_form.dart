import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../../providers/api_provider.dart';

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
        child: Text(l10n.loginMagicSent),
      );
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        TextField(
          controller: _emailController,
          keyboardType: TextInputType.emailAddress,
          autofillHints: const [AutofillHints.email],
          textInputAction: TextInputAction.send,
          onSubmitted: (_) => _submit(),
          decoration: InputDecoration(
            labelText: l10n.loginMagicPlaceholder,
            border: const OutlineInputBorder(),
          ),
        ),
        const SizedBox(height: 8),
        ElevatedButton(
          onPressed: _loading ? null : _submit,
          child: Text(_loading ? l10n.loginMagicSending : l10n.loginMagicSubmit),
        ),
      ],
    );
  }
}

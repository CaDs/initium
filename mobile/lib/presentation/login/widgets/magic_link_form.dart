import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
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

    setState(() {
      _loading = false;
      _sent = success;
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_sent) {
      return Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.green[50],
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          children: [
            Text(
              'Check your email!',
              style: TextStyle(color: Colors.green[700], fontWeight: FontWeight.w600),
            ),
            const SizedBox(height: 4),
            Text(
              'A magic link has been sent.',
              style: TextStyle(color: Colors.green[600], fontSize: 13),
            ),
          ],
        ),
      );
    }

    return Column(
      children: [
        TextField(
          controller: _emailController,
          keyboardType: TextInputType.emailAddress,
          decoration: InputDecoration(
            hintText: 'Enter your email',
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 14),
          ),
        ),
        const SizedBox(height: 12),
        FilledButton(
          onPressed: _loading ? null : _submit,
          style: FilledButton.styleFrom(
            minimumSize: const Size(double.infinity, 52),
          ),
          child: _loading
              ? const SizedBox(
                  height: 20,
                  width: 20,
                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                )
              : const Text('Send Magic Link', style: TextStyle(fontSize: 16)),
        ),
      ],
    );
  }
}

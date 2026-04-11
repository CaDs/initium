import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

class GoogleSignInButton extends ConsumerWidget {
  const GoogleSignInButton({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return OutlinedButton(
      onPressed: () async {
        // TODO: Integrate google_sign_in package
        // final googleUser = await GoogleSignIn(scopes: ['email', 'profile']).signIn();
        // if (googleUser == null) return;
        // final auth = await googleUser.authentication;
        // ref.read(authProvider.notifier).loginWithGoogle(auth.idToken!);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Google Sign-In: configure google-services.json first')),
        );
      },
      style: OutlinedButton.styleFrom(
        minimumSize: const Size(double.infinity, 52),
        side: const BorderSide(color: Colors.grey),
      ),
      child: const Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.g_mobiledata, size: 24),
          SizedBox(width: 8),
          Text('Sign in with Google', style: TextStyle(fontSize: 16)),
        ],
      ),
    );
  }
}

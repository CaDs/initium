import 'package:flutter/material.dart';
import '../../providers/api_provider.dart';

class DevModeBanner extends StatelessWidget {
  const DevModeBanner({super.key});

  @override
  Widget build(BuildContext context) {
    if (!isDevBypassAuth) return const SizedBox.shrink();

    return Container(
      width: double.infinity,
      color: Colors.amber[100],
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Text(
        'Dev Mode: Logged in as dev@initium.local',
        textAlign: TextAlign.center,
        style: TextStyle(fontSize: 12, color: Colors.amber[900]),
      ),
    );
  }
}

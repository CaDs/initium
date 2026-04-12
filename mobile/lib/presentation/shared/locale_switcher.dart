import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../providers/locale_provider.dart';

class LocaleSwitcher extends ConsumerWidget {
  const LocaleSwitcher({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final currentLocale = ref.watch(localeProvider);
    final effectiveCode = currentLocale?.languageCode ??
        Localizations.localeOf(context).languageCode;

    return PopupMenuButton<Locale>(
      icon: const Icon(Icons.language),
      tooltip: 'Language',
      onSelected: (locale) => ref.read(localeProvider.notifier).setLocale(locale),
      itemBuilder: (context) => supportedLocales.map((locale) {
        final code = locale.languageCode;
        final label = localeLabels[code] ?? code;
        return PopupMenuItem(
          value: locale,
          child: Row(
            children: [
              if (code == effectiveCode)
                Icon(Icons.check, size: 18, color: Theme.of(context).colorScheme.primary)
              else
                const SizedBox(width: 18),
              const SizedBox(width: 8),
              Text(label),
            ],
          ),
        );
      }).toList(),
    );
  }
}

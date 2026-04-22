# Screen pattern

Screens live in `presentation/<feature>/`. Default to `ConsumerWidget` (or
`ConsumerStatefulWidget` when you need local state), not plain `StatelessWidget`.

## Skeleton

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mobile/l10n/app_localizations.dart';

import '../../providers/api_provider.dart';
import '../../providers/auth_provider.dart';

class OrdersScreen extends ConsumerWidget {
  const OrdersScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);
    final l10n = AppLocalizations.of(context)!;

    if (authState is! AuthAuthenticated) {
      return const Scaffold(body: Center(child: CircularProgressIndicator.adaptive()));
    }

    return Scaffold(
      appBar: AppBar(title: Text(l10n.ordersTitle)),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // raw Material widgets — no AppBtn/AppHeader wrappers
        ],
      ),
    );
  }
}
```

## Rules

- Use raw Material widgets (`Scaffold`, `AppBar`, `ElevatedButton`,
  `OutlinedButton`, `TextField`). This template deliberately has no wrapper
  widget layer.
- Theme: `Theme.of(context).colorScheme` + `theme.textTheme`. Never
  `Colors.grey[500]` or ad-hoc `TextStyle`.
- i18n: every user-visible string comes from `l10n.xxx`. Add the key to all
  three ARB files, run `make gen:mobile`.
- Accessibility: `Semantics` on interactive elements, `tooltip` on
  `IconButton`s and icon-bearing buttons, `semanticsLabel` on icons that
  aren't decorative. Use `autofillHints` on text fields ONLY when a category
  applies (`AutofillHints.email`, `.password`, `.username`, `.newPassword`,
  `.oneTimeCode`) — omit for free-form fields like titles, bodies, or search
  queries.
- Keep screens flat. Extract a sub-widget (`class _Card extends StatelessWidget`)
  when the same pattern appears twice. Don't extract prematurely.

## Routing

- New routes go in `presentation/router/app_router.dart`.
- Decide whether the route needs an auth-based redirect; add it to the
  `redirect:` callback.
- Navigation: `context.push('/path')` for detail / drill-down screens that
  should preserve the back-stack (home → orders, orders → order detail).
  `context.go('/path')` for redirects and top-level nav replacements (login →
  home, logout → login, deep-link handoff in `verify_screen.dart`). Never
  `Navigator.push`.

## When adding a screen that mirrors a web page

Match the web page's behavior as closely as possible: same endpoint, same
empty/loading/error states, same i18n namespace. See `_shared/parity.md`.

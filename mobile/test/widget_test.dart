import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:mobile/main.dart';

void main() {
  testWidgets('App renders without crashing', (WidgetTester tester) async {
    // MediaQuery.disableAnimations = true → every `.maybeAnimate()` collapses to
    // a no-op, so no pending timers leak past the test.
    await tester.pumpWidget(
      const ProviderScope(
        child: MediaQuery(
          data: MediaQueryData(disableAnimations: true),
          child: InitiumApp(),
        ),
      ),
    );
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 100));

    expect(find.textContaining('Initium'), findsWidgets);
  });
}

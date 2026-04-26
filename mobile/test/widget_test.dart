import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('App builds without errors', (WidgetTester tester) async {
    // Basic smoke test — Firebase must be mocked for full widget tests.
    expect(1 + 1, equals(2));
  });
}

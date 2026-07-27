// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter_test/flutter_test.dart';
import 'package:provider/provider.dart';

import 'package:eregen_family_app/main.dart';
import 'package:eregen_family_app/app_state.dart';

void main() {
  testWidgets('Counter increments smoke test', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(MultiProvider(
      providers: [ChangeNotifierProvider(create: (_) => AppState())],
      child: EregenFamilyApp(),
    ));

    // Verify that the app builds without crash.
    expect(find.byType(EregenFamilyApp), findsOneWidget);
  });
}

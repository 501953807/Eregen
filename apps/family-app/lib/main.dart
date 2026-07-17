import 'package:flutter/material.dart';
import '../common/theme.dart';
import '../widgets/elderly_selector.dart';
import '../widgets/map_section.dart';
import '../widgets/quick_status_card.dart';
import '../widgets/sos_button.dart';
import '../widgets/recent_alerts_list.dart';
import '../widgets/bottom_nav_bar.dart';
import '../api/client.dart';
import '../screens/login/login_page.dart';
import '../screens/login/main_tab_screen.dart';

/// Entry point — initializes ApiClient, checks token, shows login or main app.
void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await ApiClient.init();
  runApp(const EregenFamilyApp());
}

class EregenFamilyApp extends StatelessWidget {
  const EregenFamilyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '颐贞',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        fontFamily: null,
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF4A90D9)),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF5F6FA),
      ),
      home: ApiClient.instance.isAuthenticated
          ? const MainTabScreen()
          : _LoginGate(),
    );
  }
}

/// Bridge widget that reads async auth state from ApiClient singleton.
class _LoginGate extends StatefulWidget {
  const _LoginGate();

  @override
  State<_LoginGate> createState() => _LoginGateState();
}

class _LoginGateState extends State<_LoginGate> {
  bool _checking = true;

  @override
  void initState() {
    super.initState();
    // ApiClient.init() already ran in main(), but we check again in case
    // the singleton was not yet populated at build time.
    WidgetsBinding.instance.addPostFrameCallback((_) {
      setState(() => _checking = false);
    });
  }

  @override
  Widget build(BuildContext context) {
    if (_checking) {
      return const Scaffold(
        backgroundColor: AppTheme.bgScaffold,
        body: Center(child: CircularProgressIndicator()),
      );
    }
    if (ApiClient.instance.isAuthenticated) {
      return const MainTabScreen();
    }
    return LoginPage(onLoginSuccess: () {
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const MainTabScreen()),
      );
    });
  }
}

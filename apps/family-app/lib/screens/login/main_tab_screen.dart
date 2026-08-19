import 'package:flutter/material.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../home/home_page.dart';
import '../health/health_page.dart';
import '../alerts/alerts_page.dart';
import '../medication/medication_page.dart';
import '../chronic/chronic_home_page.dart';
import '../settings/settings_page.dart';
import '../welfare_page.dart';

/// Post-login bottom-tab shell — 5 prototype pages + chronic care.
class MainTabScreen extends StatefulWidget {
  const MainTabScreen({super.key});

  @override
  State<MainTabScreen> createState() => _MainTabScreenState();
}

class _MainTabScreenState extends State<MainTabScreen> {
  int _currentIndex = 0;

  final List<Widget> _pages = const [
    HomePage(),
    HealthPage(),
    AlertsPage(),
    MedicationPage(),
    ChronicHomePage(),
    SettingsPage(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _currentIndex, children: _pages),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _currentIndex,
        onTabSelected: (i) => setState(() => _currentIndex = i),
        onSpecialTab: (idx) {
          if (idx == 6) {
            Navigator.of(context).push(MaterialPageRoute(builder: (_) => const WelfarePage()));
          }
        },
      ),
    );
  }
}

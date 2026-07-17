import 'package:flutter/material.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../home/home_page.dart';
import '../health/health_page.dart';
import '../alerts/alerts_page.dart';
import '../medication/medication_page.dart';
import '../settings/settings_page.dart';

/// Post-login bottom-tab shell — 4 prototype pages.
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
    SettingsPage(),
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _currentIndex, children: _pages),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _currentIndex,
        onTabSelected: (i) => setState(() => _currentIndex = i),
      ),
    );
  }
}

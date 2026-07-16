import 'package:flutter/material.dart';
import 'screens/home/home_page.dart';
import 'screens/health/health_page.dart';
import 'screens/alerts/alerts_page.dart';
import 'screens/medication/medication_page.dart';

void main() {
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
        fontFamily: null, // Use system font (PingFang SC on iOS)
        colorScheme: ColorScheme.fromSeed(seedColor: const Color(0xFF4A90D9)),
        useMaterial3: true,
        scaffoldBackgroundColor: const Color(0xFFF5F6FA),
      ),
      home: const MainTabScreen(),
    );
  }
}

/// Main tab screen — switches between the 4 prototype pages
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
  ];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: IndexedStack(index: _currentIndex, children: _pages),
    );
  }
}

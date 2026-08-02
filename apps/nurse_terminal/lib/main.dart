import 'package:flutter/material.dart';
import 'src/screens/login_screen.dart';
import 'src/screens/home_screen.dart';
import 'src/services/medical_wristband_ble_service.dart';
import 'common/theme.dart'; // Import unified theme

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const NurseTerminalApp());
}

class NurseTerminalApp extends StatelessWidget {
  const NurseTerminalApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: '颐贞 护士终端',
      theme: ThemeData.useMaterial3.copyWith(
        colorSchemeSeed: NurseTerminalTheme.primary, // Use amber instead of blue
        brightness: Brightness.light,
        fontFamily: 'PingFang SC', // Match web font stack
        elevationScale: 1.5,
        primaryColor: NurseTerminalTheme.primary,
        scaffoldBackground: NurseTerminalTheme.bgScaffold,
        cardTheme: CardTheme(
          margin: EdgeInsets.zero,
          padding: EdgeInsets.all(NurseTerminalTheme.spacingM),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(NurseTerminalTheme.radiusLarge)),
          elevation: 2,
        ),
        elevatedButtonTheme: NurseTerminalTheme.elevatedButtonTheme,
        textButtonTheme: NurseTerminalTheme.textButtonTheme,
        iconButtonTheme: NurseTerminalTheme.iconButtonTheme,
      ),
      darkTheme: ThemeData.brightness.copyWith(
        colorSchemeSeed: NurseTerminalTheme.primary,
        brightness: Brightness.dark,
      ),
      home: const LoginScreen(),
      routes: {
        '/': (_) => const LoginScreen(),
        '/home': (_) => const HomeScreen(),
        '/ble-scan': (_) => BleScanPage(),
      },
    );
  }
}

class BleScanPage extends StatefulWidget {
  const BleScanPage({super.key});

  @override
  State<BleScanPage> createState() => _BleScanPageState();
}

class _BleScanPageState extends State<BleScanPage> {
  final MedicalWristbandService _bleService = MedicalWristbandService();
  final List<String> _log = [];

  @override
  void dispose() {
    _bleService.dispose();
    super.dispose();
  }

  void _addLog(String msg) {
    setState(() => _log.insert(0, '${DateTime.now().toIso8601String().substring(11, 19)} $msg'));
  }

  Future<void> _startScan() async {
    _addLog('Starting BLE scan...');
    await _bleService.startScan();
  }

  Future<void> _stopScan() async {
    _addLog('Stopping BLE scan...');
    await _bleService.stopScan();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('颐贞 护士终端'),
        elevation: 2,
      ),
      body: Padding(
        padding: EdgeInsets.all(NurseTerminalTheme.spacingM),
        child: Column(
          children: [
            Container(
              decoration: NurseTerminalTheme.cardDecoration,
              padding: EdgeInsets.all(NurseTerminalTheme.spacingM),
              child: Column(
                children: [
                  ElevatedButton.icon(
                    onPressed: _startScan,
                    icon: const Icon(Icons.bluetooth_searching),
                    label: const Text('开始扫描'),
                  ),
                  const SizedBox(height: NurseTerminalTheme.spacingS),
                  ElevatedButton.icon(
                    onPressed: _stopScan,
                    icon: const Icon(Icons.bluetooth_disabled),
                    label: const Text('停止扫描'),
                  ),
                ],
              ),
            ),
            const SizedBox(height: NurseTerminalTheme.spacingL),
            Expanded(
              child: ListView.builder(
                itemCount: _log.length,
                itemBuilder: (context, i) => ListTile(
                  dense: true,
                  title: Text(_log[i], style: NurseTerminalTheme.smallLabelStyle),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
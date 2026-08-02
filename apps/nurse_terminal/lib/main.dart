import 'package:flutter/material.dart';
import 'src/screens/login_screen.dart';
import 'src/screens/home_screen.dart';
import 'src/screens/verification_screen.dart';
import 'src/services/medical_wristband_ble_service.dart';
import 'common/theme.dart';

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
      theme: ThemeData(
        colorSchemeSeed: NurseTerminalTheme.primary,
        brightness: Brightness.light,
        fontFamily: 'PingFang SC',
        primaryColor: NurseTerminalTheme.primary,
        scaffoldBackgroundColor: NurseTerminalTheme.bgScaffold,
        cardTheme: CardThemeData(
          margin: EdgeInsets.zero,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(NurseTerminalTheme.radiusLarge),
          ),
          elevation: 2,
        ),
        elevatedButtonTheme: NurseTerminalTheme.elevatedButtonTheme,
        textButtonTheme: NurseTerminalTheme.textButtonTheme,
        iconButtonTheme: NurseTerminalTheme.iconButtonTheme,
      ),
      darkTheme: ThemeData.dark(),
      themeMode: ThemeMode.light,
      home: const LoginScreen(),
      routes: {
        '/': (_) => const LoginScreen(),
        '/home': (_) => const HomeScreen(),
        '/nfc-scan': (_) => const NfcScanPage(),
      },
    );
  }
}

class NfcScanPage extends StatefulWidget {
  const NfcScanPage({super.key});

  @override
  State<NfcScanPage> createState() => _NfcScanPageState();
}

class _NfcScanPageState extends State<NfcScanPage> {
  final MedicalWristbandService _nfcService = MedicalWristbandService();
  final List<String> _log = [];

  @override
  void dispose() {
    _nfcService.dispose();
    super.dispose();
  }

  void _addLog(String msg) {
    setState(() => _log.insert(0, '${DateTime.now().toIso8601String().substring(11, 19)} $msg'));
  }

  Future<void> _startScan() async {
    if (!mounted) return;
    _addLog('Starting NFC scan...');
    try {
      final message = await _nfcService.scanWristband();
      if (message != null) {
        _addLog('NFC tag read successfully');
        final patient = _nfcService.parsePatientInfo(message);
        if (patient != null) {
          _addLog('Patient: ${patient.patientId}');
          if (mounted) {
            final result = await Navigator.push(
              context,
              MaterialPageRoute(
                builder: (_) => VerificationScreen(
                  patientId: patient.patientId,
                  patientName: patient.name,
                ),
              ),
            );
            if (result == true) {
              _addLog('Verification saved');
            }
          }
        }
      } else {
        _addLog('No wristband detected');
      }
    } catch (e) {
      _addLog('NFC error: $e');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('颐贞 护士终端'),
        elevation: 2,
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Container(
              decoration: const BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.all(Radius.circular(16)),
                boxShadow: [BoxShadow(color: Colors.black12, blurRadius: 4, offset: Offset(0, 2))],
              ),
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  ElevatedButton.icon(
                    onPressed: _startScan,
                    icon: const Icon(Icons.nfc),
                    label: const Text('NFC 读取腕带'),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 24),
            Expanded(
              child: ListView.builder(
                itemCount: _log.length,
                itemBuilder: (context, i) => ListTile(
                  dense: true,
                  title: Text(_log[i], style: const TextStyle(fontSize: 12, color: Colors.grey)),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

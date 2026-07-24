import 'package:flutter/material.dart';
import 'package:nurse_terminal/src/services/medical_wristband_ble_service.dart';

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
        colorSchemeSeed: Colors.blue,
        useMaterial3: true,
        brightness: Brightness.light,
      ),
      darkTheme: ThemeData(
        colorSchemeSeed: Colors.blue,
        useMaterial3: true,
        brightness: Brightness.dark,
      ),
      home: const BleScanPage(),
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
      appBar: AppBar(title: const Text('颐贞 护士终端')),
      body: Column(
        children: [
          Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                ElevatedButton.icon(
                  onPressed: _startScan,
                  icon: const Icon(Icons.bluetooth_searching),
                  label: const Text('开始扫描'),
                ),
                const SizedBox(height: 8),
                ElevatedButton.icon(
                  onPressed: _stopScan,
                  icon: const Icon(Icons.bluetooth_disabled),
                  label: const Text('停止扫描'),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView.builder(
              itemCount: _log.length,
              itemBuilder: (context, i) => ListTile(
                dense: true,
                title: Text(_log[i], style: const TextStyle(fontSize: 12)),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

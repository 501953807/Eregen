import 'package:flutter/material.dart';
import 'package:flutter_blue_plus/flutter_blue_plus.dart';
import 'package:uuid/uuid.dart';
import '../services/api_client.dart';
import '../services/verification_service.dart';
import '../services/medical_wristband_ble_service.dart';
import '../models/medical_models.dart';

/// Verification screen: scans medical wristband via BLE, displays result,
/// and allows the nurse to confirm and save the verification record.
class VerificationScreen extends StatefulWidget {
  final String patientId;
  final String patientName;

  const VerificationScreen({
    super.key,
    required this.patientId,
    required this.patientName,
  });

  @override
  State<VerificationScreen> createState() => _VerificationScreenState();
}

class _VerificationScreenState extends State<VerificationScreen> {
  final ApiClient _api = ApiClient();
  late final VerificationService _verificationService;
  final MedicalWristbandService _bleService = MedicalWristbandService();

  VerificationResult? _result;
  bool _scanning = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _verificationService = VerificationService(_api);
  }

  @override
  void dispose() {
    _bleService.dispose();
    super.dispose();
  }

  Future<void> _startScan() async {
    setState(() {
      _scanning = true;
      _error = null;
      _result = null;
    });

    try {
      // Start BLE scan for wristband devices
      await FlutterBluePlus.startScan(
        timeout: const Duration(seconds: 15),
        androidUsesFineLocation: true,
      );

      // Wait for a device to be discovered
      BluetoothDevice? targetDevice;
      final subscription = FlutterBluePlus.scanResults.listen((results) {
        for (final r in results) {
          final name = r.device.platformName;
          if (name.contains('Eregen') || name.contains('WB-')) {
            targetDevice = r.device;
          }
        }
      });

      // Wait up to 20 seconds for a device
      await Future.delayed(const Duration(seconds: 20));
      await subscription.cancel();
      await FlutterBluePlus.stopScan();

      if (targetDevice == null) {
        if (mounted) {
          setState(() {
            _scanning = false;
            _error = '未找到腕带设备，请确保腕带已开机并在附近';
          });
        }
        return;
      }

      // Connect to the discovered device
      setState(() => _error = '正在连接腕带...');
      final connected = await _bleService.connect(targetDevice!);
      if (!connected) {
        if (mounted) {
          setState(() {
            _scanning = false;
            _error = '腕带连接失败，请重试';
          });
        }
        return;
      }

      // Read patient info from wristband
      setState(() => _error = '正在读取腕带信息...');
      final info = await _bleService.readPatientInfo();
      if (info == null) {
        if (mounted) {
          setState(() {
            _scanning = false;
            _error = '读取腕带信息失败，请重试';
          });
        }
        return;
      }

      // Send verification request through BLE
      setState(() => _error = '正在核验...');
      final requestId = const Uuid().v4();
      final verificationResult = await _bleService.sendVerificationRequest(
        requestId: requestId,
        scanType: 'nurse_scan',
        patientId: info.patientId,
      );

      await _bleService.disconnect();

      if (mounted) {
        setState(() {
          _scanning = false;
          if (verificationResult != null) {
            _result = verificationResult;
          } else {
            _error = '核验请求超时或腕带返回为空';
          }
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _scanning = false;
          _error = '扫描失败: $e';
        });
      }
    }
  }

  Future<void> _confirm() async {
    if (_result == null) return;
    try {
      await _verificationService.create(_result!.toJson());
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('核验记录已保存')),
        );
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('保存失败: $e')),
        );
      }
    }
  }

  // ── Build ───────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('腕带扫描核验'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Current patient banner
            Card(
              color: Colors.blue.shade50,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    const Icon(Icons.person, color: Colors.blue),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            '当前患者',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey,
                            ),
                          ),
                          Text(
                            '${widget.patientName} (ID: ${widget.patientId})',
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                              fontSize: 16,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Scan button or result
            if (_scanning)
              Center(
                child: Column(
                  children: [
                    const CircularProgressIndicator(),
                    const SizedBox(height: 12),
                    Text(_error ?? '正在扫描腕带...'),
                  ],
                ),
              )
            else if (_error != null && _result == null)
              Card(
                color: Colors.red.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    children: [
                      Row(
                        children: [
                          const Icon(Icons.error_outline, color: Colors.red),
                          const SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              _error!,
                              style: const TextStyle(color: Colors.red),
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),
                      ElevatedButton.icon(
                        onPressed: _startScan,
                        icon: const Icon(Icons.refresh),
                        label: const Text('重新扫描'),
                      ),
                    ],
                  ),
                ),
              )
            else if (_result != null)
              _buildResultCard()
            else
              Center(
                child: Column(
                  children: [
                    const Icon(
                      Icons.qr_code_scanner,
                      size: 80,
                      color: Colors.blue,
                    ),
                    const SizedBox(height: 16),
                    const Text(
                      '点击下方按钮开始扫描医疗腕带',
                      style: TextStyle(color: Colors.grey),
                    ),
                    const SizedBox(height: 24),
                    ElevatedButton.icon(
                      onPressed: _startScan,
                      icon: const Icon(Icons.bluetooth_searching),
                      label: const Text('开始扫描'),
                      style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 32,
                          vertical: 16,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildResultCard() {
    final r = _result!;
    Color statusColor;
    String statusText;
    IconData statusIcon;

    switch (r.result) {
      case 'matched':
        statusColor = Colors.green;
        statusText = '匹配成功';
        statusIcon = Icons.check_circle;
        break;
      case 'unmatched':
        statusColor = Colors.orange;
        statusText = '腕带与患者不匹配';
        statusIcon = Icons.warning;
        break;
      default:
        statusColor = Colors.red;
        statusText = '腕带未在系统中找到';
        statusIcon = Icons.cancel;
    }

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(statusIcon, color: statusColor, size: 28),
                const SizedBox(width: 12),
                Expanded(
                  child: Text(
                    statusText,
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: statusColor,
                    ),
                  ),
                ),
              ],
            ),
            const Divider(height: 24),
            _buildResultField('患者 ID', r.patientId),
            _buildResultField('腕带设备 ID', r.deviceDeviceId),
            _buildResultField('核验类型', r.scanType),
            _buildResultField('核验人', r.verifiedBy),
            _buildResultField('备注', r.notes.isEmpty ? '无' : r.notes),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _confirm,
                icon: const Icon(Icons.check),
                label: const Text('确认并保存核验记录'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: statusColor,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 14),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildResultField(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 80,
            child: Text(
              label,
              style: const TextStyle(
                color: Colors.grey,
                fontSize: 13,
              ),
            ),
          ),
          Expanded(
            child: Text(value, style: const TextStyle(fontSize: 14)),
          ),
        ],
      ),
    );
  }
}

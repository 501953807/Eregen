import 'package:flutter/material.dart';
import '../services/api_client.dart';
import 'verification_screen.dart';

/// Medication screen: shows today's medication schedule and allows
/// verification of each dose via NFC wristband scan.
class MedicationScreen extends StatefulWidget {
  final String patientId;

  const MedicationScreen({super.key, required this.patientId});

  @override
  State<MedicationScreen> createState() => _MedicationScreenState();
}

class _MedicationScreenState extends State<MedicationScreen> {
  final ApiClient _api = ApiClient();

  List<dynamic> _medications = [];
  final Set<int> _verifiedIndices = {};
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadMedications();
  }

  Future<void> _loadMedications() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      // Try to load from API; fall back to sample data
      final res = await _api.get(
        '/api/v1/admin/medical/patients/${widget.patientId}/medications?date=${DateTime.now().toIso8601String().substring(0, 10)}',
      );
      final data = res['data'] as List<dynamic>? ?? [];
      setState(() => _medications = data);
    } catch (e) {
      // Fall back to sample medications for UI demonstration
      setState(() {
        _error = 'API 不可用，显示示例数据';
        _medications = [
          {
            'id': '1',
            'name': '氨氯地平片',
            'dose': '5mg',
            'time': '08:00',
            'type': '口服',
            'taken': false,
          },
          {
            'id': '2',
            'name': '阿司匹林肠溶片',
            'dose': '100mg',
            'time': '12:00',
            'type': '口服',
            'taken': false,
          },
          {
            'id': '3',
            'name': '二甲双胍片',
            'dose': '500mg',
            'time': '18:00',
            'type': '口服',
            'taken': false,
          },
        ];
      });
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _verifyMedication(int index) async {
    final navResult = await Navigator.push<bool>(
      context,
      MaterialPageRoute(
        builder: (_) => VerificationScreen(
          patientId: widget.patientId,
          patientName: _medicationName(index),
        ),
      ),
    );
    if (navResult == true && mounted) {
      setState(() => _verifiedIndices.add(index));
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('${_medicationName(index)} 已确认服用'),
          duration: const Duration(seconds: 2),
        ),
      );
    }
  }

  String _medicationName(int index) {
    if (index >= _medications.length) return '药物';
    final m = _medications[index];
    return (m['name'] ?? m['medication_name'] ?? '未命名药物').toString();
  }

  String _medicationDose(int index) {
    if (index >= _medications.length) return '';
    final m = _medications[index];
    return (m['dose'] ?? '').toString();
  }

  String _medicationTime(int index) {
    if (index >= _medications.length) return '';
    final m = _medications[index];
    return (m['time'] ?? '').toString();
  }

  String _medicationType(int index) {
    if (index >= _medications.length) return '';
    final m = _medications[index];
    return (m['type'] ?? '口服').toString();
  }

  // ── Build ───────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('用药管理'),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Date header
                  Card(
                    color: Colors.blue.shade50,
                    child: Padding(
                      padding: const EdgeInsets.all(12),
                      child: Row(
                        children: [
                          const Icon(Icons.calendar_today,
                              color: Colors.blue),
                          const SizedBox(width: 8),
                          Text(
                            '${DateTime.now().year}年${DateTime.now().month}月${DateTime.now().day}日 用药计划',
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  if (_error != null) ...[
                    const SizedBox(height: 8),
                    Card(
                      color: Colors.orange.shade50,
                      child: Padding(
                        padding: const EdgeInsets.all(8),
                        child: Text(
                          _error!,
                          style: TextStyle(
                            color: Colors.orange.shade900,
                            fontSize: 12,
                          ),
                        ),
                      ),
                    ),
                  ],
                  const SizedBox(height: 12),

                  // Medication list
                  if (_medications.isEmpty)
                    const Center(
                      child: Padding(
                        padding: EdgeInsets.all(32),
                        child: Text('今日暂无用药计划'),
                      ),
                    )
                  else
                    ..._medications.asMap().entries.map((entry) {
                      final index = entry.key;
                      final med = entry.value;
                      final isVerified = _verifiedIndices.contains(index);
                      final alreadyTaken =
                          (med['taken'] ?? false) == true || isVerified;
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: Card(
                          elevation: alreadyTaken ? 1 : 2,
                          color: alreadyTaken
                              ? Colors.green.shade50
                              : null,
                          child: ListTile(
                            leading: CircleAvatar(
                              backgroundColor: alreadyTaken
                                  ? Colors.green
                                  : Colors.blue,
                              foregroundColor: Colors.white,
                              child: Icon(
                                alreadyTaken
                                    ? Icons.check
                                    : Icons.medication,
                                size: 20,
                              ),
                            ),
                            title: Text(
                              _medicationName(index),
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                decoration: alreadyTaken
                                    ? TextDecoration.lineThrough
                                    : null,
                              ),
                            ),
                            subtitle: Text(
                              '${_medicationDose(index)} | ${_medicationTime(index)} | ${_medicationType(index)}',
                            ),
                            trailing: alreadyTaken
                                ? const Icon(
                                    Icons.check_circle,
                                    color: Colors.green,
                                  )
                                : TextButton(
                                    onPressed: () => _verifyMedication(index),
                                    child: const Text('确认服用'),
                                  ),
                          ),
                        ),
                      );
                    }),
                ],
              ),
            ),
    );
  }
}

import 'package:flutter/material.dart';
import '../services/api_client.dart';
import '../services/patient_service.dart';
import '../services/verification_service.dart';
import 'verification_screen.dart';
import 'ward_round_screen.dart';
import 'medication_screen.dart';
import 'discharge_screen.dart';

/// Full patient detail screen showing demographics, admission info,
/// wristband status, verification history, and action buttons.
class PatientDetailScreen extends StatefulWidget {
  final Map<String, dynamic> patient;

  const PatientDetailScreen({super.key, required this.patient});

  @override
  State<PatientDetailScreen> createState() => _PatientDetailScreenState();
}

class _PatientDetailScreenState extends State<PatientDetailScreen> {
  final ApiClient _api = ApiClient();
  late PatientService _patientService;
  late VerificationService _verificationService;

  Map<String, dynamic>? _patientData;
  List<dynamic> _verifications = [];
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _patientService = PatientService(_api);
    _verificationService = VerificationService(_api);
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final id = widget.patient['id'] ?? widget.patient['admission_no'] ?? '';
      if (id.isNotEmpty) {
        final data = await _patientService.getById(id);
        setState(() => _patientData = data);
      } else {
        setState(() => _patientData = widget.patient);
      }
      final verifications = await _verificationService.list(pageSize: 20);
      setState(() => _verifications = verifications);
    } catch (e) {
      setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  // ── Helpers to extract fields flexibly ──────────────────────────

  String _str(dynamic key, [dynamic fallback = '—']) {
    if (_patientData == null) return fallback.toString();
    return (_patientData![key] ?? widget.patient[key] ?? fallback)
            .toString()
        .isEmpty
        ? fallback.toString()
        : (_patientData![key] ?? widget.patient[key]).toString();
  }

  int? _int(dynamic key) {
    if (_patientData == null) return null;
    final v = _patientData![key] ?? widget.patient[key];
    if (v == null) return null;
    return int.tryParse(v.toString());
  }

  // ── Build ───────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_str('name', '患者详情')),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadData,
            tooltip: '刷新',
          ),
        ],
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.error_outline, size: 48, color: Colors.red),
                      const SizedBox(height: 16),
                      Text('加载失败: $_error'),
                      const SizedBox(height: 16),
                      ElevatedButton(
                        onPressed: _loadData,
                        child: const Text('重试'),
                      ),
                    ],
                  ),
                )
              : SingleChildScrollView(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Demographics card
                      _buildSectionCard(
                        '基本信息',
                        Icons.person_outline,
                        [
                          _buildInfoRow('姓名', _str('name')),
                          _buildInfoRow('性别', _str('gender')),
                          _buildInfoRow('年龄', '${_int('age') ?? '—'} 岁'),
                          _buildInfoRow('血型', _str('blood_type')),
                          _buildInfoRow(
                            '过敏史',
                            _str('allergies', '无'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),

                      // Admission info card
                      _buildSectionCard(
                        '入院信息',
                        Icons.local_hospital,
                        [
                          _buildInfoRow('床号', _str('bed_no', _str('bed_number'))),
                          _buildInfoRow('科室', _str('department')),
                          _buildInfoRow('诊断', _str('diagnosis')),
                          _buildInfoRow(
                            '入院时间',
                            _str('admitted_at', '—'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),

                      // Wristband status card
                      _buildSectionCard(
                        '医疗腕带状态',
                        Icons.watch,
                        [
                          _buildInfoRow(
                            '设备 ID',
                            _str('device_id', '未绑定'),
                          ),
                          _buildInfoRow(
                            '绑定时间',
                            _str('bound_at', '未绑定'),
                          ),
                          _buildInfoRow(
                            '状态',
                            _str('wristband_status', '未知'),
                          ),
                        ],
                      ),
                      const SizedBox(height: 12),

                      // Action buttons row
                      _buildActionButtons(),
                      const SizedBox(height: 12),

                      // Recent verifications card
                      _buildVerificationsCard(),
                    ],
                  ),
                ),
    );
  }

  // ── Widget builders ─────────────────────────────────────────────

  Widget _buildSectionCard(
    String title,
    IconData icon,
    List<Widget> rows,
  ) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(icon, size: 20, color: Colors.blue.shade700),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const Divider(height: 20),
            ...rows,
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 72,
            child: Text(
              label,
              style: TextStyle(
                color: Colors.grey.shade600,
                fontSize: 14,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(fontSize: 14),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            _ActionButton(
              icon: Icons.qr_code_scanner,
              label: '扫描核验',
              color: Colors.blue,
              onTap: () => _navigateToVerification(),
            ),
            _ActionButton(
              icon: Icons.assignment,
              label: '开始查房',
              color: Colors.green,
              onTap: () => _navigateToWardRound(),
            ),
            _ActionButton(
              icon: Icons.medication_liquid,
              label: '用药管理',
              color: Colors.orange,
              onTap: () => _navigateToMedication(),
            ),
            _ActionButton(
              icon: Icons.logout,
              label: '办理出院',
              color: Colors.red,
              onTap: () => _navigateToDischarge(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerificationsCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                const Icon(Icons.history, size: 20, color: Colors.blue),
                const SizedBox(width: 8),
                const Text(
                  '最近核验记录',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(height: 20),
            if (_verifications.isEmpty)
              const Padding(
                padding: EdgeInsets.symmetric(vertical: 16),
                child: Center(child: Text('暂无核验记录')),
              )
            else
              ..._verifications.take(5).map((v) {
                final result = v['result'] ?? 'unknown';
                final timestamp = v['timestamp'] ?? v['created_at'] ?? '';
                Color dotColor;
                switch (result) {
                  case 'matched':
                    dotColor = Colors.green;
                    break;
                  case 'unmatched':
                    dotColor = Colors.orange;
                    break;
                  default:
                    dotColor = Colors.grey;
                }
                return Padding(
                  padding: const EdgeInsets.only(bottom: 8),
                  child: Row(
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          color: dotColor,
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          '$result | $timestamp',
                          style: const TextStyle(fontSize: 13),
                        ),
                      ),
                    ],
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }

  // ── Navigation ──────────────────────────────────────────────────

  void _navigateToVerification() {
    final id = widget.patient['id'] ?? widget.patient['admission_no'] ?? '';
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => VerificationScreen(
          patientId: id,
          patientName: _str('name'),
        ),
      ),
    );
  }

  void _navigateToWardRound() {
    final id = widget.patient['id'] ?? widget.patient['admission_no'] ?? '';
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => WardRoundScreen(patientId: id),
      ),
    );
  }

  void _navigateToMedication() {
    final id = widget.patient['id'] ?? widget.patient['admission_no'] ?? '';
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => MedicationScreen(patientId: id),
      ),
    );
  }

  void _navigateToDischarge() {
    final id = widget.patient['id'] ?? widget.patient['admission_no'] ?? '';
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (_) => DischargeScreen(admissionId: id),
      ),
    );
  }
}

/// A compact action button for the patient detail screen.
class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return OutlinedButton.icon(
      onPressed: onTap,
      icon: Icon(icon, size: 18, color: color),
      label: Text(label),
      style: OutlinedButton.styleFrom(
        foregroundColor: color,
        side: BorderSide(color: color),
      ),
    );
  }
}

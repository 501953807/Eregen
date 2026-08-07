import 'package:flutter/material.dart';
import '../services/api_client.dart';
import '../services/ward_round_service.dart';

/// Ward round screen: collect vital signs and observations, then submit.
/// Uses hospital-api for ward round data storage.
class WardRoundScreen extends StatefulWidget {
  final String patientId;

  const WardRoundScreen({super.key, required this.patientId});

  @override
  State<WardRoundScreen> createState() => _WardRoundScreenState();
}

class _WardRoundScreenState extends State<WardRoundScreen> {
  final HospitalApiClient _hospitalApi = HospitalApiClient();
  late final WardRoundService _wardRoundService;

  // Vitals controllers
  final _bpController = TextEditingController();
  final _hrController = TextEditingController();
  final _spo2Controller = TextEditingController();
  final _tempController = TextEditingController();
  final _weightController = TextEditingController();

  // Notes
  final _notesController = TextEditingController();

  // Observation checkboxes
  bool _falls = false;
  bool _confusion = false;
  bool _pain = false;
  bool _poorAppetite = false;

  bool _submitting = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _wardRoundService = WardRoundService(_hospitalApi);
  }

  @override
  void dispose() {
    _bpController.dispose();
    _hrController.dispose();
    _spo2Controller.dispose();
    _tempController.dispose();
    _weightController.dispose();
    _notesController.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    // Validate vitals are numeric where expected
    if (_bpController.text.isEmpty &&
        _hrController.text.isEmpty &&
        _spo2Controller.text.isEmpty &&
        _tempController.text.isEmpty &&
        _weightController.text.isEmpty &&
        _notesController.text.isEmpty &&
        !_falls &&
        !_confusion &&
        !_pain &&
        !_poorAppetite) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请至少填写一项内容')),
      );
      return;
    }

    setState(() {
      _submitting = true;
      _error = null;
    });

    try {
      final observations = <String, dynamic>[];
      if (_falls) observations.add('falls_risk');
      if (_confusion) observations.add('confusion');
      if (_pain) observations.add('pain');
      if (_poorAppetite) observations.add('poor_appetite');

      final entry = <String, dynamic>{
        'patient_id': widget.patientId,
        'recorded_at': DateTime.now().toIso8601String(),
        if (_bpController.text.isNotEmpty) 'blood_pressure': _bpController.text,
        if (_hrController.text.isNotEmpty) 'heart_rate': int.parse(_hrController.text),
        if (_spo2Controller.text.isNotEmpty) 'spo2': int.parse(_spo2Controller.text),
        if (_tempController.text.isNotEmpty) 'temperature': double.parse(_tempController.text),
        if (_weightController.text.isNotEmpty) 'weight_kg': double.parse(_weightController.text),
        'observations': observations,
        if (_notesController.text.isNotEmpty) 'notes': _notesController.text.trim(),
      };

      await _wardRoundService.create(widget.patientId, entry);

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('查房记录已保存')),
        );
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error = e.toString());
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('保存失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  // ── Build ───────────────────────────────────────────────────────

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('开始查房'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Vitals section
            _buildSectionHeader('生命体征'),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    _buildTextField(
                      _bpController,
                      '血压 (mmHg)',
                      hintText: '例: 120/80',
                      keyboardType: const TextInputType.numberWithOptions(
                        signed: true,
                        decimal: true,
                      ),
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      _hrController,
                      '心率 (bpm)',
                      keyboardType: TextInputType.number,
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      _spo2Controller,
                      '血氧饱和度 (%)',
                      keyboardType: TextInputType.number,
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      _tempController,
                      '体温 (°C)',
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    ),
                    const SizedBox(height: 12),
                    _buildTextField(
                      _weightController,
                      '体重 (kg)',
                      keyboardType: const TextInputType.numberWithOptions(decimal: true),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Observations section
            _buildSectionHeader('观察项目'),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _buildCheckboxTile('跌倒风险', _falls, (v) {
                      if (v != null) setState(() => _falls = v);
                    }),
                    _buildCheckboxTile('意识混乱', _confusion, (v) {
                      if (v != null) setState(() => _confusion = v);
                    }),
                    _buildCheckboxTile('疼痛', _pain, (v) {
                      if (v != null) setState(() => _pain = v);
                    }),
                    _buildCheckboxTile('食欲不佳', _poorAppetite, (v) {
                      if (v != null) setState(() => _poorAppetite = v);
                    }),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Notes section
            _buildSectionHeader('备注'),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: TextField(
                  controller: _notesController,
                  maxLines: 4,
                  decoration: const InputDecoration(
                    hintText: '请输入查房备注...',
                    border: InputBorder.none,
                  ),
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Error display
            if (_error != null)
              Card(
                color: Colors.red.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: Text('$_error', style: const TextStyle(color: Colors.red)),
                ),
              ),

            const SizedBox(height: 16),

            // Submit button
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton(
                onPressed: _submitting ? null : _submit,
                child: _submitting
                    ? const CircularProgressIndicator()
                    : const Text('提交查房记录'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Row(
      children: [
        Container(
          width: 4,
          height: 20,
          decoration: BoxDecoration(
            color: Colors.blue.shade700,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 8),
        Text(
          title,
          style: const TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
        ),
      ],
    );
  }

  Widget _buildTextField(
    TextEditingController controller,
    String label, {
    TextInputType? keyboardType,
    String? hintText,
  }) {
    return TextField(
      controller: controller,
      keyboardType: keyboardType,
      decoration: InputDecoration(
        labelText: label,
        hintText: hintText,
        border: const OutlineInputBorder(),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 12,
        ),
      ),
    );
  }

  Widget _buildCheckboxTile(
    String label,
    bool value,
    ValueChanged<bool?> onChanged,
  ) {
    return FilterChip(
      label: Text(label),
      selected: value,
      onSelected: onChanged,
      selectedColor: Colors.blue.shade100,
    );
  }
}

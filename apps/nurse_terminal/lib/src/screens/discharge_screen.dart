import 'package:flutter/material.dart';
import '../services/api_client.dart';
import '../services/patient_service.dart';

/// Discharge screen: select discharge type, add notes, optionally
/// specify transfer destination, then call the discharge API.
class DischargeScreen extends StatefulWidget {
  final String admissionId;

  const DischargeScreen({super.key, required this.admissionId});

  @override
  State<DischargeScreen> createState() => _DischargeScreenState();
}

class _DischargeScreenState extends State<DischargeScreen> {
  final ApiClient _api = ApiClient();
  late PatientService _patientService;

  String _dischargeType = 'discharged';
  final _notesController = TextEditingController();
  final _transferredController = TextEditingController();
  bool _submitting = false;
  String? _error;

  // Discharge type options: (value, label)
  static const _dischargeOptions = [
    ('discharged', '正常出院'),
    ('transferred', '转院'),
    ('deceased', '死亡'),
  ];

  @override
  void initState() {
    super.initState();
    _patientService = PatientService(_api);
  }

  @override
  void dispose() {
    _notesController.dispose();
    _transferredController.dispose();
    super.dispose();
  }

  bool get _showTransferredTo => _dischargeType == 'transferred';

  Future<void> _submit() async {
    if (_notesController.text.isEmpty && !_showTransferredTo) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('请填写备注或选择转院')),
      );
      return;
    }

    setState(() {
      _submitting = true;
      _error = null;
    });

    try {
      final body = <String, dynamic>{
        'discharge_type': _dischargeType,
        'notes': _notesController.text.trim(),
      };
      if (_showTransferredTo && _transferredController.text.isNotEmpty) {
        body['transferred_to'] = _transferredController.text.trim();
      }

      await _patientService.discharge(
        widget.admissionId,
        _dischargeType,
        notes: _notesController.text.trim().isEmpty ? null : _notesController.text.trim(),
        transferredTo: _showTransferredTo && _transferredController.text.isNotEmpty
            ? _transferredController.text.trim()
            : null,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('出院手续已办理，腕带绑定已清除')),
        );
        Navigator.of(context).pop(true);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _error = e.toString());
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('办理失败: $e')),
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
        title: const Text('办理出院'),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Warning banner
            Card(
              color: Colors.red.shade50,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Row(
                  children: [
                    const Icon(Icons.warning_amber_rounded, color: Colors.red),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        '出院后将清除腕带绑定关系，请谨慎操作。',
                        style: TextStyle(color: Colors.red.shade900),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Discharge type selector
            _buildSectionHeader('出院类型'),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _dischargeOptions.map((option) {
                    final value = option.$1;
                    final label = option.$2;
                    final isSelected = _dischargeType == value;
                    return ChoiceChip(
                      label: Text(label),
                      selected: isSelected,
                      onSelected: (selected) {
                        if (selected) setState(() => _dischargeType = value);
                      },
                      selectedColor: Colors.blue.shade100,
                    );
                  }).toList(),
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Transfer destination (conditional)
            if (_showTransferredTo) ...[
              _buildSectionHeader('转入机构'),
              const SizedBox(height: 8),
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(12),
                  child: TextField(
                    controller: _transferredController,
                    decoration: const InputDecoration(
                      labelText: '转入医院或科室名称',
                      border: OutlineInputBorder(),
                      hintText: '例：市第一人民医院 心内科',
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 16),
            ],

            // Notes field
            _buildSectionHeader('备注'),
            const SizedBox(height: 8),
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: TextField(
                  controller: _notesController,
                  maxLines: 4,
                  decoration: const InputDecoration(
                    hintText: '出院医嘱、随访安排等备注信息...',
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
                  child: Text(
                    _error!,
                    style: const TextStyle(color: Colors.red),
                  ),
                ),
              ),

            const SizedBox(height: 16),

            // Submit button
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton(
                onPressed: _submitting ? null : _submit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.red,
                  foregroundColor: Colors.white,
                ),
                child: _submitting
                    ? const CircularProgressIndicator()
                    : const Text('确认办理出院'),
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
}

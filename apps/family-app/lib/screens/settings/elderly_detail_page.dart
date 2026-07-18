import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../api/client.dart';

/// Elderly person detail screen — health overview, devices, medication rules.
class ElderlyDetailPage extends StatefulWidget {
  const ElderlyDetailPage({super.key});

  @override
  State<ElderlyDetailPage> createState() => _ElderlyDetailPageState();
}

class _ElderlyDetailPageState extends State<ElderlyDetailPage> {
  bool _loading = true;
  String? _error;
  List<Map<String, dynamic>> _elderlyList = [];
  Map<String, dynamic>? _selectedElderly;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final resp = await ApiClient.instance.get('/admin/elderly');
      if (resp.data != null && (resp.data as Map).containsKey('data')) {
        final list = resp.data['data'] as List;
        setState(() {
          _elderlyList = list.cast<Map<String, dynamic>>();
          if (_selectedElderly == null && _elderlyList.isNotEmpty) {
            _selectedElderly = _elderlyList[0];
          }
        });
      }
    } catch (e) {
      setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _selectElderly(Map<String, dynamic> elderly) {
    setState(() => _selectedElderly = elderly);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(_selectedElderly?['name'] ?? '李秀英 奶奶'),
        backgroundColor: AppTheme.primary,
        foregroundColor: Colors.white,
      ),
      backgroundColor: AppTheme.bgScaffold,
      body: _error != null
          ? Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error_outline, size: 48, color: Colors.red),
                  const SizedBox(height: 16),
                  Text('加载失败: $_error', style: const TextStyle(color: Colors.red)),
                  const SizedBox(height: 16),
                  ElevatedButton(onPressed: _fetchData, child: const Text('重试')),
                ],
              ),
            )
          : _elderlyList.isEmpty && !_loading
              ? const Center(child: Text('暂无关联老人，请添加'))
              : CustomScrollView(
                  slivers: [
                    // Elderly selector (if multiple)
                    if (_elderlyList.length > 1)
                      SliverToBoxAdapter(
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 12),
                          child: Wrap(
                            spacing: 8,
                            runSpacing: 8,
                            children: _elderlyList.map((e) {
                              final name = e['name'] as String? ?? '未知';
                              return ChoiceChip(
                                label: Text(name),
                                selected: _selectedElderly == e,
                                onSelected: (_) => _selectElderly(e),
                                selectedColor: AppTheme.primary.withValues(alpha: 0.2),
                              );
                            }).toList(),
                          ),
                        ),
                      ),

                    // Profile card
                    SliverToBoxAdapter(
                      child: Container(
                        margin: const EdgeInsets.all(20),
                        padding: const EdgeInsets.all(20),
                        decoration: BoxDecoration(
                          color: AppTheme.bgCard,
                          borderRadius: BorderRadius.circular(14),
                          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 10)],
                        ),
                        child: Row(
                          children: [
                            CircleAvatar(
                              radius: 32,
                              backgroundColor: AppTheme.primary,
                              child: Text(
                                (_selectedElderly?['name'] ?? '李')[0],
                                style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w700, color: Colors.white),
                              ),
                            ),
                            const SizedBox(width: 16),
                            Expanded(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    _selectedElderly?['name'] ?? '李秀英 奶奶',
                                    style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
                                  ),
                                  const SizedBox(height: 4),
                                  Row(
                                    children: [
                                      _statusChip(
                                        _selectedElderly?['health_tiers'] != null && (_selectedElderly!['health_tiers'] as List).isNotEmpty
                                            ? (_selectedElderly!['health_tiers'] as List)[0]
                                            : '中风险',
                                        AppTheme.statusWarning,
                                      ),
                                      const SizedBox(width: 8),
                                      _statusChip('在线', AppTheme.statusNormal),
                                    ],
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),

                    // Health summary
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('健康概览', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                            const SizedBox(height: 8),
                            Wrap(
                              spacing: 10,
                              runSpacing: 10,
                              children: [
                                _miniStatCard('心率', '${_selectedElderly?['last_hr'] ?? 72} bpm', AppTheme.primary, Icons.favorite),
                                _miniStatCard('血氧', '${_selectedElderly?['last_spo2'] ?? 97}%', AppTheme.statusNormal, Icons.air),
                                _miniStatCard('步数', (_selectedElderly?['last_steps'] ?? 3456).toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (Match m) => '${m[1]},'), AppTheme.statusWarning, Icons.directions_walk),
                                _miniStatCard('电量', '${_selectedElderly?['last_battery'] ?? 85}%', AppTheme.statusInfo, Icons.battery_full),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SliverToBoxAdapter(child: SizedBox(height: 16)),

                    // Devices
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('关联设备', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.all(16),
                              decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14)),
                              child: Column(
                                children: [
                                  _deviceRow('颐贞手环 Plus', 'BR-0012', 'online', 'v2.1.0'),
                                  const Divider(height: 1),
                                  _deviceRow('颐贞药盒 Smart', 'PX-0008', 'online', 'v1.3.2'),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SliverToBoxAdapter(child: SizedBox(height: 16)),

                    // Medication
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.symmetric(horizontal: 20),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('用药规则', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                            const SizedBox(height: 8),
                            Container(
                              padding: const EdgeInsets.all(16),
                              decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14)),
                              child: Column(
                                children: [
                                  _medRuleRow('08:00', '降压药', '1 粒'),
                                  const Divider(height: 1),
                                  _medRuleRow('12:00', '钙片', '2 粒'),
                                  const Divider(height: 1),
                                  _medRuleRow('20:00', '维生素D', '1 粒'),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SliverToBoxAdapter(child: SizedBox(height: 24)),
                  ],
                ),
    );
  }

  Widget _statusChip(String label, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
      child: Text(label, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color)),
    );
  }

  Widget _miniStatCard(String label, String value, Color color, IconData icon) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(12)),
        child: Column(
          children: [
            Icon(icon, size: 20, color: color),
            const SizedBox(height: 4),
            Text(value, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: color)),
            Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
          ],
        ),
      ),
    );
  }

  Widget _deviceRow(String name, String deviceId, String status, String fwVersion) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Container(width: 8, height: 8, decoration: BoxDecoration(color: status == 'online' ? AppTheme.statusNormal : Colors.grey, shape: BoxShape.circle)),
          const SizedBox(width: 12),
          Expanded(child: Text(name, style: const TextStyle(fontWeight: FontWeight.w600, fontSize: 13))),
          Text(deviceId, style: const TextStyle(fontSize: 11, color: Color(0xFF999999))),
          const SizedBox(width: 8),
          Text(fwVersion, style: const TextStyle(fontSize: 11, color: Color(0xFF999999))),
        ],
      ),
    );
  }

  Widget _medRuleRow(String time, String medName, String dose) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(color: const Color(0xFFE3F2FD), borderRadius: BorderRadius.circular(8)),
            child: Text(time, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: AppTheme.primary)),
          ),
          const SizedBox(width: 12),
          Expanded(child: Text(medName, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600))),
          Text(dose, style: const TextStyle(fontSize: 12, color: Color(0xFF888888))),
        ],
      ),
    );
  }
}

/// Add elderly person page — for linking a new elderly profile.
class AddElderlyPage extends StatefulWidget {
  const AddElderlyPage({super.key});

  @override
  State<AddElderlyPage> createState() => _AddElderlyPageState();
}

class _AddElderlyPageState extends State<AddElderlyPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();
  DateTime? _selectedBirthDate;
  String? _selectedTier;

  @override
  void dispose() {
    _nameController.dispose();
    _phoneController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('添加老人'), backgroundColor: AppTheme.primary, foregroundColor: Colors.white),
      backgroundColor: AppTheme.bgScaffold,
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              TextFormField(
                controller: _nameController,
                decoration: const InputDecoration(labelText: '姓名', border: OutlineInputBorder()),
                validator: (v) => v == null || v.isEmpty ? '请输入姓名' : null,
              ),
              const SizedBox(height: 16),
              TextFormField(
                controller: _phoneController,
                decoration: const InputDecoration(labelText: '手机号', border: OutlineInputBorder(), hintText: '用于接收短信通知'),
                keyboardType: TextInputType.phone,
                validator: (v) => v == null || v.length < 11 ? '请输入有效手机号' : null,
              ),
              const SizedBox(height: 16),
              InkWell(
                onTap: _pickBirthDate,
                child: InputDecorator(
                  decoration: const InputDecoration(labelText: '出生日期', border: OutlineInputBorder()),
                  child: Text(_selectedBirthDate == null ? '选择出生日期' : '${_selectedBirthDate!.year}-${_selectedBirthDate!.month.toString().padLeft(2, '0')}-${_selectedBirthDate!.day.toString().padLeft(2, '0')}'),
                ),
              ),
              const SizedBox(height: 16),
              DropdownButtonFormField<String>(
                initialValue: _selectedTier,
                decoration: const InputDecoration(labelText: '健康等级', border: OutlineInputBorder()),
                items: ['低风险', '中风险', '高风险'].map((tier) => DropdownMenuItem(value: tier, child: Text(tier))).toList(),
                onChanged: (v) => setState(() => _selectedTier = v),
              ),
              const SizedBox(height: 24),
              SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton(
                  onPressed: () async {
                    if (_formKey.currentState!.validate()) {
                      try {
                        final birthDate = _selectedBirthDate != null
                            ? '${_selectedBirthDate!.year}-${_selectedBirthDate!.month.toString().padLeft(2, '0')}-${_selectedBirthDate!.day.toString().padLeft(2, '0')}'
                            : null;
                        final healthTiers = _selectedTier != null ? [_selectedTier] : <String>[];
                        await ApiClient.instance.post('/elderly', data: {
                          'name': _nameController.text.trim(),
                          if (birthDate != null) 'birth_date': birthDate,
                          'health_tiers': healthTiers,
                        });
                        if (mounted) {
                          Navigator.of(context).pop();
                          ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('添加成功')));
                        }
                      } catch (e) {
                        if (mounted) {
                          ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text('添加失败: $e')));
                        }
                      }
                    }
                  },
                  child: const Text('保存', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Future<void> _pickBirthDate() async {
    final date = await showDatePicker(context: context, initialDate: DateTime(1950), firstDate: DateTime(1900), lastDate: DateTime.now());
    if (date != null) setState(() => _selectedBirthDate = date);
  }
}

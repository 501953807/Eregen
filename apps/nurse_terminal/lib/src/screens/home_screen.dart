import 'package:flutter/material.dart';
import '../services/api_client.dart';
import '../services/patient_service.dart';
import '../services/medical_wristband_ble_service.dart';
import 'patient_detail_screen.dart';

/// Home screen showing the list of admitted patients.
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiClient _api = ApiClient();
  late PatientService _patientService;

  List<dynamic> patients = [];
  bool loading = true;
  final _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _patientService = PatientService(_api);
    _loadPatients();
  }

  Future<void> _loadPatients() async {
    setState(() => loading = true);
    try {
      final data = await _patientService.listAdmitted();
      setState(() => patients = data);
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载患者列表失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => loading = false);
    }
  }

  List<dynamic> get _filteredPatients {
    final query = _searchController.text.trim().toLowerCase();
    if (query.isEmpty) return patients;
    return patients.where((p) {
      final name = (p['name'] ?? p['patient_name'] ?? '').toString().toLowerCase();
      final admissionNo = (p['admission_no'] ?? p['id'] ?? '').toString().toLowerCase();
      return name.contains(query) || admissionNo.contains(query);
    }).toList();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _scanWristband() async {
    final result = await Navigator.push<String>(
      context,
      MaterialPageRoute(
        builder: (_) => _NfcScanScreen(),
      ),
    );
    if (result != null && mounted) {
      final patient = patients.firstWhere(
        (p) => p['id'] == result || p['admission_no'] == result,
        orElse: () => <String, dynamic>{'id': result, 'name': '未知患者'},
      );
      Navigator.push(
        context,
        MaterialPageRoute(
          builder: (_) => PatientDetailScreen(patient: patient),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filteredPatients;

    return Scaffold(
      appBar: AppBar(
        title: const Text('颐贞 护士终端'),
        actions: [
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await _api.logout();
              if (mounted) {
                Navigator.of(context).pushReplacementNamed('/');
              }
            },
          ),
        ],
      ),
      body: Column(children: [
        Padding(
          padding: const EdgeInsets.all(16),
          child: Row(children: [
            Expanded(
              child: TextField(
                controller: _searchController,
                decoration: const InputDecoration(
                  prefixIcon: Icon(Icons.search),
                  hintText: '搜索患者姓名或住院号',
                  border: OutlineInputBorder(),
                ),
                onChanged: (_) => setState(() {}),
              ),
            ),
            const SizedBox(width: 8),
            ElevatedButton.icon(
              onPressed: _loadPatients,
              icon: const Icon(Icons.refresh),
              label: const Text('刷新'),
            ),
          ]),
        ),
        Expanded(
          child: loading
              ? const Center(child: CircularProgressIndicator())
              : filtered.isEmpty
                  ? const Center(child: Text('暂无在院患者'))
                  : ListView.builder(
                      itemCount: filtered.length,
                      itemBuilder: (ctx, i) {
                        final p = filtered[i];
                        final name = p['name'] ?? p['patient_name'] ?? '未知';
                        final admissionNo = p['admission_no'] ?? p['id'] ?? '';
                        final dept = p['department'] ?? '—';
                        final bed = p['bed_number'] ?? p['bed_no'] ?? '—';
                        return Card(
                          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
                          child: ListTile(
                            leading: CircleAvatar(
                              child: Text(name[0]?.toUpperCase() ?? '?'),
                            ),
                            title: Text(
                              name,
                              style: const TextStyle(fontWeight: FontWeight.bold),
                            ),
                            subtitle: Text(
                              '$dept | 床号: $bed | 住院号: $admissionNo',
                            ),
                            trailing: const Icon(Icons.chevron_right),
                            onTap: () {
                              Navigator.push(
                                ctx,
                                MaterialPageRoute(
                                  builder: (_) => PatientDetailScreen(patient: p),
                                ),
                              );
                            },
                          ),
                        );
                      },
                    ),
        ),
      ]),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _scanWristband,
        icon: const Icon(Icons.nfc),
        label: const Text('扫描腕带'),
      ),
    );
  }
}

/// A simple screen that performs NFC scan and returns patient ID on success.
class _NfcScanScreen extends StatefulWidget {
  @override
  State<_NfcScanScreen> createState() => _NfcScanScreenState();
}

class _NfcScanScreenState extends State<_NfcScanScreen> {
  final MedicalWristbandService _nfcService = MedicalWristbandService();
  bool _scanning = false;
  String? _error;
  String? _patientId;

  @override
  void dispose() {
    _nfcService.dispose();
    super.dispose();
  }

  Future<void> _startScan() async {
    setState(() {
      _scanning = true;
      _error = null;
      _patientId = null;
    });
    try {
      final message = await _nfcService.scanWristband();
      if (message == null) {
        setState(() => _error = '未检测到腕带，请将腕带靠近设备背面');
        return;
      }
      final patientInfo = _nfcService.parsePatientInfo(message);
      if (patientInfo == null) {
        setState(() => _error = '腕带数据格式异常，请重试');
        return;
      }
      setState(() {
        _scanning = false;
        _patientId = patientInfo.patientId;
      });
      Navigator.pop(context, patientInfo.patientId);
    } catch (e) {
      setState(() {
        _scanning = false;
        _error = '扫描失败: $e';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('扫描医用腕带')),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (_patientId != null)
              Card(
                color: Colors.green.shade50,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Row(
                    children: [
                      const Icon(Icons.check_circle, color: Colors.green),
                      const SizedBox(width: 12),
                      Text(
                        '已识别腕带 ID: $_patientId',
                        style: const TextStyle(color: Colors.green),
                      ),
                    ],
                  ),
                ),
              )
            else if (_error != null)
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
            else
              Center(
                child: Column(
                  children: [
                    const Icon(Icons.nfc, size: 80, color: Colors.blue),
                    const SizedBox(height: 16),
                    const Text(
                      '将医用腕带靠近设备背面进行 NFC 读取',
                      style: TextStyle(color: Colors.grey),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '保持 4cm 以内距离，等待 2 秒',
                      style: TextStyle(color: Colors.grey[600], fontSize: 13),
                    ),
                    const SizedBox(height: 24),
                    if (_scanning)
                      const CircularProgressIndicator()
                    else
                      ElevatedButton.icon(
                        onPressed: _startScan,
                        icon: const Icon(Icons.nfc),
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
}

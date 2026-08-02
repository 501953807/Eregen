import 'package:flutter/material.dart';
import '../services/api_client.dart';
import 'patient_detail_screen.dart';

/// Home screen showing the list of admitted patients.
/// Supports search, refresh, and navigation to patient detail.
class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final ApiClient api = ApiClient();
  List<dynamic> patients = [];
  bool loading = true;
  final _searchController = TextEditingController();

  @override
  void initState() {
    super.initState();
    loadPatients();
  }

  Future<void> loadPatients() async {
    setState(() => loading = true);
    try {
      final res = await api.get(
        '/api/v1/admin/medical/patients?page=1&page_size=50&status=admitted',
      );
      final data = res['data'] as List<dynamic>? ?? [];
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
              final navigator = Navigator.of(context);
              await api.logout();
              if (mounted) {
                navigator.pushReplacementNamed('/');
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
              onPressed: loadPatients,
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
        onPressed: () {
          // TODO: Open NFC scan to read medical wristband
        },
        icon: const Icon(Icons.qr_code_scanner),
        label: const Text('扫描腕带'),
      ),
    );
  }
}

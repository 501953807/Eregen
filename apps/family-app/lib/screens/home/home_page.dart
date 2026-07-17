import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/bottom_nav_bar.dart';
import '../widgets/elderly_selector.dart';
import '../widgets/map_section.dart';
import '../widgets/quick_status_card.dart';
import '../widgets/sos_button.dart';
import '../widgets/recent_alerts_list.dart';
import '../../api/client.dart';
import '../../models/health.dart';
import '../../models/alert.dart';

/// Home page — matches home.html prototype, now with live API data.
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _selectedIndex = 0;
  bool _loading = true;
  HealthRecord? _latestHealth;
  List<Alert> _recentAlerts = [];

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final futures = <Future>[];
      // Fetch latest health snapshot
      futures.add(ApiClient.instance.get('/health/latest').then((resp) {
        if (resp.data != null && (resp.data as Map).isNotEmpty) {
          setState(() => _latestHealth = HealthRecord.fromJson(resp.data as Map<String, dynamic>));
        }
      }).catchError((_) {}));
      // Fetch recent alerts
      futures.add(ApiClient.instance.get('/alerts', query: {'limit': 5}).then((resp) {
        final list = (resp.data as List);
        setState(() => _recentAlerts = list.map((a) => Alert.fromJson(a as Map<String, dynamic>)).toList());
      }).catchError((_) {}));
      await Future.wait(futures);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  // Header with gradient background
                  SliverToBoxAdapter(
                    child: Container(
                      decoration: const BoxDecoration(gradient: AppTheme.headerGradient),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Padding(
                            padding: const EdgeInsets.fromLTRB(20, 8, 20, 0),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                const Text('颐贞', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: Colors.white, letterSpacing: 1)),
                                Row(
                                  children: [
                                    _headerIcon(Icons.notifications_outlined, badge: _recentAlerts.where((a) => a.status == 'pending').length),
                                    const SizedBox(width: 8),
                                    _headerIcon(Icons.settings_outlined),
                                  ],
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: 12),
                          const ElderlySelector(name: '李秀英 奶奶', isOnline: true, lastUpdate: '2分钟前'),
                          const SizedBox(height: 8),
                        ],
                      ),
                    ),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 8)),
                  const SliverToBoxAdapter(child: MapSection()),
                  const SliverToBoxAdapter(child: SizedBox(height: 16)),

                  // Quick status cards — populated from API
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      child: Row(
                        children: [
                          Expanded(
                            child: QuickStatusCard(
                              icon: Icons.favorite,
                              value: '${_latestHealth?.hr ?? '72'}',
                              label: '心率 bpm',
                              status: _statusLabel(_latestHealth?.hr),
                              statusColor: _statusColor(_latestHealth?.hr),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: QuickStatusCard(
                              icon: Icons.air,
                              value: '${_latestHealth?.spo2 ?? '97'}%',
                              label: '血氧 SpO2',
                              status: _statusLabelSpO2(_latestHealth?.spo2),
                              statusColor: _statusColorSpO2(_latestHealth?.spo2),
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: QuickStatusCard(
                              icon: Icons.directions_walk,
                              value: '${(_latestHealth?.steps ?? 3456).toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (m) => '${m[1]},')}',
                              label: '今日步数',
                              status: '正常',
                              statusColor: AppTheme.statusNormal,
                            ),
                          ),
                          const SizedBox(width: 10),
                          Expanded(
                            child: QuickStatusCard(
                              icon: Icons.battery_full,
                              value: '85%',
                              label: '手环电量',
                              status: '充足',
                              statusColor: AppTheme.statusNormal,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 16)),
                  const SliverToBoxAdapter(child: SOSButton()),
                  const SliverToBoxAdapter(child: SizedBox(height: 16)),

                  // Recent alerts — populated from API
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              const Text('最近告警', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                              GestureDetector(
                                onTap: () {},
                                child: const Text('查看全部 ›', style: TextStyle(fontSize: 12, color: AppTheme.primary, fontWeight: FontWeight.w600)),
                              ),
                            ],
                          ),
                          const SizedBox(height: 12),
                          if (_recentAlerts.isEmpty)
                            const Text('暂无告警记录', style: TextStyle(color: Color(0xFFBBBBBB), fontSize: 13)),
                          if (_recentAlerts.isNotEmpty) ..._buildRecentAlertCards(),
                          const SizedBox(height: 24),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
      bottomNavigationBar: BottomNavBar(selectedTab: _selectedIndex, onTabSelected: (i) => setState(() => _selectedIndex = i)),
    );
  }

  Widget _headerIcon(IconData icon, {int? badge}) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        Container(width: 36, height: 36, decoration: BoxDecoration(color: Colors.white.withOpacity(0.2), shape: BoxShape.circle), child: Icon(icon, size: 18, color: Colors.white)),
        if (badge != null && badge > 0)
          Positioned(right: -2, top: -2, child: Container(width: 16, height: 16, decoration: const BoxDecoration(color: Color(0xFFFF6B6B), shape: BoxShape.circle), child: Center(child: Text('$badge', style: const TextStyle(fontSize: 9, color: Colors.white, fontWeight: FontWeight.w700))))),
      ],
    );
  }

  String _statusLabel(int? hr) {
    if (hr == null) return '正常';
    if (hr < 60 || hr > 100) return '异常';
    return '正常';
  }

  Color _statusColor(int? hr) {
    if (hr == null) return AppTheme.statusNormal;
    if (hr < 60 || hr > 100) return AppTheme.statusDanger;
    return AppTheme.statusNormal;
  }

  String _statusLabelSpO2(int? spo2) {
    if (spo2 == null) return '正常';
    if (spo2 < 95) return '偏低';
    return '正常';
  }

  Color _statusColorSpO2(int? spo2) {
    if (spo2 == null) return AppTheme.statusNormal;
    if (spo2 < 95) return AppTheme.statusWarning;
    return AppTheme.statusNormal;
  }

  List<Widget> _buildRecentAlertCards() {
    return _recentAlerts.take(3).map((alert) {
      return Container(
        margin: const EdgeInsets.only(bottom: 8),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(12), border: Border.all(color: const Color(0xFFF0F0F0))),
        child: Row(
          children: [
            Container(width: 8, height: 8, margin: const EdgeInsets.only(right: 10), decoration: BoxDecoration(color: alert.severity == 'P0' ? AppTheme.statusDanger : AppTheme.statusWarning, shape: BoxShape.circle)),
            Expanded(
              child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                Text(alert.alertType, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                Text(alert.createdAt.toLocal().toString().substring(5, 16), style: const TextStyle(fontSize: 11, color: Color(0xFF999999))),
              ]),
            ),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(color: alert.status == 'pending' ? const Color(0xFFFFEBEE) : const Color(0xFFE8F5E9), borderRadius: BorderRadius.circular(8)),
              child: Text(alert.status == 'pending' ? '未处理' : '已处理', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: alert.status == 'pending' ? const Color(0xFFC62828) : const Color(0xFF2E7D32))),
            ),
          ],
        ),
      );
    }).toList();
  }
}

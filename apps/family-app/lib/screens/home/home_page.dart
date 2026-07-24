import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../widgets/elderly_selector.dart';
import '../../widgets/map_section.dart';
import '../../widgets/sos_button.dart';
import '../../api/client.dart';
import '../../models/health.dart';
import '../../models/alert.dart';
import '../alerts/alerts_page.dart';
import '../../screens/settings/settings_page.dart';
import '../../screens/welfare_page.dart';

/// Home page — v2 design: brand header, elder selector cards, SOS banner,
/// map with geofence/pin/tooltips, collapsible bottom-sheet status card, health tips.
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _selectedIndex = 0;
  bool _loading = true;
  bool _darkMode = false;
  bool _showSOSBanner = true;
  bool _showHealthTip = true;
  bool _cardExpanded = false;
  HealthRecord? _latestHealth;
  List<Alert> _recentAlerts = [];

  final List<ElderInfo> _elders = [
    ElderInfo(name: '爷爷 张三丰', title: '在线 · 手环Pro', icon: '\u{1F468}', bg: const Color(0xFFFFF3E0), online: true, tier: 'Pro'),
    ElderInfo(name: '奶奶 李秀英', title: '在线 · 手环Plus', icon: '\u{1F469}', bg: const Color(0xFFFCE7F3), online: true, tier: 'Plus'),
    ElderInfo(name: '外公 王建国', title: '离线 · 手环Starter', icon: '\u{1F468}', bg: const Color(0xFFE8EAF6), online: false, tier: 'Starter'),
  ];
  int _activeElder = 0;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final futures = <Future>[];
      futures.add(ApiClient.instance.get('/health/latest').then((resp) {
        if (resp.data != null && (resp.data as Map).isNotEmpty) {
          setState(() => _latestHealth = HealthRecord.fromJson(resp.data as Map<String, dynamic>));
        }
      }).catchError((_) {}));
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
      backgroundColor: _darkMode ? const Color(0xFF111827) : const Color(0xFFF3F4F6),
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : Column(
          children: [
            // Top Bar
            _buildTopBar(),
            // SOS Banner
            if (_showSOSBanner) _buildSOSBanner(),
            // Map Area (flexible)
            Expanded(
              child: Stack(
                children: [
                  const MapSection(),
                  // Location tooltip
                  Positioned(
                    bottom: 16,
                    left: 50,
                    child: _locationTooltip(),
                  ),
                  // Map controls
                  Positioned(
                    right: 12,
                    bottom: 140,
                    child: _mapControls(),
                  ),
                ],
              ),
            ),
            // Status Card (bottom sheet)
            _buildStatusCard(),
          ],
        ),
      ),
      bottomNavigationBar: _buildBottomNav(),
    );
  }

  // ===== Top Bar =====
  Widget _buildTopBar() {
    return Container(
      color: Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Brand row
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Row(
                children: [
                  Container(
                    width: 32,
                    height: 32,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(colors: [Color(0xFF2563EB), Color(0xFF7C3AED)]),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Center(child: Text('颐', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Colors.white))),
                  ),
                  const SizedBox(width: 8),
                  const Text('Eregen', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
                  const Text(' 颐贞', style: TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                ],
              ),
              Row(
                children: [
                  _topIconButton(_darkMode ? '\u{1F31E}' : '\u{2600}', onTap: () => setState(() => _darkMode = !_darkMode)),
                  const SizedBox(width: 6),
                  Stack(
                    clipBehavior: Clip.none,
                    children: [
                      _topIconButton('\u{1F514}'),
                      Positioned(right: -2, top: -2, child: _badge(2)),
                    ],
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 10),
          // Elder selector (horizontal scroll)
          _buildElderSelector(),
        ],
      ),
    );
  }

  Widget _buildElderSelector() {
    return SizedBox(
      height: 60,
      child: ListView.separated(
        scrollDirection: Axis.horizontal,
        itemCount: _elders.length,
        separatorBuilder: (_, __) => const SizedBox(width: 10),
        itemBuilder: (ctx, i) {
          final elder = _elders[i];
          final isActive = i == _activeElder;
          return GestureDetector(
            onTap: () => setState(() => _activeElder = i),
            child: Container(
              width: 160,
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              decoration: BoxDecoration(
                color: isActive ? const Color(0xFFDBEAFE) : Colors.white,
                border: Border.all(color: isActive ? const Color(0xFF2563EB) : const Color(0xFFE5E7EB)),
                borderRadius: BorderRadius.circular(24),
                boxShadow: isActive ? [BoxShadow(color: const Color(0xFF2563EB).withValues(alpha: 0.15), blurRadius: 12, offset: const Offset(0, 2))] : [],
              ),
              child: Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(color: elder.bg, shape: BoxShape.circle),
                    child: Center(child: Text(elder.icon, style: const TextStyle(fontSize: 20))),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(elder.name, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: isActive ? const Color(0xFF1F2937) : const Color(0xFF374151))),
                        Row(
                          children: [
                            Container(
                              width: 6,
                              height: 6,
                              decoration: BoxDecoration(
                                color: elder.online ? AppTheme.statusNormal : const Color(0xFFD1D5DB),
                                shape: BoxShape.circle,
                              ),
                            ),
                            const SizedBox(width: 4),
                            Text(elder.title, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
                          ],
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      ),
    );
  }

  Widget _topIconButton(String icon, {VoidCallback? onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 36,
        height: 36,
        decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(18)),
        child: Center(child: Text(icon, style: const TextStyle(fontSize: 16))),
      ),
    );
  }

  Widget _badge(int count) {
    return Positioned(
      right: -2,
      top: -2,
      child: Container(
        width: 16,
        height: 16,
        decoration: const BoxDecoration(color: AppTheme.statusDanger, shape: BoxShape.circle),
        child: Center(child: Text('$count', style: const TextStyle(fontSize: 9, color: Colors.white, fontWeight: FontWeight.w700))),
      ),
    );
  }

  // ===== SOS Banner =====
  Widget _buildSOSBanner() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: const BoxDecoration(
        gradient: LinearGradient(colors: [Color(0xFFDC2626), Color(0xFFEF4444)]),
      ),
      child: Row(
        children: [
          const Text('\u{26A0}', style: TextStyle(fontSize: 24)),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('爷爷 张三丰 触发SOS紧急呼叫', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.white)),
                const Text('2分钟前 · 位置：上海市浦东新区陆家嘴环路1000号', style: TextStyle(fontSize: 11, color: Colors.white70)),
              ],
            ),
          ),
          GestureDetector(
            onTap: () => setState(() => _showSOSBanner = false),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(20)),
              child: const Text('查看处理', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Color(0xFFEF4444))),
            ),
          ),
        ],
      ),
    );
  }

  // ===== Map Controls =====
  Widget _mapControls() {
    return Column(
      children: [
        _mapCtrlBtn('\u{1F50D}'),
        const SizedBox(height: 8),
        _mapCtrlBtn('\u{2795}'),
        const SizedBox(height: 8),
        _mapCtrlBtn('\u{1F4CD}'),
      ],
    );
  }

  Widget _mapCtrlBtn(String icon) {
    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.95), borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 8)]),
      child: Center(child: Text(icon, style: const TextStyle(fontSize: 18))),
    );
  }

  Widget _locationTooltip() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.95), borderRadius: BorderRadius.circular(20), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 12)]),
      child: const Text('\u{1F4CD} <strong>陆家嘴金融中心</strong> · 距家 2.3km', style: TextStyle(fontSize: 12, color: Color(0xFF374151))),
    );
  }

  // ===== Status Card (Bottom Sheet) =====
  Widget _buildStatusCard() {
    return StatefulBuilder(
      builder: (context, setInnerState) {
        return GestureDetector(
          onTap: () => setInnerState(() => _cardExpanded = !_cardExpanded),
          child: Container(
            width: double.infinity,
            decoration: BoxDecoration(
              color: _darkMode ? const Color(0xFF111827) : Colors.white,
              borderRadius: const BorderRadius.vertical(top: Radius.circular(24)),
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.08), blurRadius: 24, offset: const Offset(0, -4))],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Handle bar
                Container(width: 36, height: 4, margin: const EdgeInsets.only(top: 10), decoration: BoxDecoration(color: const Color(0xFFD1D5DB), borderRadius: BorderRadius.circular(2))),
                // Health tip
                if (_showHealthTip) _healthTip(),
                // Card header
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 4),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('健康状态', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
                      Text('最后更新：<span style="color:#16A34A;font-weight:600;">2分钟前</span>', style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
                    ],
                  ),
                ),
                // Stats grid
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                  child: GridView.count(
                    shrinkWrap: true,
                    crossAxisCount: _cardExpanded ? 3 : 3,
                    crossAxisSpacing: 10,
                    mainAxisSpacing: 10,
                    childAspectRatio: 1.1,
                    physics: const NeverScrollableScrollPhysics(),
                    children: _statItems(),
                  ),
                ),
                // Battery row
                _batteryRow(),
                // Quick actions
                if (_cardExpanded) _quickActions(),
                // Expand/collapse indicator
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  child: Text(
                    _cardExpanded ? '收起 ▲' : '展开 ▼',
                    style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _healthTip() {
    return Container(
      margin: const EdgeInsets.fromLTRB(16, 4, 16, 8),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        gradient: const LinearGradient(colors: [Color(0xFFECFDF5), Color(0xFFD1FAE5)]),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        children: [
          const Text('\u{1F4A1}', style: TextStyle(fontSize: 16)),
          const SizedBox(width: 8),
          const Expanded(child: Text('今日步数偏低，建议饭后散步15分钟', style: TextStyle(fontSize: 12, color: Color(0xFF065F46)))),
          GestureDetector(
            onTap: () => setState(() => _showHealthTip = false),
            child: const Text('\u{2715}', style: TextStyle(fontSize: 14, color: Color(0xFF065F46), opacity: 0.5)),
          ),
        ],
      ),
    );
  }

  List<Widget> _statItems() {
    final hr = _latestHealth?.hr ?? 72;
    final spo2 = _latestHealth?.spo2 ?? 98;
    final steps = _latestHealth?.steps ?? 3456;
    return [
      _statItem('\u{2764}', '${hr} bpm', '心率', success: true, trend: '+2', trendUp: true),
      _statItem('\u{1FAC7}', '${spo2}%', '血氧', success: true),
      _statItem('\u{1F6B6}', _formatSteps(steps), '今日步数', trend: '-12%', trendUp: false),
      _statItem('\u{1F50B}', '85%', '电池电量'),
      _statItem('\u{1F6F6}', '2.1 km', '今日距离'),
      _statItem('\u{26A0}', '${_recentAlerts.where((a) => a.status == 'pending').length}', '今日告警', danger: true),
    ];
  }

  String _formatSteps(int steps) {
    return steps.toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (m) => '${m[1]},');
  }

  Widget _statItem(String icon, String value, String label,
      {bool success = false, bool danger = false, String? trend, bool? trendUp}) {
    final bgColor = danger
        ? const Color(0xFFFEF2F2)
        : (success ? const Color(0xFFF0FDF4) : const Color(0xFFF9FAFB));
    final valueColor = danger ? AppTheme.statusDanger : (success ? AppTheme.statusNormal : const Color(0xFF1F2937));
    return Container(
      decoration: BoxDecoration(color: bgColor, borderRadius: BorderRadius.circular(14)),
      padding: const EdgeInsets.symmetric(vertical: 10, horizontal: 6),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (trend != null)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
              decoration: BoxDecoration(color: (trendUp ?? true) ? const Color(0xFFF0FDF4) : const Color(0xFFFEF2F2), borderRadius: BorderRadius.circular(4)),
              child: Text(trend, style: TextStyle(fontSize: 9, fontWeight: FontWeight.w600, color: (trendUp ?? true) ? AppTheme.statusNormal : AppTheme.statusDanger)),
            ),
          const SizedBox(height: 2),
          Text(icon, style: const TextStyle(fontSize: 20)),
          const SizedBox(height: 2),
          Text(value, style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: valueColor)),
          const SizedBox(height: 2),
          Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
        ],
      ),
    );
  }

  Widget _batteryRow() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 4, 16, 8),
      child: Row(
        children: [
          // Battery icon
          Container(
            width: 48,
            height: 22,
            decoration: BoxDecoration(border: Border.all(color: const Color(0xFF9CA3AF)), borderRadius: BorderRadius.circular(4)),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(2),
              child: Container(width: 34, color: AppTheme.statusNormal),
            ),
          ),
          const SizedBox(width: 6),
          const Text('|', style: TextStyle(fontSize: 14, color: Color(0xFF9CA3AF))),
          const SizedBox(width: 6),
          const Text('剩余约 3 天续航', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
        ],
      ),
    );
  }

  Widget _quickActions() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Row(
        children: [
          Expanded(
            child: _actionBtn('\u{26A0} 一键呼叫', AppTheme.statusDanger, Colors.white),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: _actionBtn('\u{1F4DE} 语音通话', AppTheme.primary, Colors.white),
          ),
          const SizedBox(width: 10),
          Expanded(
            child: _actionBtn('\u{1F4CD} 历史轨迹', const Color(0xFFF3F4F6), const Color(0xFF374151)),
          ),
        ],
      ),
    );
  }

  Widget _actionBtn(String text, Color bg, Color fg) {
    return ElevatedButton(
      onPressed: () {},
      style: ElevatedButton.styleFrom(backgroundColor: bg, foregroundColor: fg, padding: const EdgeInsets.symmetric(vertical: 10), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
      child: Text(text, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
    );
  }

  // ===== Bottom Nav =====
  Widget _buildBottomNav() {
    return Container(
      decoration: BoxDecoration(color: _darkMode ? const Color(0xFF111827) : Colors.white, border: Border(top: BorderSide(color: const Color(0xFFE5E7EB)))),
      padding: const EdgeInsets.only(bottom: 6),
      child: Row(
        children: [
          _navItem('\u{1F3E0}', '首页', 0),
          _navItem('\u{1F7EA}', '健康', 1),
          _navItem('\u{26A0}', '告警', 2, badge: _recentAlerts.where((a) => a.status == 'pending').length),
          _navItem('\u{1F48A}', '用药', 3),
          _navItem('\u{1F464}', '我的', 4),
        ],
      ),
    );
  }

  Widget _navItem(String icon, String label, int index, {int? badge}) {
    final isActive = index == _selectedIndex;
    return Expanded(
      child: GestureDetector(
        onTap: () {
          setState(() => _selectedIndex = index);
          if (index == 4) {
            Navigator.of(context).push(MaterialPageRoute(builder: (_) => const SettingsPage()));
          } else if (index == 5) {
            Navigator.of(context).push(MaterialPageRoute(builder: (_) => const WelfarePage()));
          }
        },
        child: Stack(
          alignment: Alignment.topCenter,
          children: [
            Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(icon, style: TextStyle(fontSize: 22, color: isActive ? AppTheme.primary : const Color(0xFF9CA3AF))),
                Text(label, style: TextStyle(fontSize: 10, fontWeight: isActive ? FontWeight.w600 : FontWeight.w400, color: isActive ? AppTheme.primary : const Color(0xFF9CA3AF))),
              ],
            ),
            if (badge != null && badge > 0)
              Positioned(
                top: 0,
                right: 24,
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                  decoration: const BoxDecoration(color: AppTheme.statusDanger, shape: BoxShape.circle),
                  child: Text('$badge', style: const TextStyle(fontSize: 9, color: Colors.white, fontWeight: FontWeight.w700)),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class ElderInfo {
  final String name, title, icon, tier;
  final Color bg;
  final bool online;
  ElderInfo({required this.name, required this.title, required this.icon, required this.bg, required this.online, required this.tier});
}

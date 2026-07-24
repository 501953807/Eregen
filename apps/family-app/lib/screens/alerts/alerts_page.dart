import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:pull_to_refresh/pull_to_refresh.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../../app_state.dart';
import '../../models/alert.dart';
import '../../services/ws_alert.dart';
import '../../services/offline_cache.dart';

/// Alerts center — v2 design: stats row, SOS quick action, filter chips,
/// priority tabs with counts, enriched alert cards with actions, dark mode.
class AlertsPage extends StatefulWidget {
  const AlertsPage({super.key});

  @override
  State<AlertsPage> createState() => _AlertsPageState();
}

class _AlertsPageState extends State<AlertsPage> {
  int _selectedIndex = 2;
  bool _loading = true;
  List<Alert> _allAlerts = [];
  String _activePriority = 'all'; // all | critical | warning | info
  bool _darkMode = false;
  late RefreshController _refreshController;
  AlertWebSocket? _ws;
  bool _wsConnected = false;

  String get _elderlyId => context.read<AppState>().elderlyId ?? '';
  String get _userId => context.read<AppState>().userId ?? '';

  @override
  void initState() {
    super.initState();
    _refreshController = RefreshController();
    _connectWebSocket();
    _populateFromCache();
    _fetchData();
  }

  @override
  void dispose() {
    _ws?.disconnect();
    _refreshController.dispose();
    super.dispose();
  }

  void _connectWebSocket() {
    if (_userId.isEmpty) return;
    final wsUrl = '${ApiClient.instance.baseUrl}/ws/alerts?user_id=$_userId';
    _ws = AlertWebSocket(
      wsUrl: wsUrl,
      onAlert: (alertJson) => _handleRealtimeAlert(alertJson),
      onDisconnected: () => setState(() => _wsConnected = false),
    );
    _ws!.connect();
    setState(() => _wsConnected = true);
  }

  void _handleRealtimeAlert(Map<String, dynamic> alertJson) {
    final alert = Alert.fromJson(alertJson);
    setState(() {
      if (!_allAlerts.any((a) => a.id == alert.id)) {
        _allAlerts.insert(0, alert);
      }
    });
    if (_elderlyId.isNotEmpty) {
      OfflineCache.cacheAlert(_elderlyId, alert.toJson());
    }
    if (mounted) {
      _showToast('新告警: ${alert.alertType}', color: AppTheme.statusDanger);
    }
  }

  void _populateFromCache() {
    if (_elderlyId.isEmpty) return;
    try {
      final cached = OfflineCache.getCachedAlerts(_elderlyId);
      if (cached.isNotEmpty) {
        final alerts = cached.map((a) => Alert.fromJson(a)).toList();
        setState(() => _allAlerts = alerts);
      }
    } catch (_) {}
  }

  Future<void> _fetchData() async {
    try {
      final resp = await ApiClient.instance.get('/alerts', query: {'limit': 50});
      final list = resp.data as List;
      final alerts = list.map((a) => Alert.fromJson(a as Map<String, dynamic>)).toList();

      if (_elderlyId.isNotEmpty) {
        for (final alert in alerts) {
          OfflineCache.cacheAlert(_elderlyId, alert.toJson());
        }
      }

      setState(() {
        _allAlerts = alerts;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  Future<void> _onRefresh() async {
    await _fetchData();
    _refreshController.refreshCompleted();
  }

  Future<void> _handleAlert(Alert alert) async {
    try {
      await ApiClient.instance.post('/alerts/${alert.id}/handle');
      setState(() {
        final idx = _allAlerts.indexWhere((a) => a.id == alert.id);
        if (idx >= 0) _allAlerts[idx] = Alert(
          id: alert.id, elderlyId: alert.elderlyId, alertType: alert.alertType,
          severity: alert.severity, status: 'resolved', metadata: alert.metadata,
          createdAt: alert.createdAt, resolvedAt: DateTime.now(),
        );
      });
      if (mounted) _showToast('已标记为处理');
    } catch (e) {
      if (mounted) _showToast('操作失败');
    }
  }

  // --- Filtering ---
  List<Alert> get _filtered {
    var list = _allAlerts;
    if (_activePriority == 'critical') list = list.where((a) => a.severity == 'P0').toList();
    else if (_activePriority == 'warning') list = list.where((a) => a.severity == 'P1').toList();
    else if (_activePriority == 'info') list = list.where((a) => a.severity == 'P2').toList();
    return list;
  }

  int get _p0Count => _allAlerts.where((a) => a.severity == 'P0' && a.status == 'pending').length;
  int get _p1Count => _allAlerts.where((a) => a.severity == 'P1' && a.status == 'pending').length;
  int get _p2Count => _allAlerts.where((a) => a.severity == 'P2' && a.status == 'pending').length;
  int get _resolvedCount => _allAlerts.where((a) => a.status == 'resolved').length;

  void _showToast(String msg, {Color? color}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), duration: const Duration(seconds: 1), backgroundColor: color),
    );
  }

  // --- Mock enrichment data for display ---
  String _alertDescription(Alert alert) {
    return alert.metadata?['description'] ?? '暂无详细描述';
  }

  String _alertLocation(Alert alert) {
    return alert.metadata?['location'] ?? '未知位置';
  }

  String _alertIcon(Alert alert) {
    final type = alert.alertType.toLowerCase();
    if (type.contains('sos')) return '\u{26A0}';
    if (type.contains('跌倒') || type.contains('fall')) return '\u{1F982}';
    if (type.contains('围栏') || type.contains('geofence')) return '\u{1F3E0}';
    if (type.contains('心率') || type.contains('heart')) return '\u{2764}';
    if (type.contains('用药') || type.contains('med')) return '\u{1F48A}';
    if (type.contains('电量') || type.contains('battery')) return '\u{1F50B}';
    if (type.contains('离线') || type.contains('offline')) return '\u{1F50C}';
    return '\u{1F4AC}';
  }

  Color _alertIconBg(Alert alert) {
    if (alert.severity == 'P0') return const Color(0xFFFEF2F2);
    if (alert.severity == 'P1') return const Color(0xFFFFFBEB);
    return const Color(0xFFDBEAFE);
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filtered;
    return Scaffold(
      backgroundColor: _darkMode ? const Color(0xFF111827) : const Color(0xFFF3F4F6),
      body: SmartRefresher(
        controller: _refreshController,
        onRefresh: _onRefresh,
        enablePullDown: true,
        enablePullUp: false,
        child: _loading && _allAlerts.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : Column(
          children: [
            _buildTopBar(),
            Expanded(
              child: SingleChildScrollView(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    _buildStatsRow(),
                    const SizedBox(height: 12),
                    _buildSOSQuickAction(),
                    const SizedBox(height: 12),
                    _buildFilterChips(),
                    const SizedBox(height: 8),
                    _buildPriorityTabs(),
                    const SizedBox(height: 12),
                    ...filtered.map((alert) => Padding(
                      padding: const EdgeInsets.only(bottom: 10),
                      child: _buildAlertItem(alert),
                    )),
                    if (filtered.isEmpty) const Center(child: Text('暂无告警', style: TextStyle(color: Color(0xFF9CA3AF)))),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _selectedIndex,
        onTabSelected: (i) => setState(() => _selectedIndex = i),
      ),
    );
  }

  // ===== Top Bar =====
  Widget _buildTopBar() {
    return Container(
      color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
      padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              GestureDetector(
                onTap: () => Navigator.of(context).pop(),
                child: Container(
                  width: 36, height: 36, decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(18)),
                  child: const Icon(Icons.arrow_back_ios_new, size: 16),
                ),
              ),
              const Expanded(
                child: Center(child: Text('告警中心', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
              ),
              Row(
                children: [
                  GestureDetector(
                    onTap: () => setState(() => _darkMode = !_darkMode),
                    child: Container(width: 36, height: 36, decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(18)),
                      child: Center(child: Text(_darkMode ? '\u{1F31E}' : '\u{2600}', style: const TextStyle(fontSize: 14))),
                    ),
                  ),
                  const SizedBox(width: 6),
                  GestureDetector(
                    onTap: () {
                      setState(() => _allAlerts = _allAlerts.where((a) => a.status == 'pending').toList());
                      _showToast('已清除已处理告警');
                    },
                    child: const Text('全部已读', style: TextStyle(fontSize: 13, color: AppTheme.primary, fontWeight: FontWeight.w600)),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 10),
          // Elder selector
          SizedBox(
            height: 56,
            child: ListView.separated(
              scrollDirection: Axis.horizontal,
              itemCount: 2,
              separatorBuilder: (_, __) => const SizedBox(width: 10),
              itemBuilder: (ctx, i) {
                final names = ['爷爷 张三丰', '奶奶 李秀英'];
                final icons = ['\u{1F468}', '\u{1F469}'];
                final bgs = [const Color(0xFFFFF3C7), const Color(0xFFFCE7F3)];
                return Container(
                  width: 130,
                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                  decoration: BoxDecoration(
                    color: i == 0 ? const Color(0xFFDBEAFE) : Colors.white,
                    border: Border.all(color: i == 0 ? AppTheme.primary : const Color(0xFFE5E7EB)),
                    borderRadius: BorderRadius.circular(24),
                  ),
                  child: Row(
                    children: [
                      Container(width: 36, height: 36, decoration: BoxDecoration(color: bgs[i], shape: BoxShape.circle),
                        child: Center(child: Text(icons[i], style: const TextStyle(fontSize: 18))),
                      ),
                      const SizedBox(width: 8),
                      Text(names[i], style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    ],
                  ),
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  // ===== Stats Row =====
  Widget _buildStatsRow() {
    return Row(
      children: [
        _statCard('$_p0Count', '紧急', AppTheme.statusDanger, const Color(0xFFFEF2F2)),
        const SizedBox(width: 8),
        _statCard('$_p1Count', '警告', AppTheme.statusWarning, const Color(0xFFFFFBEB)),
        const SizedBox(width: 8),
        _statCard('$_resolvedCount', '已处理', AppTheme.statusNormal, const Color(0xFFF0FDF4)),
      ],
    );
  }

  Widget _statCard(String num, String label, Color numColor, Color bgColor) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 10),
        decoration: BoxDecoration(color: bgColor, borderRadius: BorderRadius.circular(14), border: Border(top: BorderSide(color: numColor, width: 3))),
        child: Column(
          children: [
            Text(num, style: TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: numColor)),
            Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
          ],
        ),
      ),
    );
  }

  // ===== SOS Quick Action =====
  Widget _buildSOSQuickAction() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        gradient: const LinearGradient(colors: [Color(0xFFFEF2F2), Color(0xFFFEE2E2)]),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFFfecaca)),
      ),
      child: Row(
        children: [
          Container(
            width: 44, height: 44, decoration: BoxDecoration(color: AppTheme.statusDanger, borderRadius: BorderRadius.circular(22)),
            child: const Center(child: Text('\u{26A1}', style: TextStyle(fontSize: 22))),
          ),
          const SizedBox(width: 12),
          const Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('一键紧急呼叫', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: Color(0xFF991B1B))),
                Text('立即联系家属和急救中心', style: TextStyle(fontSize: 12, color: Color(0xFFB91C1C))),
              ],
            ),
          ),
          ElevatedButton(
            onPressed: () => _showToast('正在发起紧急呼叫...'),
            style: ElevatedButton.styleFrom(backgroundColor: AppTheme.statusDanger, foregroundColor: Colors.white, padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
            child: const Text('紧急呼叫', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700)),
          ),
        ],
      ),
    );
  }

  // ===== Filter Chips =====
  Widget _buildFilterChips() {
    final chips = ['\u{1F50D} 全部', '\u{1F512} 仅未读', '\u{1F4CB} 待处理', '\u{1F4C0} 按时间'];
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: chips.map((c) {
          final isActive = c.contains('全部');
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
              decoration: BoxDecoration(
                color: isActive ? const Color(0xFFDBEAFE) : Colors.white,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: isActive ? AppTheme.primary : const Color(0xFFE5E7EB)),
              ),
              child: Text(c, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: isActive ? AppTheme.primary : const Color(0xFF4B5563))),
            ),
          );
        }).toList(),
      ),
    );
  }

  // ===== Priority Tabs =====
  Widget _buildPriorityTabs() {
    return Container(
      decoration: BoxDecoration(color: const Color(0xFFE5E7EB), borderRadius: BorderRadius.circular(12)),
      padding: const EdgeInsets.all(3),
      child: Row(
        children: [
          _priorityTab('全部', _p0Count + _p1Count + _p2Count, 'all', isActive: true),
          const SizedBox(width: 4),
          _priorityTab('\u{1F534} P0', _p0Count, 'critical'),
          const SizedBox(width: 4),
          _priorityTab('\u{26A0} P1', _p1Count, 'warning'),
          const SizedBox(width: 4),
          _priorityTab('\u{1F4DC} P2', _p2Count, 'info'),
        ],
      ),
    );
  }

  Widget _priorityTab(String label, int count, String priority, {bool isActive = false}) {
    return Expanded(
      child: GestureDetector(
        onTap: () => setState(() => _activePriority = priority),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 8),
          decoration: BoxDecoration(
            color: isActive ? Colors.white : null,
            borderRadius: BorderRadius.circular(10),
            boxShadow: isActive ? [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4)] : [],
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(label, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: isActive ? const Color(0xFF1F2937) : const Color(0xFF6B7280))),
              if (count > 0) ...[
                const SizedBox(width: 4),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                  decoration: BoxDecoration(color: isActive ? (priority == 'critical' ? AppTheme.statusDanger : (priority == 'warning' ? AppTheme.statusWarning : AppTheme.primary)) : const Color(0xFF9CA3AF), borderRadius: BorderRadius.circular(8)),
                  child: Text('$count', style: const TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.w700)),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  // ===== Alert Item =====
  Widget _buildAlertItem(Alert alert) {
    final isPending = alert.status == 'pending';
    final isRead = alert.status == 'resolved';
    final isCritical = alert.severity == 'P0';
    final borderColor = isCritical ? AppTheme.statusDanger : (alert.severity == 'P1' ? AppTheme.statusWarning : AppTheme.primary);
    final bg = !isRead ? const Color(0xFFF9FAFB) : Colors.white;

    return Container(
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : bg,
        borderRadius: BorderRadius.circular(16),
        border: Border(left: BorderSide(color: borderColor, width: 4)),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, 1))],
      ),
      padding: const EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                width: 40, height: 40, decoration: BoxDecoration(color: _alertIconBg(alert), borderRadius: BorderRadius.circular(20)),
                child: Center(child: Text(_alertIcon(alert), style: const TextStyle(fontSize: 20))),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(alert.alertType, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
                        const SizedBox(width: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                          decoration: BoxDecoration(
                            color: isCritical ? const Color(0xFFFEF2F2) : (alert.severity == 'P1' ? const Color(0xFFFFFBEB) : const Color(0xFFDBEAFE)),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(alert.severity, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: isCritical ? AppTheme.statusDanger : (alert.severity == 'P1' ? const Color(0xFFD97706) : AppTheme.primary))),
                        ),
                        if (!isRead) ...[
                          const SizedBox(width: 6),
                          Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                            decoration: BoxDecoration(color: const Color(0xFFFEF2F2), borderRadius: BorderRadius.circular(8)),
                            child: const Text('未读', style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: AppTheme.statusDanger)),
                          ),
                        ],
                      ],
                    ),
                    const SizedBox(height: 2),
                    Text(_alertDescription(alert), style: TextStyle(fontSize: 12, color: _darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280), height: 1.4)),
                  ],
                ),
              ),
              // Unread dot
              if (!isRead)
                Container(width: 8, height: 8, decoration: const BoxDecoration(color: AppTheme.statusDanger, shape: BoxShape.circle)),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Text('\u{1F4CD} ${_alertLocation(alert)}', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
              const SizedBox(width: 12),
              Text('\u{1F550} ${_timeAgo(alert.createdAt)}', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
            ],
          ),
          if (isPending) ...[
            const SizedBox(height: 12),
            Divider(color: const Color(0xFFF3F4F6), height: 1),
            const SizedBox(height: 12),
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                _alertActionButton('\u{1F4DE} 电话回拨', AppTheme.primary, Colors.white),
                _alertActionButton('\u{1F4CD} 查看位置', null, const Color(0xFF4B5563)),
                _alertActionButton('标记已读', null, const Color(0xFF6B7280), onTap: () => _handleAlert(alert)),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _alertActionButton(String label, Color? bg, Color fg, {VoidCallback? onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        flex: 1,
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(
          color: bg ?? Colors.white,
          borderRadius: BorderRadius.circular(10),
          border: Border.all(color: bg != null ? bg : const Color(0xFFE5E7EB)),
        ),
        child: Center(child: Text(label, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: fg ?? const Color(0xFF4B5563)))),
      ),
    );
  }

  String _timeAgo(DateTime dt) {
    final now = DateTime.now();
    final diff = now.difference(dt);
    if (diff.inMinutes < 1) return '刚刚';
    if (diff.inHours < 1) return '${diff.inMinutes}分钟前';
    if (diff.inDays < 1) return '${diff.inHours}小时前';
    return '${diff.inDays}天前';
  }
}

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

/// Alerts center page — fetches from GET /alerts, supports POST /alerts/:id/handle and pull-to-refresh.
/// Integrates real-time WebSocket alerts and Hive offline cache.
class AlertsPage extends StatefulWidget {
  const AlertsPage({super.key});

  @override
  State<AlertsPage> createState() => _AlertsPageState();
}

class _AlertsPageState extends State<AlertsPage> {
  int _selectedIndex = 2;
  String _activeFilter = '全部';
  bool _loading = true;
  List<Alert> _allAlerts = [];
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

  List<Alert> get _filtered {
    if (_activeFilter == '全部') return _allAlerts;
    if (_activeFilter == '未处理') return _allAlerts.where((a) => a.status == 'pending').toList();
    if (_activeFilter == 'SOS') return _allAlerts.where((a) => a.alertType.contains('SOS')).toList();
    if (_activeFilter == '跌倒') return _allAlerts.where((a) => a.alertType.contains('跌倒')).toList();
    if (_activeFilter == '健康') return _allAlerts.where((a) => !a.alertType.contains('SOS') && !a.alertType.contains('跌倒')).toList();
    return _allAlerts;
  }

  int get _p0Count => _allAlerts.where((a) => a.severity == 'P0' && a.status == 'pending').length;
  int get _p1Count => _allAlerts.where((a) => a.severity == 'P1' && a.status == 'pending').length;
  int get _p2Count => _allAlerts.where((a) => a.severity == 'P2' && a.status == 'pending').length;

  void _showToast(String msg, {Color? color}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), duration: const Duration(seconds: 1), backgroundColor: color),
    );
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filtered;
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SmartRefresher(
        controller: _refreshController,
        onRefresh: _onRefresh,
        enablePullDown: true,
        enablePullUp: false,
        child: _loading && _allAlerts.isEmpty
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  _buildHeader(),
                  _buildStatsCards(),
                  _buildFilterTabs(),
                  SliverPadding(
                    padding: const EdgeInsets.symmetric(horizontal: 20),
                    sliver: SliverList(
                      delegate: SliverChildBuilderDelegate((ctx, i) {
                        if (i >= filtered.length) return null;
                        return _buildAlertItem(filtered[i]);
                      }, childCount: filtered.length),
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

  Widget _buildHeader() {
    return SliverToBoxAdapter(
      child: Container(
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
        color: AppTheme.bgCard,
        child: Row(
          children: [
            IconButton(
              icon: const Icon(Icons.arrow_back_ios_new, size: 18),
              onPressed: () => Navigator.of(context).pop(),
            ),
            const Expanded(
              child: Text('告警中心', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
            ),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (_wsConnected)
                  Container(
                    width: 8,
                    height: 8,
                    decoration: const BoxDecoration(
                      color: Color(0xFF4ADE80),
                      shape: BoxShape.circle,
                    ),
                  ),
                const SizedBox(width: 8),
                IconButton(
                  icon: const Icon(Icons.search),
                  onPressed: () => _showSearchDialog(context),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatsCards() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        child: Row(
          children: [
            _statCard('$_p0Count', 'P0 紧急', const Color(0xFFFF6B6B), const Color(0xFFEE5A24)),
            const SizedBox(width: 10),
            _statCard('$_p1Count', 'P1 重要', const Color(0xFFFFA726), const Color(0xFFFB8C00)),
            const SizedBox(width: 10),
            _statCard('$_p2Count', 'P2 通知', const Color(0xFF42A5F5), const Color(0xFF1E88E5)),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterTabs() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.only(left: 20, right: 20, bottom: 12),
        child: Row(
          children: ['全部', '未处理', 'SOS', '跌倒', '健康'].map((f) {
            final isActive = f == _activeFilter;
            return Padding(
              padding: const EdgeInsets.only(right: 16),
              child: GestureDetector(
                onTap: () => setState(() => _activeFilter = f),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      f,
                      style: TextStyle(
                        fontSize: 13,
                        fontWeight: FontWeight.w600,
                        color: isActive ? const Color(0xFFFF6B6B) : const Color(0xFF888888),
                      ),
                    ),
                    if (isActive)
                      Container(
                        width: double.infinity,
                        height: 2,
                        color: const Color(0xFFFF6B6B),
                      ),
                  ],
                ),
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _buildAlertItem(Alert alert) {
    final isPending = alert.status == 'pending';
    final isResolved = !isPending;
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isResolved ? const Color(0xFFFAFBFE) : const Color(0xFFFFF5F5),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFF0F0F0)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.04),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
                decoration: BoxDecoration(
                  color: const Color(0xFFE3F2FD),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  alert.alertType,
                  style: const TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                    color: Color(0xFF1565C0),
                  ),
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: alert.severity == 'P0'
                      ? const Color(0xFFFF6B6B)
                      : (alert.severity == 'P1' ? const Color(0xFFFFA726) : const Color(0xFF90CAF9)),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  alert.severity,
                  style: const TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(alert.alertType, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700)),
          const SizedBox(height: 4),
          Text(
            alert.metadata?['description'] ?? '暂无详细描述',
            style: const TextStyle(fontSize: 12, color: Color(0xFF666666), height: 1.5),
          ),
          const SizedBox(height: 8),
          Text(
            '🕐 ${alert.createdAt.toLocal().toString().substring(5, 16)}',
            style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA)),
          ),
          if (isPending) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              children: [
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFEBEE),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Text(
                    '📞 立即呼叫',
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFFFF6B6B)),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                  decoration: BoxDecoration(
                    color: const Color(0xFFF0F0F5),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: const Text(
                    '📍 查看位置',
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFF666666)),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
                  decoration: BoxDecoration(
                    color: const Color(0xFFE8F5E9),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: GestureDetector(
                    onTap: () => _handleAlert(alert),
                    child: const Text(
                      '✓ 标记处理',
                      style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFF2E7D32)),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  void _showSearchDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('搜索告警'),
        content: TextField(
          decoration: const InputDecoration(hintText: '输入关键词搜索...'),
          onSubmitted: (keyword) {
            Navigator.of(ctx).pop();
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text('搜索: $keyword')),
            );
          },
        ),
        actions: [
          TextButton(onPressed: () => Navigator.of(ctx).pop(), child: const Text('取消')),
        ],
      ),
    );
  }

  Widget _statCard(String num, String label, Color startColor, Color endColor) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(12),
          gradient: LinearGradient(colors: [startColor, endColor]),
        ),
        child: Column(
          children: [
            Text(num, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: Colors.white)),
            Text(
              label,
              style: const TextStyle(fontSize: 10, color: Colors.white),
            ),
          ],
        ),
      ),
    );
  }
}

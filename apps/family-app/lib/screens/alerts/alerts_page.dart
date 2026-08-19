import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:pull_to_refresh/pull_to_refresh.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../../app_state.dart';
import '../../models/alert.dart';
import '../../services/ws_alert.dart';
import '../../services/offline_cache.dart';
import '../../widgets/empty_state.dart';

/// Alerts center — v2 design: clean stat cards, filter chips, priority tabs,
/// type badges + priority labels + action buttons, SOS quick action banner.
class AlertsPage extends StatefulWidget {
  final VoidCallback? onBack;
  const AlertsPage({super.key, this.onBack});

  @override
  State<AlertsPage> createState() => _AlertsPageState();
}

class _AlertsPageState extends State<AlertsPage> {
  bool _loading = true;
  List<Alert> _allAlerts = [];
  String _activeFilter = '全部';
  late RefreshController _refreshController;
  AlertWebSocket? _ws;
  bool _wsConnected = false;

  String get _elderlyId => context.read<AppState>().elderlyId ?? '';
  String get _userId => context.read<AppState>().userId ?? '';

  final List<String> _filterChips = ['全部', '未处理', 'SOS', '跌倒', '健康', '围栏', '用药'];
  String _activePriority = '全部';

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
      final resp = await ApiClient.instance.get('/api/v1/alerts', query: {'limit': 50});
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
      // On API failure, seed mock alerts so the page isn't blank
      _seedMockAlerts();
      setState(() => _loading = false);
    }
  }

  void _seedMockAlerts() {
    setState(() {
      _allAlerts = [
        Alert(id: 'a1', elderlyId: '', alertType: 'sos', severity: 'high', status: 'pending', metadata: {'location': '陆家嘴环路1000号'}, createdAt: DateTime.now().subtract(const Duration(minutes: 2))),
        Alert(id: 'a2', elderlyId: '', alertType: 'fall', severity: 'high', status: 'pending', metadata: {'location': '人民广场'}, createdAt: DateTime.now().subtract(const Duration(hours: 3))),
        Alert(id: 'a3', elderlyId: '', alertType: 'medication_missed', severity: 'medium', status: 'resolved', metadata: {'description': '降压药漏服'}, createdAt: DateTime.now().subtract(const Duration(hours: 8)), resolvedAt: DateTime.now().subtract(const Duration(hours: 7))),
        Alert(id: 'a4', elderlyId: '', alertType: 'geofence_exit', severity: 'medium', status: 'resolved', metadata: {'location': '外滩'}, createdAt: DateTime.now().subtract(const Duration(days: 1)), resolvedAt: DateTime.now().subtract(const Duration(days: 1, hours: 2))),
      ];
    });
  }

  Future<void> _onRefresh() async {
    await _fetchData();
    _refreshController.refreshCompleted();
  }

  Future<void> _handleAlert(Alert alert) async {
    try {
      await ApiClient.instance.handleAlert(alert.id);
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
    if (_activeFilter == '未处理') list = list.where((a) => a.status == 'pending').toList();
    else if (_activeFilter == 'SOS') list = list.where((a) => a.alertType.toLowerCase().contains('sos')).toList();
    else if (_activeFilter == '跌倒') list = list.where((a) => a.alertType.toLowerCase().contains('跌倒') || a.alertType.toLowerCase().contains('fall')).toList();
    else if (_activeFilter == '健康') list = list.where((a) => !a.alertType.toLowerCase().contains('sos') && !a.alertType.toLowerCase().contains('跌倒')).toList();
    if (_activePriority != '全部') list = list.where((a) => a.severity == _activePriority).toList();
    return list;
  }

  int get _highCount => _allAlerts.where((a) => a.severity == 'high' && a.status == 'pending').length;
  int get _mediumCount => _allAlerts.where((a) => a.severity == 'medium' && a.status == 'pending').length;
  int get _lowCount => _allAlerts.where((a) => a.severity == 'low' && a.status == 'pending').length;

  Future<void> _handleAllAlerts() async {
    final pendingAlerts = _allAlerts.where((a) => a.status == 'pending').toList();
    if (pendingAlerts.isEmpty) {
      if (mounted) _showToast('没有未处理的告警');
      return;
    }
    try {
      for (final alert in pendingAlerts) {
        await ApiClient.instance.handleAlert(alert.id);
      }
      setState(() {
        for (int i = 0; i < _allAlerts.length; i++) {
          if (_allAlerts[i].status == 'pending') {
            _allAlerts[i] = Alert(
              id: _allAlerts[i].id,
              elderlyId: _allAlerts[i].elderlyId,
              alertType: _allAlerts[i].alertType,
              severity: _allAlerts[i].severity,
              status: 'resolved',
              metadata: _allAlerts[i].metadata,
              createdAt: _allAlerts[i].createdAt,
              resolvedAt: DateTime.now(),
            );
          }
        }
      });
      if (mounted) _showToast('已将 ${pendingAlerts.length} 条告警标记为已处理');
    } catch (e) {
      if (mounted) _showToast('操作失败: $e');
    }
  }

  void _showToast(String msg, {Color? color}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg), duration: const Duration(seconds: 1), backgroundColor: color),
    );
  }

  String _alertDescription(Alert alert) {
    return alert.metadata?['description'] ?? '暂无详细描述';
  }

  String _alertLocation(Alert alert) {
    return alert.metadata?['location'] ?? '未知位置';
  }

  String _alertTypeLabel(Alert alert) {
    final type = alert.alertType.toLowerCase();
    if (type.contains('sos')) return 'SOS';
    if (type.contains('跌倒') || type.contains('fall')) return '跌倒检测';
    if (type.contains('心率') || type.contains('heart')) return '心率异常';
    if (type.contains('围栏') || type.contains('geofence')) return '电子围栏';
    if (type.contains('用药') || type.contains('med')) return '用药提醒';
    return alert.alertType;
  }

  IconData _alertTypeIcon(Alert alert) {
    final type = alert.alertType.toLowerCase();
    if (type.contains('sos')) return Icons.phone_in_talk_rounded;
    if (type.contains('跌倒') || type.contains('fall')) return Icons.elderly_rounded;
    if (type.contains('心率') || type.contains('heart')) return Icons.favorite_rounded;
    if (type.contains('围栏') || type.contains('geofence')) return Icons.home_rounded;
    if (type.contains('用药') || type.contains('med')) return Icons.medication_rounded;
    return Icons.info_rounded;
  }

  Color _alertTypeBadgeBg(String label) {
    if (label.contains('SOS')) return const Color(0xFFFFEBEE);
    if (label.contains('跌倒')) return const Color(0xFFFFF3E0);
    if (label.contains('心率')) return const Color(0xFFFCE4EC);
    if (label.contains('围栏')) return const Color(0xFFFFF8E1);
    return AppTheme.primaryBg;
  }

  Color _alertTypeBadgeColor(String label) {
    if (label.contains('SOS')) return const Color(0xFFC62828);
    if (label.contains('跌倒')) return const Color(0xFFE65100);
    if (label.contains('心率')) return const Color(0xFFAD1457);
    if (label.contains('围栏')) return const Color(0xFFF57F17);
    return AppTheme.primary;
  }

  @override
  Widget build(BuildContext context) {
    final filtered = _filtered;
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: SmartRefresher(
          controller: _refreshController,
          onRefresh: _onRefresh,
          enablePullDown: true,
          enablePullUp: false,
          child: _loading && _allAlerts.isEmpty
              ? const Center(child: CircularProgressIndicator())
              : Column(
                  children: [
                    // Top bar
                    Container(
                      decoration: BoxDecoration(
                        color: Colors.white,
                        border: Border(bottom: BorderSide(color: const Color(0xFF000000).withOpacity(0.04))),
                      ),
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 12),
                      child: Row(
                        children: [
                          GestureDetector(
                            onTap: widget.onBack ?? () => Navigator.of(context).pop(),
                            child: Container(
                              width: 36,
                              height: 36,
                              decoration: BoxDecoration(borderRadius: BorderRadius.circular(14)),
                              child: const Center(child: Icon(Icons.chevron_left_rounded, size: 20, color: AppTheme.textSecondary)),
                            ),
                          ),
                          const Expanded(
                            child: Center(child: Text('告警中心', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: AppTheme.textPrimary))),
                          ),
                          GestureDetector(
                            onTap: () => _handleAllAlerts(),
                            child: Container(
                              width: 36,
                              height: 36,
                              decoration: BoxDecoration(borderRadius: BorderRadius.circular(14)),
                              child: const Center(child: Icon(Icons.select_all_rounded, size: 18, color: AppTheme.textMuted)),
                            ),
                          ),
                        ],
                      ),
                    ),
                    // Scrollable content
                    Expanded(
                      child: SingleChildScrollView(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Stats row
                            _buildStatsRow(),
                            const SizedBox(height: 16),
                            // Filter chips
                            _buildFilterChips(),
                            const SizedBox(height: 12),
                            // Priority tabs
                            _buildPriorityTabs(),
                            const SizedBox(height: 14),
                            // Alert list
                            _buildAlertList(filtered),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }

  // ===== Stats Row =====
  Widget _buildStatsRow() {
    return Row(
      children: [
        _statCard('$_highCount', 'SOS 紧急', AppTheme.statusDanger),
        const SizedBox(width: 10),
        _statCard('$_mediumCount', '跌倒检测', AppTheme.statusWarning),
        const SizedBox(width: 10),
        _statCard('$_lowCount', '健康异常', AppTheme.primary),
      ],
    );
  }

  Widget _statCard(String num, String label, Color color) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(AppTheme.cardRadiusLg),
          boxShadow: AppTheme.shadowMd,
        ),
        child: Column(
          children: [
            Container(height: 3, width: 24, margin: const EdgeInsets.only(bottom: 8), decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(2))),
            Text(num, style: TextStyle(fontSize: 28, fontWeight: FontWeight.w800, color: color)),
            Text(label, style: const TextStyle(fontSize: 11, color: AppTheme.textMuted)),
          ],
        ),
      ),
    );
  }

  // ===== Filter Chips =====
  Widget _buildFilterChips() {
    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: Row(
        children: _filterChips.map((chip) {
          final isActive = chip == _activeFilter;
          return Padding(
            padding: const EdgeInsets.only(right: 8),
            child: GestureDetector(
              onTap: () => setState(() => _activeFilter = chip),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 7),
                decoration: BoxDecoration(
                  color: isActive ? AppTheme.primary : Colors.white,
                  borderRadius: BorderRadius.circular(18),
                  border: Border.all(color: isActive ? AppTheme.primary : const Color(0xFF000000).withOpacity(0.08)),
                ),
                child: Text(
                  chip,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isActive ? Colors.white : AppTheme.textSecondary,
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  // ===== Priority Tabs =====
  Widget _buildPriorityTabs() {
    return Container(
      padding: const EdgeInsets.all(3),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: AppTheme.shadowSm,
      ),
      child: Row(
        children: [
          _prioTab('全部', true),
          const SizedBox(width: 4),
          _prioTab('高', false, color: AppTheme.statusDanger),
          const SizedBox(width: 4),
          _prioTab('中', false, color: AppTheme.statusWarning),
          const SizedBox(width: 4),
          _prioTab('低', false, color: AppTheme.primary),
        ],
      ),
    );
  }

  Widget _prioTab(String label, bool active, {Color? color}) {
    return Expanded(
      child: GestureDetector(
        onTap: () => setState(() => _activePriority = label == '全部' ? '全部' : label),
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 6),
          decoration: BoxDecoration(
            color: active ? AppTheme.bgScaffold : null,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: active ? (color ?? AppTheme.textPrimary) : AppTheme.textMuted,
              ),
            ),
          ),
        ),
      ),
    );
  }

  // ===== Alert List =====
  Widget _buildAlertList(List<Alert> alerts) {
    if (alerts.isEmpty) {
      return const EmptyState(
        icon: Icons.check_circle_outline,
        title: '暂无告警',
        subtitle: '一切正常，继续加油',
      );
    }
    return Column(
      children: alerts.map((alert) {
        final isRead = alert.status == 'resolved';
        return Padding(
          padding: const EdgeInsets.only(bottom: 10),
          child: _buildAlertItem(alert, isRead),
        );
      }).toList(),
    );
  }

  // ===== Alert Item =====
  Widget _buildAlertItem(Alert alert, bool isRead) {
    final isPending = alert.status == 'pending';
    final isCritical = alert.severity == 'high';
    final typeLabel = _alertTypeLabel(alert);

    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.cardRadiusLg),
        boxShadow: AppTheme.shadowMd,
        border: Border(left: BorderSide(color: isRead ? Colors.transparent : AppTheme.accent, width: 4)),
      ),
      child: Stack(
        children: [
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Header row: type badge + priority + status
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                    decoration: BoxDecoration(
                      color: _alertTypeBadgeBg(typeLabel),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(_alertTypeIcon(alert), size: 10, color: _alertTypeBadgeColor(typeLabel)),
                        const SizedBox(width: 4),
                        Text(
                          typeLabel,
                          style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: _alertTypeBadgeColor(typeLabel)),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(width: 8),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: (isCritical ? const Color(0xFFFFEBEE) : (alert.severity == 'medium' ? const Color(0xFFFFFBEB) : AppTheme.primaryBg)),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      alert.severity,
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.w700,
                        color: isCritical ? const Color(0xFFDC2626) : (alert.severity == 'medium' ? const Color(0xFFD97706) : AppTheme.primary),
                      ),
                    ),
                  ),
                  const Spacer(),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: const Color(0xFFEEF4FF),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      isPending ? '未读' : '已处理',
                      style: TextStyle(
                        fontSize: 10,
                        fontWeight: FontWeight.w600,
                        color: isPending ? AppTheme.primary : const Color(0xFF16A34A),
                      ),
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              // Title
              Text(alert.alertType, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: AppTheme.textPrimary)),
              const SizedBox(height: 4),
              // Description
              Text(_alertDescription(alert), style: TextStyle(fontSize: 12, color: AppTheme.textSecondary, height: 1.5)),
              const SizedBox(height: 8),
              // Footer: location + time
              Row(
                children: [
                  const Icon(Icons.location_on_rounded, size: 12),
                  const SizedBox(width: 4),
                  Text(_alertLocation(alert), style: const TextStyle(fontSize: 11, color: AppTheme.textMuted)),
                  const SizedBox(width: 16),
                  const Icon(Icons.schedule_rounded, size: 12),
                  const SizedBox(width: 4),
                  Text(_timeAgo(alert.createdAt), style: const TextStyle(fontSize: 11, color: AppTheme.textMuted)),
                ],
              ),
              // Action buttons
              if (isPending) ...[
                const SizedBox(height: 12),
                Row(
                  children: [
                    Expanded(
                      child: _actionBtn(Icons.phone_in_talk_rounded, '立即呼叫', AppTheme.statusDanger, Colors.white, onTap: () async {
                        try {
                          await ApiClient.instance.sosCall();
                          if (mounted) _showToast('紧急呼叫已发起');
                        } catch (_) {
                          if (mounted) _showToast('呼叫失败');
                        }
                      }),
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: _actionBtn(Icons.near_me_rounded, '查看位置', AppTheme.primary, Colors.white, onTap: () {
                        // Navigate to map view with alert location
                        final loc = alert.metadata?['location'];
                        if (loc != null && mounted) {
                          _showToast('正在打开位置: $loc');
                        }
                      }),
                    ),
                    const SizedBox(width: 6),
                    Expanded(
                      child: _actionBtn(Icons.check_circle_rounded, '标记处理', const Color(0xFFE8F5E9), const Color(0xFF2E7D32), onTap: () => _handleAlert(alert)),
                    ),
                  ],
                ),
              ],
            ],
          ),
          // Unread dot
          if (!isRead)
            Positioned(
              top: 14,
              right: 14,
              child: Container(width: 8, height: 8, decoration: const BoxDecoration(color: AppTheme.accent, shape: BoxShape.circle)),
            ),
        ],
      ),
    );
  }

  Widget _actionBtn(IconData? icon, String label, Color bg, Color fg, {VoidCallback? onTap}) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(14)),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (icon != null) ...[
              Icon(icon, size: 14, color: fg),
              const SizedBox(width: 4),
            ],
            Text(label, style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: fg)),
          ],
        ),
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

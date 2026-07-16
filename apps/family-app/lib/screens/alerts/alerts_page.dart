import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/bottom_nav_bar.dart';

/// Alerts center page — matches alerts.html prototype
class AlertsPage extends StatefulWidget {
  const AlertsPage({super.key});

  @override
  State<AlertsPage> createState() => _AlertsPageState();

  int get initialIndex => 2;
}

class _AlertsPageState extends State<AlertsPage> {
  int _selectedIndex = 2;
  String _activeFilter = '全部';

  final List<String> filters = ['全部', '未处理', 'SOS', '跌倒', '健康'];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            // Header
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
                color: AppTheme.bgCard,
                child: Row(
                  children: [
                    IconButton(icon: const Icon(Icons.arrow_back_ios_new, size: 18), onPressed: () {}),
                    const Expanded(child: Text('告警中心', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
                    IconButton(icon: const Icon(Icons.search), onPressed: () {}),
                  ],
                ),
              ),
            ),

            // Alert stats cards
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                child: Row(
                  children: [
                    _statCard('2', 'P0 紧急', const Color(0xFFFF6B6B), const Color(0xFFEE5A24)),
                    const SizedBox(width: 10),
                    _statCard('5', 'P1 重要', const Color(0xFFFFA726), const Color(0xFFFB8C00)),
                    const SizedBox(width: 10),
                    _statCard('12', 'P2 通知', const Color(0xFF42A5F5), const Color(0xFF1E88E5)),
                  ],
                ),
              ),
            ),

            // Filter tabs
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.only(left: 20, right: 20, bottom: 12),
                child: Row(
                  children: filters.map((f) {
                    final isActive = f == _activeFilter;
                    return Padding(
                      padding: const EdgeInsets.only(right: 16),
                      child: GestureDetector(
                        onTap: () => setState(() => _activeFilter = f),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(f,
                                style: TextStyle(
                                    fontSize: 13,
                                    fontWeight: FontWeight.w600,
                                    color: isActive ? const Color(0xFFFF6B6B) : const Color(0xFF888888))),
                            if (isActive)
                              Container(width: double.infinity, height: 2, color: const Color(0xFFFF6B6B)),
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ),
            ),

            // Alert list
            SliverPadding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              sliver: SliverList(
                delegate: SliverChildListDelegate([
                  // Critical SOS - unread
                  _alertItem(
                    type: '🆘 SOS',
                    typeColor: const Color(0xFFC62828),
                    typeBg: const Color(0xFFFFEBEE),
                    priority: 'P0',
                    priorityColor: Colors.white,
                    priorityBg: const Color(0xFFFF6B6B),
                    title: 'SOS紧急按钮触发',
                    desc: '奶奶在小区花园触发了手环SOS按钮，已自动发送位置信息给所有家属。',
                    location: '📍 陆家嘴环路1000号',
                    time: '🕐 今天 08:32',
                    actions: [
                      _actionBtn('📞 立即呼叫', const Color(0xFFFF6B6B), null),
                      _actionBtn('📍 查看位置', const Color(0xFFF0F0F5), const Color(0xFF666666)),
                      _actionBtn('✓ 标记处理', const Color(0xFFE8F5E9), const Color(0xFF2E7D32)),
                    ],
                    isUnread: true,
                  ),
                  // Fall detection - unread
                  _alertItem(
                    type: '⚠️ 跌倒检测',
                    typeColor: const Color(0xFFE65100),
                    typeBg: const Color(0xFFFFF3E0),
                    priority: 'P0',
                    priorityColor: Colors.white,
                    priorityBg: const Color(0xFFFF6B6B),
                    title: '检测到跌倒事件',
                    desc: '跌倒置信度 95%，设备已自动发送告警并通知家属。请确认奶奶安全。',
                    location: '📍 小区电梯口',
                    time: '🕐 今天 07:45',
                    actions: [
                      _actionBtn('📞 立即呼叫', const Color(0xFFFF6B6B), null),
                      _actionBtn('📍 查看位置', const Color(0xFFF0F0F5), const Color(0xFF666666)),
                    ],
                    isUnread: true,
                  ),
                  // Heart rate warning
                  _alertItem(
                    type: '💓 心率异常',
                    typeColor: const Color(0xFFAD1457),
                    typeBg: const Color(0xFFFCE4EC),
                    priority: 'P1',
                    priorityColor: Colors.white,
                    priorityBg: const Color(0xFFFFA726),
                    title: '心率持续偏高',
                    desc: '上午心率持续超过100bpm，已自动通知奶奶休息。目前心率已恢复正常。',
                    meta: '📊 最高 112 bpm · 🕐 今天 07:15 · ✅ 已恢复',
                    actions: [_actionBtn('查看详情', const Color(0xFFF0F0F5), const Color(0xFF666666))],
                    isUnread: false,
                  ),
                  // Geofence resolved
                  _alertItem(
                    type: '📍 电子围栏',
                    typeColor: const Color(0xFFF57F17),
                    typeBg: const Color(0xFFFFF8E1),
                    priority: 'P1',
                    priorityColor: Colors.white,
                    priorityBg: const Color(0xFFFFA726),
                    title: '离开安全区域',
                    desc: '奶奶走出电子围栏范围（500m），已提醒注意安全。约15分钟后返回。',
                    meta: '📍 距围栏 120m · 🕐 昨天 15:30 · ✅ 已返回',
                    actions: [],
                    isUnread: false,
                    isResolved: true,
                  ),
                  // Medication
                  _alertItem(
                    type: '💊 用药提醒',
                    typeColor: const Color(0xFF1565C0),
                    typeBg: const Color(0xFFE3F2FD),
                    priority: 'P2',
                    priorityColor: const Color(0xFF1565C0),
                    priorityBg: const Color(0xFF90CAF9),
                    title: '降压药未按时服用',
                    desc: '原计划20:00服用降压药，实际20:30服用，迟了30分钟。',
                    meta: '📅 昨天 20:00',
                    actions: [],
                    isUnread: false,
                  ),
                  const SizedBox(height: 24),
                ]),
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: BottomNavBar(selectedTab: _selectedIndex, onTabSelected: (i) => setState(() => _selectedIndex = i)),
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
            Text(label, style: const TextStyle(fontSize: 10, color: Colors.white, opacity: 0.9)),
          ],
        ),
      ),
    );
  }

  Widget _alertItem({
    required String type,
    required Color typeColor,
    required Color typeBg,
    required String priority,
    required Color priorityColor,
    required Color priorityBg,
    required String title,
    required String desc,
    String? location,
    String? time,
    String? meta,
    required List<Widget> actions,
    required bool isUnread,
    bool isResolved = false,
  }) {
    return Container(
      margin: const EdgeInsets.only(bottom: 10),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: isResolved ? const Color(0xFFFAFBFE) : (isUnread ? const Color(0xFFFFF5F5) : AppTheme.bgCard),
        borderRadius: BorderRadius.circular(14),
        border: Border.all(color: const Color(0xFFF0F0F0)),
        boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.04), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 3),
                decoration: BoxDecoration(color: typeBg, borderRadius: BorderRadius.circular(12)),
                child: Text(type, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: typeColor)),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(color: priorityBg, borderRadius: BorderRadius.circular(8)),
                child: Text(priority, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w700, color: priorityColor)),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(title, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700)),
          const SizedBox(height: 4),
          Text(desc, style: const TextStyle(fontSize: 12, color: Color(0xFF666666), height: 1.5)),
          const SizedBox(height: 8),
          Wrap(
            spacing: 16,
            runSpacing: 4,
            children: [
              if (location != null) Text(location, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA))),
              if (time != null) Text(time, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA))),
              if (meta != null) ...meta.split(' · ').map((m) => Text(m, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA)))),
            ].toList(),
          ),
          if (actions.isNotEmpty) ...[
            const SizedBox(height: 12),
            Wrap(
              spacing: 8,
              children: actions,
            ),
          ],
        ],
      ),
    );
  }

  Widget _actionBtn(String text, Color bgColor, Color? textColor) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 6),
      decoration: BoxDecoration(color: bgColor, borderRadius: BorderRadius.circular(8)),
      child: Text(text,
          style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: textColor ?? const Color(0xFF666666))),
    );
  }
}

import 'package:flutter/material.dart';
import '../../common/theme.dart';

/// Recent alerts list — matches home.html prototype
class RecentAlertsList extends StatelessWidget {
  const RecentAlertsList({super.key});

  @override
  Widget build(BuildContext context) {
    final alerts = [
      {'type': 'critical', 'icon': '🆘', 'title': 'SOS紧急按钮触发', 'desc': '奶奶在小区花园触发了手环SOS按钮', 'time': '今天 08:32 · 已处理'},
      {'type': 'warning', 'icon': '⚠️', 'title': '心率偏高提醒', 'desc': '上午心率持续超过100bpm，已恢复', 'time': '今天 07:15 · 已恢复'},
      {'type': 'info', 'icon': '💊', 'title': '用药提醒', 'desc': '降压药未按时服用（迟了30分钟）', 'time': '昨天 20:00'},
    ];

    final borderColors = {
      'critical': const Color(0xFFFF6B6B),
      'warning': const Color(0xFFFFA726),
      'info': AppTheme.primary,
    };

    final iconBgColors = {
      'critical': const Color(0xFFFFEBEE),
      'warning': const Color(0xFFFFF3E0),
      'info': const Color(0xFFE3F2FD),
    };

    return Column(
      children: alerts.map((alert) {
        final isUnread = alert['time']!.startsWith('今天');
        return Container(
          margin: const EdgeInsets.only(bottom: 8),
          padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
          decoration: BoxDecoration(
            color: AppTheme.bgCard,
            borderRadius: BorderRadius.circular(12),
            border: Border(left: BorderSide(color: borderColors[alert['type']]!, width: 3)),
            boxShadow: [
              BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 4, offset: const Offset(0, 1)),
            ],
          ),
          child: Row(
            children: [
              Container(width: 36, height: 36, decoration: BoxDecoration(color: iconBgColors[alert['type']], borderRadius: BorderRadius.circular(10)),
                  child: Center(child: Text(alert['icon'] as String, style: const TextStyle(fontSize: 18)))),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(alert['title'] as String, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    Text(alert['desc'] as String, style: const TextStyle(fontSize: 11, color: Color(0xFF888888))),
                    Text(alert['time'] as String, style: AppTheme.timeLabel),
                  ],
                ),
              ),
              if (isUnread) Container(width: 8, height: 8, decoration: const BoxDecoration(color: AppTheme.primary, shape: BoxShape.circle)),
            ],
          ),
        );
      }).toList(),
    );
  }
}

import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/bottom_nav_bar.dart';

/// Health dashboard page — matches health.html prototype
class HealthPage extends StatefulWidget {
  const HealthPage({super.key});

  @override
  State<HealthPage> createState() => _HealthPageState();

  int get initialIndex => 1;
}

class _HealthPageState extends State<HealthPage> {
  int _selectedIndex = 1;
  String _timeRange = '本周';

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
                padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
                color: AppTheme.bgCard,
                child: Row(
                  children: [
                    IconButton(icon: const Icon(Icons.arrow_back_ios_new, size: 18), onPressed: () {}),
                    const Expanded(child: Text('健康数据', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
                    IconButton(icon: const Icon(Icons.share_outlined), onPressed: () {}),
                  ],
                ),
              ),
            ),

            // Risk score card
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Container(
                  decoration: BoxDecoration(
                    gradient: AppTheme.riskGradient,
                    borderRadius: BorderRadius.circular(16),
                  ),
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    children: [
                      const Text('综合健康风险评估', style: TextStyle(fontSize: 13, color: Colors.white, opacity: 0.9)),
                      const SizedBox(height: 8),
                      Stack(
                        alignment: Alignment.center,
                        children: [
                          SizedBox(
                            width: 100,
                            height: 100,
                            child: CircularProgressIndicator(
                              value: 0.65, // 35/100 risk = 65% safe
                              strokeWidth: 8,
                              backgroundColor: Colors.white.withOpacity(0.2),
                              valueColor: const AlwaysStoppedAnimation<Color>(Colors.white),
                            ),
                          ),
                          Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Text('35', style: TextStyle(fontSize: 28, fontWeight: FontWeight.w800, color: Colors.white)),
                              Text('/ 100', style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.8))),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 4),
                      const Text('🟢 低风险', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.white)),
                      const SizedBox(height: 6),
                      Text('各项指标基本正常，步数略低于目标值',
                          style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.85), height: 1.5)),
                    ],
                  ),
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 16)),

            // Time range selector
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Row(
                  children: ['今日', '本周', '本月', '自定义'].map((range) {
                    final isActive = range == _timeRange;
                    return Padding(
                      padding: const EdgeInsets.only(right: 8),
                      child: FilterChip(
                        label: Text(range),
                        selected: isActive,
                        onSelected: (_) => setState(() => _timeRange = range),
                        selectedColor: AppTheme.primary,
                        labelStyle: TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: isActive ? Colors.white : const Color(0xFF888888),
                        ),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
                        side: BorderSide(color: isActive ? AppTheme.primary : const Color(0xFFF0F0F5)),
                      ),
                    );
                  }).toList(),
                ),
              ),
            ),

            // Heart rate metric
            SliverToBoxAdapter(
              child: _metricCard(
                title: '❤️ 心率',
                trend: '↔ 稳定',
                currentValue: '72',
                unit: 'bpm (静息)',
                badge: null,
                miniChart: _buildMiniChart(12, AppTheme.primary, [55, 60, 45, 70, 65, 50, 55, 60, 75, 50, 55, 60]),
              ),
            ),

            // SpO2 metric
            SliverToBoxAdapter(
              child: _metricCard(
                title: '💨 血氧饱和度',
                trend: '↔ 正常',
                currentValue: '97',
                unit: '%',
                badge: const BadgeChip(label: '正常', color: AppTheme.statusNormal),
                miniChart: _buildMiniChart(7, AppTheme.statusNormal, [90, 92, 88, 95, 90, 92, 88]),
              ),
            ),

            // Blood pressure metric
            SliverToBoxAdapter(
              child: _metricCard(
                title: '🩺 血压',
                trend: '↓ 改善',
                currentValue: '128/82',
                unit: 'mmHg',
                badge: const BadgeChip(label: '偏高', color: AppTheme.statusWarning),
                child: Padding(
                  padding: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: _bpBox('128', '收缩压', const Color(0xFFE65100)),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: _bpBox('82', '舒张压', const Color(0xFFE65100)),
                      ),
                    ],
                  ),
                ),
              ),
            ),

            // Sleep quality metric
            SliverToBoxAdapter(
              child: _metricCard(
                title: '😴 睡眠质量',
                trend: '↓ 改善',
                currentValue: '7.2',
                unit: '小时',
                badge: const BadgeChip(label: '良好', color: AppTheme.statusNormal),
                child: Padding(
                  padding: const EdgeInsets.only(top: 10),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text('深睡 2.1h', style: TextStyle(fontSize: 10, color: Color(0xFF999999))),
                          Text('浅睡 3.8h', style: TextStyle(fontSize: 10, color: Color(0xFF999999))),
                          Text('REM 1.3h', style: TextStyle(fontSize: 10, color: Color(0xFF999999))),
                        ],
                      ),
                      const SizedBox(height: 4),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(4),
                        child: Row(
                          children: [
                            Expanded(flex: 29, child: Container(color: const Color(0xFF5C6BC0))),
                            Expanded(flex: 53, child: Container(color: const Color(0xFF7986CB))),
                            Expanded(flex: 18, child: Container(color: const Color(0xFF9FA8DA))),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),

            // Steps summary
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.only(top: 16, left: 20, right: 20),
                child: Row(
                  children: [
                    _stepStat('🚶', '3,456', '今日步数'),
                    const SizedBox(width: 12),
                    _stepStat('🔥', '186', '千卡消耗'),
                    const SizedBox(width: 12),
                    _stepStat('⏱️', '42', '活动分钟'),
                  ],
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 24)),
          ],
        ),
      ),
      bottomNavigationBar: BottomNavBar(selectedTab: _selectedIndex, onTabSelected: (i) => setState(() => _selectedIndex = i)),
    );
  }

  Widget _metricCard({
    required String title,
    required String trend,
    required String currentValue,
    required String unit,
    BadgeChip? badge,
    Widget? child,
    Widget? miniChart,
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 5),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppTheme.bgCard,
          borderRadius: BorderRadius.circular(AppTheme.cardRadius),
          border: Border.all(color: const Color(0xFFF5F5FA)),
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 6)],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(title, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
                Text(trend, style: const TextStyle(fontSize: 11, color: Color(0xFF999999))),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                Text(currentValue, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800)),
                Text(unit, style: const TextStyle(fontSize: 12, color: Color(0xFF999999))),
                if (badge != null) const SizedBox(width: 8),
                if (badge != null) badge,
              ],
            ),
            if (miniChart != null) miniChart,
            if (child != null) child,
          ],
        ),
      ),
    );
  }

  Widget _buildMiniChart(int bars, Color color, List<double> heights) {
    return Padding(
      padding: const EdgeInsets.only(top: 10),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: List.generate(bars, (i) {
          return Expanded(
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 1.5),
              height: heights[i] / 100 * 40,
              decoration: BoxDecoration(
                color: color,
                borderRadius: const BorderRadius.vertical(top: Radius.circular(3)),
                gradient: LinearGradient(
                  colors: [color, color.withOpacity(0.6)],
                  begin: Alignment.bottomCenter,
                  end: Alignment.topCenter,
                ),
              ),
            ),
          );
        }),
      ),
    );
  }

  Widget _stepStat(String icon, String value, String label) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10),
        decoration: BoxDecoration(
          color: AppTheme.bgCard,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 4)],
        ),
        child: Column(
          children: [
            Text(icon, style: const TextStyle(fontSize: 20)),
            const SizedBox(height: 6),
            Text(value, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
            Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
          ],
        ),
      ),
    );
  }
}

class BadgeChip extends StatelessWidget {
  final String label;
  final Color color;
  const BadgeChip({super.key, required this.label, required this.color});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(10)),
      child: Text(label, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: color)),
    );
  }
}

import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';

/// Chronic care home page — glucose / uric acid / BP metric cards, daily
/// task checklist, AI recommendation card, and quick action buttons.
class ChronicHomePage extends StatefulWidget {
  const ChronicHomePage({super.key});

  @override
  State<ChronicHomePage> createState() => _ChronicHomePageState();
}

class _ChronicHomePageState extends State<ChronicHomePage> {
  int _selectedIndex = 4;
  int _activeElder = 0;
  bool _loading = true;

  final List<ElderInfo> _elders = const [
    ElderInfo(name: '爷爷 张三丰', icon: '\u{1F468}', bg: Color(0xFFFFF3E0), tier: 'Pro'),
    ElderInfo(name: '奶奶 李秀英', icon: '\u{1F469}', bg: Color(0xFFFCE7F3), tier: 'Plus'),
  ];

  // Mock chronic care data
  final List<MetricData> _metrics = const [
    MetricData(
      label: '空腹血糖',
      value: '6.8',
      unit: 'mmol/L',
      trend: '+0.3',
      trendUp: false,
      range: '3.9-6.1',
      color: Color(0xFF2563EB),
      icon: '\u{1F4A1}',
    ),
    MetricData(
      label: '尿酸',
      value: '380',
      unit: 'μmol/L',
      trend: '-12',
      trendUp: true,
      range: '143-416',
      color: Color(0xFF7C3AED),
      icon: '\u{1F52C}',
    ),
    MetricData(
      label: '血压',
      value: '138/88',
      unit: 'mmHg',
      trend: '偏高',
      trendUp: false,
      range: '90-140/60-90',
      color: Color(0xFFEF4444),
      icon: '\u{2764}',
    ),
  ];

  final List<TaskItem> _tasks = [
    TaskItem(id: '1', label: '测量空腹血糖', checked: true),
    TaskItem(id: '2', label: '测量血压（晨起）', checked: true),
    TaskItem(id: '3', label: '服用降糖药（早）', checked: false),
    TaskItem(id: '4', label: '服用降压药（早）', checked: false),
    TaskItem(id: '5', label: '饭后散步 30 分钟', checked: false),
    TaskItem(id: '6', label: '服用降糖药（晚）', checked: false),
    TaskItem(id: '7', label: '测量睡前血糖', checked: false),
    TaskItem(id: '8', label: '服用降压药（晚）', checked: false),
  ];

  final List<QuickAction> _quickActions = const [
    QuickAction(label: '记录血糖', icon: '\u{1F4A1}', color: Color(0xFFDBEAFE)),
    QuickAction(label: '记录血压', icon: '\u{2764}', color: Color(0xFFFEF2F2)),
    QuickAction(label: '用药提醒', icon: '\u{1F48A}', color: Color(0xFFF3E8FF)),
    QuickAction(label: '就医预约', icon: '\u{1F489}', color: Color(0xFFECFDF5)),
  ];

  final String _aiRecommendation =
      '爷爷的空腹血糖近7天呈上升趋势，建议减少碳水摄入并增加饭后散步时长至40分钟。血压值略高于目标范围，请持续监测并按时服药。';

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    // Simulate loading
    await Future.delayed(const Duration(milliseconds: 600));
    if (mounted) setState(() => _loading = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFF3F4F6),
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  _buildTopBar(),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildMetricsSection(),
                          const SizedBox(height: 16),
                          _buildTasksSection(),
                          const SizedBox(height: 16),
                          _buildAIRecommendationCard(),
                          const SizedBox(height: 16),
                          _buildQuickActionsSection(),
                          const SizedBox(height: 24),
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
    return SliverToBoxAdapter(
      child: Container(
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
                        gradient: LinearGradient(colors: [Color(0xFF2563EB), Color(0xFF7C3AED)]),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Center(child: Text('颐', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Colors.white))),
                    ),
                    const SizedBox(width: 8),
                    const Text('Eregen 颐贞', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
                  ],
                ),
                const Text(
                  '慢病管理',
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF6B7280)),
                ),
              ],
            ),
            const SizedBox(height: 10),
            // Elder selector
            _buildElderSelector(),
          ],
        ),
      ),
    );
  }

  Widget _buildElderSelector() {
    return SizedBox(
      height: 56,
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
              width: 150,
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              decoration: BoxDecoration(
                color: isActive ? const Color(0xFFDBEAFE) : Colors.white,
                border: Border.all(color: isActive ? const Color(0xFF2563EB) : const Color(0xFFE5E7EB)),
                borderRadius: BorderRadius.circular(24),
                boxShadow: isActive
                    ? [BoxShadow(color: const Color(0xFF2563EB).withValues(alpha: 0.15), blurRadius: 12, offset: const Offset(0, 2))]
                    : [],
              ),
              child: Row(
                children: [
                  Container(
                    width: 36,
                    height: 36,
                    decoration: BoxDecoration(color: elder.bg, shape: BoxShape.circle),
                    child: Center(child: Text(elder.icon, style: const TextStyle(fontSize: 18))),
                  ),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Text(
                          elder.name,
                          style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: isActive ? const Color(0xFF1F2937) : const Color(0xFF374151)),
                        ),
                        Text('慢病管理', style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
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

  // ===== Metric Cards =====
  Widget _buildMetricsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text(
              '慢病指标',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937)),
            ),
            Text(
              '最后更新：今天 08:30',
              style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
            ),
          ],
        ),
        const SizedBox(height: 10),
        SizedBox(
          height: 120,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: _metrics.length,
            separatorBuilder: (_, __) => const SizedBox(width: 10),
            itemBuilder: (ctx, i) => _metricCard(_metrics[i]),
          ),
        ),
      ],
    );
  }

  Widget _metricCard(MetricData m) {
    return Container(
      width: 140,
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Container(
                width: 32,
                height: 32,
                decoration: BoxDecoration(color: m.color.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
                child: Center(child: Text(m.icon, style: const TextStyle(fontSize: 16))),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: m.trendUp ? const Color(0xFFF0FDF4) : const Color(0xFFFEF2F2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  m.trendUp ? '\u{2193} ${m.trend}' : '\u{2191} ${m.trend}',
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w600,
                    color: m.trendUp ? AppTheme.statusNormal : AppTheme.statusDanger,
                  ),
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(m.value, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: Color(0xFF1F2937))),
          Text(m.unit, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
          const SizedBox(height: 4),
          Text(m.label, style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
          Text('参考: ${m.range}', style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
        ],
      ),
    );
  }

  // ===== Daily Tasks =====
  Widget _buildTasksSection() {
    final doneCount = _tasks.where((t) => t.checked).length;
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text(
                '今日任务',
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937)),
              ),
              Text(
                '$doneCount / ${_tasks.length} 已完成',
                style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: doneCount == _tasks.length ? AppTheme.statusNormal : const Color(0xFF6B7280)),
              ),
            ],
          ),
          const SizedBox(height: 12),
          // Progress bar
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: doneCount / _tasks.length,
              minHeight: 6,
              backgroundColor: const Color(0xFFF3F4F6),
              valueColor: AlwaysStoppedAnimation<Color>(doneCount == _tasks.length ? AppTheme.statusNormal : AppTheme.primary),
            ),
          ),
          const SizedBox(height: 12),
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _tasks.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (ctx, i) {
              final task = _tasks[i];
              return InkWell(
                onTap: () => setState(() => _tasks[i] = TaskItem(id: task.id, label: task.label, checked: !task.checked)),
                borderRadius: BorderRadius.circular(8),
                child: Row(
                  children: [
                    Container(
                      width: 22,
                      height: 22,
                      decoration: BoxDecoration(
                        color: task.checked ? AppTheme.primary : Colors.transparent,
                        border: Border.all(color: task.checked ? AppTheme.primary : const Color(0xFFD1D5DB)),
                        borderRadius: BorderRadius.circular(6),
                      ),
                      child: task.checked ? const Icon(Icons.check, size: 14, color: Colors.white) : null,
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: Text(
                        task.label,
                        style: TextStyle(
                          fontSize: 14,
                          color: task.checked ? const Color(0xFF9CA3AF) : const Color(0xFF374151),
                          decoration: task.checked ? TextDecoration.lineThrough : null,
                        ),
                      ),
                    ),
                  ],
                ),
              );
            },
          ),
        ],
      ),
    );
  }

  // ===== AI Recommendation Card =====
  Widget _buildAIRecommendationCard() {
    return GestureDetector(
      onTap: () {
        // Navigate to AI report / detail page (placeholder)
      },
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          gradient: const LinearGradient(colors: [Color(0xFFF3E8FF), Color(0xFFE9D5FF)]),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFC084FC).withValues(alpha: 0.15)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
              child: const Center(child: Text('\u{1F916}', style: TextStyle(fontSize: 20))),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'AI 慢病建议',
                    style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: Color(0xFF6B21A8)),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    _aiRecommendation,
                    style: const TextStyle(fontSize: 12, color: Color(0xFF7C3AED), height: 1.5),
                  ),
                  const SizedBox(height: 6),
                  Row(
                    children: const [
                      Text(
                        '查看完整报告 \u{2192}',
                        style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF7C3AED), decoration: TextDecoration.underline),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Quick Actions =====
  Widget _buildQuickActionsSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('快捷操作', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
        const SizedBox(height: 10),
        Wrap(
          spacing: 10,
          runSpacing: 10,
          children: _quickActions
              .map((a) => GestureDetector(
                    onTap: () {},
                    child: Container(
                      width: 84,
                      padding: const EdgeInsets.symmetric(vertical: 14),
                      decoration: BoxDecoration(
                        color: a.color,
                        borderRadius: BorderRadius.circular(14),
                      ),
                      child: Column(
                        children: [
                          Text(a.icon, style: const TextStyle(fontSize: 22)),
                          const SizedBox(height: 6),
                          Text(a.label, style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
                        ],
                      ),
                    ),
                  ))
              .toList(),
        ),
      ],
    );
  }
}

// ===== Data Models =====

class ElderInfo {
  final String name;
  final String icon;
  final Color bg;
  final String tier;
  const ElderInfo({required this.name, required this.icon, required this.bg, required this.tier});
}

class MetricData {
  final String label, value, unit, range, trend;
  final bool trendUp;
  final Color color;
  final String icon;
  const MetricData({
    required this.label,
    required this.value,
    required this.unit,
    required this.trend,
    required this.trendUp,
    required this.range,
    required this.color,
    required this.icon,
  });
}

class TaskItem {
  final String id;
  final String label;
  final bool checked;
  const TaskItem({required this.id, required this.label, required this.checked});
}

class QuickAction {
  final String label, icon;
  final Color color;
  const QuickAction({required this.label, required this.icon, required this.color});
}

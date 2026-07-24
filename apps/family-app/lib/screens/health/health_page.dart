import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../ai/ai_report_page.dart';

/// Health dashboard — v2 design: date range pills, AI insight banner, gradient
/// health-score card with SVG ring, summary cards (SpO2 + steps), bar chart for
/// daily steps, anomaly alerts, intergenerational comparison toggle, dark mode.
class HealthPage extends StatefulWidget {
  const HealthPage({super.key});

  @override
  State<HealthPage> createState() => _HealthPageState();
}

class _HealthPageState extends State<HealthPage> {
  int _selectedIndex = 1;
  String _activeDateRange = '近7天';
  bool _showCompare = false;
  bool _darkMode = false;

  // Mock data matching prototype
  final healthScore = 80;
  final healthScoreLabel = '优秀';
  final scoreTrend = '+3分';
  final scoreDetails = '心率正常 · 血氧良好\n步数偏低 · 睡眠充足';
  final avgSpo2 = 97;
  final dailySteps = 3200;
  final stepsTrend = '-12%';
  final deviceBattery = 85;
  final todayDistance = 2.1;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _darkMode ? const Color(0xFF111827) : const Color(0xFFF3F4F6),
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            _buildTopBar(),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildDateRangeSelector(),
                    const SizedBox(height: 12),
                    _buildAIInsightBanner(),
                    const SizedBox(height: 12),
                    _buildHealthScoreCard(),
                    const SizedBox(height: 12),
                    _buildHeartRateChart(),
                    const SizedBox(height: 12),
                    _buildSummaryRow1(),
                    _buildStepsChart(),
                    const SizedBox(height: 12),
                    _buildSummaryRow2(),
                    const SizedBox(height: 16),
                    _buildAnomalyAlerts(),
                    const SizedBox(height: 12),
                    _buildIntergenerationalComparison(),
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
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
        boxShadows: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4)],
        child: Column(
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
                  child: Center(child: Text('健康数据看板', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
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
                      onTap: () {},
                      child: Container(width: 36, height: 36, decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(18)),
                        child: const Icon(Icons.swap_horiz, size: 18),
                      ),
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
                  final active = i == 0;
                  return Container(
                    width: 140,
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: active ? const Color(0xFFDBEAFE) : (_darkMode ? const Color(0xFF1F2937) : Colors.white),
                      border: Border.all(color: active ? AppTheme.primary : const Color(0xFFE5E7EB)),
                      borderRadius: BorderRadius.circular(24),
                      boxShadow: active ? [BoxShadow(color: AppTheme.primary.withValues(alpha: 0.15), blurRadius: 12)] : [],
                    ),
                    child: Row(
                      children: [
                        Container(width: 36, height: 36, decoration: BoxDecoration(color: bgs[i], shape: BoxShape.circle),
                          child: Center(child: Text(icons[i], style: const TextStyle(fontSize: 18))),
                        ),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            mainAxisAlignment: MainAxisAlignment.center,
                            children: [
                              Text(names[i], style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: active ? AppTheme.primary : (_darkMode ? Colors.white : const Color(0xFF1F2937)))),
                              Text('今日数据完整', style: TextStyle(fontSize: 10, color: const Color(0xFF6B7280))),
                            ],
                          ),
                        ),
                      ],
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Date Range Selector =====
  Widget _buildDateRangeSelector() {
    final ranges = ['今日', '近7天', '近30天', '自定义'];
    return Container(
      decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(12)),
      padding: const EdgeInsets.all(4),
      child: Row(
        children: ranges.map((r) {
          final isActive = r == _activeDateRange;
          return Expanded(
            child: GestureDetector(
              onTap: () => setState(() => _activeDateRange = r),
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 8),
                decoration: BoxDecoration(
                  color: isActive ? (_darkMode ? const Color(0xFF374151) : Colors.white) : null,
                  borderRadius: BorderRadius.circular(10),
                  boxShadow: isActive ? [BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 8)] : [],
                ),
                child: Text(r, textAlign: TextAlign.center, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: isActive ? (_darkMode ? const Color(0xFF60A5FA) : AppTheme.primary) : const Color(0xFF6B7280))),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  // ===== AI Insight Banner =====
  Widget _buildAIInsightBanner() {
    return GestureDetector(
      onTap: () {
        Navigator.of(context).push(
          MaterialPageRoute(builder: (_) => const AIReportPage()),
        );
      },
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          gradient: const LinearGradient(colors: [Color(0xFFF3E8FF), Color(0xFFE9D5FF)]),
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: const Color(0xFFC084FC).withValues(alpha: 0.1)),
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text('\u{1F916}', style: TextStyle(fontSize: 20)),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('AI 健康建议', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: Color(0xFF6B21A8))),
                  const SizedBox(height: 4),
                  const Text(
                    '基于近7天数据，爷爷的心率变异性(HRV)呈上升趋势，表明自主神经调节能力改善。建议继续保持当前运动量。',
                    style: TextStyle(fontSize: 12, color: Color(0xFF7C3AED), height: 1.5),
                  ),
                  const SizedBox(height: 4),
                  Text('查看详细报告 →', style: const TextStyle(fontSize: 12, color: Color(0xFF7C3AED), fontWeight: FontWeight.w600, decoration: TextDecoration.underline)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Health Score Card =====
  Widget _buildHealthScoreCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Color(0xFF2563EB), Color(0xFF7C3AED)]),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: const Color(0xFF2563EB).withValues(alpha: 0.3), blurRadius: 16)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('综合健康评分', style: TextStyle(fontSize: 14, color: Color(0xFFFFFFFF), opacity: 0.9)),
              Text('本周平均', style: TextStyle(fontSize: 12, color: Color(0xFFFFFFFF), fontWeight: FontWeight.w600)),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              // SVG-like ring using CircularProgressIndicator
              SizedBox(
                width: 100,
                height: 100,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    CircularProgressIndicator(
                      value: 1,
                      strokeWidth: 8,
                      backgroundColor: const Color(0xFFFFFFFF).withValues(alpha: 0.2),
                    ),
                    CircularProgressIndicator(
                      value: healthScore / 100,
                      strokeWidth: 8,
                      backgroundColor: Colors.transparent,
                      valueColor: const AlwaysStoppedAnimation<Color>(Colors.white),
                      transform: Matrix4.rotationZ(0),
                    ),
                    Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('$healthScore', style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: Colors.white)),
                        Text(healthScoreLabel, style: const TextStyle(fontSize: 10, color: Color(0xFFFFFFFF), opacity: 0.7)),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 24),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    RichText(
                      text: TextSpan(
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Colors.white),
                        children: [
                          const TextSpan(text: '82分 '),
                          TextSpan(text: scoreTrend, style: const TextStyle(opacity: 0.9)),
                        ],
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(scoreDetails, style: const TextStyle(fontSize: 12, color: Color(0xFFFFFFFF), opacity: 0.9, height: 1.6)),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Wrap(
            spacing: 6,
            runSpacing: 6,
            children: [
              _scorePill(true, '心率正常'),
              _scorePill(true, '血氧良好'),
              _scorePill(false, '步数偏低'),
              _scorePill(true, '睡眠充足'),
            ],
          ),
        ],
      ),
    );
  }

  Widget _scorePill(bool good, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: good ? const Color(0xFF16A34A).withValues(alpha: 0.2) : const Color(0xFFF59E0B).withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(12),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(good ? '\u{2713}' : '\u{26A0}', style: TextStyle(fontSize: 10, color: good ? const Color(0xFFFFFFFF) : const Color(0xFFFDE68A))),
          const SizedBox(width: 4),
          Text(label, style: const TextStyle(fontSize: 11, color: Colors.white)),
        ],
      ),
    );
  }

  // ===== Heart Rate Chart =====
  Widget _buildHeartRateChart() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('\u{2764} 心率趋势', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
              Text('单位: bpm', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
            ],
          ),
          const SizedBox(height: 12),
          Row(
            children: [
              _chartLegend(const Color(0xFFEF4444), '静息心率'),
              const SizedBox(width: 12),
              _chartLegend(const Color(0xFFF59E0B), '活动心率'),
              const SizedBox(width: 12),
              _chartLegend(const Color(0xFFD1D5DB), '正常范围'),
            ],
          ),
          const SizedBox(height: 12),
          SizedBox(
            height: 160,
            child: CustomPaint(painter: _HeartRateChartPainter(darkMode: _darkMode), size: const Size(double.infinity, 160)),
          ),
        ],
      ),
    );
  }

  Widget _chartLegend(Color color, String label) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 8, height: 8, decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF4B5563))),
      ],
    );
  }

  // ===== Summary Cards Row 1 =====
  Widget _buildSummaryRow1() {
    return Row(
      children: [
        _summaryCard('\u{1F9C7} 血氧平均值', '${avgSpo2}%', '正常范围', true, suffix: '%'),
        const SizedBox(width: 10),
        _summaryCard('\u{1F6B6} 日均步数', '3.2k', '较上周 $stepsTrend', false, trendDown: true),
      ],
    );
  }

  // ===== Steps Bar Chart =====
  Widget _buildStepsChart() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('\u{1F6B6} 每日步数', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
              Text('目标: 5,000步', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
            ],
          ),
          const SizedBox(height: 12),
          SizedBox(
            height: 160,
            child: CustomPaint(painter: _StepsBarChartPainter(darkMode: _darkMode), size: const Size(double.infinity, 160)),
          ),
        ],
      ),
    );
  }

  // ===== Summary Cards Row 2 =====
  Widget _buildSummaryRow2() {
    return Row(
      children: [
        _summaryCard('\u{1F50B} 今日电量', '$deviceBattery%', '预计3天', neutral: true, suffix: '%'),
        const SizedBox(width: 10),
        _summaryCard('\u{1F3DA} 今日距离', '$todayDistance km', '较昨日 +0.3km', true),
      ],
    );
  }

  Widget _summaryCard(String label, String value, String trend, bool isGood, {bool trendDown = false, bool neutral = false, String? suffix}) {
    final trendColor = trendDown ? AppTheme.statusDanger : (isGood ? AppTheme.statusNormal : (neutral ? const Color(0xFF6B7280) : AppTheme.statusNormal));
    final trendIcon = trendDown ? '\u{2193}' : (trend.startsWith('+') ? '\u{2191}' : '\u{1F50B}');
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16, offset: const Offset(0, 2))],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(label, style: TextStyle(fontSize: 12, color: _darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280))),
            const SizedBox(height: 6),
            RichText(
              text: TextSpan(
                style: TextStyle(fontSize: 28, fontWeight: FontWeight.w700, color: _darkMode ? Colors.white : const Color(0xFF1F2937)),
                children: [
                  TextSpan(text: value.split(suffix ?? '').first),
                  if (suffix != null) TextSpan(text: suffix, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.normal, color: Color(0xFF6B7280))),
                ],
              ),
            ),
            const SizedBox(height: 6),
            Row(
              children: [
                Text(trendIcon, style: TextStyle(fontSize: 11, color: trendColor, fontWeight: FontWeight.w600)),
                const SizedBox(width: 2),
                Text(trend, style: TextStyle(fontSize: 11, color: trendColor, fontWeight: FontWeight.w600)),
              ],
            ),
          ],
        ),
      ),
    );
  }

  // ===== Anomaly Alerts =====
  Widget _buildAnomalyAlerts() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: const [
            Text('异常提醒', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
            SizedBox(width: 6),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
              decoration: BoxDecoration(color: AppTheme.statusDanger, borderRadius: BorderRadius.circular(10)),
              child: Text('3', style: TextStyle(fontSize: 10, color: Colors.white, fontWeight: FontWeight.w700)),
            ),
          ],
        ),
        const SizedBox(height: 10),
        _anomalyItem(
          severity: 'critical',
          icon: '\u{26A0}',
          title: '心率异常偏高',
          desc: '周四 14:30 检测到活动心率 142bpm，持续15分钟超过安全阈值',
          time: '4天前',
          actionLabel: '查看',
        ),
        _anomalyItem(
          severity: 'warning',
          icon: '\u{1F989}',
          title: '步数连续3日偏低',
          desc: '近3日平均步数低于个人基线40%，建议关注活动量',
          time: '持续中',
          actionLabel: '详情',
        ),
        _anomalyItem(
          severity: 'info',
          icon: '\u{1F4CD}',
          title: '血氧轻微下降',
          desc: '周日血氧均值97%，较周均下降1个百分点',
          time: '1天前',
          actionLabel: '了解',
        ),
      ],
    );
  }

  Widget _anomalyItem({required String severity, required String icon, required String title, required String desc, required String time, required String actionLabel}) {
    final bg = _darkMode ? const Color(0xFF1F2937) : Colors.white;
    final iconBgMap = {
      'critical': const Color(0xFFFEF2F2),
      'warning': const Color(0xFFFFFBEB),
      'info': const Color(0xFFDBEAFE),
    };
    final borderColorMap = {
      'critical': AppTheme.statusDanger,
      'warning': AppTheme.statusWarning,
      'info': AppTheme.primary,
    };
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(14),
        border: Border(left: BorderSide(color: borderColorMap[severity]!, width: 4)),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4)],
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 36, height: 36, decoration: BoxDecoration(color: iconBgMap[severity], borderRadius: BorderRadius.circular(18)),
            child: Center(child: Text(icon, style: const TextStyle(fontSize: 18))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
                const SizedBox(height: 2),
                Text(desc, style: TextStyle(fontSize: 12, color: _darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280), height: 1.4)),
                Text(time, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
              ],
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
            decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(8)),
            child: Text(actionLabel, style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: Color(0xFF4B5563))),
          ),
        ],
      ),
    );
  }

  // ===== Intergenerational Comparison =====
  Widget _buildIntergenerationalComparison() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('\u{1F4CA} 代际对比', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
              GestureDetector(
                onTap: () => setState(() => _showCompare = !_showCompare),
                child: Text(_showCompare ? '收起 ^' : '点击展开同年龄段对比数据 →', style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
              ),
            ],
          ),
          if (_showCompare) ...[
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(border: Border.all(color: const Color(0xFFE5E7EB)), borderRadius: BorderRadius.circular(12)),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: const [
                        Text('爷爷 张三丰', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
                        SizedBox(height: 4),
                        Text('3.2k', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700)),
                        Text('步/日', style: TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: Container(
                    padding: const EdgeInsets.all(14),
                    decoration: BoxDecoration(border: Border.all(color: const Color(0xFFE5E7EB)), borderRadius: BorderRadius.circular(12)),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: const [
                        Text('同龄人平均', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
                        SizedBox(height: 4),
                        Text('5.1k', style: TextStyle(fontSize: 20, fontWeight: FontWeight.w700, color: Color(0xFF9CA3AF))),
                        Text('步/日', style: TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
                      ],
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
}

// ===== Custom Painters for Charts =====

class _HeartRateChartPainter extends CustomPainter {
  final bool darkMode;
  _HeartRateChartPainter({required this.darkMode});

  @override
  void paint(Canvas canvas, Size size) {
    final paintGrid = Paint()
      ..color = darkMode ? const Color(0xFF374151) : const Color(0xFFE5E7EB)
      ..strokeWidth = 0.5
      ..strokeStyle = PaintingStyle.stroke;
    final dashPaint = Paint()
      ..color = darkMode ? const Color(0xFF374151) : const Color(0xFFE5E7EB)
      ..strokeWidth = 1
      ..strokeDashArray = const [6, 3]
      ..strokeStyle = PaintingStyle.stroke;

    // Grid lines
    for (double y = 20; y < size.height - 30; y += 35) {
      canvas.drawLine(Offset(40, y), Offset(size.width - 10, y), paintGrid);
    }

    // Normal range band
    final normalBand = Paint()..color = const Color(0xFF16A34A).withValues(alpha: 0.06)..style = PaintingStyle.fill;
    canvas.drawRect(Rect.fromLTWH(40, size.height * 0.35, size.width - 50, size.height * 0.22), normalBand);

    // Target line
    canvas.drawLine(Offset(40, size.height * 0.3), Offset(size.width - 10, size.height * 0.3), dashPaint);

    // Resting HR line (red)
    final restPaint = Paint()
      ..color = const Color(0xFFEF4444)
      ..strokeWidth = 2.5
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;
    final activityPaint = Paint()
      ..color = const Color(0xFFF59E0B)
      ..strokeWidth = 2.5
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;

    final w = size.width;
    final h = size.height;
    final restingPoints = [
      Offset(40, h * 0.55),
      Offset(w * 0.18, h * 0.50),
      Offset(w * 0.36, h * 0.53),
      Offset(w * 0.54, h * 0.45),
      Offset(w * 0.72, h * 0.48),
      Offset(w * 0.86, h * 0.42),
      Offset(w - 12, h * 0.46),
    ];
    final activityPoints = [
      Offset(40, h * 0.30),
      Offset(w * 0.18, h * 0.33),
      Offset(w * 0.36, h * 0.28),
      Offset(w * 0.54, h * 0.24),
      Offset(w * 0.72, h * 0.27),
      Offset(w * 0.86, h * 0.20),
      Offset(w - 12, h * 0.23),
    ];

    // Area fills
    final areaRest = Paint()..color = const Color(0xFFEF4444).withValues(alpha: 0.1)..style = PaintingStyle.fill;
    final areaActivity = Paint()..color = const Color(0xFFF59E0B).withValues(alpha: 0.1)..style = PaintingStyle.fill;

    Path restPath = Path()..moveTo(40, h * 0.65);
    for (final p in restingPoints) restPath.lineTo(p.dx, p.dy);
    restPath.lineTo(restingPoints.last.dx, h * 0.65);
    restPath.close();
    canvas.drawPath(restPath, areaRest);
    canvas.drawPath(Path()..addPolygon(restingPoints, false), restPaint);

    Path actPath = Path()..moveTo(40, h * 0.65);
    for (final p in activityPoints) actPath.lineTo(p.dx, p.dy);
    actPath.lineTo(activityPoints.last.dx, h * 0.65);
    actPath.close();
    canvas.drawPath(actPath, areaActivity);
    canvas.drawPath(Path()..addPolygon(activityPoints, false), activityPaint);

    // Data dots
    final dotPaint = Paint()..style = PaintingStyle.fill;
    final strokeDot = Paint()..style = PaintingStyle.stroke..strokeWidth = 2..color = Colors.white;
    for (int i = 2; i < restingPoints.length; i += 2) {
      canvas.drawCircle(restingPoints[i], 4, dotPaint..color = const Color(0xFFEF4444));
      canvas.drawCircle(restingPoints[i], 4, strokeDot);
    }

    // Y-axis labels
    final labelStyle = TextStyle(fontSize: 10, color: darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280));
    final yLabels = ['100', '80', '60', '40'];
    for (int i = 0; i < yLabels.length; i++) {
      final y = 20 + i * 35;
      TextSpan(yLabels[i], labelStyle).toTextPainter()
        ..layout(minWidth: 0, maxWidth: 40)
        ..paint(canvas, Offset(2, y - 4));
    }

    // X-axis labels
    final days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    for (int i = 0; i < days.length; i++) {
      final x = 40 + i * ((w - 52) / 6);
      TextPainter(
        text: TextSpan(text: days[i], style: labelStyle),
        textDirection: TextDirection.ltr,
      )..layout()
        ..paint(canvas, Offset(x - 15, h - 14));
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _StepsBarChartPainter extends CustomPainter {
  final bool darkMode;
  _StepsBarChartPainter({required this.darkMode});

  @override
  void paint(Canvas canvas, Size size) {
    final w = size.width;
    final h = size.height;
    final days = ['周一', '周二', '周三', '周四', '周五', '周六', '周日'];
    final values = [3.8, 4.2, 3.1, 4.8, 3.5, 5.2, 2.8];
    final target = 5.0;
    final maxVal = 6.0;
    final barWidth = (w - 80) / days.length;
    const labelStyle = TextStyle(fontSize: 10, color: Color(0xFF6B7280));

    // Target line
    final targetY = h - 30 - (target / maxVal) * (h - 50);
    canvas.drawLine(
      Offset(40, targetY),
      Offset(w - 10, targetY),
      Paint()..color = const Color(0xFFEF4444)..strokeWidth = 1..strokeDashArray = const [6, 3],
    );
    TextPainter(
      text: TextSpan(text: '目标 5k', style: const TextStyle(fontSize: 9, color: Color(0xFFEF4444), fontWeight: FontWeight.w600)),
      textDirection: TextDirection.ltr,
    )..layout(minWidth: 0, maxWidth: 300)
      ..paint(canvas, Offset(w - 50, targetY - 12));

    // Bars
    for (int i = 0; i < days.length; i++) {
      final barH = (values[i] / maxVal) * (h - 50);
      final x = 40 + i * barWidth + barWidth * 0.15;
      final barW = barWidth * 0.7;
      final y = h - 30 - barH;

      canvas.drawRRect(
        RRect.fromRectAndRadius(Rect.fromLTWH(x, y, barW, barH), const Radius.circular(4)),
        Paint()
          ..color = values[i] >= target ? const Color(0xFF2563EB) : const Color(0xFFDBEAFE),
      );

      // Value label
      TextPainter(
        text: TextSpan(
          text: '${values[i].toStringAsFixed(1)}k',
          style: TextStyle(
            fontSize: 9,
            fontWeight: values[i] >= target ? FontWeight.w700 : FontWeight.w600,
            color: values[i] >= target ? const Color(0xFF2563EB) : const Color(0xFF6B7280),
          ),
        ),
        textDirection: TextDirection.ltr,
      )..layout()
        ..paint(canvas, Offset(x + barW / 2 - 12, y - 14));

      // Day label
      TextPainter(
        text: TextSpan(text: days[i], style: TextStyle(fontSize: 10, color: values[i] >= target ? const Color(0xFF2563EB) : const Color(0xFF6B7280), fontWeight: values[i] >= target ? FontWeight.w700 : FontWeight.normal)),
        textDirection: TextDirection.ltr,
      )..layout()
        ..paint(canvas, Offset(x + barW / 2 - 10, h - 14));
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

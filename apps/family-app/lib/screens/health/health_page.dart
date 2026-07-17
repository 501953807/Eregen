import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../../models/health.dart';

/// Health dashboard page — fetches live data from GET /health/records and risk score from GET /health/risk-score.
class HealthPage extends StatefulWidget {
  const HealthPage({super.key});

  @override
  State<HealthPage> createState() => _HealthPageState();
}

class _HealthPageState extends State<HealthPage> {
  int _selectedIndex = 1;
  String _timeRange = '本周';
  bool _loading = true;
  List<HealthRecord> _records = [];
  double _riskScore = 0;
  String _riskLevel = '加载中...';
  Color _riskColor = Colors.white;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      // Fetch health records
      final healthResp = await ApiClient.instance.get('/health/records', query: {'range': _timeRange});
      final list = (healthResp.data as List);
      final records = list.map((r) => HealthRecord.fromJson(r as Map<String, dynamic>)).toList();

      // Fetch risk score from API
      double riskScore = 0;
      String riskLevel = '暂无数据';
      Color riskColor = Colors.white;
      try {
        final riskResp = await ApiClient.instance.get('/health/risk-score');
        if (riskResp.data != null) {
          final riskData = riskResp.data as Map<String, dynamic>;
          riskScore = (riskData['score'] ?? 0).toDouble();
          final level = (riskData['level'] ?? '未知').toString().toLowerCase();
          if (level.contains('低')) {
            riskLevel = '低风险';
            riskColor = const Color(0xFF4CAF50);
          } else if (level.contains('中') || level.contains('moderate')) {
            riskLevel = '中风险';
            riskColor = const Color(0xFFFFA726);
          } else {
            riskLevel = '高风险';
            riskColor = const Color(0xFFFF5252);
          }
        }
      } catch (_) {
        // Risk score endpoint may not be available — fall back to computed score
        riskScore = _computeRiskScore(records);
        riskLevel = _riskLabel(riskScore);
        riskColor = _riskColorForScore(riskScore);
      }

      setState(() {
        _records = records;
        _loading = false;
        _riskScore = riskScore;
        _riskLevel = riskLevel;
        _riskColor = riskColor;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  /// Compute a simple risk score from health records when API is unavailable.
  double _computeRiskScore(List<HealthRecord> records) {
    if (records.isEmpty) return 0;
    final latest = records.first;
    double score = 0;
    if (latest.hr != null && (latest.hr! < 60 || latest.hr! > 100)) score += 0.2;
    if (latest.spo2 != null && latest.spo2! < 95) score += 0.3;
    if (latest.bpSystolic != null && latest.bpSystolic! > 140) score += 0.25;
    if (latest.bpDiastolic != null && latest.bpDiastolic! > 90) score += 0.15;
    if (latest.sleepHours != null && latest.sleepHours! < 6) score += 0.1;
    return (score * 100).clamp(0, 100);
  }

  String _riskLabel(double score) {
    if (score < 30) return '低风险';
    if (score < 60) return '中风险';
    return '高风险';
  }

  Color _riskColorForScore(double score) {
    if (score < 30) return const Color(0xFF4CAF50);
    if (score < 60) return const Color(0xFFFFA726);
    return const Color(0xFFFF5252);
  }

  // Derive current values from fetched records
  int? get _latestHr => _records.isNotEmpty ? _records.first.hr : null;
  int? get _latestSpo2 => _records.isNotEmpty ? _records.first.spo2 : null;
  int? get _latestSteps => _records.isNotEmpty ? _records.first.steps : null;
  double? get _latestSleep => _records.isNotEmpty ? _records.first.sleepHours : null;
  int? get _latestSys => _records.isNotEmpty ? _records.first.bpSystolic : null;
  int? get _latestDia => _records.isNotEmpty ? _records.first.bpDiastolic : null;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  SliverToBoxAdapter(
                    child: Container(
                      padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
                      color: AppTheme.bgCard,
                      child: Row(
                        children: [
                          IconButton(icon: const Icon(Icons.arrow_back_ios_new, size: 18), onPressed: () => Navigator.of(context).pop()),
                          const Expanded(child: Text('健康数据', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
                          IconButton(icon: const Icon(Icons.share_outlined), onPressed: () {}),
                        ],
                      ),
                    ),
                  ),

                  // Risk score card — now fetched from API or computed locally
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 20),
                      child: Container(
                        decoration: BoxDecoration(gradient: LinearGradient(colors: [const Color(0xFF4A90D9), const Color(0xFF357ABD)]), borderRadius: BorderRadius.circular(16)),
                        padding: const EdgeInsets.all(20),
                        child: Column(children: [
                          const Text('综合健康风险评估', style: TextStyle(fontSize: 13, color: Colors.white, opacity: 0.9)),
                          const SizedBox(height: 8),
                          Stack(alignment: Alignment.center, children: [
                            SizedBox(width: 100, height: 100, child: CircularProgressIndicator(value: _riskScore / 100, strokeWidth: 8, backgroundColor: Colors.white.withOpacity(0.2), valueColor: AlwaysStoppedAnimation<Color>(_riskColor))),
                            Column(mainAxisSize: MainAxisSize.min, children: [
                              Text('${_riskScore.toInt()}', style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w800, color: Colors.white)),
                              Text('/ 100', style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.8))),
                            ]),
                          ]),
                          const SizedBox(height: 4),
                          Text(_riskLevel, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: _riskColor)),
                          const SizedBox(height: 6),
                          Text(_riskSummary(), style: TextStyle(fontSize: 11, color: Colors.white.withOpacity(0.85), height: 1.5)),
                        ]),
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
                          return Padding(padding: const EdgeInsets.only(right: 8), child: FilterChip(label: Text(range), selected: isActive, onSelected: (_) { setState(() => _timeRange = range); _fetchData(); }, selectedColor: AppTheme.primary, labelStyle: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: isActive ? Colors.white : const Color(0xFF888888)), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)), side: BorderSide(color: isActive ? AppTheme.primary : const Color(0xFFF0F0F5))));
                        }).toList(),
                      ),
                    ),
                  ),

                  // Heart rate metric
                  SliverToBoxAdapter(
                    child: _metricCard(
                      title: '❤️ 心率',
                      trend: _trendLabel(_latestHr),
                      currentValue: '${_latestHr ?? '--'}',
                      unit: 'bpm (静息)',
                      miniChart: _buildMiniChart(7, AppTheme.primary, [55, 60, 45, 70, 65, 50, 55]),
                    ),
                  ),

                  // SpO2 metric
                  SliverToBoxAdapter(
                    child: _metricCard(
                      title: '💨 血氧饱和度',
                      trend: _trendLabelSpO2(_latestSpo2),
                      currentValue: '${_latestSpo2 ?? '--'}',
                      unit: '%',
                      badge: _latestSpo2 != null ? BadgeChip(label: _latestSpo2! >= 95 ? '正常' : '偏低', color: _latestSpo2! >= 95 ? AppTheme.statusNormal : AppTheme.statusWarning) : null,
                      miniChart: _buildMiniChart(7, AppTheme.statusNormal, [90, 92, 88, 95, 90, 92, 88]),
                    ),
                  ),

                  // Blood pressure metric
                  if (_latestSys != null && _latestDia != null)
                    SliverToBoxAdapter(
                      child: _metricCard(
                        title: '🩺 血压',
                        trend: '↓ 改善',
                        currentValue: '$_latestSys/$_latestDia',
                        unit: 'mmHg',
                        badge: const BadgeChip(label: '偏高', color: AppTheme.statusWarning),
                        child: Padding(
                          padding: const EdgeInsets.only(top: 10),
                          child: Row(children: [
                            Expanded(child: _bpBox('${_latestSys}', '收缩压', const Color(0xFFE65100))),
                            const SizedBox(width: 8),
                            Expanded(child: _bpBox('${_latestDia}', '舒张压', const Color(0xFFE65100))),
                          ]),
                        ),
                      ),
                    ),

                  // Sleep quality metric
                  if (_latestSleep != null)
                    SliverToBoxAdapter(
                      child: _metricCard(
                        title: '😴 睡眠质量',
                        trend: '↓ 改善',
                        currentValue: _latestSleep!.toStringAsFixed(1),
                        unit: '小时',
                        badge: const BadgeChip(label: '良好', color: AppTheme.statusNormal),
                        child: Padding(
                          padding: const EdgeInsets.only(top: 10),
                          child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                            const Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [Text('深睡 2.1h', style: TextStyle(fontSize: 10, color: Color(0xFF999999))), Text('浅睡 3.8h', style: TextStyle(fontSize: 10, color: Color(0xFF999999))), Text('REM 1.3h', style: TextStyle(fontSize: 10, color: Color(0xFF999999)))]),
                            const SizedBox(height: 4),
                            ClipRRect(borderRadius: BorderRadius.circular(4), child: Row(children: [Expanded(flex: 29, child: Container(color: const Color(0xFF5C6BC0))), Expanded(flex: 53, child: Container(color: const Color(0xFF7986CB))), Expanded(flex: 18, child: Container(color: const Color(0xFF9FA8DA)))])),
                          ]),
                        ),
                      ),
                    ),

                  // Steps summary
                  SliverToBoxAdapter(
                    child: Padding(padding: const EdgeInsets.only(top: 16, left: 20, right: 20), child: Row(children: [
                      _stepStat('🚶', '${(_latestSteps ?? 0).toString().replaceAllMapped(RegExp(r'(\d{1,3})(?=(\d{3})+(?!\d))'), (m) => '${m[1]},')}', '今日步数'),
                      const SizedBox(width: 12),
                      _stepStat('🔥', '186', '千卡消耗'),
                      const SizedBox(width: 12),
                      _stepStat('⏱️', '42', '活动分钟'),
                    ])),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 24)),
                ],
              ),
      bottomNavigationBar: BottomNavBar(selectedTab: _selectedIndex, onTabSelected: (i) => setState(() => _selectedIndex = i)),
    );
  }

  String _riskSummary() {
    if (_latestHr != null && (_latestHr! < 60 || _latestHr! > 100)) return '心率略高于/低于正常范围，建议持续监测';
    if (_latestSpo2 != null && _latestSpo2! < 95) return '血氧偏低，请注意休息并咨询医生';
    return '各项指标基本正常，步数略低于目标值';
  }

  String _trendLabel(int? hr) {
    if (hr == null) return '暂无数据';
    if (hr < 60 || hr > 100) return '↗ 偏高';
    return '↔ 稳定';
  }

  String _trendLabelSpO2(int? spo2) {
    if (spo2 == null) return '暂无数据';
    if (spo2 < 95) return '↘ 偏低';
    return '↔ 正常';
  }

  Widget _metricCard({required String title, required String trend, required String currentValue, required String unit, BadgeChip? badge, Widget? child, Widget? miniChart}) {
    return Padding(padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 5), child: Container(padding: const EdgeInsets.all(16), decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(AppTheme.cardRadius), border: Border.all(color: const Color(0xFFF5F5FA)), boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 6)]), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [Text(title, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600)), Text(trend, style: const TextStyle(fontSize: 11, color: Color(0xFF999999)))]),
      const SizedBox(height: 12),
      Row(children: [Text(currentValue, style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800)), Text(unit, style: const TextStyle(fontSize: 12, color: Color(0xFF999999))), if (badge != null) const SizedBox(width: 8), if (badge != null) badge]),
      if (miniChart != null) miniChart,
      if (child != null) child,
    ]));
  }

  Widget _buildMiniChart(int bars, Color color, List<double> heights) {
    return Padding(padding: const EdgeInsets.only(top: 10), child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: List.generate(bars, (i) => Expanded(child: Container(margin: const EdgeInsets.symmetric(horizontal: 1.5), height: heights[i] / 100 * 40, decoration: BoxDecoration(color: color, borderRadius: const BorderRadius.vertical(top: Radius.circular(3)), gradient: LinearGradient(colors: [color, color.withOpacity(0.6)], begin: Alignment.bottomCenter, end: Alignment.topCenter))))));
  }

  Widget _stepStat(String icon, String value, String label) {
    return Expanded(child: Container(padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 10), decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 4)]), child: Column(children: [Text(icon, style: const TextStyle(fontSize: 20)), const SizedBox(height: 6), Text(value, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)), Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999)))]));
  }

  Widget _bpBox(String value, String label, Color color) {
    return Container(padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 6), decoration: BoxDecoration(color: color.withOpacity(0.08), borderRadius: BorderRadius.circular(8)), child: Column(children: [Text(value, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: color)), Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999)))]));
  }
}

class BadgeChip extends StatelessWidget {
  final String label;
  final Color color;
  const BadgeChip({super.key, required this.label, required this.color});
  @override
  Widget build(BuildContext context) => Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3), decoration: BoxDecoration(color: color.withOpacity(0.1), borderRadius: BorderRadius.circular(10)), child: Text(label, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: color)));
}

import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../../models/health.dart';
import '../../services/offline_cache.dart';
import '../../app_state.dart';

/// AI health analysis report page — deep-dive insights from health data trends.
class AIReportPage extends StatefulWidget {
  const AIReportPage({super.key});

  @override
  State<AIReportPage> createState() => _AIReportPageState();
}

class _AIReportPageState extends State<AIReportPage> {
  int _selectedIndex = 1; // matches '健康' tab
  bool _loading = true;
  List<HealthRecord> _records = [];
  double _riskScore = 0;
  String _riskLevel = '加载中...';
  Color _riskColor = AppTheme.statusWarning;
  String _summary = '';
  List<Insight> _insights = [];

  String get _elderlyId => context.read<AppState>().elderlyId ?? '';

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final healthResp = await ApiClient.instance.get('/health/records', query: {'range': '本月'});
      final list = (healthResp.data as List);
      final records = list.map((r) => HealthRecord.fromJson(r as Map<String, dynamic>)).toList();

      if (_elderlyId.isNotEmpty) {
        for (final record in records) {
          OfflineCache.cacheHealth(_elderlyId, record.toJson());
        }
      }

      double riskScore = 0;
      String riskLevel = '暂无数据';
      Color riskColor = AppTheme.statusNormal;
      try {
        final riskResp = await ApiClient.instance.get('/health/risk-score');
        if (riskResp.data != null) {
          final riskData = riskResp.data as Map<String, dynamic>;
          riskScore = (riskData['score'] ?? 0).toDouble();
          final level = (riskData['level'] ?? '未知').toString().toLowerCase();
          if (level.contains('低')) {
            riskLevel = '低风险';
            riskColor = AppTheme.statusNormal;
          } else if (level.contains('中') || level.contains('moderate')) {
            riskLevel = '中风险';
            riskColor = AppTheme.statusWarning;
          } else {
            riskLevel = '高风险';
            riskColor = AppTheme.statusDanger;
          }
        }
      } catch (_) {
        riskScore = _computeRisk(records);
        riskLevel = _riskLabel(riskScore);
        riskColor = _riskColorForScore(riskScore);
      }

      final latest = records.isNotEmpty ? records.first : null;
      final summary = _generateSummary(latest, records);
      final insights = _generateInsights(latest, records);

      setState(() {
        _records = records;
        _riskScore = riskScore;
        _riskLevel = riskLevel;
        _riskColor = riskColor;
        _summary = summary;
        _insights = insights;
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  double _computeRisk(List<HealthRecord> records) {
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
    if (score < 30) return AppTheme.statusNormal;
    if (score < 60) return AppTheme.statusWarning;
    return AppTheme.statusDanger;
  }

  String _generateSummary(HealthRecord? latest, List<HealthRecord> records) {
    final parts = <String>[];
    if (latest != null) {
      if (latest.hr != null && latest.hr! >= 60 && latest.hr! <= 100) parts.add('心率稳定');
      if (latest.spo2 != null && latest.spo2! >= 95) parts.add('血氧水平良好');
      if (latest.steps != null && latest.steps! >= 5000) parts.add('日常活动量达标');
      if (latest.sleepHours != null && latest.sleepHours! >= 6) parts.add('睡眠质量基本正常');
    }
    if (parts.isEmpty) parts.add('数据不足，建议持续佩戴手环采集健康数据');
    return parts.join('，');
  }

  List<Insight> _generateInsights(HealthRecord? latest, List<HealthRecord> records) {
    final insights = <Insight>[];

    // Trend analysis
    if (records.length >= 2) {
      final hrTrend = records[0].hr != null && records[1].hr != null
          ? (records[0].hr! > records[1].hr! ? '上升' : (records[0].hr! < records[1].hr! ? '下降' : '平稳'))
          : '暂无趋势';
      insights.add(Insight(
        icon: Icons.favorite,
        title: '心率趋势',
        desc: '近${records.length}日静息心率呈$hrTrend趋势',
        color: AppTheme.primary,
      ));

      final stepTotal = records.fold<int>(0, (s, r) => s + (r.steps ?? 0));
      final avgSteps = stepTotal / records.length;
      if (avgSteps < 3000) {
        insights.add(Insight(
          icon: Icons.directions_walk,
          title: '运动建议',
          desc: '日均步数仅${avgSteps.toInt()}，建议每日散步30分钟以上',
          color: AppTheme.statusWarning,
        ));
      } else {
        insights.add(Insight(
          icon: Icons.directions_walk,
          title: '运动达标',
          desc: '日均步数${avgSteps.toInt()}，运动量充足',
          color: AppTheme.statusNormal,
        ));
      }
    }

    // Sleep insight
    if (latest?.sleepHours != null && latest!.sleepHours! < 6) {
      insights.add(Insight(
        icon: Icons.nightlight,
        title: '睡眠提醒',
        desc: '最近睡眠${latest.sleepHours!.toStringAsFixed(1)}小时，低于推荐值6小时',
        color: AppTheme.statusWarning,
      ));
    }

    // BP insight
    if (latest?.bpSystolic != null && latest!.bpSystolic! > 140) {
      insights.add(Insight(
        icon: Icons.warning_amber,
        title: '血压偏高',
        desc: '收缩压${latest.bpSystolic}mmHg，建议低盐饮食并咨询医生',
        color: AppTheme.statusDanger,
      ));
    }

    // SpO2 insight
    if (latest?.spo2 != null && latest!.spo2! < 95) {
      insights.add(Insight(
        icon: Icons.air,
        title: '血氧偏低',
        desc: 'SpO2 ${latest.spo2}%，注意通风并减少剧烈运动',
        color: AppTheme.statusWarning,
      ));
    }

    if (insights.isEmpty) {
      insights.add(Insight(
        icon: Icons.check_circle,
        title: '各项指标正常',
        desc: '当前健康数据未见异常，继续保持健康生活方式',
        color: AppTheme.statusNormal,
      ));
    }

    return insights;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  _buildHeader(),
                  _buildRiskGauge(),
                  const SliverToBoxAdapter(child: SizedBox(height: 16)),
                  _buildSummaryCard(),
                  const SliverToBoxAdapter(child: SizedBox(height: 16)),
                  _buildInsightsList(),
                  const SliverToBoxAdapter(child: SizedBox(height: 24)),
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
        padding: const EdgeInsets.fromLTRB(20, 12, 20, 20),
        color: AppTheme.bgCard,
        child: Row(
          children: [
            IconButton(
              icon: const Icon(Icons.arrow_back_ios_new, size: 18),
              onPressed: () => Navigator.of(context).pop(),
            ),
            const Expanded(
              child: Text('AI 健康分析报告', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
            ),
            IconButton(
              icon: const Icon(Icons.auto_awesome),
              color: AppTheme.primary,
              onPressed: () {},
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRiskGauge() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(colors: [AppTheme.primary, AppTheme.accent]),
            borderRadius: BorderRadius.circular(16),
          ),
          padding: const EdgeInsets.all(24),
          child: Column(
            children: [
              const Text('综合健康风险评估', style: TextStyle(fontSize: 14, color: Colors.white, fontWeight: FontWeight.w600)),
              const SizedBox(height: 12),
              SizedBox(
                width: 120,
                height: 120,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    SizedBox(
                      width: 120,
                      height: 120,
                      child: CircularProgressIndicator(
                        value: _riskScore / 100,
                        strokeWidth: 10,
                        backgroundColor: Colors.white.withValues(alpha: 0.2),
                        valueColor: AlwaysStoppedAnimation<Color>(_riskColor),
                      ),
                    ),
                    Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('${_riskScore.toInt()}', style: const TextStyle(fontSize: 32, fontWeight: FontWeight.w800, color: Colors.white)),
                        Text('/ 100', style: TextStyle(fontSize: 12, color: Colors.white.withValues(alpha: 0.8))),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 8),
              Text(_riskLevel, style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: _riskColor)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSummaryCard() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20),
        child: Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: AppTheme.bgCard,
            borderRadius: BorderRadius.circular(14),
            border: Border.all(color: const Color(0xFFF0E8E3)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 28,
                    height: 28,
                    decoration: BoxDecoration(
                      color: AppTheme.primary.withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: const Icon(Icons.lightbulb, size: 16, color: AppTheme.primary),
                  ),
                  const SizedBox(width: 8),
                  const Text('AI 总结', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700)),
                ],
              ),
              const SizedBox(height: 12),
              Text(_summary, style: const TextStyle(fontSize: 13, color: Color(0xFF374151), height: 1.6)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildInsightsList() {
    return SliverPadding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      sliver: SliverList(
        delegate: SliverChildBuilderDelegate((ctx, i) {
          if (i >= _insights.length) return null;
          final insight = _insights[i];
          return Container(
            margin: const EdgeInsets.only(bottom: 10),
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: AppTheme.bgCard,
              borderRadius: BorderRadius.circular(14),
              border: Border.all(color: const Color(0xFFF0F0F0)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Container(
                  width: 36,
                  height: 36,
                  decoration: BoxDecoration(
                    color: insight.color.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(10),
                  ),
                  child: Icon(insight.icon, size: 20, color: insight.color),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(insight.title, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w700)),
                      const SizedBox(height: 4),
                      Text(insight.desc, style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280), height: 1.5)),
                    ],
                  ),
                ),
              ],
            ),
          );
        }, childCount: _insights.length),
      ),
    );
  }
}

class Insight {
  final IconData icon;
  final String title;
  final String desc;
  final Color color;
  const Insight({required this.icon, required this.title, required this.desc, required this.color});
}

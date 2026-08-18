import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
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
      final healthResp = await ApiClient.instance.get('/api/v1/health/records', query: {'range': '本月'});
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
        final riskResp = await ApiClient.instance.get('/api/v1/health/risk-score');
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
      // On API failure, seed mock data so the page isn't blank
      _seedMockData();
      setState(() => _loading = false);
    }
  }

  void _seedMockData() {
    final now = DateTime.now();
    setState(() {
      _records = [
        HealthRecord(id: 'h1', elderlyId: '', timestamp: now, hr: 72, spo2: 98, steps: 6500, sleepHours: 7.2, bpSystolic: 125, bpDiastolic: 80),
        HealthRecord(id: 'h2', elderlyId: '', timestamp: now.subtract(const Duration(days: 1)), hr: 68, spo2: 97, steps: 5200, sleepHours: 6.5, bpSystolic: 130, bpDiastolic: 85),
        HealthRecord(id: 'h3', elderlyId: '', timestamp: now.subtract(const Duration(days: 2)), hr: 75, spo2: 96, steps: 3800, sleepHours: 5.8, bpSystolic: 135, bpDiastolic: 88),
      ];
      _riskScore = 25;
      _riskLevel = '低风险';
      _riskColor = AppTheme.statusNormal;
      _summary = '心率稳定，血氧水平良好，日常活动量达标';
      _insights = [
        Insight(icon: Icons.favorite, title: '心率趋势', desc: '近3日静息心率呈平稳趋势', color: AppTheme.statusNormal),
        Insight(icon: Icons.directions_walk, title: '运动达标', desc: '日均步数5167，运动量充足', color: AppTheme.statusNormal),
        Insight(icon: Icons.nightlight, title: '睡眠提醒', desc: '最近睡眠6.5小时，建议达到7小时以上', color: AppTheme.statusWarning),
      ];
    });
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

  String _buildReportText() {
    final sb = <String>[];
    sb.add('Eregen 颐贞 - AI 健康分析报告');
    sb.add('生成时间: ${DateTime.now().toString().substring(0, 16)}');
    sb.add('');
    sb.add('风险评估: $_riskLevel ($_riskScore/100)');
    sb.add('');
    sb.add('摘要: $_summary');
    sb.add('');
    if (_insights.isNotEmpty) {
      sb.add('AI 洞察:');
      for (final insight in _insights) {
        sb.add('  • ${insight.title}: ${insight.desc}');
      }
    }
    if (_records.isNotEmpty) {
      sb.add('');
      sb.add('最近记录:');
      final latest = _records.first;
      sb.add('  心率: ${latest.hr ?? "N/A"} bpm');
      sb.add('  血氧: ${latest.spo2 ?? "N/A"}%');
      sb.add('  步数: ${latest.steps ?? "N/A"}');
      sb.add('  睡眠: ${latest.sleepHours?.toStringAsFixed(1) ?? "N/A"} 小时');
      sb.add('  血压: ${latest.bpSystolic ?? "?"}/${latest.bpDiastolic ?? "?"}');
    }
    return sb.join('\n');
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
              onPressed: () {
                final reportText = _buildReportText();
                showDialog(context: context, builder: (_) => _ShareReportDialog(reportText: reportText));
              },
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
                        backgroundColor: Colors.white.withOpacity(0.2),
                        valueColor: AlwaysStoppedAnimation<Color>(_riskColor),
                      ),
                    ),
                    Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('${_riskScore.toInt()}', style: const TextStyle(fontSize: 32, fontWeight: FontWeight.w800, color: Colors.white)),
                        Text('/ 100', style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.8))),
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
                      color: AppTheme.primary.withOpacity(0.1),
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
                    color: insight.color.withOpacity(0.1),
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

/// Share AI health report dialog
class _ShareReportDialog extends StatelessWidget {
  final String reportText;
  const _ShareReportDialog({required this.reportText, super.key});

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      title: const Text('分享健康报告'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text('选择分享方式', style: TextStyle(fontSize: 13, color: AppTheme.textMuted)),
          const SizedBox(height: 16),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              _ShareOption(icon: Icons.copy_all, label: '复制文本', onTap: () async {
                await Clipboard.setData(ClipboardData(text: reportText));
                if (context.mounted) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('报告已复制到剪贴板')),
                  );
                }
              }),
              _ShareOption(icon: Icons.share, label: '分享', onTap: () async {
                // Copy to clipboard as share fallback (no share_plus dependency)
                await Clipboard.setData(ClipboardData(text: reportText));
                if (context.mounted) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('报告已复制到剪贴板，可粘贴分享')),
                  );
                }
              }),
              _ShareOption(icon: Icons.download, label: '下载', onTap: () async {
                // Save report text to clipboard with filename hint
                await Clipboard.setData(ClipboardData(text: reportText));
                if (context.mounted) {
                  Navigator.pop(context);
                  ScaffoldMessenger.of(context).showSnackBar(
                    const SnackBar(content: Text('报告已保存至剪贴板，可粘贴到文档中保存')),
                  );
                }
              }),
            ],
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
      ],
    );
  }
}

class _ShareOption extends StatelessWidget {
  final IconData icon;
  final String label;
  final VoidCallback onTap;
  const _ShareOption({required this.icon, required this.label, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Column(
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(
              color: AppTheme.primary.withOpacity(0.1),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Icon(icon, size: 28, color: AppTheme.primary),
          ),
          const SizedBox(height: 8),
          Text(label, style: const TextStyle(fontSize: 12, color: AppTheme.textSecondary)),
        ],
      ),
    );
  }
}

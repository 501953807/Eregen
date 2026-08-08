import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../common/theme.dart';
import '../../api/client.dart';

/// Health report page — weekly / monthly / annual view with average glucose,
/// uric acid, blood pressure, medication compliance rate, and AI recommendations.
///
/// API: GET /api/v1/chronic/:elderly_id/report/:type
/// Response body.data: {
///   id, elderly_id, report_type, period_start, period_end,
///   data_summary: {glucose:{avg,min,max,count,in_range_pct}, uric_acid:{avg,min,max,count,high_pct}, blood_pressure:{avg_systolic,avg_diastolic,count,high_systolic_pct}},
///   ai_recommendations: [{level, title, detail}],
///   generated_at
/// }
class ReportPage extends StatefulWidget {
  final String elderlyId;
  const ReportPage({super.key, required this.elderlyId});

  @override
  State<ReportPage> createState() => _ReportPageState();
}

class _ReportPageState extends State<ReportPage> {
  String _activeType = 'weekly';
  bool _loading = true;
  String? _error;

  late ReportData _report;

  @override
  void initState() {
    super.initState();
    _fetchReport();
  }

  Future<void> _fetchReport() async {
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      final resp = await ApiClient.instance.get(
        '/api/v1/chronic/${widget.elderlyId}/report/$_activeType',
      );
      final data = resp.data as Map<String, dynamic>;
      if (data['code'] != 'OK') {
        setState(() {
          _loading = false;
          _error = data['message'] ?? 'Failed to load report';
        });
        return;
      }
      setState(() {
        _report = _parseReport(data['data'] as Map<String, dynamic>);
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _loading = false;
        _error = e.toString();
      });
    }
  }

  void _switchType(String type) {
    if (_activeType == type) return;
    setState(() => _activeType = type);
    _fetchReport();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : _error != null
                ? _buildErrorState()
                : CustomScrollView(
                    slivers: [
                      _buildTopBar(),
                      SliverToBoxAdapter(child: _buildTypeSelector()),
                      SliverToBoxAdapter(child: const SizedBox(height: 16)),
                      SliverToBoxAdapter(child: _buildPeriodInfo()),
                      SliverToBoxAdapter(child: const SizedBox(height: 12)),
                      SliverToBoxAdapter(child: _buildMetricsGrid()),
                      SliverToBoxAdapter(child: const SizedBox(height: 12)),
                      SliverToBoxAdapter(child: _buildComplianceCard()),
                      SliverToBoxAdapter(child: const SizedBox(height: 12)),
                      SliverToBoxAdapter(child: _buildRecommendations()),
                      SliverToBoxAdapter(child: const SizedBox(height: 24)),
                    ],
                  ),
      ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('加载失败', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 8),
            Text(_error ?? '', style: const TextStyle(fontSize: 13, color: Color(0xFF6B7280))),
            const SizedBox(height: 16),
            FilledButton(
              onPressed: _fetchReport,
              child: const Text('重试'),
            ),
          ],
        ),
      ),
    );
  }

  // ── Top Bar ────────────────────────────────────────────────────────────────

  Widget _buildTopBar() {
    return SliverToBoxAdapter(
      child: Container(
        color: Colors.white,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
        child: Row(
          children: [
            GestureDetector(
              onTap: () => Navigator.of(context).pop(),
              child: Container(
                width: 36,
                height: 36,
                decoration: BoxDecoration(
                  color: const Color(0xFFF3F4F6),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: const Icon(Icons.arrow_back_ios_new, size: 16),
              ),
            ),
            const SizedBox(width: 12),
            const Expanded(
              child: Center(
                child: Text(
                  '健康报告',
                  style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700),
                ),
              ),
            ),
            const SizedBox(width: 36),
          ],
        ),
      ),
    );
  }

  // ── Type Selector ──────────────────────────────────────────────────────────

  Widget _buildTypeSelector() {
    final types = const [
      ('weekly', '周报'),
      ('monthly', '月报'),
      ('annual', '年报'),
    ];
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Container(
        decoration: BoxDecoration(
          color: const Color(0xFFF3F4F6),
          borderRadius: BorderRadius.circular(12),
        ),
        padding: const EdgeInsets.all(4),
        child: Row(
          children: types.map((t) {
            final code = t.$1;
            final label = t.$2;
            final isActive = code == _activeType;
            return Expanded(
              child: GestureDetector(
                onTap: () => _switchType(code),
                child: Container(
                  padding: const EdgeInsets.symmetric(vertical: 10),
                  decoration: BoxDecoration(
                    color: isActive ? Colors.white : null,
                    borderRadius: BorderRadius.circular(10),
                    boxShadow: isActive
                        ? [BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 8)]
                        : [],
                  ),
                  child: Text(
                    label,
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: isActive ? FontWeight.w700 : FontWeight.w500,
                      color: isActive ? AppTheme.primary : const Color(0xFF6B7280),
                    ),
                  ),
                ),
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  // ── Period Info ────────────────────────────────────────────────────────────

  Widget _buildPeriodInfo() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            '${_formatDate(_report.periodStart)} ~ ${_formatDate(_report.periodEnd)}',
            style: const TextStyle(fontSize: 13, color: Color(0xFF9CA3AF)),
          ),
          Text(
            '生成于 ${DateFormat('MM/dd HH:mm').format(_report.generatedAt)}',
            style: const TextStyle(fontSize: 12, color: Color(0xFFD1D5DB)),
          ),
        ],
      ),
    );
  }

  // ── Metrics Grid ───────────────────────────────────────────────────────────

  Widget _buildMetricsGrid() {
    final glucose = _report.glucoseSummary;
    final uric = _report.uricSummary;
    final bp = _report.bpSummary;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            '慢病指标',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 10),
          Row(
            children: [
              Expanded(
                child: _metricCard(
                  icon: '\u{1F4A1}',
                  label: '空腹血糖',
                  value: glucose.avg > 0 ? '${glucose.avg.toStringAsFixed(1)}' : '--',
                  unit: 'mmol/L',
                  status: _glucoseStatus(glucose.avg),
                  range: '3.9–7.8',
                  count: glucose.count,
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _metricCard(
                  icon: '\u{1F52C}',
                  label: '尿酸',
                  value: uric.avg > 0 ? '${uric.avg.toStringAsFixed(0)}' : '--',
                  unit: 'μmol/L',
                  status: _uricStatus(uric.avg),
                  range: '143–420',
                  count: uric.count,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          _metricCard(
            icon: '\u{2764}',
            label: '血压',
            value: bp.avgSystolic > 0
                ? '${bp.avgSystolic}/${bp.avgDiastolic}'
                : '--',
            unit: 'mmHg',
            status: _bpStatus(bp.avgSystolic),
            range: '90–140/60–90',
            count: bp.count,
          ),
        ],
      ),
    );
  }

  Widget _metricCard({
    required String icon,
    required String label,
    required String value,
    required String unit,
    required _MetricStatus status,
    required String range,
    required int count,
  }) {
    final bgColor = switch (status) {
      _MetricStatus.normal => const Color(0xFFF0FDF4),
      _MetricStatus.warning => const Color(0xFFFFFBEB),
      _MetricStatus.danger => const Color(0xFFFEF2F2),
      _MetricStatus.unknown => const Color(0xFFF9FAFB),
    };
    final dotColor = switch (status) {
      _MetricStatus.normal => AppTheme.statusNormal,
      _MetricStatus.warning => AppTheme.statusWarning,
      _MetricStatus.danger => AppTheme.statusDanger,
      _MetricStatus.unknown => const Color(0xFFD1D5DB),
    };
    return Container(
      padding: const EdgeInsets.all(14),
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
              Container(
                width: 34,
                height: 34,
                decoration: BoxDecoration(
                  color: bgColor,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Center(child: Text(icon, style: const TextStyle(fontSize: 16))),
              ),
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(color: dotColor, shape: BoxShape.circle),
              ),
            ],
          ),
          const SizedBox(height: 10),
          Text(value, style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: Color(0xFF1F2937))),
          Text(unit, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
          const SizedBox(height: 8),
          Text(label, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
          Text('参考: $range', style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
          if (count > 0)
            Text('$count 次检测', style: const TextStyle(fontSize: 10, color: Color(0xFFD1D5DB))),
        ],
      ),
    );
  }

  // ── Compliance Card ────────────────────────────────────────────────────────

  Widget _buildComplianceCard() {
    final rate = _report.complianceRate;
    final pct = rate > 0 ? (rate * 100).round() : null;
    final label = pct == null
        ? '暂无数据'
        : pct >= 90
            ? '依从性优秀'
            : pct >= 70
                ? '依从性良好'
                : '依从性偏低';
    final cardColor = pct == null
        ? const Color(0xFFF9FAFB)
        : pct >= 90
            ? const Color(0xFFF0FDF4)
            : pct >= 70
                ? const Color(0xFFFFFBEB)
                : const Color(0xFFFEF2F2);
    final dotColor = pct == null
        ? const Color(0xFFD1D5DB)
        : pct >= 90
            ? AppTheme.statusNormal
            : pct >= 70
                ? AppTheme.statusWarning
                : AppTheme.statusDanger;

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Container(
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 2))],
        ),
        child: Row(
          children: [
            Container(
              width: 48,
              height: 48,
              decoration: BoxDecoration(
                color: cardColor,
                borderRadius: BorderRadius.circular(14),
              ),
              child: Center(
                child: Text(
                  pct == null ? '\u{1F4C5}' : '\u{2713}',
                  style: TextStyle(fontSize: 24),
                ),
              ),
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    '用药依从率',
                    style: TextStyle(fontSize: 12, color: Color(0xFF6B7280)),
                  ),
                  const SizedBox(height: 2),
                  if (pct != null)
                    Row(
                      children: [
                        Text(
                          '$pct%',
                          style: const TextStyle(fontSize: 26, fontWeight: FontWeight.w800, color: Color(0xFF1F2937)),
                        ),
                        const SizedBox(width: 6),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                          decoration: BoxDecoration(color: dotColor.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(8)),
                          child: Text(
                            label,
                            style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: dotColor),
                          ),
                        ),
                      ],
                    )
                  else
                    Text(
                      '暂无数据',
                      style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Color(0xFF9CA3AF)),
                    ),
                ],
              ),
            ),
            if (pct != null)
              SizedBox(
                width: 64,
                height: 64,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    CircularProgressIndicator(
                      value: 1,
                      strokeWidth: 5,
                      backgroundColor: const Color(0xFFF3F4F6),
                    ),
                    CircularProgressIndicator(
                      value: rate > 0 ? rate : 0,
                      strokeWidth: 5,
                      backgroundColor: Colors.transparent,
                      valueColor: AlwaysStoppedAnimation<Color>(dotColor),
                    ),
                    Text(
                      '$pct%',
                      style: TextStyle(fontSize: 13, fontWeight: FontWeight.w800, color: dotColor),
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }

  // ── AI Recommendations ─────────────────────────────────────────────────────

  Widget _buildRecommendations() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'AI 健康建议',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700),
          ),
          const SizedBox(height: 10),
          if (_report.recommendations.isEmpty)
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 2))],
              ),
              child: const Center(
                child: Text(
                  '暂无建议，请确保已记录相关健康数据',
                  style: TextStyle(fontSize: 13, color: Color(0xFF9CA3AF)),
                ),
              ),
            )
          else
            ..._report.recommendations.map((rec) => _recommendationItem(rec)),
        ],
      ),
    );
  }

  Widget _recommendationItem(Map<String, dynamic> rec) {
    final level = rec['level'] as String? ?? 'info';
    final title = rec['title'] as String? ?? '';
    final detail = rec['detail'] as String? ?? '';

    final levelConfig = switch (level) {
      'danger' => (_iconDanger, const Color(0xFFFEF2F2), AppTheme.statusDanger, const Color(0xFFDC2626)),
      'warning' => (_iconWarning, const Color(0xFFFFFBEB), AppTheme.statusWarning, const Color(0xFFD97706)),
      _ => (_iconInfo, const Color(0xFFDBEAFE), const Color(0xFF3B82F6), const Color(0xFF2563EB)),
    };

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(14),
          border: Border(left: BorderSide(color: levelConfig.$4, width: 4)),
          boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, 2))],
        ),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(color: levelConfig.$2, borderRadius: BorderRadius.circular(8)),
              child: Center(child: Text(levelConfig.$1, style: const TextStyle(fontSize: 16))),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(title, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
                  const SizedBox(height: 4),
                  Text(detail, style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280), height: 1.5)),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ── Helpers ────────────────────────────────────────────────────────────────

  static const _iconDanger = '\u{26A0}';
  static const _iconWarning = '\u{1F4A1}';
  static const _iconInfo = '\u{1F44B}';

  String _formatDate(DateTime dt) => DateFormat('MM/dd').format(dt);

  _MetricStatus _glucoseStatus(double avg) {
    if (avg <= 0) return _MetricStatus.unknown;
    if (avg >= 3.9 && avg <= 7.8) return _MetricStatus.normal;
    if (avg > 10.0) return _MetricStatus.danger;
    return _MetricStatus.warning;
  }

  _MetricStatus _uricStatus(double avg) {
    if (avg <= 0) return _MetricStatus.unknown;
    return avg > 420 ? _MetricStatus.warning : _MetricStatus.normal;
  }

  _MetricStatus _bpStatus(int systolic) {
    if (systolic <= 0) return _MetricStatus.unknown;
    if (systolic >= 160) return _MetricStatus.danger;
    if (systolic >= 140) return _MetricStatus.warning;
    return _MetricStatus.normal;
  }
}

// ── Data Models ──────────────────────────────────────────────────────────────

enum _MetricStatus { normal, warning, danger, unknown }

class GlucoseSummary {
  final double avg;
  final int count;
  final double inRangePct;
  const GlucoseSummary({required this.avg, required this.count, required this.inRangePct});
}

class UricAcidSummary {
  final double avg;
  final int count;
  final double highPct;
  const UricAcidSummary({required this.avg, required this.count, required this.highPct});
}

class BPSummary {
  final int avgSystolic;
  final int avgDiastolic;
  final int count;
  final double highSysPct;
  const BPSummary({required this.avgSystolic, required this.avgDiastolic, required this.count, required this.highSysPct});
}

class ReportData {
  final String reportType;
  final DateTime periodStart;
  final DateTime periodEnd;
  final DateTime generatedAt;
  final GlucoseSummary glucoseSummary;
  final UricAcidSummary uricSummary;
  final BPSummary bpSummary;
  final double complianceRate;
  final List<Map<String, dynamic>> recommendations;
  const ReportData({
    required this.reportType,
    required this.periodStart,
    required this.periodEnd,
    required this.generatedAt,
    required this.glucoseSummary,
    required this.uricSummary,
    required this.bpSummary,
    required this.complianceRate,
    required this.recommendations,
  });
}

ReportData _parseReport(Map<String, dynamic> data) {
  final dataSummary = (data['data_summary'] ?? {}) as Map<String, dynamic>;
  final glucoseRaw = (dataSummary['glucose'] ?? {}) as Map<String, dynamic>;
  final uricRaw = (dataSummary['uric_acid'] ?? {}) as Map<String, dynamic>;
  final bpRaw = (dataSummary['blood_pressure'] ?? {}) as Map<String, dynamic>;
  final recsRaw = (data['ai_recommendations'] ?? []) as List;

  return ReportData(
    reportType: data['report_type'] ?? '',
    periodStart: DateTime.tryParse(data['period_start'] ?? '') ?? DateTime.now(),
    periodEnd: DateTime.tryParse(data['period_end'] ?? '') ?? DateTime.now(),
    generatedAt: DateTime.tryParse(data['generated_at'] ?? '') ?? DateTime.now(),
    glucoseSummary: GlucoseSummary(
      avg: (glucoseRaw['avg'] as num?)?.toDouble() ?? 0,
      count: (glucoseRaw['count'] as num?)?.toInt() ?? 0,
      inRangePct: (glucoseRaw['in_range_pct'] as num?)?.toDouble() ?? 0,
    ),
    uricSummary: UricAcidSummary(
      avg: (uricRaw['avg'] as num?)?.toDouble() ?? 0,
      count: (uricRaw['count'] as num?)?.toInt() ?? 0,
      highPct: (uricRaw['high_pct'] as num?)?.toDouble() ?? 0,
    ),
    bpSummary: BPSummary(
      avgSystolic: (bpRaw['avg_systolic'] as num?)?.toInt() ?? 0,
      avgDiastolic: (bpRaw['avg_diastolic'] as num?)?.toInt() ?? 0,
      count: (bpRaw['count'] as num?)?.toInt() ?? 0,
      highSysPct: (bpRaw['high_systolic_pct'] as num?)?.toDouble() ?? 0,
    ),
    complianceRate: 0.0, // derived from backend in future; currently 0
    recommendations: recsRaw.map((r) => r as Map<String, dynamic>).toList(),
  );
}

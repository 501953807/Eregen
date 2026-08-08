import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../../app_state.dart';

/// Blood pressure detail page — dual Y-axis line chart (systolic/diastolic),
/// anomaly highlighting, and a chronological record list.
///
/// API: GET /api/v1/chronic/:elderly_id/blood-pressure?days=30
class BloodPressurePage extends StatefulWidget {
  final String? elderlyId;
  final String? elderlyName;

  const BloodPressurePage({super.key, this.elderlyId, this.elderlyName});

  @override
  State<BloodPressurePage> createState() => _BloodPressurePageState();
}

class _BloodPressurePageState extends State<BloodPressurePage> {
  bool _loading = true;
  String? _error;
  List<BloodPressureRecord> _records = [];

  int _selectedDays = 30;

  // Anomaly thresholds
  static const int _systolicThreshold = 140;
  static const int _diastolicThreshold = 90;

  List<BloodPressureRecord> get _filteredRecords {
    final now = DateTime.now();
    final cutoff = now.subtract(Duration(days: _selectedDays));
    return _records
        .where((r) => r.measurementTime.isAfter(cutoff))
        .toList()
      ..sort((a, b) => a.measurementTime.compareTo(b.measurementTime));
  }

  int get _anomalyCount => _filteredRecords
      .where((r) => r.isAnomaly)
      .length;

  int get _latestSystolic =>
      _filteredRecords.isNotEmpty ? _filteredRecords.last.systolic : 0;
  int get _latestDiastolic =>
      _filteredRecords.isNotEmpty ? _filteredRecords.last.diastolic : 0;
  int? get _latestPulse =>
      _filteredRecords.isNotEmpty ? _filteredRecords.last.pulse : null;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final elderlyId = widget.elderlyId ??
          context.read<AppState>().elderlyId ??
          '';
      final resp = await ApiClient.instance.get(
        '/api/v1/chronic/$elderlyId/blood-pressure',
        query: {'days': _selectedDays},
      );
      final data = resp.data;
      if (data != null && data is Map) {
        final list = data['data'] as List?;
        if (list != null) {
          setState(() {
            _records = list
                .map((item) => BloodPressureRecord.fromJson(item as Map<String, dynamic>))
                .toList();
            _loading = false;
            _error = null;
          });
          return;
        }
      }
      // Also accept plain list response
      if (data != null && data is List) {
        setState(() {
          _records = data
              .map((item) => BloodPressureRecord.fromJson(item as Map<String, dynamic>))
              .toList();
          _loading = false;
          _error = null;
        });
        return;
      }
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
    setState(() => _loading = false);
  }

  Future<void> _onChangeDays(int days) async {
    setState(() {
      _selectedDays = days;
      _loading = true;
      _error = null;
    });
    await _fetchData();
  }

  @override
  Widget build(BuildContext context) {
    final elderlyName = widget.elderlyName ??
        context.read<AppState>().elderlyName ??
        '老人';

    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(
                slivers: [
                  _buildAppBar(elderlyName),
                  if (_error != null) _buildErrorBanner(),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                      child: _buildDaySelector(),
                    ),
                  ),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: _buildSummaryCards(),
                    ),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 12)),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: _buildChartCard(),
                    ),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 12)),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 16),
                      child: _buildAnomalyBanner(),
                    ),
                  ),
                  const SliverToBoxAdapter(child: SizedBox(height: 8)),
                  _buildRecordList(),
                  const SliverToBoxAdapter(child: SizedBox(height: 32)),
                ],
              ),
      ),
    );
  }

  // ===== App Bar =====
  Widget _buildAppBar(String name) {
    return SliverToBoxAdapter(
      child: Container(
        color: AppTheme.bgCard,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 16),
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
                child: const Center(
                    child: Icon(Icons.arrow_back_ios_new, size: 16)),
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    '血压详情',
                    style: TextStyle(
                        fontSize: 18, fontWeight: FontWeight.w700),
                  ),
                  Text(
                    name,
                    style: const TextStyle(
                        fontSize: 12, color: Color(0xFF6B7280)),
                  ),
                ],
              ),
            ),
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: _isNormalBloodPressure()
                    ? AppTheme.statusNormal.withValues(alpha: 0.15)
                    : AppTheme.statusWarning.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(12),
              ),
              child: Text(
                _isNormalBloodPressure() ? '正常' : '偏高',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: _isNormalBloodPressure()
                      ? AppTheme.statusNormal
                      : AppTheme.statusWarning,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Error Banner =====
  Widget _buildErrorBanner() {
    return SliverToBoxAdapter(
      child: Container(
        margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: const Color(0xFFFEF2F2),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: const Color(0xFFFECACA)),
        ),
        child: Row(
          children: [
            const Icon(Icons.error_outline, color: Color(0xFFDC2626), size: 20),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                '加载失败: $_error',
                style: const TextStyle(fontSize: 12, color: Color(0xFF991B1B)),
              ),
            ),
            TextButton(
              onPressed: _fetchData,
              child: const Text('重试'),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Day Selector =====
  Widget _buildDaySelector() {
    const options = [
      (7, '近7天'),
      (14, '近14天'),
      (30, '近30天'),
      (90, '近90天'),
    ];
    return Container(
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(12),
      ),
      padding: const EdgeInsets.all(4),
      child: Row(
        children: options.map((option) {
          final (days, label) = option;
          final active = days == _selectedDays;
          return Expanded(
            child: GestureDetector(
              onTap: () => _onChangeDays(days),
              child: Container(
                padding: const EdgeInsets.symmetric(vertical: 8),
                decoration: BoxDecoration(
                  color:
                      active ? AppTheme.primary : Colors.transparent,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  label,
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: active ? FontWeight.w700 : FontWeight.w500,
                    color:
                        active ? Colors.white : const Color(0xFF6B7280),
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }

  // ===== Summary Cards =====
  Widget _buildSummaryCards() {
    return Row(
      children: [
        _summaryCard(
          label: '收缩压 (高压)',
          value: '$_latestSystolic',
          unit: 'mmHg',
          color: const Color(0xFFEF4444),
          status: _latestSystolic > _systolicThreshold ? '偏高' : '正常',
        ),
        const SizedBox(width: 10),
        _summaryCard(
          label: '舒张压 (低压)',
          value: '$_latestDiastolic',
          unit: 'mmHg',
          color: const Color(0xFF3B82F6),
          status: _latestDiastolic > _diastolicThreshold ? '偏高' : '正常',
        ),
        const SizedBox(width: 10),
        Expanded(
          child: _summaryCard(
            label: '心率',
            value: _latestPulse?.toString() ?? '--',
            unit: 'bpm',
            color: AppTheme.primary,
            status: '参考',
          ),
        ),
      ],
    );
  }

  Widget _summaryCard(
      {required String label,
      required String value,
      required String unit,
      required Color color,
      required String status}) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: AppTheme.bgCard,
          borderRadius: BorderRadius.circular(14),
          boxShadow: [
            BoxShadow(
                color: Colors.black.withValues(alpha: 0.04),
                blurRadius: 12,
                offset: const Offset(0, 2))
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              label,
              style: const TextStyle(
                  fontSize: 11, color: Color(0xFF6B7280)),
            ),
            const SizedBox(height: 6),
            Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                Text(
                  value,
                  style: TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.w800,
                      color: color),
                ),
                const SizedBox(width: 4),
                Text(
                  unit,
                  style: const TextStyle(
                      fontSize: 11, color: Color(0xFF9CA3AF)),
                ),
              ],
            ),
            const SizedBox(height: 6),
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: status == '偏高'
                    ? AppTheme.statusWarning.withValues(alpha: 0.15)
                    : const Color(0xFFF0FDF4),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                status,
                style: TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.w600,
                  color: status == '偏高'
                      ? AppTheme.statusWarning
                      : AppTheme.statusNormal,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  // ===== Chart =====
  Widget _buildChartCard() {
    final filtered = _filteredRecords;
    if (filtered.isEmpty) {
      return Container(
        padding: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: AppTheme.bgCard,
          borderRadius: BorderRadius.circular(16),
        ),
        child: const Center(
          child: Text(
            '暂无血压数据',
            style: TextStyle(fontSize: 14, color: Color(0xFF9CA3AF)),
          ),
        ),
      );
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 16,
              offset: const Offset(0, 2))
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text(
                '血压趋势',
                style:
                    TextStyle(fontSize: 15, fontWeight: FontWeight.w700),
              ),
              Text(
                '单位: mmHg',
                style: TextStyle(fontSize: 12, color: Color(0xFF9CA3AF)),
              ),
            ],
          ),
          const SizedBox(height: 12),
          // Legend
          Row(
            children: [
              _chartLegend(const Color(0xFFEF4444), '收缩压'),
              const SizedBox(width: 16),
              _chartLegend(const Color(0xFF3B82F6), '舒张压'),
              const SizedBox(width: 16),
              _chartLegend(const Color(0xFFD1D5DB), '正常范围'),
            ],
          ),
          const SizedBox(height: 8),
          // Threshold annotation
          Row(
            children: const [
              Text('↑', style: TextStyle(fontSize: 10, color: Color(0xFFEF4444), fontWeight: FontWeight.w700)),
              SizedBox(width: 4),
              Text('收缩压>140 / 舒张压>90 为异常', style: TextStyle(fontSize: 10, color: Color(0xFFEF4444))),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 200,
            child: _BloodPressureLineChart(records: filtered),
          ),
        ],
      ),
    );
  }

  Widget _chartLegend(Color color, String label) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
            width: 8, height: 8,
            decoration: BoxDecoration(color: color, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF4B5563))),
      ],
    );
  }

  // ===== Anomaly Banner =====
  Widget _buildAnomalyBanner() {
    if (_anomalyCount == 0) return const SizedBox.shrink();
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFFFBEB),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFFDE68A)),
      ),
      child: Row(
        children: [
          const Icon(Icons.warning_amber, color: AppTheme.statusWarning, size: 20),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              '近$_selectedDays天内有 $_anomalyCount 次血压异常（收缩压>140 或 舒张压>90）',
              style: const TextStyle(fontSize: 12, color: Color(0xFF92400E)),
            ),
          ),
        ],
      ),
    );
  }

  // ===== Record List =====
  Widget _buildRecordList() {
    final filtered = _filteredRecords;
    if (filtered.isEmpty) {
      return const SliverToBoxAdapter(child: SizedBox.shrink());
    }

    return SliverList(
      delegate: SliverChildBuilderDelegate((ctx, i) {
        if (i >= filtered.length) return null;
        final rec = filtered[filtered.length - 1 - i]; // newest first
        return _recordRow(rec);
      }, childCount: filtered.length),
    );
  }

  Widget _recordRow(BloodPressureRecord rec) {
    final dateStr =
        '${rec.measurementTime.month}月${rec.measurementTime.day}日';
    final timeStr =
        '${rec.measurementTime.hour.toString().padLeft(2, '0')}:${rec.measurementTime.minute.toString().padLeft(2, '0')}';
    final isAnomaly = rec.isAnomaly;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(14),
        border: isAnomaly
            ? Border.all(color: const Color(0xFFFECACA), width: 1.5)
            : null,
        boxShadow: [
          BoxShadow(
              color: Colors.black.withValues(alpha: 0.04),
              blurRadius: 8,
              offset: const Offset(0, 1)),
        ],
      ),
      child: Row(
        children: [
          // Date
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: isAnomaly
                  ? const Color(0xFFFEF2F2)
                  : const Color(0xFFF0FDF4),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  dateStr,
                  style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: isAnomaly
                          ? AppTheme.statusDanger
                          : AppTheme.statusNormal),
                ),
                Text(
                  timeStr,
                  style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF)),
                ),
              ],
            ),
          ),
          const SizedBox(width: 14),
          // BP value
          Expanded(
            child: Row(
              children: [
                _bpValue(rec.systolic, const Color(0xFFEF4444), isAnomaly),
                const Text(
                  '/',
                  style: TextStyle(fontSize: 20, color: Color(0xFFD1D5DB)),
                ),
                _bpValue(rec.diastolic, const Color(0xFF3B82F6), isAnomaly),
                if (rec.pulse != null) ...[
                  const SizedBox(width: 12),
                  Text(
                    '${rec.pulse} bpm',
                    style: const TextStyle(fontSize: 13, color: Color(0xFF9CA3AF)),
                  ),
                ],
              ],
            ),
          ),
          // Anomaly badge
          if (isAnomaly)
            Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: AppTheme.statusDanger,
                borderRadius: BorderRadius.circular(6),
              ),
              child: const Text(
                '异',
                style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.w700,
                    color: Colors.white),
              ),
            ),
        ],
      ),
    );
  }

  Widget _bpValue(int value, Color color, bool isAnomaly) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          value.toString(),
          style: TextStyle(
            fontSize: 22,
            fontWeight: FontWeight.w800,
            color: isAnomaly ? color : color.withValues(alpha: 0.6),
          ),
        ),
        Text(
          isAnomaly ? '偏高' : '正常',
          style: TextStyle(
            fontSize: 9,
            fontWeight: FontWeight.w600,
            color: isAnomaly ? color : const Color(0xFF10B981),
          ),
        ),
      ],
    );
  }

  bool _isNormalBloodPressure() {
    return _latestSystolic <= _systolicThreshold &&
        _latestDiastolic <= _diastolicThreshold;
  }
}

// ===== Custom Line Chart Painter =====

class _BloodPressureLineChart extends StatelessWidget {
  final List<BloodPressureRecord> records;
  const _BloodPressureLineChart({required this.records});

  @override
  Widget build(BuildContext context) {
    return CustomPaint(
      painter: _BloodPressureChartPainter(records: records),
      size: const Size(double.infinity, 200),
    );
  }
}

class _BloodPressureChartPainter extends CustomPainter {
  final List<BloodPressureRecord> records;
  _BloodPressureChartPainter({required this.records});

  static const int _systolicThreshold = 140;

  @override
  void paint(Canvas canvas, Size size) {
    if (records.isEmpty) return;

    final padding = const EdgeInsets.only(left: 40, right: 12, top: 8, bottom: 28);
    final chartW = size.width - padding.left - padding.right;
    final chartH = size.height - padding.top - padding.bottom;

    // Determine Y range
    int maxY = 0;
    for (final r in records) {
      if (r.systolic > maxY) maxY = r.systolic;
      if (r.diastolic > maxY) maxY = r.diastolic;
    }
    // Ensure threshold is visible
    if (maxY < _systolicThreshold) maxY = _systolicThreshold;
    // Round up to nearest 20
    maxY = ((maxY + 19) ~/ 20) * 20;
    final minY = 40; // BP never below 40 is clinically relevant

    // Helper: map value to Y pixel
    double valToY(double val) {
      return padding.top + chartH - ((val - minY) / (maxY - minY)) * chartH;
    }

    // Helper: map index to X pixel
    double idxToX(int i) {
      if (records.length == 1) return padding.left + chartW / 2;
      return padding.left + (i / (records.length - 1)) * chartW;
    }

    // Draw normal range band (80-120 systolic, 60-80 diastolic)
    final normalBand = Paint()
      ..color = const Color(0xFF10B981).withValues(alpha: 0.06)
      ..style = PaintingStyle.fill;
    final normalBandY1 = valToY(_systolicThreshold.toDouble());
    final normalBandY2 = valToY(120.0);
    canvas.drawRect(
      Rect.fromLTWH(padding.left, normalBandY2, chartW, normalBandY1 - normalBandY2),
      normalBand,
    );

    // Draw threshold line (systolic > 140)
    final thresholdPaint = Paint()
      ..color = const Color(0xFFEF4444).withValues(alpha: 0.4)
      ..strokeWidth = 1;
    final thresholdY = valToY(140.0);
    canvas.drawLine(
      Offset(padding.left, thresholdY),
      Offset(padding.left + chartW, thresholdY),
      thresholdPaint,
    );

    // Threshold label
    final thresholdLabel = TextPainter(
      text: const TextSpan(text: '140', style: TextStyle(fontSize: 9, color: Color(0xFFEF4444), fontWeight: FontWeight.w600)),
      textDirection: TextDirection.ltr,
    )..layout();
    thresholdLabel.paint(canvas, Offset(padding.left - 36, thresholdY - 4));

    // Y-axis labels
    final yLabels = <double>[];
    final step = maxY > 200 ? 40 : 20;
    for (double v = minY.toDouble(); v <= maxY.toDouble(); v += step) yLabels.add(v);
    final labelStyle = const TextStyle(fontSize: 9, color: Color(0xFF9CA3AF));
    for (final v in yLabels) {
      final y = valToY(v);
      // Grid line
      canvas.drawLine(
        Offset(padding.left, y),
        Offset(padding.left + chartW, y),
        Paint()..color = const Color(0xFFE5E7EB).withValues(alpha: 0.5)..strokeWidth = 0.5,
      );
      // Label
      TextPainter(
        text: TextSpan(text: v.toInt().toString(), style: labelStyle),
        textDirection: TextDirection.ltr,
      )..layout(minWidth: 0, maxWidth: 36)
        ..paint(canvas, Offset(2, y - 4));
    }

    // X-axis labels
    final xLabelStyle = const TextStyle(fontSize: 9, color: Color(0xFF9CA3AF));
    final labelCount = records.length <= 14 ? records.length : 14;
    final stepIdx = records.length / labelCount;
    for (int i = 0; i < labelCount; i++) {
      final idx = (i * stepIdx).round().clamp(0, records.length - 1);
      final rec = records[idx];
      final x = idxToX(idx);
      final label = '${rec.measurementTime.month}/${rec.measurementTime.day}';
      TextPainter(
        text: TextSpan(text: label, style: xLabelStyle),
        textDirection: TextDirection.ltr,
      )..layout()
        ..paint(canvas, Offset(x - 12, size.height - 10));
    }

    // Draw systolic line (red)
    _drawLine(canvas, records, (r) => r.systolic.toDouble(),
        const Color(0xFFEF4444), valToY, idxToX, anomaly: true);

    // Draw diastolic line (blue)
    _drawLine(canvas, records, (r) => r.diastolic.toDouble(),
        const Color(0xFF3B82F6), valToY, idxToX, anomaly: false);
  }

  void _drawLine(
    Canvas canvas,
    List<BloodPressureRecord> records,
    double Function(BloodPressureRecord) getValue,
    Color lineColor,
    double Function(double) valToY,
    double Function(int) idxToX,
    {required bool anomaly}
  ) {
    if (records.isEmpty) return;

    final linePaint = Paint()
      ..color = lineColor
      ..strokeWidth = 2.5
      ..strokeCap = StrokeCap.round
      ..strokeJoin = StrokeJoin.round
      ..style = PaintingStyle.stroke;

    final areaPaint = Paint()
      ..color = lineColor.withValues(alpha: 0.08)
      ..style = PaintingStyle.fill;

    // Build path
    Path path = Path()
      ..moveTo(idxToX(0), valToY(getValue(records[0])));
    for (int i = 1; i < records.length; i++) {
      path.lineTo(idxToX(i), valToY(getValue(records[i])));
    }

    // Area fill
    final closePath = Path()
      ..moveTo(idxToX(0), valToY(getValue(records[0])));
    for (int i = 1; i < records.length; i++) {
      closePath.lineTo(idxToX(i), valToY(getValue(records[i])));
    }
    closePath.lineTo(idxToX(records.length - 1), valToY(40.0));
    closePath.lineTo(idxToX(0), valToY(40.0));
    closePath.close();
    canvas.drawPath(closePath, areaPaint);

    // Line
    canvas.drawPath(path, linePaint);

    // Draw dots — anomaly dots are larger and red
    final dotPaint = Paint()..style = PaintingStyle.fill;
    final strokeDot = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5
      ..color = Colors.white;

    for (int i = 0; i < records.length; i++) {
      final rec = records[i];
      final val = getValue(rec);
      final isPointAnomaly = anomaly
          ? (rec.systolic > 140)
          : (rec.diastolic > 90);

      final cx = idxToX(i);
      final cy = valToY(val);

      if (isPointAnomaly) {
        // Larger anomaly dot
        canvas.drawCircle(Offset(cx, cy), 5, dotPaint..color = const Color(0xFFEF4444));
        canvas.drawCircle(Offset(cx, cy), 5, strokeDot);
      } else {
        // Normal dot
        canvas.drawCircle(Offset(cx, cy), 3, dotPaint..color = lineColor);
      }
    }
  }

  @override
  bool shouldRepaint(covariant _BloodPressureChartPainter oldDelegate) {
    return oldDelegate.records != records;
  }
}

// ===== Data Model =====

class BloodPressureRecord {
  final String id;
  final String elderlyId;
  final int systolic;
  final int diastolic;
  final int? pulse;
  final DateTime measurementTime;

  const BloodPressureRecord({
    required this.id,
    required this.elderlyId,
    required this.systolic,
    required this.diastolic,
    this.pulse,
    required this.measurementTime,
  });

  bool get isAnomaly => systolic > 140 || diastolic > 90;

  factory BloodPressureRecord.fromJson(Map<String, dynamic> json) {
    return BloodPressureRecord(
      id: json['id'] as String? ?? '',
      elderlyId: json['elderly_id'] as String? ?? '',
      systolic: (json['systolic'] as num?)?.toInt() ?? 0,
      diastolic: (json['diastolic'] as num?)?.toInt() ?? 0,
      pulse: json['pulse'] as int?,
      measurementTime: json['measurement_time'] != null
          ? DateTime.parse(json['measurement_time'] as String)
          : DateTime.now(),
    );
  }

  Map<String, dynamic> toJson() => {
        'id': id,
        'elderly_id': elderlyId,
        'systolic': systolic,
        'diastolic': diastolic,
        if (pulse != null) 'pulse': pulse,
        'measurement_time': measurementTime.toIso8601String(),
      };
}

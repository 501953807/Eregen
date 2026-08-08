import 'package:dio/dio.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../api/client.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';

/// Uric acid detail page — line chart with normal-range banding,
/// daily measurement records list, and current reading summary.
class UricAcidPage extends StatefulWidget {
  final String elderlyId;
  final String elderlyName;
  const UricAcidPage({super.key, required this.elderlyId, required this.elderlyName});

  @override
  State<UricAcidPage> createState() => _UricAcidPageState();
}

class _UricAcidPageState extends State<UricAcidPage> {
  int _selectedIndex = 4;
  bool _loading = true;
  String? _error;

  // Normal range: 143–416 μmol/L (male), 89–357 μmol/L (female)
  // Using conservative low: 143, high: 416
  static const double _normalLow = 143.0;
  static const double _normalHigh = 416.0;
  static const double _chartMinY = 0.0;
  static const double _chartMaxY = 600.0;
  static const String _unit = 'μmol/L';

  double? _latestValue;
  DateTime? _latestTime;

  final List<FlSpot> _spots = [];
  final List<_Record> _records = [];

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    setState(() { _loading = true; _error = null; });
    try {
      final resp = await ApiClient.instance.get(
        '/api/v1/chronic/${widget.elderlyId}/uric-acid',
        query: {'days': 30},
      );
      final data = resp.data as Map<String, dynamic>;
      final List<dynamic> items = data['data'] ?? data['records'] ?? [];

      setState(() {
        _spots.clear();
        _records.clear();
        for (int i = 0; i < items.length; i++) {
          final Map<String, dynamic> r = items[i] as Map<String, dynamic>;
          final ts = DateTime.parse(r['timestamp'] as String).millisecondsSinceEpoch / 1000;
          final val = double.tryParse((r['value'] ?? r['uric_acid'] ?? '0').toString()) ?? 0.0;
          _spots.add(FlSpot(ts, val));
          _records.add(_Record(
            time: r['timestamp'] as String? ?? '',
            value: val,
            type: (r['type'] ?? r['label'] ?? '测量').toString(),
            note: r['note']?.toString(),
          ));
        }
        if (_records.isNotEmpty) {
          _latestValue = _records.first.value;
          _latestTime = DateTime.tryParse(_records.first.time);
        }
        _loading = false;
      });
    } on DioException catch (e) {
      setState(() {
        _loading = false;
        _error = e.response?.data?['message'] ?? e.message ?? '加载失败';
      });
    } catch (e) {
      setState(() { _loading = false; _error = '加载失败，请稍后重试'; });
    }
  }

  Color _valueColor(double v) {
    if (v > _normalHigh) return AppTheme.statusDanger;
    if (v < _normalLow) return AppTheme.statusInfo;
    return AppTheme.statusNormal;
  }

  String _valueLabel(double v) {
    if (v > _normalHigh) return '偏高';
    if (v < _normalLow) return '偏低';
    return '正常';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      appBar: AppBar(
        title: Text('${widget.elderlyName} · 尿酸详情'),
        backgroundColor: Colors.white,
        foregroundColor: const Color(0xFF1F2937),
        elevation: 0,
        bottom: const PreferredSize(preferredSize: Size.fromHeight(1), child: Divider(height: 1, color: Color(0xFFF0F0F5))),
        iconTheme: const IconThemeData(color: Color(0xFF374151)),
      ),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _selectedIndex,
        onTabSelected: (i) => setState(() => _selectedIndex = i),
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(_error!, style: const TextStyle(color: Color(0xFF6B7280))),
                      const SizedBox(height: 12),
                      ElevatedButton(onPressed: _fetchData, child: const Text('重试')),
                    ],
                  ),
                )
              : CustomScrollView(
                  slivers: [
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(16, 16, 16, 0),
                        child: _buildSummaryCard(),
                      ),
                    ),
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(16, 12, 16, 0),
                        child: _buildChart(),
                      ),
                    ),
                    SliverToBoxAdapter(
                      child: Padding(
                        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
                        child: const Text(
                          '记录列表',
                          style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937)),
                        ),
                      ),
                    ),
                    _records.isEmpty
                        ? const SliverToBoxAdapter(child: Center(child: Text('暂无记录', style: TextStyle(color: Color(0xFF9CA3AF)))))
                        : SliverList(
                            delegate: SliverChildBuilderDelegate(
                              (ctx, i) => _RecordTile(record: _records[i]),
                              childCount: _records.length,
                            ),
                          ),
                    const SliverToBoxAdapter(child: SizedBox(height: 24)),
                  ],
                ),
    );
  }

  Widget _buildSummaryCard() {
    if (_latestValue == null) return const SizedBox.shrink();
    final color = _valueColor(_latestValue!);
    final timeStr = _latestTime != null ? DateFormat('MM-dd HH:mm').format(_latestTime!) : '';
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 12, offset: const Offset(0, 2))],
      ),
      child: Row(
        children: [
          Container(
            width: 56,
            height: 56,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(16)),
            child: const Center(child: Text('\u{1F52C}', style: TextStyle(fontSize: 26))),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('当前尿酸', style: TextStyle(fontSize: 13, color: const Color(0xFF6B7280))),
                const SizedBox(height: 4),
                RichText(
                  text: TextSpan(
                    style: const TextStyle(fontSize: 28, fontWeight: FontWeight.w800, color: Color(0xFF1F2937)),
                    children: [
                      TextSpan(text: _latestValue!.toStringAsFixed(0)),
                      TextSpan(text: ' $_unit', style: const TextStyle(fontSize: 13, fontWeight: FontWeight.normal, color: Color(0xFF9CA3AF))),
                    ],
                  ),
                ),
                const SizedBox(height: 4),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                  decoration: BoxDecoration(color: color.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(8)),
                  child: Text(
                    _valueLabel(_latestValue!),
                    style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: color),
                  ),
                ),
              ],
            ),
          ),
          Text(timeStr, style: const TextStyle(fontSize: 12, color: Color(0xFF9CA3AF))),
        ],
      ),
    );
  }

  Widget _buildChart() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 12, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('尿酸趋势（近30天）', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
              Text('单位: μmol/L', style: TextStyle(fontSize: 12, color: Color(0xFF9CA3AF))),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: const [
              _LegendChip(color: AppTheme.statusNormal, label: '正常范围'),
              SizedBox(width: 12),
              _LegendChip(color: AppTheme.statusDanger, label: '偏高'),
              SizedBox(width: 12),
              _LegendChip(color: AppTheme.statusInfo, label: '偏低'),
            ],
          ),
          const SizedBox(height: 16),
          SizedBox(
            height: 220,
            child: _spots.isEmpty
                ? const Center(child: Text('暂无数据', style: TextStyle(color: Color(0xFF9CA3AF))))
                : LineChart(_chartData()),
          ),
        ],
      ),
    );
  }

  LineChartData _chartData() {
    final minY = _chartMinY;
    final maxY = _chartMaxY;
    // Y axis labels: 0, 100, 200, 300, 400, 500, 600
    final yLabels = ['600', '500', '400', '300', '200', '100', '0'];
    final yLabelCount = yLabels.length;
    final yStep = (maxY - minY) / (yLabelCount - 1);

    return LineChartData(
      minX: _spots.isEmpty ? 0 : _spots.map((s) => s.x).reduce((a, b) => a < b ? a : b),
      maxX: _spots.isEmpty ? 1 : _spots.map((s) => s.x).reduce((a, b) => a > b ? a : b),
      minY: minY,
      maxY: maxY,
      gridData: FlGridData(
        show: true,
        drawVerticalLine: false,
        horizontalInterval: yStep,
        getDrawingHorizontalLine: (_) => FlLine(color: const Color(0xFFF3F4F6), strokeWidth: 1),
        getDrawingVerticalLine: (_) => const FlLine(color: Colors.transparent),
      ),
      titlesData: FlTitlesData(
        leftTitles: AxisTitles(
          sideTitles: SideTitles(
            showTitles: true,
            interval: yStep,
            reservedSize: 36,
            getTitlesWidget: (double v, TitleMeta meta) {
              final idx = ((v - minY) / yStep).round();
              if (idx < 0 || idx >= yLabels.length) return const Text('');
              return Text(yLabels[idx], style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF)));
            },
          ),
        ),
        bottomTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
        topTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
        rightTitles: const AxisTitles(sideTitles: SideTitles(showTitles: false)),
      ),
      borderData: FlBorderData(show: false),
      lineBarsData: [
        LineChartBarData(
          spots: _spots,
          isCurved: true,
          curveSmoothness: 0.15,
          color: const Color(0xFF7C3AED),
          barWidth: 2.5,
          dotData: const FlDotData(show: false),
          belowBarData: BarAreaData(
            show: true,
            color: const Color(0xFF7C3AED).withValues(alpha: 0.08),
          ),
        ),
      ],
    );
  }
}

class _Record {
  final String time;
  final double value;
  final String type;
  final String? note;
  const _Record({required this.time, required this.value, required this.type, this.note});
}

class _RecordTile extends StatelessWidget {
  final _Record record;
  const _RecordTile({required this.record});

  @override
  Widget build(BuildContext context) {
    final color = record.value > _UricAcidPageState._normalHigh
        ? AppTheme.statusDanger
        : (record.value < _UricAcidPageState._normalLow
            ? AppTheme.statusInfo
            : AppTheme.statusNormal);
    final timeStr = record.time.isNotEmpty
        ? DateFormat('MM-dd HH:mm').format(DateTime.parse(record.time))
        : '';
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8, offset: const Offset(0, 1))],
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: color.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(12)),
            child: Center(child: Text('\u{1F52C}', style: const TextStyle(fontSize: 18))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(record.type, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
                    const SizedBox(width: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(color: color.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
                      child: Text(
                        record.value < _UricAcidPageState._normalLow ? '偏低' : (record.value > _UricAcidPageState._normalHigh ? '偏高' : '正常'),
                        style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: color),
                      ),
                    ),
                  ],
                ),
                if (record.note != null && record.note!.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(record.note!, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                ],
              ],
            ),
          ),
          Text(
            '${record.value.toStringAsFixed(0)}',
            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: color),
          ),
          const SizedBox(width: 10),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(timeStr, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
              const Text('μmol/L', style: TextStyle(fontSize: 9, color: Color(0xFFD1D5DB))),
            ],
          ),
        ],
      ),
    );
  }
}

class _LegendChip extends StatelessWidget {
  final Color color;
  final String label;
  const _LegendChip({required this.color, required this.label});

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 8, height: 8, decoration: const BoxDecoration(color: Colors.blue, shape: BoxShape.circle)),
        const SizedBox(width: 4),
        Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
      ],
    );
  }
}

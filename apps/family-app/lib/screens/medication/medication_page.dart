import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../../models/medication.dart';

/// Medication management page — fetches rules from GET /medication/rules, confirms taken via POST /medication/:rule_id/take.
class MedicationPage extends StatefulWidget {
  const MedicationPage({super.key});

  @override
  State<MedicationPage> createState() => _MedicationPageState();
}

class _MedicationPageState extends State<MedicationPage> {
  int _selectedIndex = 3;
  bool _loading = true;
  List<MedicationRule> _rules = [];
  Set<String> _takenIds = {}; // track locally taken rule IDs

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  Future<void> _fetchData() async {
    try {
      final resp = await ApiClient.instance.get('/medication/rules');
      final list = resp.data as List;
      setState(() {
        _rules = list.map((r) => MedicationRule.fromJson(r as Map<String, dynamic>)).toList();
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  Future<void> _confirmTaken(String ruleId) async {
    try {
      await ApiClient.instance.post('/medication/$ruleId/take');
      setState(() => _takenIds.add(ruleId));
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: const Text('已确认服用'), duration: const Duration(seconds: 1)));
    } catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(const SnackBar(content: Text('操作失败'), duration: Duration(seconds: 1)));
    }
  }

  // Derive today's schedule from fetched rules
  List<MedicationRule> get _todayRules => _rules.where((r) => r.active).toList()
    ..sort((a, b) => a.scheduleTime.compareTo(b.scheduleTime));

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : CustomScrollView(slivers: [
                // Header
                SliverToBoxAdapter(
                  child: Container(padding: const EdgeInsets.fromLTRB(20, 12, 20, 12), color: AppTheme.bgCard, child: Row(children: [
                    IconButton(icon: const Icon(Icons.arrow_back_ios_new, size: 18), onPressed: () {}),
                    const Expanded(child: Text('用药管理', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
                    IconButton(icon: const Icon(Icons.settings_outlined), onPressed: () {}),
                  ])),
                ),

                // Adherence overview card
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.symmetric(horizontal: 16), child: Container(
                    decoration: BoxDecoration(gradient: const LinearGradient(colors: [Color(0xFF66BB6A), Color(0xFF43A047)]), borderRadius: BorderRadius.circular(16)),
                    padding: const EdgeInsets.all(20),
                    child: Column(children: [
                      const Text('今日服药进度', style: TextStyle(fontSize: 13, color: Colors.white)),
                      const SizedBox(height: 12),
                      Stack(alignment: Alignment.center, children: [
                        SizedBox(width: 110, height: 110, child: CircularProgressIndicator(value: _todayRules.isEmpty ? 0 : _takenIds.length / _todayRules.length.toDouble(), strokeWidth: 10, backgroundColor: Colors.white.withValues(alpha: 0.2), valueColor: const AlwaysStoppedAnimation<Color>(Colors.white))),
                        Column(mainAxisSize: MainAxisSize.min, children: [
                          Text('${_todayRules.isEmpty ? 0 : _takenIds.length}/${_todayRules.length}', style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w800, color: Colors.white)),
                          Text('次', style: TextStyle(fontSize: 12, color: Colors.white.withValues(alpha: 0.8))),
                        ]),
                      ]),
                      const SizedBox(height: 10),
                      Text(_todayRules.isEmpty ? '暂无用药规则' : '已服 ${_takenIds.length} / ${_todayRules.length} 次', style: const TextStyle(fontSize: 12, color: Colors.white)),
                    ]),
                  )),
                ),

                // Today's schedule section title
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.only(left: 20, right: 20, bottom: 12), child: Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                    const Text('今日用药计划', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                    GestureDetector(onTap: () {}, child: const Text('+ 添加', style: TextStyle(fontSize: 12, color: AppTheme.primary, fontWeight: FontWeight.w600))),
                  ])),
                ),

                if (_todayRules.isEmpty)
                  const SliverToBoxAdapter(child: Center(child: Text('暂无用药规则，请先添加', style: TextStyle(color: Color(0xFFBBBBBB)))))
                else
                  SliverList(delegate: SliverChildBuilderDelegate((ctx, i) {
                    if (i >= _todayRules.length) return null;
                    final rule = _todayRules[i];
                    final isTaken = _takenIds.contains(rule.id);
                    return _timeSlot(
                      time: rule.scheduleTime.substring(0, 5),
                      period: _periodFromTime(rule.scheduleTime),
                      dotColor: isTaken ? AppTheme.statusNormal : (DateTime.now().isAfter(DateTime.parse('2024-01-01T${rule.scheduleTime}')) ? const Color(0xFFFF6B6B) : const Color(0xFFFFA726)),
                      dotBg: isTaken ? const Color(0xFFE8F5E9) : (isTaken ? const Color(0xFFE8F5E9) : const Color(0xFFFFF8E1)),
                      lineColor: isTaken ? AppTheme.statusNormal : null,
                      borderColor: isTaken ? AppTheme.statusNormal : (isTaken ? const Color(0xFFFF6B6B) : const Color(0xFFFFA726)),
                      medNames: rule.pillType,
                      dosage: '每次 ${rule.doseCount} 粒',
                      status: isTaken ? '✓ 已服用' : '⏳ 待服用',
                      statusColor: isTaken ? const Color(0xFF2E7D32) : const Color(0xFFEF6C00),
                      statusBg: isTaken ? const Color(0xFFE8F5E9) : const Color(0xFFFFF3E0),
                      note: isTaken ? null : '🔔 智能药盒将语音播报提醒',
                      onConfirm: isTaken ? null : () => _confirmTaken(rule.id),
                    );
                  }, childCount: _todayRules.length)),

                // Inventory tracking section
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.only(left: 20, right: 20, top: 16, bottom: 12), child: const Text('药品库存', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700))),
                ),
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.symmetric(horizontal: 20), child: Column(children: [
                    _inventoryBar('降压药', 45, 60, const Color(0xFF4CAF50)),
                    const SizedBox(height: 8),
                    _inventoryBar('维生素D', 12, 30, const Color(0xFFFFA726)),
                    const SizedBox(height: 8),
                    _inventoryBar('钙片', 3, 60, const Color(0xFFFF6B6B)),
                    const SizedBox(height: 8),
                    _inventoryBar('降糖药', 0, 30, const Color(0xFFFF6B6B)),
                  ])),
                ),
                const SliverToBoxAdapter(child: SizedBox(height: 16)),

                // Adherence heatmap
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.only(left: 20, right: 20, bottom: 12), child: const Text('服药依从性', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700))),
                ),
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.symmetric(horizontal: 20), child: Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14), border: Border.all(color: const Color(0xFFF0F0F0))),
                    child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                      const Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                        Text('近7日服药记录', style: TextStyle(fontSize: 12, color: Color(0xFF999999))),
                        Text('3/4 已服', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Color(0xFF4CAF50))),
                      ]),
                      const SizedBox(height: 12),
                      Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                        ...'一二三四五六日'.split('').map((d) => SizedBox(width: 38, child: Center(child: Text(d, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA)))))),
                      ]),
                      const SizedBox(height: 6),
                      Row(mainAxisAlignment: MainAxisAlignment.spaceBetween, children: [
                        ...List.generate(7, (i) => _heatCell(i % 3 == 0 ? 'missed' : (i % 3 == 1 ? 'partial' : 'full'))),
                      ]),
                    ]),
                  )),
                ),

                // Remote config section placeholder
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.only(left: 20, right: 20, top: 16, bottom: 12), child: const Text('远程配置', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700))),
                ),
                SliverToBoxAdapter(
                  child: Padding(padding: const EdgeInsets.symmetric(horizontal: 20), child: Container(padding: const EdgeInsets.all(16), decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.02), blurRadius: 6)]), child: const Column(children: [Text('配置项将在后续版本开放', style: TextStyle(fontSize: 13, color: Color(0xFF999999)))]))),
                ),
                const SliverToBoxAdapter(child: SizedBox(height: 24)),
              ]),
      ),
      bottomNavigationBar: BottomNavBar(
        selectedTab: _selectedIndex,
        onTabSelected: (i) => setState(() => _selectedIndex = i),
      ),
    );
  }

  String _periodFromTime(String time) {
    final hour = int.tryParse(time.split(':').first) ?? 0;
    if (hour < 10) return '上午';
    if (hour < 13) return '中午';
    if (hour < 17) return '下午';
    if (hour < 21) return '晚上';
    return '睡前';
  }

  Widget _timeSlot({
    required String time, required String period, required Color dotColor, required Color dotBg,
    Color? lineColor, required Color borderColor, required String medNames, required String dosage,
    required String status, required Color statusColor, required Color statusBg, String? note, VoidCallback? onConfirm,
  }) {
    final hasLine = lineColor != null;
    return Padding(padding: const EdgeInsets.only(bottom: 0), child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
      Container(width: 56, padding: const EdgeInsets.only(top: 14), child: Column(children: [Text(time, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)), Text(period, style: const TextStyle(fontSize: 10, color: Color(0xFF999999)))])),
      const SizedBox(width: 14),
      Container(width: 2, height: 100, margin: const EdgeInsets.only(top: 0), child: Stack(children: [
        Positioned(top: 0, child: Container(width: 14, height: 14, decoration: BoxDecoration(color: dotBg, shape: BoxShape.circle, border: Border.all(color: dotColor, width: 3)))),
        if (hasLine) Positioned(top: 14, left: -6, child: Container(width: 2, height: 86, color: lineColor)),
      ])),
      const SizedBox(width: 14),
      Expanded(child: Container(padding: const EdgeInsets.all(12), decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(12), border: Border(left: BorderSide(color: borderColor, width: 3)), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.02), blurRadius: 4)]), child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
        Text(medNames, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
        Text(dosage, style: const TextStyle(fontSize: 11, color: Color(0xFF888888))),
        const SizedBox(height: 6),
        Container(padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2), decoration: BoxDecoration(color: statusBg, borderRadius: BorderRadius.circular(8)), child: Text(status, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: statusColor))),
        if (note != null) Text(note, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA))),
        if (onConfirm != null) ...[
          const SizedBox(height: 8),
          SizedBox(width: double.infinity, height: 32, child: ElevatedButton(onPressed: onConfirm, style: ElevatedButton.styleFrom(backgroundColor: AppTheme.statusNormal, foregroundColor: Colors.white, padding: EdgeInsets.zero, shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8))), child: const Text('确认服用', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600)))),
        ],
      ]))),
    ]));
  }

  /// Inventory stock bar — v2 prototype enhancement
  Widget _inventoryBar(String name, int remaining, int total, Color levelColor) {
    final pct = total > 0 ? remaining / total : 0.0;
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFFF0F0F0)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(name, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
              Text('$remaining / $total', style: const TextStyle(fontSize: 12, color: Color(0xFF888888))),
            ],
          ),
          const SizedBox(height: 8),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: pct,
              minHeight: 6,
              backgroundColor: const Color(0xFFF0F0F5),
              valueColor: AlwaysStoppedAnimation<Color>(levelColor),
            ),
          ),
          const SizedBox(height: 4),
          Text(
            pct > 0.5 ? '库存充足' : (pct > 0 ? '库存偏低，建议补货' : '库存耗尽，请及时购买'),
            style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: levelColor),
          ),
        ],
      ),
    );
  }

  /// Heatmap cell for adherence calendar
  Widget _heatCell(String state) {
    final colors = {'full': const Color(0xFF4CAF50), 'partial': const Color(0xFFFFA726), 'missed': const Color(0xFFFF6B6B)};
    final size = 12.0;
    return SizedBox(
      width: size,
      height: size,
      child: Container(
        decoration: BoxDecoration(
          color: colors[state]!.withValues(alpha: 0.3),
          borderRadius: BorderRadius.circular(3),
          border: Border.all(color: colors[state]!, width: 1),
        ),
      ),
    );
  }
}

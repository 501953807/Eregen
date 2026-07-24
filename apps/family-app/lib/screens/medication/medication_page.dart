import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';
import '../../models/medication.dart';

/// Medication management — v2 design: today summary card with ring, period tabs,
/// timeline with icons/status badges/actions, inventory bars, remote config toggles,
/// adherence heatmap calendar, dark mode.
class MedicationPage extends StatefulWidget {
  const MedicationPage({super.key});

  @override
  State<MedicationPage> createState() => _MedicationPageState();
}

class _MedicationPageState extends State<MedicationPage> {
  int _selectedIndex = 3;
  bool _loading = true;
  List<MedicationRule> _rules = [];
  Set<String> _takenIds = {};
  String _activePeriod = '上午';
  bool _darkMode = false;
  bool _showSuccess = false;
  String? _confirmMedName;
  String? _confirmDose;

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

  Future<void> _confirmTaken(String ruleId, String medName, String dose) async {
    setState(() {
      _confirmMedName = medName;
      _confirmDose = dose;
      _showSuccess = true;
    });
    try {
      await ApiClient.instance.post('/medication/$ruleId/take');
      setState(() => _takenIds.add(ruleId));
    } catch (_) {}
    Future.delayed(const Duration(milliseconds: 1200), () {
      setState(() {
        _showSuccess = false;
        _confirmMedName = null;
        _confirmDose = null;
      });
    });
  }

  List<MedicationRule> get _todayRules => _rules.where((r) => r.active).toList()
    ..sort((a, b) => a.scheduleTime.compareTo(b.scheduleTime));

  List<MedicationRule> get _filteredRules => _todayRules.where((r) {
    final period = _periodFromTime(r.scheduleTime);
    return period == _activePeriod;
  }).toList();

  int get _todayTotal => _todayRules.length;
  int get _todayTaken => _takenIds.length;
  double get _adherenceRate => _todayTotal > 0 ? _todayTaken / _todayTotal : 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: _darkMode ? const Color(0xFF111827) : const Color(0xFFF3F4F6),
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator())
            : Stack(
          children: [
            CustomScrollView(
              slivers: [
                _buildTopBar(),
                SliverToBoxAdapter(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        _buildTodaySummaryCard(),
                        const SizedBox(height: 12),
                        _buildPeriodTabs(),
                        const SizedBox(height: 12),
                        ..._filteredRules.map((rule) => Padding(
                          padding: const EdgeInsets.only(bottom: 8),
                          child: _buildMedItem(rule),
                        )),
                        if (_filteredRules.isEmpty) const Center(child: Text('该时段暂无用药', style: TextStyle(color: Color(0xFF9CA3AF)))),
                        const SizedBox(height: 16),
                        _buildInventoryCard(),
                        const SizedBox(height: 12),
                        _buildRemoteConfig(),
                        const SizedBox(height: 12),
                        _buildAdherenceHistory(),
                      ],
                    ),
                  ),
                ),
              ],
            ),
            // Success overlay
            if (_showSuccess)
              Container(
                color: AppTheme.statusNormal.withValues(alpha: 0.1),
                child: Center(
                  child: Container(
                    width: 80, height: 80, decoration: BoxDecoration(color: AppTheme.statusNormal, shape: BoxShape.circle),
                    child: const Center(child: Text('\u{2713}', style: TextStyle(fontSize: 40, color: Colors.white))),
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
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                GestureDetector(
                  onTap: () => Navigator.of(context).pop(),
                  child: Container(width: 36, height: 36, decoration: BoxDecoration(color: const Color(0xFFF3F4F6), borderRadius: BorderRadius.circular(18)),
                    child: const Icon(Icons.arrow_back_ios_new, size: 16),
                  ),
                ),
                const Expanded(child: Center(child: Text('用药管理', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700)))),
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
                      onTap: () => _showAddMedDialog(),
                      child: Container(width: 36, height: 36, decoration: BoxDecoration(color: AppTheme.primary, borderRadius: BorderRadius.circular(18)),
                        child: const Icon(Icons.add, size: 20, color: Colors.white),
                      ),
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 10),
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
                    width: 130,
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                    decoration: BoxDecoration(
                      color: active ? const Color(0xFFDBEAFE) : (_darkMode ? const Color(0xFF1F2937) : Colors.white),
                      border: Border.all(color: active ? AppTheme.primary : const Color(0xFFE5E7EB)),
                      borderRadius: BorderRadius.circular(24),
                    ),
                    child: Row(
                      children: [
                        Container(width: 36, height: 36, decoration: BoxDecoration(color: bgs[i], shape: BoxShape.circle),
                          child: Center(child: Text(icons[i], style: const TextStyle(fontSize: 18))),
                        ),
                        const SizedBox(width: 8),
                        Text(names[i], style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
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

  // ===== Today Summary Card =====
  Widget _buildTodaySummaryCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        gradient: const LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Color(0xFF66BB6A), Color(0xFF43A047)]),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: const Color(0xFF16A34A).withValues(alpha: 0.3), blurRadius: 16)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('今日服药进度', style: TextStyle(fontSize: 14, color: Color(0xFFFFFFFF), opacity: 0.9)),
              Text('2026年7月24日 周五', style: TextStyle(fontSize: 11, color: Color(0xFFFFFFFF), opacity: 0.7)),
            ],
          ),
          const SizedBox(height: 16),
          Row(
            children: [
              SizedBox(
                width: 90, height: 90,
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    CircularProgressIndicator(value: 1, strokeWidth: 8, backgroundColor: const Color(0xFFFFFFFF).withValues(alpha: 0.2)),
                    if (_todayTotal > 0)
                      CircularProgressIndicator(
                        value: _adherenceRate,
                        strokeWidth: 8,
                        backgroundColor: Colors.transparent,
                        valueColor: const AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                    Column(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text('$_todayTaken/$_todayTotal', style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w800, color: Colors.white)),
                        Text('已服用', style: const TextStyle(fontSize: 9, color: Color(0xFFFFFFFF), opacity: 0.7)),
                      ],
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 20),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    RichText(
                      text: TextSpan(
                        style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Colors.white),
                        children: [
                          TextSpan(text: '${(_adherenceRate * 100).toInt()}%', style: const TextStyle(opacity: 1)),
                          const TextSpan(text: ' 服药率'),
                        ],
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text('漏服 0 次 · 延迟 1 次', style: const TextStyle(fontSize: 12, color: Color(0xFFFFFFFF), opacity: 0.9)),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          Divider(color: const Color(0xFFFFFFFF).withValues(alpha: 0.2)),
          const SizedBox(height: 8),
          Wrap(
            spacing: 0,
            runSpacing: 4,
            children: [
              _reminderRow('\u{2713}', '早餐药 - 已播报'),
              _reminderRow('\u{2713}', '午餐药 - 已播报'),
              _reminderRow('\u{2713}', '晚餐药 - 已播报'),
              _reminderRow('\u{25EF}', '睡前药 - 未到时', dim: true),
            ],
          ),
        ],
      ),
    );
  }

  Widget _reminderRow(String check, String text, {bool dim = false}) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 2),
      child: Row(
        children: [
          Text(check, style: TextStyle(fontSize: 14, color: dim ? const Color(0xFFFFFFFF).withValues(alpha: 0.4) : const Color(0xFFFFFFFF))),
          const SizedBox(width: 6),
          Text(text, style: TextStyle(fontSize: 12, color: dim ? const Color(0xFFFFFFFF).withValues(alpha: 0.5) : const Color(0xFFFFFFFF))),
        ],
      ),
    );
  }

  // ===== Period Tabs =====
  Widget _buildPeriodTabs() {
    final periods = [
      (label: '上午', count: '3项'),
      (label: '中午', count: '2项'),
      (label: '下午', hasMissed: true),
      (label: '晚上', count: '2项'),
      (label: '睡前', count: '1项'),
    ];
    return Row(
      children: periods.map((p) {
        final isActive = p.label == _activePeriod;
        return Padding(
          padding: const EdgeInsets.only(right: 6),
          child: GestureDetector(
            onTap: () => setState(() => _activePeriod = p.label),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
              decoration: BoxDecoration(
                color: isActive ? AppTheme.primary : (_darkMode ? const Color(0xFF374151) : Colors.white),
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: p.hasMissed && !isActive ? AppTheme.statusDanger : (isActive ? AppTheme.primary : const Color(0xFFE5E7EB))),
                boxShadow: isActive ? [BoxShadow(color: AppTheme.primary.withValues(alpha: 0.2), blurRadius: 8)] : [],
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(p.label, style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: isActive ? Colors.white : (_darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280)))),
                  if (p.count != null) ...[
                    const SizedBox(width: 4),
                    Text(p.count!, style: TextStyle(fontSize: 10, color: isActive ? const Color(0xFFFFFFFF).withValues(alpha: 0.7) : const Color(0xFF9CA3AF))),
                  ],
                  if (p.hasMissed && !isActive) ...[
                    const SizedBox(width: 4),
                    Container(width: 6, height: 6, decoration: const BoxDecoration(color: AppTheme.statusDanger, shape: BoxShape.circle)),
                  ],
                ],
              ),
            ),
          ),
        );
      }).toList(),
    );
  }

  // ===== Medication Item =====
  Widget _buildMedItem(MedicationRule rule) {
    final isTaken = _takenIds.contains(rule.id);
    final hour = int.tryParse(rule.scheduleTime.substring(0, 2)) ?? 0;
    final isMissed = hour < DateTime.now().hour && !isTaken;
    final timeStr = rule.scheduleTime.substring(0, 5);
    final period = _periodFromTime(rule.scheduleTime);

    return Container(
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : (isTaken ? const Color(0xFFFFFFFF).withValues(alpha: 0.7) : (isMissed ? const Color(0xFFFFF5F5) : Colors.white)),
        borderRadius: BorderRadius.circular(14),
        border: isMissed ? Border.all(color: const Color(0xFFEF4444).withValues(alpha: 0.15)) : null,
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4)],
      ),
      padding: const EdgeInsets.all(14),
      child: Row(
        children: [
          // Time column
          Container(
            width: 52,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Text(timeStr, style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: isMissed ? AppTheme.statusDanger : (_darkMode ? Colors.white : const Color(0xFF1F2937)))),
                Text(period, style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
              ],
            ),
          ),
          const SizedBox(width: 12),
          // Icon
          Container(
            width: 44, height: 44, decoration: BoxDecoration(color: const Color(0xFFFFF3C7), borderRadius: BorderRadius.circular(22)),
            child: const Center(child: Text('\u{1F48A}', style: TextStyle(fontSize: 22))),
          ),
          const SizedBox(width: 12),
          // Body
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(isMissed ? rule.pillType : rule.pillType, style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700, color: isMissed ? AppTheme.statusDanger : (_darkMode ? Colors.white : const Color(0xFF374151)))),
                Text(rule.doseCount != null ? '每次 ${rule.doseCount} 粒' : '', style: TextStyle(fontSize: 12, color: _darkMode ? const Color(0xFF9CA3AF) : const Color(0xFF6B7280))),
                if (!isTaken && !isMissed)
                  Text('预计 $timeStr 语音播报提醒', style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                if (isMissed)
                  Text('\u{26A0} 已超时未服用，已发送短信提醒', style: const TextStyle(fontSize: 11, color: AppTheme.statusDanger)),
              ],
            ),
          ),
          // Status + action
          Container(
            width: 44,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                if (isTaken)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                    decoration: BoxDecoration(color: const Color(0xFFF0FDF4), borderRadius: BorderRadius.circular(8)),
                    child: const Text('✓', style: TextStyle(fontSize: 10, color: Color(0xFF16A34A), fontWeight: FontWeight.w700)),
                  )
                else if (isMissed)
                  Column(
                    children: [
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                        decoration: BoxDecoration(color: const Color(0xFFFEF2F2), borderRadius: BorderRadius.circular(8)),
                        child: const Text('漏服', style: TextStyle(fontSize: 10, color: Color(0xFFEF4444), fontWeight: FontWeight.w700)),
                      ),
                      const SizedBox(height: 4),
                      GestureDetector(
                        onTap: () => _confirmTaken(rule.id, rule.pillType, '${rule.doseCount}粒'),
                        child: Container(
                          width: 36, height: 36, decoration: BoxDecoration(color: AppTheme.statusWarning, borderRadius: BorderRadius.circular(18)),
                          child: const Center(child: Text('补', style: TextStyle(fontSize: 14, color: Colors.white, fontWeight: FontWeight.w700))),
                        ),
                      ),
                    ],
                  )
                else
                  GestureDetector(
                    onTap: () => _confirmTaken(rule.id, rule.pillType, '${rule.doseCount}粒'),
                    child: Container(
                      width: 36, height: 36, decoration: BoxDecoration(color: AppTheme.statusNormal, borderRadius: BorderRadius.circular(18)),
                      child: const Center(child: Text('\u{2713}', style: TextStyle(fontSize: 18, color: Colors.white))),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  // ===== Inventory Card =====
  Widget _buildInventoryCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('\u{1F48E} 药品库存', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
              Text('查看全部 →', style: TextStyle(fontSize: 12, color: Color(0xFF2563EB), fontWeight: FontWeight.w600)),
            ],
          ),
          const SizedBox(height: 12),
          _invItem('\u{1F48A}', '降压药 (氨氯地平)', 75, 100, 'ok', '22天'),
          const SizedBox(height: 10),
          _invItem('\u{1F48A}', '钙片', 35, 100, 'low', '约5天用完'),
          const SizedBox(height: 10),
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(border: Border.all(color: const Color(0xFFEF4444).withValues(alpha: 0.15)), borderRadius: BorderRadius.circular(12)),
            child: Row(
              children: [
                const Text('\u{1F48A}', style: TextStyle(fontSize: 24)),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('降糖药 (二甲双胍)', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFFEF4444))),
                      const SizedBox(height: 4),
                      ClipRRect(
                        borderRadius: BorderRadius.circular(3),
                        child: LinearProgressIndicator(value: 0.1, minHeight: 6, backgroundColor: const Color(0xFFE5E7EB), valueColor: AlwaysStoppedAnimation<Color>(AppTheme.statusDanger)),
                      ),
                      const Text('\u{26A0} 仅剩2天量，建议尽快购买', style: TextStyle(fontSize: 10, color: Color(0xFFEF4444), fontWeight: FontWeight.w600)),
                    ],
                  ),
                ),
                const SizedBox(width: 10),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.end,
                  children: const [
                    Text('2', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w700, color: Color(0xFFEF4444))),
                    Text('天', style: TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _invItem(String icon, String name, int pct, int max, String level, String label) {
    final barColor = level == 'ok' ? AppTheme.statusNormal : (level == 'low' ? AppTheme.statusWarning : AppTheme.statusDanger);
    return Row(
      children: [
        Text(icon, style: const TextStyle(fontSize: 24)),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(name, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
              const SizedBox(height: 4),
              ClipRRect(
                borderRadius: BorderRadius.circular(3),
                child: LinearProgressIndicator(value: pct / 100, minHeight: 6, backgroundColor: const Color(0xFFE5E7EB), valueColor: AlwaysStoppedAnimation<Color>(barColor)),
              ),
            ],
          ),
        ),
        const SizedBox(width: 10),
        Column(
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text('$pct', style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w700)),
            Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))),
          ],
        ),
      ],
    );
  }

  // ===== Remote Config =====
  Widget _buildRemoteConfig() {
    final configs = [
      ('语音播报提醒', '药盒到点自动语音通知服药', true),
      ('家属推送通知', '漏服/延迟时推送消息给家属', true),
      ('短信兜底提醒', '超时2小时未服发送短信', true),
      ('库存预警通知', '药品低于3天量时提醒补货', false),
    ];
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('\u{2699}\u{FE0F} 远程配置', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
          const SizedBox(height: 12),
          ...configs.asMap().entries.map((e) {
            final idx = e.key;
            final cfg = e.value;
            return Column(
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(cfg.$2, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                          Text(cfg.$3, style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
                        ],
                      ),
                    ),
                    GestureDetector(
                      onTap: () {},
                      child: Container(
                        width: 48, height: 28, decoration: BoxDecoration(color: cfg.$4 ? AppTheme.statusNormal : const Color(0xFFD1D5DB), borderRadius: BorderRadius.circular(14)),
                        child: AnimatedAlign(
                          duration: const Duration(milliseconds: 200),
                          alignment: cfg.$4 ? Alignment.centerRight : Alignment.centerLeft,
                          child: Container(width: 22, height: 22, margin: EdgeInsets.only(left: cfg.$4 ? 24 : 3), decoration: const BoxDecoration(color: Colors.white, shape: BoxShape.circle), shadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 3)]),
                        ),
                      ),
                    ),
                  ],
                ),
                if (idx < configs.length - 1) const SizedBox(height: 12),
              ],
            );
          }).toList(),
        ],
      ),
    );
  }

  // ===== Adherence History =====
  Widget _buildAdherenceHistory() {
    final days = ['一', '二', '三', '四', '五', '六', '日'];
    final cells = [
      ('full', '24'), ('full', '23'), ('full', '22'), ('full', '21'), ('partial', '20'), ('full', '19'), ('full', '18'),
      ('full', '17'), ('full', '16'), ('missed', '15'), ('full', '14'), ('partial', '13'), ('full', '12'), ('full', '11'),
      ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''),
      ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''), ('empty', ''),
    ];
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: _darkMode ? const Color(0xFF1F2937) : Colors.white,
        borderRadius: BorderRadius.circular(16),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 16)],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('\u{1F4CA} 服药依从性', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
              Text('2026年7月', style: TextStyle(fontSize: 12, color: Color(0xFF2563EB), fontWeight: FontWeight.w600)),
            ],
          ),
          const SizedBox(height: 10),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: const [
              Text('本月服药率', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
              Text('连续准时 12 天', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280))),
            ],
          ),
          const SizedBox(height: 2),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              const Text('', style: TextStyle(fontSize: 12)),
              RichText(
                text: const TextSpan(
                  children: [
                    TextSpan(text: '92%', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF16A34A))),
                  ],
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          // Weekday labels
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: days.map((d) => Center(child: Text(d, style: const TextStyle(fontSize: 10, color: Color(0xFF9CA3AF))))).toList(),
          ),
          const SizedBox(height: 4),
          // Calendar grid
          Wrap(
            spacing: 4,
            runSpacing: 4,
            children: cells.map((c) {
              final color = c.$1 == 'full' ? AppTheme.statusNormal : (c.$1 == 'partial' ? AppTheme.statusWarning : (c.$1 == 'missed' ? AppTheme.statusDanger : const Color(0xFFF3F4F6)));
              final textColor = c.$1 == 'empty' ? const Color(0xFFD1D5DB) : (c.$1 == 'full' ? AppTheme.statusNormal : (c.$1 == 'partial' ? const Color(0xFFB45309) : AppTheme.statusDanger));
              return Container(
                width: (MediaQuery.of(context).size.width - 80) / 7 - 4,
                height: (MediaQuery.of(context).size.width - 80) / 7 - 4,
                decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(6)),
                child: Center(child: Text(c.$2, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: textColor))),
              );
            }).toList(),
          ),
          const SizedBox(height: 10),
          Row(
            children: const [
              _legendChip(Color(0xFFDCFCE7), AppTheme.statusNormal, '全按时'),
              SizedBox(width: 12),
              _legendChip(Color(0xFFFEF9C3), AppTheme.statusWarning, '部分延迟'),
              SizedBox(width: 12),
              _legendChip(Color(0xFFFEE2E2), AppTheme.statusDanger, '有漏服'),
              SizedBox(width: 12),
              _legendChip(Color(0xFFF3F4F6), Color(0xFFD1D5DB), '无数据'),
            ],
          ),
        ],
      ),
    );
  }

  Widget _legendChip(Color bg, Color text, String label) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(width: 10, height: 10, decoration: BoxDecoration(color: bg, borderRadius: BorderRadius.circular(2))),
        const SizedBox(width: 4),
        Text(label, style: TextStyle(fontSize: 10, color: text)),
      ],
    );
  }

  // ===== Add Medicine Dialog =====
  void _showAddMedDialog() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(24))),
      builder: (ctx) => Container(
        padding: const EdgeInsets.all(20),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Center(child: Container(width: 36, height: 4, decoration: BoxDecoration(color: const Color(0xFFD1D5DB), borderRadius: BorderRadius.circular(2)))),
            const SizedBox(height: 16),
            const Text('添加用药规则', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
            const SizedBox(height: 16),
            _configField('药品名称', '如：降压药'),
            const SizedBox(height: 12),
            Row(
              children: [
                Expanded(child: _configField('剂量', '1 粒')),
                const SizedBox(width: 8),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const Text('频次', style: TextStyle(fontSize: 12, color: Color(0xFF6B7280), fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      DropdownButtonFormField<String>(
                        value: '每日3次',
                        items: const ['每日1次', '每日2次', '每日3次', '必要时'].map((v) => DropdownMenuItem(value: v, child: Text(v))).toList(),
                        onChanged: (_) {},
                        decoration: const InputDecoration(border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(10))), contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8)),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),
            _configField('服用时间', '', isTime: true),
            const SizedBox(height: 12),
            _configField('备注', '如：餐后服用'),
            const SizedBox(height: 16),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () => Navigator.of(context).pop(),
                style: ElevatedButton.styleFrom(backgroundColor: AppTheme.statusNormal, foregroundColor: Colors.white, padding: const EdgeInsets.symmetric(vertical: 14), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                child: const Text('保存用药规则', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w700)),
              ),
            ),
            const SizedBox(height: 8),
            SizedBox(
              width: double.infinity,
              child: OutlinedButton(
                onPressed: () => Navigator.of(context).pop(),
                style: OutlinedButton.styleFrom(padding: const EdgeInsets.symmetric(vertical: 14), shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12))),
                child: const Text('取消', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600)),
              ),
            ),
            const SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _configField(String label, String placeholder, {bool isTime = false}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 12, color: Color(0xFF6B7280), fontWeight: FontWeight.w600)),
        const SizedBox(height: 4),
        isTime
            ? const TextField(decoration: InputDecoration(border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(10))), contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 8)))
            : TextField(decoration: InputDecoration(hintText: placeholder, border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(10))), contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10))),
      ],
    );
  }

  String _periodFromTime(String time) {
    final hour = int.tryParse(time.substring(0, 2)) ?? 0;
    if (hour < 10) return '上午';
    if (hour < 13) return '中午';
    if (hour < 17) return '下午';
    if (hour < 21) return '晚上';
    return '睡前';
  }
}

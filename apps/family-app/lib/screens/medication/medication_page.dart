import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/bottom_nav_bar.dart';

/// Medication management page — matches medication.html prototype
class MedicationPage extends StatefulWidget {
  const MedicationPage({super.key});

  @override
  State<MedicationPage> createState() => _MedicationPageState();

  int get initialIndex => 3;
}

class _MedicationPageState extends State<MedicationPage> {
  int _selectedIndex = 3;
  bool _pushEnabled = true;
  bool _voiceEnabled = true;
  bool _vibrationEnabled = false;
  bool _emailEnabled = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            // Header
            SliverToBoxAdapter(
              child: Container(
                padding: const EdgeInsets.fromLTRB(20, 12, 20, 12),
                color: AppTheme.bgCard,
                child: Row(
                  children: [
                    IconButton(icon: const Icon(Icons.arrow_back_ios_new, size: 18), onPressed: () {}),
                    const Expanded(child: Text('用药管理', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700))),
                    IconButton(icon: const Icon(Icons.settings_outlined), onPressed: () {}),
                  ],
                ),
              ),
            ),

            // Adherence overview card
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Container(
                  decoration: BoxDecoration(
                    gradient: const LinearGradient(colors: [Color(0xFF66BB6A), Color(0xFF43A047)]),
                    borderRadius: BorderRadius.circular(16),
                  ),
                  padding: const EdgeInsets.all(20),
                  child: Column(
                    children: [
                      const Text('本周服药依从性',
                          style: TextStyle(fontSize: 13, color: Colors.white, opacity: 0.9)),
                      const SizedBox(height: 12),
                      Stack(
                        alignment: Alignment.center,
                        children: [
                          SizedBox(
                            width: 110,
                            height: 110,
                            child: CircularProgressIndicator(
                              value: 0.85,
                              strokeWidth: 10,
                              backgroundColor: Colors.white.withOpacity(0.2),
                              valueColor: const AlwaysStoppedAnimation<Color>(Colors.white),
                            ),
                          ),
                          Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              const Text('85', style: TextStyle(fontSize: 30, fontWeight: FontWeight.w800, color: Colors.white)),
                              Text('%', style: TextStyle(fontSize: 12, color: Colors.white.withOpacity(0.8))),
                            ],
                          ),
                        ],
                      ),
                      const SizedBox(height: 10),
                      const Text('本周已服 21/24 次药 · 连续2天达标',
                          style: TextStyle(fontSize: 12, color: Colors.white, opacity: 0.9)),
                    ],
                  ),
                ),
              ),
            ),

            // Med stats row
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
                child: Row(
                  children: [
                    _medStat('21', '已服用', const Color(0xFF4CAF50)),
                    const SizedBox(width: 10),
                    _medStat('2', '漏服', const Color(0xFFFF6B6B), sub: '昨日降压药'),
                    const SizedBox(width: 10),
                    _medStat('1', '迟到', const Color(0xFFFFA726), sub: '迟30分钟'),
                  ],
                ),
              ),
            ),

            // Today's schedule
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.only(left: 20, right: 20, bottom: 12),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('今日用药计划', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                    GestureDetector(
                      onTap: () {},
                      child: const Text('+ 添加', style: TextStyle(fontSize: 12, color: AppTheme.primary, fontWeight: FontWeight.w600)),
                    ),
                  ],
                ),
              ),
            ),

            // Morning - taken
            SliverToBoxAdapter(
              child: _timeSlot(
                time: '08:00',
                period: '早餐后',
                dotColor: AppTheme.statusNormal,
                dotBg: const Color(0xFFE8F5E9),
                lineColor: AppTheme.statusNormal,
                borderColor: AppTheme.statusNormal,
                bgColor: null,
                medNames: '降压药 + 阿司匹林',
                dosage: '缬沙坦 80mg × 1 | 阿司匹林 100mg × 1',
                status: '✓ 已服用 08:05',
                statusColor: const Color(0xFF2E7D32),
                statusBg: const Color(0xFFE8F5E9),
                note: null,
              ),
            ),

            // Noon - taken
            SliverToBoxAdapter(
              child: _timeSlot(
                time: '12:00',
                period: '午餐后',
                dotColor: AppTheme.statusNormal,
                dotBg: const Color(0xFFE8F5E9),
                lineColor: AppTheme.statusNormal,
                borderColor: AppTheme.statusNormal,
                bgColor: null,
                medNames: '降糖药 (二甲双胍)',
                dosage: '二甲双胍 500mg × 1',
                status: '✓ 已服用 12:10',
                statusColor: const Color(0xFF2E7D32),
                statusBg: const Color(0xFFE8F5E9),
                note: null,
              ),
            ),

            // Evening - missed
            SliverToBoxAdapter(
              child: _timeSlot(
                time: '20:00',
                period: '晚餐后',
                dotColor: const Color(0xFFFF6B6B),
                dotBg: const Color(0xFFFFF0F0),
                lineColor: null,
                borderColor: const Color(0xFFFF6B6B),
                bgColor: null,
                medNames: '维生素D + 钙片',
                dosage: '维生素D 400IU × 1 | 碳酸钙 500mg × 1',
                status: '✗ 未服用（已超2小时）',
                statusColor: const Color(0xFFC62828),
                statusBg: const Color(0xFFFFEBEE),
                note: '💡 已发送短信提醒',
              ),
            ),

            // Night - pending
            SliverToBoxAdapter(
              child: _timeSlot(
                time: '21:00',
                period: '睡前',
                dotColor: const Color(0xFFFFA726),
                dotBg: const Color(0xFFFFF8E1),
                lineColor: null,
                borderColor: const Color(0xFFFFA726),
                bgColor: null,
                medNames: '辅酶Q10',
                dosage: '辅酶Q10 100mg × 1',
                status: '⏳ 待服用',
                statusColor: const Color(0xFFEF6C00),
                statusBg: const Color(0xFFFFF3E0),
                note: '🔔 智能药盒将语音播报提醒',
              ),
            ),

            // Remote config section
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.only(left: 20, right: 20, top: 16, bottom: 12),
                child: const Text('远程配置', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
              ),
            ),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppTheme.bgCard,
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 6)],
                  ),
                  child: Column(
                    children: [
                      _configRow('📱 APP推送提醒', _pushEnabled, (v) => setState(() => _pushEnabled = v)),
                      _divider(),
                      _configRow('🔊 药盒语音播报', _voiceEnabled, (v) => setState(() => _voiceEnabled = v)),
                      _divider(),
                      _configRow('📳 震动提醒', _vibrationEnabled, (v) => setState(() => _vibrationEnabled = v)),
                      _divider(),
                      _configRow('📧 每日用药报告邮件', _emailEnabled, (v) => setState(() => _emailEnabled = v)),
                    ],
                  ),
                ),
              ),
            ),
            SliverToBoxAdapter(child: const SizedBox(height: 12)),
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: GestureDetector(
                  onTap: () {},
                  child: Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      border: Border.all(color: const Color(0xFFD0D0D0), style: BorderStyle.solid, strokeAlign: 0),
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: const Center(child: Text('＋ 新增用药规则', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF888888)))),
                  ),
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 24)),
          ],
        ),
      ),
      bottomNavigationBar: BottomNavBar(selectedTab: _selectedIndex, onTabSelected: (i) => setState(() => _selectedIndex = i)),
    );
  }

  Widget _medStat(String value, String label, Color color, {String? sub}) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 12, horizontal: 8),
        decoration: BoxDecoration(
          color: AppTheme.bgCard,
          borderRadius: BorderRadius.circular(12),
          boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 4)],
        ),
        child: Column(
          children: [
            Text(value, style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: color)),
            Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
            if (sub != null) Text(sub, style: const TextStyle(fontSize: 9, color: Color(0xFFCCCCCC))),
          ],
        ),
      ),
    );
  }

  Widget _timeSlot({
    required String time,
    required String period,
    required Color dotColor,
    required Color dotBg,
    Color? lineColor,
    required Color borderColor,
    Color? bgColor,
    required String medNames,
    required String dosage,
    required String status,
    required Color statusColor,
    required Color statusBg,
    String? note,
  }) {
    final hasLine = lineColor != null;
    return Padding(
      padding: const EdgeInsets.only(bottom: 0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Time label
          Container(width: 56, padding: const EdgeInsets.only(top: 14), child: Column(
            children: [
              Text(time, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
              Text(period, style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
            ],
          )),
          const SizedBox(width: 14),
          // Connector
          Container(width: 2, height: 100, margin: const EdgeInsets.only(top: 0), child: Stack(
            children: [
              Positioned(top: 0, child: Container(
                width: 14, height: 14, decoration: BoxDecoration(color: dotBg, shape: BoxShape.circle, border: Border.all(color: dotColor, width: 3)),
              )),
              if (hasLine) Positioned(top: 14, left: -6, child: Container(width: 2, height: 86, color: lineColor!)),
            ],
          )),
          const SizedBox(width: 14),
          // Med card
          Expanded(
            child: Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: AppTheme.bgCard,
                borderRadius: BorderRadius.circular(12),
                border: Border(left: BorderSide(color: borderColor, width: 3)),
                boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.02), blurRadius: 4)],
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(medNames, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                  Text(dosage, style: const TextStyle(fontSize: 11, color: Color(0xFF888888))),
                  const SizedBox(height: 6),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(color: statusBg, borderRadius: BorderRadius.circular(8)),
                    child: Text(status, style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: statusColor)),
                  ),
                  if (note != null) Text(note, style: const TextStyle(fontSize: 10, color: Color(0xFFAAAAAA), marginTop: 4)),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _configRow(String label, bool value, ValueChanged<bool> onChanged) {
    return Column(
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label, style: const TextStyle(fontSize: 13, color: Color(0xFF555555))),
            GestureDetector(
              onTap: () => onChanged(!value),
              child: Container(
                width: 44,
                height: 24,
                decoration: BoxDecoration(
                  color: value ? AppTheme.statusNormal : const Color(0xFFDDDDDD),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: AnimatedAlign(
                  duration: const Duration(milliseconds: 200),
                  alignment: value ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    width: 20,
                    height: 20,
                    margin: EdgeInsets.only(right: value ? 2 : 0, left: value ? 0 : 2),
                    decoration: BoxDecoration(color: Colors.white, shape: BoxShape.circle, boxShadow: [
                      BoxShadow(color: Colors.black.withOpacity(0.2), blurRadius: 1),
                    ]),
                  ),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _divider() {
    return const Divider(height: 1, color: Color(0xFFF5F5FA));
  }
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../../app_state.dart';
import '../../widgets/bottom_nav_bar.dart';

/// Exercise tracking page — exercise type selector, duration input, calories display,
/// submit via POST /api/v1/chronic/:elderly_id/exercise, list past entries with step data.
class ExercisePage extends StatefulWidget {
  const ExercisePage({super.key});

  @override
  State<ExercisePage> createState() => _ExercisePageState();
}

class _ExercisePageState extends State<ExercisePage> {
  int _selectedIndex = 4;
  bool _loading = true;
  String? _elderlyId;

  // Exercise types
  final List<_ExerciseType> _exerciseTypes = [
    _ExerciseType(label: '散步', icon: '\u{1F6B6}', color: const Color(0xFFE8F5E9), border: const Color(0xFF4CAF50), calPerMin: 4.0),
    _ExerciseType(label: '慢跑', icon: '\u{1F3C3}', color: const Color(0xFFE3F2FD), border: const Color(0xFF2196F3), calPerMin: 8.0),
    _ExerciseType(label: '太极拳', icon: '\u{26A1}', color: const Color(0xFFFFF8E1), border: const Color(0xFFFF9800), calPerMin: 3.0),
    _ExerciseType(label: '骑车', icon: '\u{1F6B2}', color: const Color(0xFFF3E5F5), border: const Color(0xFF9C27B0), calPerMin: 6.0),
    _ExerciseType(label: '游泳', icon: '\u{1F3CA}', color: const Color(0xFFE0F7FA), border: const Color(0xFF00BCD4), calPerMin: 7.0),
    _ExerciseType(label: '其他', icon: '\u{1F3AF}', color: const Color(0xFFFCE4EC), border: const Color(0xFFE91E63), calPerMin: 5.0),
  ];
  int _selectedType = 0;

  final TextEditingController _durationController = TextEditingController();
  int get _duration => int.tryParse(_durationController.text) ?? 0;
  double get _calories => _exerciseTypes[_selectedType].calPerMin * _duration;

  List<ExerciseEntry> _entries = [];
  int? _todaySteps;

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  @override
  void dispose() {
    _durationController.dispose();
    super.dispose();
  }

  Future<void> _fetchData() async {
    final id = context.read<AppState>().elderlyId;
    if (id == null || id.isEmpty) {
      setState(() => _loading = false);
      return;
    }
    _elderlyId = id;
    try {
      final resp = await ApiClient.instance.get('/api/v1/chronic/$id/exercise');
      final data = resp.data as Map<String, dynamic>;
      final list = data['data'] as List? ?? [];
      setState(() {
        _entries = list.map((e) => ExerciseEntry.fromJson(e as Map<String, dynamic>)).toList();
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
    // Fetch today's step data from bracelet
    _fetchSteps();
  }

  Future<void> _fetchSteps() async {
    final id = context.read<AppState>().elderlyId;
    if (id == null) return;
    try {
      final resp = await ApiClient.instance.get('/api/v1/health/history', query: {
        'elderly_id': id,
        'from': DateTime.now().subtract(const Duration(days: 1)).toIso8601String(),
        'to': DateTime.now().toIso8601String(),
      });
      final list = resp.data as List? ?? [];
      if (list.isNotEmpty) {
        final last = list.last as Map<String, dynamic>;
        setState(() => _todaySteps = last['steps'] as int?);
      }
    } catch (e) {
      // Steps fetch is best-effort
    }
  }

  Future<void> _submit() async {
    if (_elderlyId == null) return;
    if (_duration <= 0) {
      _showToast('请输入运动时长', Colors.orange);
      return;
    }
    try {
      await ApiClient.instance.post('/api/v1/chronic/$_elderlyId/exercise', data: {
        'type': _exerciseTypes[_selectedType].label,
        'duration_min': _duration,
        'calories': _calories,
      });
      _durationController.clear();
      _showToast('记录成功', AppTheme.statusNormal);
      await _fetchData();
    } catch (e) {
      _showToast('提交失败', AppTheme.statusDanger);
    }
  }

  void _showToast(String msg, Color color) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(msg, style: const TextStyle(color: Colors.white)),
        backgroundColor: color,
        duration: const Duration(seconds: 1),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: _loading
            ? const Center(child: CircularProgressIndicator(color: AppTheme.primary))
            : CustomScrollView(
                slivers: [
                  _buildTopBar(),
                  SliverToBoxAdapter(
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          _buildStepsBanner(),
                          const SizedBox(height: 16),
                          _buildTypeSelector(),
                          const SizedBox(height: 16),
                          _buildFormCard(),
                          const SizedBox(height: 16),
                          _buildHistorySection(),
                          const SizedBox(height: 24),
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

  Widget _buildTopBar() {
    return SliverToBoxAdapter(
      child: Container(
        color: Colors.white,
        padding: const EdgeInsets.fromLTRB(16, 12, 16, 8),
        child: Row(
          children: [
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                gradient: LinearGradient(colors: [const Color(0xFF2563EB), const Color(0xFF7C3AED)]),
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Center(child: Text('颐', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Colors.white))),
            ),
            const SizedBox(width: 8),
            const Text('Eregen 颐贞', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
            const Spacer(),
            const Text('运动记录', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
          ],
        ),
      ),
    );
  }

  Widget _buildStepsBanner() {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        gradient: const LinearGradient(colors: [Color(0xFFE8F5E9), Color(0xFFC8E6C9)]),
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
      ),
      child: Row(
        children: [
          Container(
            width: 44,
            height: 44,
            decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12)),
            child: const Center(child: Text('\u{1F6B6}', style: TextStyle(fontSize: 22))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text('今日步数', style: TextStyle(fontSize: 12, color: Color(0xFF4B5563))),
                Text(
                  _todaySteps != null ? '$_todaySteps 步' : '-- 步',
                  style: const TextStyle(fontSize: 22, fontWeight: FontWeight.w800, color: Color(0xFF1F2937)),
                ),
              ],
            ),
          ),
          if (_todaySteps != null)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(8)),
              child: Text(
                _todaySteps! >= 6000 ? '达标' : '加油',
                style: TextStyle(fontSize: 11, fontWeight: FontWeight.w600, color: _todaySteps! >= 6000 ? AppTheme.statusNormal : AppTheme.statusWarning),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildTypeSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('运动类型', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
        const SizedBox(height: 10),
        SizedBox(
          height: 60,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: _exerciseTypes.length,
            separatorBuilder: (_, __) => const SizedBox(width: 8),
            itemBuilder: (ctx, i) {
              final t = _exerciseTypes[i];
              final active = i == _selectedType;
              return GestureDetector(
                onTap: () => setState(() => _selectedType = i),
                child: Container(
                  width: 72,
                  padding: const EdgeInsets.symmetric(vertical: 6),
                  decoration: BoxDecoration(
                    color: active ? t.color : Colors.white,
                    border: Border.all(color: active ? t.border : const Color(0xFFE5E7EB)),
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: active
                        ? [BoxShadow(color: t.border.withValues(alpha: 0.2), blurRadius: 8, offset: const Offset(0, 2))]
                        : [],
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(t.icon, style: const TextStyle(fontSize: 20)),
                      const SizedBox(height: 4),
                      Text(
                        t.label,
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: active ? FontWeight.w700 : FontWeight.w500,
                          color: active ? const Color(0xFF1F2937) : const Color(0xFF6B7280),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            },
          ),
        ),
      ],
    );
  }

  Widget _buildFormCard() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusLarge),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 12, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(_exerciseTypes[_selectedType].icon, style: const TextStyle(fontSize: 24)),
              const SizedBox(width: 8),
              Text(
                _exerciseTypes[_selectedType].label,
                style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937)),
              ),
              const Spacer(),
              Text(
                '${_exerciseTypes[_selectedType].calPerMin.toInt()} kcal/min',
                style: const TextStyle(fontSize: 12, color: Color(0xFF9CA3AF)),
              ),
            ],
          ),
          const SizedBox(height: 16),
          _buildInputRow(
            label: '运动时长（分钟）',
            child: TextField(
              controller: _durationController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: const InputDecoration(
                hintText: '输入运动分钟数',
                hintStyle: TextStyle(color: Color(0xFF9CA3AF)),
                prefixIcon: Icon(Icons.access_time, size: 20, color: Color(0xFF9CA3AF)),
                suffixText: 'min',
                border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(AppTheme.radiusSmall))),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
            ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: const Color(0xFFFFF8E1),
              borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNutritionItem('消耗热量', '${_calories.toInt()} kcal', '\u{1F525}'),
                _buildNutritionItem('运动类型', _exerciseTypes[_selectedType].label, _exerciseTypes[_selectedType].icon),
                _buildNutritionItem('时长', '${_duration} min', '\u{23F1}'),
              ],
            ),
          ),
          const SizedBox(height: 16),
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _submit,
              style: ElevatedButton.styleFrom(
                backgroundColor: AppTheme.primary,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 14),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppTheme.radiusMedium)),
                elevation: 0,
              ),
              child: const Text('提交记录', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700)),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInputRow({required String label, required Widget child}) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
        const SizedBox(height: 6),
        child,
      ],
    );
  }

  Widget _buildNutritionItem(String label, String value, String icon) {
    return Column(
      children: [
        Text(icon, style: const TextStyle(fontSize: 18)),
        const SizedBox(height: 4),
        Text(value, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
        Text(label, style: const TextStyle(fontSize: 11, color: Color(0xFF6B7280))),
      ],
    );
  }

  Widget _buildHistorySection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            const Text('历史运动记录', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
            Text('${_entries.length} 条', style: const TextStyle(fontSize: 12, color: Color(0xFF9CA3AF))),
          ],
        ),
        const SizedBox(height: 10),
        if (_entries.isEmpty)
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(AppTheme.radiusLarge)),
            child: const Center(child: Text('暂无运动记录', style: TextStyle(color: Color(0xFF9CA3AF)))),
          )
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _entries.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (ctx, i) => _exerciseEntryCard(_entries[i]),
          ),
      ],
    );
  }

  Widget _exerciseEntryCard(ExerciseEntry e) {
    final typeIdx = _exerciseTypes.indexWhere((t) => t.label == e.type);
    final t = typeIdx >= 0 ? _exerciseTypes[typeIdx] : _exerciseTypes.last;
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
        border: Border(left: BorderSide(color: t.border, width: 4)),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 8, offset: const Offset(0, 1))],
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: t.color, borderRadius: BorderRadius.circular(10)),
            child: Center(child: Text(t.icon, style: const TextStyle(fontSize: 18))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(e.type, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF1F2937))),
                    const SizedBox(width: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(color: t.color, borderRadius: BorderRadius.circular(6)),
                      child: Text('${e.durationMin}min', style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(e.recordedAt.toString().substring(5, 16), style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF))),
              ],
            ),
          ),
          if (e.calories != null)
            Text(
              '${e.calories!.toInt()}kcal',
              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: AppTheme.primary),
            ),
        ],
      ),
    );
  }
}

class _ExerciseType {
  final String label, icon;
  final Color color, border;
  final double calPerMin;
  const _ExerciseType({required this.label, required this.icon, required this.color, required this.border, required this.calPerMin});
}

class ExerciseEntry {
  final String id;
  final String elderlyId;
  final String type;
  final int? durationMin;
  final double? calories;
  final DateTime recordedAt;

  ExerciseEntry({required this.id, required this.elderlyId, required this.type,
                  this.durationMin, this.calories, required this.recordedAt});

  factory ExerciseEntry.fromJson(Map<String, dynamic> json) => ExerciseEntry(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    type: json['type'] as String,
    durationMin: json['duration_min'] as int?,
    calories: json['calories'] as double?,
    recordedAt: DateTime.parse(json['recorded_at'] as String),
  );
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../../app_state.dart';
import '../../widgets/bottom_nav_bar.dart';

/// Diet recording page — meal type selector, food name input, portion display,
/// carbs calculation, submit via POST /api/v1/chronic/:elderly_id/diet, list past entries.
class DietPage extends StatefulWidget {
  const DietPage({super.key});

  @override
  State<DietPage> createState() => _DietPageState();
}

class _DietPageState extends State<DietPage> {
  int _selectedIndex = 4;
  bool _loading = true;
  String? _elderlyId;

  // Form state
  int _mealTypeIndex = 0;
  final List<String> _mealTypes = ['breakfast', 'lunch', 'dinner', 'snack'];
  final List<String> _mealLabels = ['早餐', '午餐', '晚餐', '加餐'];
  final List<String> _mealIcons = ['\u{1F305}', '\u{1F35C}', '\u{1F37D}', '\u{1F36A}'];
  final List<Color> _mealColors = [
    const Color(0xFFFFF8E1),
    const Color(0xFFFFF3E0),
    const Color(0xFFEDE7F6),
    const Color(0xFFF3E5F5),
  ];
  final List<Color> _mealBorders = [
    const Color(0xFFFFD54F),
    const Color(0xFFFFB74D),
    const Color(0xFFCE93D8),
    const Color(0xFFBA68C8),
  ];

  final TextEditingController _foodController = TextEditingController();
  final TextEditingController _portionController = TextEditingController();

  double get _carbs => _calcCarbs();
  double get _calories => _calcCalories();

  List<DietEntry> _entries = [];

  @override
  void initState() {
    super.initState();
    _fetchData();
  }

  @override
  void dispose() {
    _foodController.dispose();
    _portionController.dispose();
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
      final resp = await ApiClient.instance.get('/api/v1/chronic/$id/diet');
      final data = resp.data as Map<String, dynamic>;
      final list = data['data'] as List? ?? [];
      setState(() {
        _entries = list.map((e) => DietEntry.fromJson(e as Map<String, dynamic>)).toList();
        _loading = false;
      });
    } catch (e) {
      setState(() => _loading = false);
    }
  }

  double _calcCarbs() {
    final portion = double.tryParse(_portionController.text) ?? 0;
    return double.parse((portion * 0.25).toStringAsFixed(1));
  }

  double _calcCalories() {
    final portion = double.tryParse(_portionController.text) ?? 0;
    return double.parse((portion * 1.0).toStringAsFixed(0));
  }

  Future<void> _submit() async {
    if (_elderlyId == null) return;
    final food = _foodController.text.trim();
    final portionText = _portionController.text.trim();
    if (food.isEmpty) {
      _showToast('请输入食物名称', Colors.orange);
      return;
    }
    final portion = double.tryParse(portionText) ?? 150;
    try {
      await ApiClient.instance.post('/api/v1/chronic/$_elderlyId/diet', data: {
        'meal_type': _mealTypes[_mealTypeIndex],
        'food_items': food,
        'portion_g': portion,
        'total_carbs': _carbs,
        'total_calories': _calories,
      });
      _foodController.clear();
      _portionController.clear();
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
                          _buildMealSelector(),
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
            const Text('饮食记录', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
          ],
        ),
      ),
    );
  }

  Widget _buildMealSelector() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        const Text('餐次选择', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF6B7280))),
        const SizedBox(height: 10),
        SizedBox(
          height: 64,
          child: ListView.separated(
            scrollDirection: Axis.horizontal,
            itemCount: _mealTypes.length,
            separatorBuilder: (_, __) => const SizedBox(width: 10),
            itemBuilder: (ctx, i) {
              final active = i == _mealTypeIndex;
              return GestureDetector(
                onTap: () => setState(() => _mealTypeIndex = i),
                child: Container(
                  width: 80,
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  decoration: BoxDecoration(
                    color: active ? _mealColors[i] : Colors.white,
                    border: Border.all(color: active ? _mealBorders[i] : const Color(0xFFE5E7EB)),
                    borderRadius: BorderRadius.circular(14),
                    boxShadow: active
                        ? [BoxShadow(color: _mealBorders[i].withValues(alpha: 0.2), blurRadius: 8, offset: const Offset(0, 2))]
                        : [],
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(_mealIcons[i], style: const TextStyle(fontSize: 22)),
                      const SizedBox(height: 4),
                      Text(
                        _mealLabels[i],
                        style: TextStyle(
                          fontSize: 12,
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
          const Text('食物记录', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
          const SizedBox(height: 14),
          _buildInputRow(
            label: '食物名称',
            child: TextField(
              controller: _foodController,
              decoration: const InputDecoration(
                hintText: '例如：米饭、面条、馒头',
                hintStyle: TextStyle(color: Color(0xFF9CA3AF)),
                prefixIcon: Icon(Icons.restaurant_menu, size: 20, color: Color(0xFF9CA3AF)),
                border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(AppTheme.radiusSmall))),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
            ),
          ),
          const SizedBox(height: 12),
          _buildInputRow(
            label: '份量（克）',
            child: TextField(
              controller: _portionController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: const InputDecoration(
                hintText: '输入克数，默认150g',
                hintStyle: TextStyle(color: Color(0xFF9CA3AF)),
                prefixIcon: Icon(Icons.restaurant, size: 20, color: Color(0xFF9CA3AF)),
                suffixText: 'g',
                border: OutlineInputBorder(borderRadius: BorderRadius.all(Radius.circular(AppTheme.radiusSmall))),
                contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 10),
              ),
            ),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(12),
            decoration: BoxDecoration(
              color: AppTheme.primaryLight,
              borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildNutritionItem('碳水化合物', '${_carbs}g', '\u{1F33E}'),
                _buildNutritionItem('热量', '${_calories.toInt()}kcal', '\u{1F525}'),
                _buildNutritionItem('餐次', _mealLabels[_mealTypeIndex], _mealIcons[_mealTypeIndex]),
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
            const Text('历史饮食记录', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: Color(0xFF1F2937))),
            Text('${_entries.length} 条', style: const TextStyle(fontSize: 12, color: Color(0xFF9CA3AF))),
          ],
        ),
        const SizedBox(height: 10),
        if (_entries.isEmpty)
          Container(
            padding: const EdgeInsets.all(24),
            decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(AppTheme.radiusLarge)),
            child: const Center(child: Text('暂无饮食记录', style: TextStyle(color: Color(0xFF9CA3AF)))),
          )
        else
          ListView.separated(
            shrinkWrap: true,
            physics: const NeverScrollableScrollPhysics(),
            itemCount: _entries.length,
            separatorBuilder: (_, __) => const SizedBox(height: 8),
            itemBuilder: (ctx, i) => _dietEntryCard(_entries[i]),
          ),
      ],
    );
  }

  Widget _dietEntryCard(DietEntry e) {
    final idx = _mealTypes.indexOf(e.mealType);
    final color = idx >= 0 ? _mealColors[idx] : const Color(0xFFF5F5F5);
    final border = idx >= 0 ? _mealBorders[idx] : const Color(0xFFE5E7EB);
    final icon = idx >= 0 ? _mealIcons[idx] : '\u{1F37D}';
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(AppTheme.radiusMedium),
        border: Border(left: BorderSide(color: border, width: 4)),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.03), blurRadius: 8, offset: const Offset(0, 1))],
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(10)),
            child: Center(child: Text(icon, style: const TextStyle(fontSize: 18))),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Text(e.foodItems, style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF1F2937))),
                    const SizedBox(width: 6),
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(6)),
                      child: Text(_mealLabels[idx >= 0 ? idx : 3], style: const TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: Color(0xFF374151))),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Text(
                  '碳水 ${e.totalCarbs?.toStringAsFixed(1) ?? '?'}g  ·  ${e.recordedAt.toString().substring(5, 16)}',
                  style: const TextStyle(fontSize: 11, color: Color(0xFF9CA3AF)),
                ),
              ],
            ),
          ),
          if (e.totalCalories != null)
            Text(
              '${e.totalCalories!.toInt()}kcal',
              style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w700, color: AppTheme.primary),
            ),
        ],
      ),
    );
  }
}

class DietEntry {
  final String id;
  final String elderlyId;
  final String mealType;
  final String foodItems;
  final double? totalCarbs;
  final double? totalCalories;
  final DateTime recordedAt;

  DietEntry({required this.id, required this.elderlyId, required this.mealType, required this.foodItems,
              this.totalCarbs, this.totalCalories, required this.recordedAt});

  factory DietEntry.fromJson(Map<String, dynamic> json) => DietEntry(
    id: json['id'] as String,
    elderlyId: json['elderly_id'] as String,
    mealType: json['meal_type'] as String,
    foodItems: json['food_items'] as String,
    totalCarbs: json['total_carbs'] as double?,
    totalCalories: json['total_calories'] as double?,
    recordedAt: DateTime.parse(json['recorded_at'] as String),
  );
}

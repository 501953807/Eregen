import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../widgets/bottom_nav_bar.dart';
import '../../api/client.dart';

/// Welfare page — v2 design with parent card, welfare tag grid, payment history, sign-in reminder.
class WelfarePage extends StatefulWidget {
  const WelfarePage({super.key});

  @override
  State<WelfarePage> createState() => _WelfarePageState();
}

class _WelfarePageState extends State<WelfarePage> {
  int _selectedIndex = 5;
  bool _loading = true;
  bool _showNotification = true;

  final parentName = '张秀兰';
  final parentAge = 76;
  final parentGender = '女';
  final parentStatus = '腕带在线';

  final List<WelfareTag> welfareTags = [
    WelfareTag(icon: '\u{1F3E0}', name: '孤寡老人', validUntil: '2028-12-31', active: true),
    WelfareTag(icon: '\u{1F4B0}', name: '特困一级', validUntil: '2028-12-31', active: true),
    WelfareTag(icon: '\u{267F}', name: '残疾二级', validUntil: '2027-05-31', active: true),
    WelfareTag(icon: '\u{1F68C}', name: '公交优惠', validUntil: '2026-07-01', active: false),
    WelfareTag(icon: '\u{1F3E5}', name: '医疗救助', validUntil: '2028-12-31', active: true),
    WelfareTag(icon: '\u{1F4CB}', name: '特病门诊', validUntil: '2027-01-15', active: true),
  ];

  final List<PaymentRecord> payments = [
    PaymentRecord(type: '特困补助', period: '2026年7月', date: '2026-07-01 发放', amount: 800.00, status: PaymentStatus.received),
    PaymentRecord(type: '残疾补贴', period: '2026年7月', date: '2026-07-01 发放', amount: 400.00, status: PaymentStatus.received),
    PaymentRecord(type: '医疗救助', period: '2026年6月', date: '预计 07-25 发放', amount: 1000.00, status: PaymentStatus.pending),
    PaymentRecord(type: '公交补贴', period: '2026年6月', date: '银行处理失败', amount: 30.00, status: PaymentStatus.failed),
  ];

  final needSignIn = true;
  final signinHospital = '社区医院 A';

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFFEEF6FF),
      body: SafeArea(
        child: Stack(
          children: [
            CustomScrollView(
              slivers: [
                _buildHeader(),
                const SliverToBoxAdapter(child: SizedBox(height: 16)),
                _buildParentCard(),
                const SliverToBoxAdapter(child: SizedBox(height: 16)),
                _buildWelfareTagsSection(),
                const SliverToBoxAdapter(child: SizedBox(height: 16)),
                _buildPaymentsSection(),
                const SliverToBoxAdapter(child: SizedBox(height: 16)),
                if (needSignIn) _buildSignInReminder(),
                const SliverToBoxAdapter(child: SizedBox(height: 24)),
              ],
            ),
            if (_showNotification) _buildNotification(),
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
        decoration: const BoxDecoration(
          gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Color(0xFF4A90D9), Color(0xFF357ABD)]),
        ),
        padding: const EdgeInsets.fromLTRB(20, 15, 20, 40),
        child: Row(
          children: [
            GestureDetector(
              onTap: () => Navigator.of(context).pop(),
              child: const Text('←', style: TextStyle(fontSize: 22, color: Colors.white)),
            ),
            const Expanded(
              child: Center(child: Text('父母福利', style: TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: Colors.white))),
            ),
            const SizedBox(width: 22),
          ],
        ),
      ),
    );
  }

  Widget _buildNotification() {
    return Positioned(
      top: 0,
      left: 16,
      right: 16,
      child: Material(
        elevation: 4,
        color: const Color(0xFFFDF6EC),
        borderRadius: BorderRadius.circular(8),
        borderOnForeground: true,
        child: Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          decoration: BoxDecoration(
            color: const Color(0xFFFDF6EC),
            borderRadius: BorderRadius.circular(8),
            border: Border.all(color: const Color(0xFFFAECD8)),
          ),
          child: Row(
            children: [
              const Text('⚠️', style: TextStyle(fontSize: 18)),
              const SizedBox(width: 8),
              Expanded(
                child: Text(
                  '您的母亲有 1 个福利标签即将到期，请及时更新',
                  style: const TextStyle(fontSize: 13, color: Color(0xFFE6A23C)),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close, size: 18, color: Color(0xFFE6A23C)),
                onPressed: () => setState(() => _showNotification = false),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildParentCard() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Container(
          margin: const EdgeInsets.only(top: -32),
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 12, offset: const Offset(0, 4))],
          ),
          child: Row(
            children: [
              Container(
                width: 64,
                height: 64,
                decoration: BoxDecoration(
                  gradient: const LinearGradient(colors: [Color(0xFF66B1FF), Color(0xFF4A90D9)]),
                  shape: BoxShape.circle,
                ),
                child: Center(child: Text(parentName[0], style: const TextStyle(fontSize: 28, color: Colors.white, fontWeight: FontWeight.w700))),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(parentName, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w600, color: Color(0xFF303133))),
                    const SizedBox(height: 2),
                    Text('$parentAge 岁 · $parentGender', style: const TextStyle(fontSize: 13, color: Color(0xFF909399))),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
                decoration: BoxDecoration(color: const Color(0xFFF0F9EB), borderRadius: BorderRadius.circular(12)),
                child: const Text('● 腕带在线', style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Color(0xFF67C23A))),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildWelfareTagsSection() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8)]),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('福利标签', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Color(0xFF303133)),
                    textAlign: TextAlign.left,
                  ),
                  GestureDetector(
                    onTap: () {},
                    child: const Text('管理 →', style: TextStyle(fontSize: 13, color: Color(0xFF4A90D9), fontWeight: FontWeight.w600))),
                ],
              ),
              const SizedBox(height: 16),
              GridView.count(
                shrinkWrap: true,
                crossAxisCount: 3,
                crossAxisSpacing: 12,
                mainAxisSpacing: 12,
                childAspectRatio: 1.0,
                physics: const NeverScrollableScrollPhysics(),
                children: welfareTags.map((tag) => _welfareTagItem(tag)).toList(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _welfareTagItem(WelfareTag tag) {
    return Container(
      decoration: BoxDecoration(
        color: tag.active ? const Color(0xFFFAFAFA) : Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: tag.active ? const Color(0xFFEBEEF5) : const Color(0xFFFDE2E2)),
      ),
      padding: const EdgeInsets.symmetric(vertical: 16),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Text(tag.icon, style: const TextStyle(fontSize: 28)),
          const SizedBox(height: 8),
          Text(tag.name, style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: tag.active ? const Color(0xFF303133) : const Color(0xFFBBBBBB))),
          const SizedBox(height: 4),
          Text(
            tag.active ? '有效至 ${tag.validUntil}' : '已过期 (${tag.validUntil})',
            style: TextStyle(fontSize: 11, color: tag.active ? const Color(0xFF909399) : const Color(0xFFF56C6C)),
          ),
        ],
      ),
    );
  }

  Widget _buildPaymentsSection() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(color: Colors.white, borderRadius: BorderRadius.circular(12), boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8)]),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('补助领取记录', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                  GestureDetector(onTap: () {}, child: const Text('全部 →', style: TextStyle(fontSize: 13, color: Color(0xFF4A90D9), fontWeight: FontWeight.w600))),
                ],
              ),
              const SizedBox(height: 16),
              ...payments.map((p) => _paymentItem(p)),
            ],
          ),
        ),
      ),
    );
  }

  Widget _paymentItem(PaymentRecord p) {
    final borderColor = p.status == PaymentStatus.received
        ? AppTheme.statusNormal
        : (p.status == PaymentStatus.pending ? AppTheme.statusWarning : AppTheme.statusDanger);
    final amountColor = p.status == PaymentStatus.received
        ? AppTheme.statusNormal
        : (p.status == PaymentStatus.pending ? AppTheme.statusWarning : AppTheme.statusDanger);
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: const Color(0xFFFAFAFA),
        borderRadius: BorderRadius.circular(8),
        border: Border(left: BorderSide(color: borderColor, width: 3)),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text('${p.type} (${p.period})', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF303133))),
                const SizedBox(height: 2),
                Text(p.date, style: const TextStyle(fontSize: 12, color: Color(0xFF909399))),
              ],
            ),
          ),
          Text('¥${p.amount.toStringAsFixed(2)}', style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold, color: amountColor)),
        ],
      ),
    );
  }

  Widget _buildSignInReminder() {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 16),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 8)],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('签到提醒', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(colors: [Color(0xFFEEF6FF), Color(0xFFD9ECFF)]),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  children: [
                    const Text('🔔', style: TextStyle(fontSize: 32)),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text('今日福利签到未完成', style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: Color(0xFF303133))),
                          const SizedBox(height: 2),
                          Text('请在$signinHospital完成今日签到以激活补助', style: const TextStyle(fontSize: 12, color: Color(0xFF909399))),
                        ],
                      ),
                    ),
                    const SizedBox(width: 12),
                    ElevatedButton(
                      onPressed: () {},
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppTheme.primary,
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                      ),
                      child: const Text('去签到', style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

class WelfareTag {
  final String icon, name, validUntil;
  final bool active;
  WelfareTag({required this.icon, required this.name, required this.validUntil, required this.active});
}

enum PaymentStatus { received, pending, failed }

class PaymentRecord {
  final String type, period, date;
  final double amount;
  final PaymentStatus status;
  PaymentRecord({required this.type, required this.period, required this.date, required this.amount, required this.status});
}

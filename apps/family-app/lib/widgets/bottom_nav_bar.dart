import 'package:flutter/material.dart';
import '../../common/theme.dart';

/// Bottom navigation bar matching prototype
class BottomNavBar extends StatelessWidget {
  final int selectedTab;
  final ValueChanged<int> onTabSelected;
  const BottomNavBar({super.key, required this.selectedTab, required this.onTabSelected});

  static const List<_TabItem> tabs = [
    _TabItem('首页', Icons.home_outlined, Icons.home),
    _TabItem('健康', Icons.monitor_heart_outlined, Icons.monitor_heart),
    _TabItem('告警', Icons.notifications_none_rounded, Icons.notifications_active),
    _TabItem('用药', Icons.medication_outlined, Icons.medication),
    _TabItem('我的', Icons.person_outline, Icons.person),
  ];

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.only(top: 10),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        border: Border(top: BorderSide(color: const Color(0xFFF0F0F5))),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceAround,
        children: tabs.map((tab) => _NavItem(tab)).toList(),
      ),
    );
  }
}

class _TabItem {
  final String label;
  final IconData inactiveIcon;
  final IconData activeIcon;
  const _TabItem(this.label, this.inactiveIcon, this.activeIcon);
}

class _NavItem extends StatelessWidget {
  final _TabItem tab;
  const _NavItem(this.tab, {super.key});

  @override
  Widget build(BuildContext context) {
    final parent = context.findAncestorWidgetOfExactType<BottomNavBar>();
    final isSelected = parent?.selectedTab == BottomNavBar.tabs.indexOf(tab);
    final idx = BottomNavBar.tabs.indexOf(tab);

    return Expanded(
      child: GestureDetector(
        onTap: () => parent?.onTabSelected(idx),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            if (idx == 0 && isSelected)
              Container(
                width: 48,
                height: 48,
                margin: const EdgeInsets.only(bottom: -24),
                decoration: BoxDecoration(
                  color: AppTheme.primary,
                  shape: BoxShape.circle,
                  boxShadow: [BoxShadow(color: AppTheme.primary.withValues(alpha:0.3), blurRadius: 12)],
                ),
                child: Icon(tab.activeIcon, color: Colors.white, size: 22),
              )
            else
              Icon(isSelected ? tab.activeIcon : tab.inactiveIcon,
                  size: 22, color: isSelected ? AppTheme.primary : const Color(0xFF999999)),
            Text(tab.label, style: TextStyle(fontSize: 10, color: isSelected ? AppTheme.primary : const Color(0xFF999999))),
          ],
        ),
      ),
    );
  }
}

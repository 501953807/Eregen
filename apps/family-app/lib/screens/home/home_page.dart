import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../widgets/elderly_selector.dart';
import '../widgets/map_section.dart';
import '../widgets/quick_status_card.dart';
import '../widgets/sos_button.dart';
import '../widgets/recent_alerts_list.dart';
import '../widgets/bottom_nav_bar.dart';

/// Home page — matches home.html prototype
class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  int _selectedIndex = 0;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: CustomScrollView(
          slivers: [
            // Header with gradient background
            SliverToBoxAdapter(
              child: Container(
                decoration: const BoxDecoration(gradient: AppTheme.headerGradient),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // Status bar + header
                    Padding(
                      padding: const EdgeInsets.fromLTRB(20, 8, 20, 0),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            '颐贞',
                            style: TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.w700,
                              color: Colors.white,
                              letterSpacing: 1,
                            ),
                          ),
                          Row(
                            children: [
                              _headerIcon(Icons.notifications_outlined, badge: 3),
                              const SizedBox(width: 8),
                              _headerIcon(Icons.settings_outlined),
                            ],
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 12),
                    // Elderly selector
                    const ElderlySelector(
                      name: '李秀英 奶奶',
                      isOnline: true,
                      lastUpdate: '2分钟前',
                    ),
                    const SizedBox(height: 8),
                  ],
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 8)),

            // Map section
            const SliverToBoxAdapter(child: MapSection()),
            const SliverToBoxAdapter(child: SizedBox(height: 16)),

            // Quick status cards
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Row(
                  children: [
                    Expanded(
                      child: QuickStatusCard(
                        icon: Icons.favorite,
                        value: '72',
                        label: '心率 bpm',
                        status: '正常',
                        statusColor: AppTheme.statusNormal,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: QuickStatusCard(
                        icon: Icons.air,
                        value: '97%',
                        label: '血氧 SpO2',
                        status: '正常',
                        statusColor: AppTheme.statusNormal,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: QuickStatusCard(
                        icon: Icons.directions_walk,
                        value: '3,456',
                        label: '今日步数',
                        status: '偏少',
                        statusColor: AppTheme.statusWarning,
                      ),
                    ),
                    const SizedBox(width: 10),
                    Expanded(
                      child: QuickStatusCard(
                        icon: Icons.battery_full,
                        value: '85%',
                        label: '手环电量',
                        status: '充足',
                        statusColor: AppTheme.statusNormal,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SliverToBoxAdapter(child: SizedBox(height: 16)),

            // SOS button
            const SliverToBoxAdapter(child: SOSButton()),
            const SliverToBoxAdapter(child: SizedBox(height: 16)),

            // Recent alerts
            SliverToBoxAdapter(
              child: Padding(
                padding: const EdgeInsets.symmetric(horizontal: 20),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('最近告警',
                            style: TextStyle(
                                fontSize: 16, fontWeight: FontWeight.w700)),
                        GestureDetector(
                          onTap: () {},
                          child: const Text('查看全部 ›',
                              style: TextStyle(
                                  fontSize: 12,
                                  color: AppTheme.primary,
                                  fontWeight: FontWeight.w600)),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    RecentAlertsList(),
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
        onTabSelected: (index) {
          setState(() => _selectedIndex = index);
        },
      ),
    );
  }

  Widget _headerIcon(IconData icon, {int? badge}) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        Container(
          width: 36,
          height: 36,
          decoration: BoxDecoration(
            color: Colors.white.withOpacity(0.2),
            shape: BoxShape.circle,
          ),
          child: Icon(icon, size: 18, color: Colors.white),
        ),
        if (badge != null && badge > 0)
          Positioned(
            right: -2,
            top: -2,
            child: Container(
              width: 16,
              height: 16,
              decoration: const BoxDecoration(
                color: Color(0xFFFF6B6B),
                shape: BoxShape.circle,
              ),
              child: Center(
                child: Text(
                  '$badge',
                  style: const TextStyle(
                    fontSize: 9,
                    color: Colors.white,
                    fontWeight: FontWeight.w700,
                  ),
                ),
              ),
            ),
          ),
      ],
    );
  }
}

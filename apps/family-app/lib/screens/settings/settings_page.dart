import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../login/login_page.dart';
import '../bind-device/bind_device_page.dart';
import 'elderly_detail_page.dart';

/// Profile & settings page — account info, elderly management, app settings.
class SettingsPage extends StatefulWidget {
  const SettingsPage({super.key});

  @override
  State<SettingsPage> createState() => _SettingsPageState();
}

class _SettingsPageState extends State<SettingsPage> {
  bool _notificationsEnabled = true;
  bool _autoRefresh = true;
  String _selectedElderly = '李秀英 奶奶';

  // Firmware version check
  List<Map<String, dynamic>> _devices = [];
  Map<String, String?> _latestVersions = {}; // device_id -> latest version
  bool _checkingFirmware = false;
  String? _otaJobId;

  @override
  void initState() {
    super.initState();
    _checkFirmwareVersions();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: CustomScrollView(
        slivers: [
          // Header
          SliverToBoxAdapter(
            child: Container(
              decoration: const BoxDecoration(gradient: AppTheme.headerGradient),
              padding: const EdgeInsets.fromLTRB(20, 16, 20, 24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text('我的', style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700, color: Colors.white)),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      CircleAvatar(radius: 28, backgroundColor: Colors.white.withValues(alpha: 0.3), child: Icon(Icons.person, size: 28, color: Colors.white)),
                      const SizedBox(width: 12),
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text('张先生', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600, color: Colors.white)),
                          Text('138****5678', style: TextStyle(fontSize: 12, color: Colors.white.withValues(alpha: 0.8))),
                        ],
                      ),
                      const Spacer(),
                      IconButton(icon: const Icon(Icons.qr_code_2_outlined, color: Colors.white), onPressed: () {}),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 16)),

          // Elderly selector
          SliverToBoxAdapter(
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 20),
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14)),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text('关联老人', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600)),
                      GestureDetector(
                        onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const AddElderlyPage())),
                        child: const Text('+ 添加', style: TextStyle(fontSize: 13, color: AppTheme.primary, fontWeight: FontWeight.w600)),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  ListTile(
                    leading: const CircleAvatar(radius: 20, backgroundColor: Color(0xFFE3F2FD), child: Icon(Icons.person, size: 20, color: AppTheme.primary)),
                    title: Text(_selectedElderly, style: const TextStyle(fontWeight: FontWeight.w600)),
                    subtitle: const Text('中风险 · 在线'),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFFCCCCCC)),
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => const ElderlyDetailPage())),
                  ),
                  const Divider(height: 1),
                  ListTile(
                    leading: const CircleAvatar(radius: 20, backgroundColor: Color(0xFFF3E5F5), child: Icon(Icons.person_outline, size: 20, color: Color(0xFF9C27B0))),
                    title: const Text('王德明 爷爷'),
                    subtitle: const Text('高风险 · 离线'),
                    trailing: const Icon(Icons.chevron_right, color: Color(0xFFCCCCCC)),
                    onTap: () => setState(() => _selectedElderly = '王德明 爷爷'),
                  ),
                ],
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 16)),

          // Settings sections
          SliverToBoxAdapter(
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 20),
              decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14)),
              child: Column(
                children: [
                  _SettingsRow(
                    icon: Icons.notifications_active,
                    title: '推送通知',
                    trailing: Switch(value: _notificationsEnabled, onChanged: (v) => setState(() => _notificationsEnabled = v)),
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.refresh,
                    title: '自动刷新',
                    subtitle: '每30秒更新数据',
                    trailing: Switch(value: _autoRefresh, onChanged: (v) => setState(() => _autoRefresh = v)),
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.lock_outline,
                    title: '修改密码',
                    onTap: () {},
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.phone_android,
                    title: '绑定设备',
                    subtitle: '已绑定 2 台设备',
                    onTap: () => Navigator.of(context).push(MaterialPageRoute(builder: (_) => BindDevicePage(onBound: () {}))),
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.system_update,
                    title: '固件版本',
                    subtitle: _checkingFirmware ? '检查中...' : (_latestVersions.isEmpty ? '未绑定设备' : '已是最新'),
                    trailing: _checkingFirmware
                        ? SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2))
                        : const Icon(Icons.chevron_right, color: Color(0xFFCCCCCC)),
                    onTap: _checkingFirmware ? null : () => _checkFirmwareVersions(),
                  ),
                ],
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 16)),

          // Support section
          SliverToBoxAdapter(
            child: Container(
              margin: const EdgeInsets.symmetric(horizontal: 20),
              decoration: BoxDecoration(color: AppTheme.bgCard, borderRadius: BorderRadius.circular(14)),
              child: Column(
                children: [
                  _SettingsRow(
                    icon: Icons.help_outline,
                    title: '帮助中心',
                    onTap: () {},
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.star_outline,
                    title: '给我们评分',
                    onTap: () {},
                  ),
                  const Divider(height: 1),
                  _SettingsRow(
                    icon: Icons.info_outline,
                    title: '关于颐贞',
                    subtitle: '版本 v1.0.0 (2026.07)',
                    onTap: () {},
                  ),
                ],
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 24)),

          // Logout button
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: 20),
              child: SizedBox(
                width: double.infinity,
                height: 48,
                child: ElevatedButton.icon(
                  onPressed: () => _showLogoutDialog(context),
                  icon: const Icon(Icons.logout, size: 18),
                  label: const Text('退出登录', style: TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: AppTheme.statusDanger)),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.white,
                    elevation: 0,
                    side: BorderSide(color: AppTheme.statusDanger.withValues(alpha: 0.3)),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                ),
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 24)),
        ],
      ),
    );
  }

  Future<void> _checkFirmwareVersions() async {
    setState(() => _checkingFirmware = true);
    try {
      // Fetch bound devices
      final devicesResp = await ApiClient.instance.get('/devices');
      final devicesList = (devicesResp.data as List)?.map((d) => {
        'id': d['device_id'] as String? ?? '',
        'type': d['device_type'] as String? ?? '',
        'tier': d['tier'] as String? ?? '',
        'fw_version': d['fw_version'] as String? ?? '',
      }).toList() ?? [];
      setState(() => _devices = devicesList);

      // Check latest firmware for each device type+tier combo
      final versions = <String, String?>{};
      for (final dev in devicesList) {
        final key = '${dev['type']}/${dev['tier']}';
        if (versions.containsKey(key)) continue;
        try {
          final fwResp = await ApiClient.instance.listFirmware(
            deviceType: dev['type'] as String?,
            tier: dev['tier'] as String?,
          );
          final items = (fwResp.data as Map<String, dynamic>)['data'] as List?;
          if (items != null && items.isNotEmpty) {
            versions[key] = items.first['version'] as String?;
          }
        } catch (_) {
          // skip
        }
      }
      setState(() {
        _latestVersions = versions;
        _checkingFirmware = false;
      });
    } catch (_) {
      setState(() => _checkingFirmware = false);
    }
  }

  void _showLogoutDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('确认退出'),
        content: const Text('确定要退出登录吗？'),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('取消')),
          ElevatedButton(
            onPressed: () {
              Navigator.pop(ctx);
              ApiClient.instance.clearAuth();
              Navigator.of(context).pushAndRemoveUntil(
                MaterialPageRoute(builder: (_) => LoginPage(onLoginSuccess: () {})),
                (_) => false,
              );
            },
            style: ElevatedButton.styleFrom(backgroundColor: AppTheme.statusDanger),
            child: const Text('退出', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }
}

class _SettingsRow extends StatelessWidget {
  final IconData icon;
  final String title;
  final String? subtitle;
  final Widget? trailing;
  final VoidCallback? onTap;

  const _SettingsRow({required this.icon, required this.title, this.subtitle, this.trailing, this.onTap});

  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: CircleAvatar(radius: 16, backgroundColor: const Color(0xFFF5F6FA), child: Icon(icon, size: 18, color: AppTheme.primary)),
      title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
      subtitle: subtitle != null ? Text(subtitle!, style: const TextStyle(fontSize: 11, color: Color(0xFF999999))) : null,
      trailing: trailing ?? const Icon(Icons.chevron_right, color: Color(0xFFCCCCCC)),
      onTap: onTap,
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
    );
  }
}

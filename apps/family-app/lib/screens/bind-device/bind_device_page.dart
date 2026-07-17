import 'package:flutter/material.dart';
import '../../common/theme.dart';
import '../../api/client.dart';

/// Manual device binding page — enter BR-XXXX / PX-XXXX device ID.
class BindDevicePage extends StatefulWidget {
  final VoidCallback onBound;
  const BindDevicePage({super.key, required this.onBound});

  @override
  State<BindDevicePage> createState() => _BindDevicePageState();
}

class _BindDevicePageState extends State<BindDevicePage> {
  final _ctrl = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  // Regex: BR- or PX- followed by exactly 4 alphanumeric chars
  static final _deviceRe = RegExp(r'^(BR|PX)-[A-Za-z0-9]{4}$');

  Future<void> _bind() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() { _loading = true; _error = null; });
    try {
      await ApiClient.instance.post('/devices/bind', data: {
        'device_id': _ctrl.text.trim().toUpperCase(),
      });
      if (!mounted) return;
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (ctx) => AlertDialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          icon: const Icon(Icons.check_circle, color: AppTheme.statusNormal, size: 56),
          title: const Text('绑定成功'),
          content: Text('设备 ${_ctrl.text.trim().toUpperCase()} 已成功绑定'),
          actions: [
            TextButton(
              onPressed: () {
                Navigator.of(ctx).pop();
                widget.onBound();
              },
              child: const Text('确定'),
            ),
          ],
        ),
      );
    } catch (e) {
      setState(() => _error = '绑定失败，请检查设备ID是否正确');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      appBar: AppBar(
        title: const Text('绑定设备'),
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              Container(
                padding: const EdgeInsets.all(16),
                decoration: BoxDecoration(
                  color: const Color(0xFFE3F2FD),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Row(
                  children: [
                    Icon(Icons.info_outline, color: Color(0xFF1565C0)),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        '请输入设备背面的设备编号。手环格式为 BR-XXXX，药盒格式为 PX-XXXX，X 为字母或数字。',
                        style: TextStyle(fontSize: 13, color: Color(0xFF1565C0)),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),

              TextFormField(
                controller: _ctrl,
                textCapitalization: TextCapitalization.characters,
                enabled: !_loading,
                decoration: InputDecoration(
                  labelText: '设备编号',
                  hintText: '例：BR-A1B2 或 PX-C3D4',
                  prefixIcon: const Icon(Icons.qr_code_scanner_outlined),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(12)),
                ),
                validator: (v) {
                  if (v == null || v.trim().isEmpty) return '请输入设备编号';
                  if (!_deviceRe.hasMatch(v.trim()))
                    return '格式不正确，应为 BR-XXXX 或 PX-XXXX';
                  return null;
                },
              ),
              const SizedBox(height: 8),

              if (_error != null)
                Container(
                  padding: const EdgeInsets.all(12),
                  margin: const EdgeInsets.only(bottom: 16),
                  decoration: BoxDecoration(
                    color: const Color(0xFFFFEBEE),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(_error!, style: const TextStyle(color: Color(0xFFC62828), fontSize: 13)),
                ),

              SizedBox(
                height: 48,
                child: ElevatedButton(
                  onPressed: _loading ? null : _bind,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppTheme.primary,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                  ),
                  child: _loading
                      ? const SizedBox(
                          height: 20, width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2, valueColor: AlwaysStoppedAnimation<Color>(Colors.white)),
                        )
                      : const Text('绑 定', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

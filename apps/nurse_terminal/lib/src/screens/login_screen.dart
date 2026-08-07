import 'package:flutter/material.dart';
import '../services/api_client.dart';
import 'home_screen.dart';

/// Login screen for the nurse terminal.
/// Supports two auth modes:
/// 1. Admin API JWT login (legacy)
/// 2. Hospital API key login (B2B institutional access)
class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _apiKeyController = TextEditingController();
  final _institutionIdController = TextEditingController();

  // Auth mode: 'admin' for JWT login, 'hospital' for API key login
  String _authMode = 'hospital';
  bool _loading = false;

  final ApiClient _adminApi = ApiClient();
  final HospitalApiClient _hospitalApi = HospitalApiClient();

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _apiKeyController.dispose();
    _institutionIdController.dispose();
    super.dispose();
  }

  Future<void> _handleAdminLogin() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);
    try {
      final result = await _adminApi.post('/api/v1/admin/login', {
        'username': _usernameController.text,
        'password': _passwordController.text,
      });
      final token = result['data']?['token'] ?? result['token'] ?? '';
      if (token.toString().isNotEmpty) {
        await _adminApi.saveToken(token);
        if (mounted) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const HomeScreen()),
          );
        }
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('登录失败：未获取到令牌')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('登录失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  Future<void> _handleHospitalLogin() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() => _loading = true);
    try {
      // Verify the API key is valid by calling the institution endpoint
      await _hospitalApi.get(
        '/api/v2/b2b/institutions/${_institutionIdController.text}',
      );
      await _hospitalApi.saveCredentials(
        _apiKeyController.text,
        _institutionIdController.text,
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('医院API连接成功')),
        );
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(builder: (_) => const HomeScreen()),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('API Key验证失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            colors: [Color(0xFF2563EB), Color(0xFF7C3AED)],
          ),
        ),
        child: Center(
          child: Card(
            elevation: 8,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(16),
            ),
            child: Padding(
              padding: const EdgeInsets.all(32),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.medical_services, size: 64, color: Colors.blue),
                    const SizedBox(height: 16),
                    const Text(
                      '颐贞 护士终端',
                      style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),
                    const Text(
                      '医院医疗腕带管理系统',
                      style: TextStyle(color: Colors.grey),
                    ),
                    const SizedBox(height: 24),

                    // Auth mode toggle
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        ChoiceChip(
                          label: const Text('医院 API Key'),
                          selected: _authMode == 'hospital',
                          onSelected: (_) => setState(() => _authMode = 'hospital'),
                          selectedColor: Colors.blue.shade100,
                        ),
                        const SizedBox(width: 8),
                        ChoiceChip(
                          label: const Text('管理员账号'),
                          selected: _authMode == 'admin',
                          onSelected: (_) => setState(() => _authMode = 'admin'),
                          selectedColor: Colors.blue.shade100,
                        ),
                      ],
                    ),
                    const SizedBox(height: 24),

                    // Hospital API Key form
                    if (_authMode == 'hospital') ...[
                      TextFormField(
                        controller: _apiKeyController,
                        decoration: const InputDecoration(
                          labelText: 'API Key',
                          prefixIcon: Icon(Icons.key),
                          border: OutlineInputBorder(),
                        ),
                        validator: (v) => v?.isEmpty == true ? '请输入 API Key' : null,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _institutionIdController,
                        decoration: const InputDecoration(
                          labelText: '机构 ID',
                          prefixIcon: Icon(Icons.business),
                          border: OutlineInputBorder(),
                        ),
                        validator: (v) => v?.isEmpty == true ? '请输入机构 ID' : null,
                      ),
                    ],

                    // Admin login form
                    if (_authMode == 'admin') ...[
                      TextFormField(
                        controller: _usernameController,
                        decoration: const InputDecoration(
                          labelText: '用户名',
                          prefixIcon: Icon(Icons.person),
                          border: OutlineInputBorder(),
                        ),
                        validator: (v) => v?.isEmpty == true ? '请输入用户名' : null,
                      ),
                      const SizedBox(height: 16),
                      TextFormField(
                        controller: _passwordController,
                        obscureText: true,
                        decoration: const InputDecoration(
                          labelText: '密码',
                          prefixIcon: Icon(Icons.lock),
                          border: OutlineInputBorder(),
                        ),
                        validator: (v) => v?.isEmpty == true ? '请输入密码' : null,
                      ),
                    ],

                    const SizedBox(height: 24),
                    SizedBox(
                      width: double.infinity,
                      height: 48,
                      child: ElevatedButton(
                        onPressed: _loading ? null :
                          _authMode == 'hospital' ? _handleHospitalLogin : _handleAdminLogin,
                        child: _loading
                            ? const CircularProgressIndicator()
                            : Text(_authMode == 'hospital' ? '连接医院API' : '登录'),
                      ),
                    ),
                    const SizedBox(height: 16),
                    Text(
                      _authMode == 'hospital'
                          ? '使用医院机构 API Key 接入 B2B 服务'
                          : '使用管理员账号登录管理后台',
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

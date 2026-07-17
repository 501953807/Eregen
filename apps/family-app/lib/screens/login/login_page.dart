import 'dart:async';

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../../common/theme.dart';
import '../../api/client.dart';
import '../../app_state.dart';

/// Phone + OTP login page with 60s countdown and form validation.
class LoginPage extends StatefulWidget {
  final VoidCallback onLoginSuccess;
  const LoginPage({super.key, required this.onLoginSuccess});

  @override
  State<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends State<LoginPage> {
  final _phoneCtrl = TextEditingController();
  final _otpCtrl = TextEditingController();
  final _formKey = GlobalKey<FormState>();
  bool _loading = false;
  int _countdown = 0;
  Timer? _timer;
  String? _error;

  @override
  void dispose() {
    _timer?.cancel();
    _phoneCtrl.dispose();
    _otpCtrl.dispose();
    super.dispose();
  }

  bool get _canSendOtp => _countdown == 0 && _phoneCtrl.text.trim().isNotEmpty;

  Future<void> _sendOtp() async {
    if (_phoneCtrl.text.trim().length != 11) {
      _showError('请输入正确的11位手机号');
      return;
    }
    try {
      await ApiClient.instance.sendOtp(_phoneCtrl.text.trim());
      setState(() => _countdown = 60);
      _startCountdown();
    } catch (e) {
      _showError('发送验证码失败，请重试');
    }
  }

  void _startCountdown() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (!mounted) return t.cancel();
      setState(() {
        _countdown--;
        if (_countdown <= 0) t.cancel();
      });
    });
  }

  Future<void> _login() async {
    if (!_formKey.currentState!.validate()) return;
    setState(() { _loading = true; _error = null; });
    try {
      final result = await ApiClient.instance.login(
        phone: _phoneCtrl.text.trim(),
        otp: _otpCtrl.text.trim(),
      );
      // Store user ID in global app state for WebSocket connection
      final userId = result['user_id'] as String? ?? result['id'] as String?;
      if (userId != null && mounted) {
        context.read<AppState>().setAuth(userId: userId);
      }
      if (mounted) widget.onLoginSuccess();
    } catch (e) {
      _showError('登录失败，请检查验证码');
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  void _showError(String msg) {
    setState(() => _error = msg);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.bgScaffold,
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 60),
                // Logo / brand
                const Text(
                  '颐贞',
                  textAlign: TextAlign.center,
                  style: TextStyle(
                    fontSize: 32, fontWeight: FontWeight.w800,
                    color: AppTheme.primary, letterSpacing: 2,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  '家属端 · 守护每一位长者',
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 14, color: Color(0xFF888888)),
                ),
                const SizedBox(height: 48),

                // Phone field
                TextFormField(
                  controller: _phoneCtrl,
                  keyboardType: TextInputType.phone,
                  maxLength: 11,
                  enabled: !_loading,
                  decoration: InputDecoration(
                    labelText: '手机号码',
                    hintText: '请输入11位手机号',
                    prefixIcon: const Icon(Icons.phone_outlined),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    counterText: '',
                  ),
                  validator: (v) {
                    if (v == null || v.trim().length != 11)
                      return '请输入正确的11位手机号';
                    return null;
                  },
                ),
                const SizedBox(height: 16),

                // OTP field + send button row
                Row(
                  children: [
                    Expanded(
                      child: TextFormField(
                        controller: _otpCtrl,
                        keyboardType: TextInputType.number,
                        maxLength: 6,
                        enabled: !_loading,
                        decoration: InputDecoration(
                          labelText: '验证码',
                          hintText: '输入6位数字',
                          prefixIcon: const Icon(Icons.security_outlined),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                          counterText: '',
                        ),
                        validator: (v) {
                          if (v == null || v.trim().length != 6)
                            return '请输入6位验证码';
                          return null;
                        },
                      ),
                    ),
                    const SizedBox(width: 12),
                    SizedBox(
                      width: 100,
                      child: ElevatedButton(
                        onPressed: _canSendOtp ? _sendOtp : null,
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppTheme.primary,
                          foregroundColor: Colors.white,
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                        ),
                        child: _countdown > 0
                            ? Text('$_countdown s', style: const TextStyle(fontSize: 13))
                            : const Text('获取验证码', style: TextStyle(fontSize: 13)),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),

                // Error message
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

                // Login button
                SizedBox(
                  height: 48,
                  child: ElevatedButton(
                    onPressed: _loading ? null : _login,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppTheme.primary,
                      foregroundColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2, valueColor: AlwaysStoppedAnimation<Color>(Colors.white)),
                          )
                        : const Text('登 录', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w700)),
                  ),
                ),
                const SizedBox(height: 24),
                Text(
                  '登录即表示同意《用户协议》和《隐私政策》',
                  textAlign: TextAlign.center,
                  style: const TextStyle(fontSize: 11, color: Color(0xFFBBBBBB)),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

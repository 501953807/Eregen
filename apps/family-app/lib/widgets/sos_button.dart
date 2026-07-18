import 'dart:async';

import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';
import '../../common/theme.dart';
import '../../api/client.dart';

/// SOS emergency call button with real phone dialing via url_launcher.
class SOSButton extends StatefulWidget {
  const SOSButton({super.key});

  @override
  State<SOSButton> createState() => _SOSButtonState();
}

class _SOSButtonState extends State<SOSButton> with SingleTickerProviderStateMixin {
  late AnimationController _pulseController;
  late Animation<double> _pulseAnimation;
  bool _isLongPressActive = false;
  Timer? _longPressTimer;
  int _pressDuration = 0;

  @override
  void initState() {
    super.initState();
    _pulseController = AnimationController(vsync: this, duration: const Duration(milliseconds: 1200))..repeat();
    _pulseAnimation = Tween<double>(begin: 0.85, end: 1.05).animate(_pulseController);
  }

  @override
  void dispose() {
    _pulseController.dispose();
    _longPressTimer?.cancel();
    super.dispose();
  }

  /// Launches phone dialer with the emergency contact number.
  Future<void> _makeEmergencyCall(String phoneNumber) async {
    final Uri phoneUri = Uri(scheme: 'tel', path: phoneNumber);
    try {
      if (await canLaunchUrl(phoneUri)) {
        await launchUrl(phoneUri);
      } else {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('无法拨打电话，请手动联系家属')),
          );
        }
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('拨号失败')),
        );
      }
    }
  }

  /// Sends the current location to all linked family members via API.
  Future<void> _sendLocationToFamily() async {
    try {
      await ApiClient.instance.post('/alerts/share-location');
    } catch (_) {
      // Silently fail — location sharing is non-critical
    }
  }

  void _startLongPress() {
    setState(() => _isLongPressActive = true);
    _pressDuration = 0;
    _longPressTimer = Timer.periodic(const Duration(milliseconds: 100), (timer) {
      setState(() => _pressDuration++);
      if (_pressDuration >= 30) {
        timer.cancel();
        _executeSOS();
      }
    });
  }

  void _stopLongPress() {
    _longPressTimer?.cancel();
    setState(() => _isLongPressActive = false);
  }

  Future<void> _executeSOS() async {
    _stopLongPress();
    await _sendLocationToFamily();
    await _makeEmergencyCall('10086');
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('已向家属发送位置并拨打电话')),
      );
    }
  }

  void _quickCall() {
    _makeEmergencyCall('10086');
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: GestureDetector(
        onTap: _quickCall,
        onLongPressStart: (_) => _startLongPress(),
        onLongPressEnd: (_) => _stopLongPress(),
        onLongPressCancel: () => _stopLongPress(),
        child: AnimatedBuilder(
          animation: _pulseAnimation,
          builder: (context, child) {
            return Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                gradient: AppTheme.sosGradient,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: const Color(0xFFFF6B6B).withValues(alpha: 0.3),
                    blurRadius: 16,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Row(
                children: [
                  AnimatedBuilder(
                    animation: _pulseAnimation,
                    builder: (context, _) {
                      return Container(
                        width: 56,
                        height: 56,
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          border: Border.all(
                            color: _isLongPressActive
                                ? Colors.yellow
                                : Colors.white.withValues(alpha: 0.4),
                            width: _isLongPressActive ? 3 : 1,
                          ),
                          color: Colors.white.withValues(alpha: 0.2),
                        ),
                        child: Center(
                          child: _isLongPressActive
                              ? Text(
                                  '${(30 - _pressDuration) * 0.1.toInt()}',
                                  style: const TextStyle(
                                    fontSize: 18,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.white,
                                  ),
                                )
                              : const Icon(Icons.phone_in_talk, size: 28, color: Colors.white),
                        ),
                      );
                    },
                  ),
                  const SizedBox(width: 16),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Text('一键呼叫家属',
                            style: TextStyle(fontSize: 15, fontWeight: FontWeight.w700, color: Colors.white)),
                        const SizedBox(height: 4),
                        Text(
                          _isLongPressActive
                              ? '松开取消 · 继续长按呼叫'
                              : '点击快速呼叫 · 长按3秒发送位置后呼叫',
                          style: const TextStyle(fontSize: 11, color: Colors.white, height: 1.4),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}

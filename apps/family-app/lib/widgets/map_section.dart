import 'package:flutter/material.dart';
import '../../common/theme.dart';

/// Map section showing elderly location — matches home.html prototype
class MapSection extends StatelessWidget {
  const MapSection({super.key});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 20, right: 20, top: 0),
      child: Container(
        height: 200,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          gradient: const LinearGradient(
            colors: [Color(0xFFE8F4FD), Color(0xFFD1ECF9)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF4A90D9).withOpacity(0.15),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Stack(
          children: [
            // Grid lines
            Positioned.fill(
              child: CustomPaint(painter: _GridPainter()),
            ),
            // Geofence ring
            Positioned(
              left: 100,
              top: 70,
              child: SizedBox(
                width: 80,
                height: 80,
                child: CustomPaint(painter: _GeofenceRingPainter()),
              ),
            ),
            // Location pin
            Positioned(
              left: 108,
              top: 50,
              child: TweenAnimationBuilder<double>(
                duration: const Duration(seconds: 2),
                tween: Tween<double>(begin: 0, end: 1),
                builder: (context, value, _) {
                  return Transform.translate(
                    offset: Offset(0, -6 * (1 - value)),
                    child: Transform.scale(
                      scale: 1.0,
                      child: Container(
                        width: 36,
                        height: 44,
                        decoration: BoxDecoration(
                          color: AppTheme.primary,
                          shape: BoxShape.circle,
                          boxShadow: [
                            BoxShadow(
                              color: const Color(0xFF4A90D9).withOpacity(0.4),
                              blurRadius: 12,
                            ),
                          ],
                        ),
                        child: const Center(child: Text('👵', style: TextStyle(fontSize: 14))),
                      ),
                    ),
                  );
                },
              ),
            ),
            // Location info card at bottom
            Positioned(
              left: 12,
              right: 12,
              bottom: 12,
              child: Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.95),
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [
                    BoxShadow(color: Colors.black.withOpacity(0.08), blurRadius: 8),
                  ],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('📍 上海市浦东新区陆家嘴环路1000号',
                        style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    const Text('安全区域：电子围栏半径 500m',
                        style: TextStyle(fontSize: 11, color: Color(0xFF666666))),
                    const Text('更新时间：2026-07-16 09:39:22',
                        style: TextStyle(fontSize: 10, color: Color(0xFF999999))),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _GridPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFF4A90D9).withOpacity(0.15)
      ..strokeWidth = 0.5;
    for (double x = 0; x < size.width; x += 30) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }
    for (double y = 0; y < size.height; y += 30) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

class _GeofenceRingPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = const Color(0xFF4A90D9).withOpacity(0.4)
      ..strokeWidth = 2
      ..style = PaintingStyle.stroke;
    canvas.drawCircle(size.center(Offset.zero), size.width / 2, paint);
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}

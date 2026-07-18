import 'package:flutter/material.dart';
/// Elderly selector card shown under app header
class ElderlySelector extends StatelessWidget {
  final String name;
  final bool isOnline;
  final String lastUpdate;
  const ElderlySelector({
    super.key,
    required this.name,
    this.isOnline = true,
    this.lastUpdate = '2分钟前',
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 20),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: Colors.white.withValues(alpha:0.15),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: const Color(0xFFFFD93D),
                shape: BoxShape.circle,
                border: Border.all(color: Colors.white.withValues(alpha:0.5), width: 2),
              ),
              child: const Center(child: Text('👵', style: TextStyle(fontSize: 20))),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(name, style: const TextStyle(fontSize: 15, fontWeight: FontWeight.w600, color: Colors.white)),
                  const SizedBox(height: 2),
                  Row(
                    children: [
                      Container(
                        width: 8,
                        height: 8,
                        decoration: BoxDecoration(
                          color: isOnline ? const Color(0xFF4ADE80) : const Color(0xFF9CA3AF),
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Text('设备${isOnline ? "在线" : "离线"} · 最后更新 $lastUpdate',
                          style: TextStyle(fontSize: 11, color: Colors.white.withValues(alpha:0.9))),
                    ],
                  ),
                ],
              ),
            ),
            Icon(Icons.arrow_drop_down, color: Colors.white.withValues(alpha:0.7)),
          ],
        ),
      ),
    );
  }
}

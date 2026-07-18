import 'package:flutter/material.dart';
import '../../common/theme.dart';

/// Quick status card — heart rate, SpO2, steps, battery
class QuickStatusCard extends StatelessWidget {
  final IconData icon;
  final String value;
  final String label;
  final String status;
  final Color statusColor;

  const QuickStatusCard({
    super.key,
    required this.icon,
    required this.value,
    required this.label,
    required this.status,
    required this.statusColor,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(vertical: 14, horizontal: 8),
      decoration: BoxDecoration(
        color: AppTheme.bgCard,
        borderRadius: BorderRadius.circular(AppTheme.cardRadius),
        border: Border.all(color: const Color(0xFFF0F0F5)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.02),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        children: [
          Icon(icon, size: 24, color: AppTheme.primary),
          const SizedBox(height: 6),
          Text(value, style: const TextStyle(fontSize: 18, fontWeight: FontWeight.w700)),
          Text(label, style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
          const SizedBox(height: 4),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
            decoration: BoxDecoration(
              color: statusColor.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(8),
            ),
            child: Text(status,
                style: TextStyle(fontSize: 9, fontWeight: FontWeight.w600, color: statusColor)),
          ),
        ],
      ),
    );
  }
}

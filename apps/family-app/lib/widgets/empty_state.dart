import 'package:flutter/material.dart';
import '../../common/theme.dart';

/// Reusable empty/error state widget used across all pages.
/// Displays a large muted icon, title, subtitle, and optional retry action.
class EmptyState extends StatelessWidget {
  final IconData icon;
  final String title;
  final String subtitle;
  final String? buttonText;
  final VoidCallback? onRetry;

  const EmptyState({
    super.key,
    required this.icon,
    required this.title,
    required this.subtitle,
    this.buttonText,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: AppTheme.textMuted),
            const SizedBox(height: 16),
            Text(title, style: const TextStyle(fontSize: 15, color: AppTheme.textSecondary)),
            const SizedBox(height: 4),
            Text(subtitle, style: const TextStyle(fontSize: 12, color: AppTheme.textMuted), textAlign: TextAlign.center),
            if (buttonText != null && onRetry != null) ...[
              const SizedBox(height: 16),
              SizedBox(
                width: 120,
                height: 36,
                child: OutlinedButton(
                  onPressed: onRetry,
                  style: OutlinedButton.styleFrom(
                    side: BorderSide(color: AppTheme.primary),
                    shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                  ),
                  child: Text(buttonText!, style: const TextStyle(fontSize: 13, color: AppTheme.primary)),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

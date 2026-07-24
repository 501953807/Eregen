import 'package:flutter/material.dart';

/// Eregen 颐贞 brand theme — 暖色调品牌体系
class AppTheme {
  // Primary palette — warm orange brand
  static const Color primary = Color(0xFFE8734A);
  static const Color primaryDark = Color(0xFFD9622E);
  static const Color accent = Color(0xFFF59E0B);

  // Status colors
  static const Color statusNormal = Color(0xFF10B981);
  static const Color statusWarning = Color(0xFFF59E0B);
  static const Color statusDanger = Color(0xFFEF4444);
  static const Color statusInfo = Color(0xFF3B82F6);

  // Backgrounds
  static const Color bgScaffold = Color(0xFFFFF9F5);
  static const Color bgCard = Colors.white;

  // Header gradient — warm orange
  static const LinearGradient headerGradient = LinearGradient(
    colors: [Color(0xFFE8734A), Color(0xFFF59E0B)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // SOS gradient — warm red-orange
  static const LinearGradient sosGradient = LinearGradient(
    colors: [Color(0xFFEF4444), Color(0xFFD9622E)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // Card radius
  static const double cardRadius = 16.0;
  static const double smallRadius = 12.0;

  // Spacing
  static const double paddingHorizontal = 20.0;
  static const double paddingVertical = 16.0;

  // Font styles
  static const TextStyle pageTitle = TextStyle(
    fontSize: 18,
    fontWeight: FontWeight.w700,
    color: Color(0xFF1F2937),
  );

  static const TextStyle sectionTitle = TextStyle(
    fontSize: 16,
    fontWeight: FontWeight.w700,
    color: Color(0xFF1F2937),
  );

  static const TextStyle labelText = TextStyle(
    fontSize: 13,
    color: Color(0xFF6B7280),
  );

  static const TextStyle timeLabel = TextStyle(
    fontSize: 10,
    color: Color(0xFF9CA3AF),
  );
}

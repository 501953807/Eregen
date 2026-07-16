import 'package:flutter/material.dart';

/// Eregen brand theme — matches UI prototypes exactly
class AppTheme {
  // Primary palette
  static const Color primary = Color(0xFF4A90D9);
  static const Color primaryDark = Color(0xFF357ABD);
  static const Color accent = Color(0xFF4A90D9);

  // Status colors
  static const Color statusNormal = Color(0xFF4CAF50);
  static const Color statusWarning = Color(0xFFFFA726);
  static const Color statusDanger = Color(0xFFFF6B6B);
  static const Color statusInfo = Color(0xFF42A5F5);

  // Backgrounds
  static const Color bgScaffold = Color(0xFFF5F6FA);
  static const Color bgCard = Colors.white;

  // Header gradient
  static const LinearGradient headerGradient = LinearGradient(
    colors: [AppTheme.primary, AppTheme.primaryDark],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // SOS gradient
  static const LinearGradient sosGradient = LinearGradient(
    colors: [Color(0xFFFF6B6B), Color(0xFFEE5A24)],
    begin: Alignment.topLeft,
    end: Alignment.bottomRight,
  );

  // Card radius
  static const double cardRadius = 14.0;
  static const double smallRadius = 10.0;

  // Spacing
  static const double paddingHorizontal = 20.0;
  static const double paddingVertical = 16.0;

  // Font styles
  static const TextStyle pageTitle = TextStyle(
    fontSize: 18,
    fontWeight: FontWeight.w700,
    color: Color(0xFF333333),
  );

  static const TextStyle sectionTitle = TextStyle(
    fontSize: 16,
    fontWeight: FontWeight.w700,
    color: Color(0xFF333333),
  );

  static const TextStyle labelText = TextStyle(
    fontSize: 13,
    color: Color(0xFF888888),
  );

  static const TextStyle timeLabel = TextStyle(
    fontSize: 10,
    color: Color(0xFFBBBBBB),
  );
}

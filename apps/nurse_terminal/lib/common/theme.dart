import 'package:flutter/material.dart';

/// Eregen 颐贞 nurse terminal unified theme — Warm amber primary with consistent design system
class NurseTerminalTheme {
  // Primary brand color — single warm amber (unified across all platforms)
  static const Color primary = Color(0xFFE8734A);
  static const Color primaryDark = Color(0xFFD9622E);
  static const Color primaryLight = Color(0xFFFBEAF0);

  // Secondary neutral palette — single gray family (consistent with web and app)
  static const Color gray50 = Color(0xFAFAFA);
  static const Color gray100 = Color(0xF5F5F5);
  static const Color gray200 = Color(0xE5E5EB);
  static const Color gray300 = Color(0xD1D5DB);
  static const Color gray400 = Color(0x9CA3AF);
  static const Color gray500 = Color(0x6B7280);
  static const Color gray600 = Color(0x4B5563);
  static const Color gray700 = Color(0x374151);
  static const Color gray800 = Color(0x1F2937);
  static const Color gray900 = Color(0x111827);

  // Status colors (desaturated and harmonious)
  static const Color statusNormal = Color(0xFF10B981);
  static const Color statusWarning = Color(0xFFE8734A); // Same as primary
  static const Color statusDanger = Color(0xFFEF4444);
  static const Color statusInfo = Color(0xFF3B82F6);

  // Surface backgrounds
  static const Color bgScaffold = Color(0xFFF9F5); // Warm off-white
  static const Color bgCard = Color(0xFFFFFFFF);
  static const Color bgSurface = Color(0xFAFAFA);

  // Component radii — match web values
  static const double radiusSmallest = 6.0;
  static const double radiusSmall = 8.0;
  static const double radiusMedium = 12.0;
  static const double radiusLarge = 16.0;
  static const double radiusXL = 20.0;

  // Spacing scale
  static const double spacingXS = 4.0;
  static const double spacingS = 8.0;
  static const double spacingM = 16.0;
  static const double spacingL = 24.0;
  static const double spacingXL = 32.0;
  static const double spacingXXL = 48.0;

  // Typography: prefer system fonts with Noto Sans SC for Chinese clarity
  static const TextStyle headlineStyle = TextStyle(
    fontSize: 24.0,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.015,
    color: gray800,
    fontFamily: 'PingFang SC',
  );

  static const TextStyle titleStyle = TextStyle(
    fontSize: 18.0,
    fontWeight: FontWeight.w600,
    color: gray800,
    fontFamily: 'PingFang SC',
  );

  static const TextStyle bodyStyle = TextStyle(
    fontSize: 15.0,
    fontWeight: FontWeight.w400,
    color: gray700,
    lineHeight: 1.6,
    fontFamily: 'PingFang SC',
  );

  static const TextStyle labelStyle = TextStyle(
    fontSize: 14.0,
    fontWeight: FontWeight.w500,
    color: gray600,
    letterSpacing: 0.02,
    fontFamily: 'PingFang SC',
  );

  static const TextStyle smallLabelStyle = TextStyle(
    fontSize: 12.0,
    fontWeight: FontWeight.w400,
    color: gray500,
    letterSpacing: 0.04,
    fontFamily: 'PingFang SC',
  );

  // Button color scheme — upgraded hover/ripple effects
  static const ElevatedButtonThemeData elevatedButtonTheme = ElevatedButtonThemeData(
    style: ElevatedButton.styleFrom(
      foregroundColor: Colors.white,
      backgroundColor: primary,
      elevation: 0,
      shadowColor: Colors.transparent,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radiusMedium)),
      padding: EdgeInsets.symmetric(horizontal: 24.0, vertical: 12.0),
      textStyle: const TextStyle(fontWeight: FontWeight.w600),
    ),
  );

  static const TextButtonThemeData textButtonTheme = TextButtonThemeData(
    style: TextButton.styleFrom(
      foregroundColor: primary,
      padding: EdgeInsets.symmetric(horizontal: 16.0, vertical: 10.0),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(radiusMedium)),
    ),
  );

  static const IconButtonThemeData iconButtonTheme = IconButtonThemeData(
    padding: EdgeInsets.all(8.0),
    splashRadius: 24.0,
  );

  // Card decoration with subtle shadow and brand-aligned border
  static const BoxDecoration cardDecoration = BoxDecoration(
    color: bgCard,
    borderRadius: BorderRadius.circular(radiusLarge),
    boxShadow: [
      BoxShadow(
        color: Color(0x1A000000),
        blurRadius: 8.0,
        offset: Offset(0, 2),
      ),
    ],
  );

  // Loading indicator color — amber spinner matching brand
  static const Color loadingIndicatorColor = Color(0xFFE8734A);

  // Scaffold bottom padding for notched screens (iPhone X+)
  static const double safeBottomPadding = 34.0;

  // AppBar elevation override
  static const AppBar appBar = AppBar(
    elevation: 2,
    surfaceTintColor: Colors.transparent,
  );
}
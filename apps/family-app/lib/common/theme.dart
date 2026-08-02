import 'package:flutter/material.dart';

/// Eregen 颐贞 unified brand theme — Warm amber primary with consistent design system
class AppTheme {
  // Primary brand color — single warm amber (unified across all platforms)
  static const Color primary = Color(0xF59E0B);       /* #F59E0B, softened saturation */
  static const Color primaryDark = Color(0xD97706);   /* Darker amber for hover/deeper state */
  static const Color primaryLight = Color(0xFFFEF3C7); /* Very light amber background */

  // Secondary neutral palette — single gray family (warm-toned grays)
  static const Color gray50 = Color(0xFAFAFA);
  static const Color gray100 = Color(0xF5F5F5);
  static const Color gray200 = Color(0xE5E5EB);
  static const Color gray300 = Color(0xD1D5DB);
  static const Color gray400 = Color(0xB3B3B3);
  static const Color gray500 = Color(0x6B7280);
  static const Color gray600 = Color(0x4B5563);
  static const Color gray700 = Color(0x374151);
  static const Color gray800 = Color(0x1F2937);
  static const Color gray900 = Color(0x111827);

  // Status colors (desaturated to blend with neutrals)
  static const Color statusNormal = Color(0x10B981);  /* Emerald, desaturated */
  static const Color statusWarning = Color(0xF59E0B); /* Same as primary (natural amber warning) */
  static const Color statusDanger = Color(0xEF4444);  /* Red, desaturated */
  static const Color statusInfo = Color(0x3B82F6);    /* Blue for informational only when needed */

  // Surface backgrounds
  static const Color bgScaffold = Color(0xFFF9F5);    /* Warm off-white for main background */
  static const Color bgCard = Color(0xFFFFFFFF);      /* Pure white for cards */
  static const Color bgSurface = Color(0xFAFAFA);     /* Slightly raised surface */

  // Component radii — unified across web and app
  static const double radiusSmallest = 6.0;
  static const double radiusSmall = 8.0;
  static const double radiusMedium = 12.0;
  static const double radiusLarge = 16.0;
  static const double radiusXL = 20.0;

  // Spacing scale — optical rhythm
  static const double spacingXS = 4.0;
  static const double spacingS = 8.0;
  static const double spacingM = 16.0;
  static const double spacingL = 24.0;
  static const double spacingXL = 32.0;
  static const double spacingXXL = 48.0;

  // Typography font family — Noto Sans SC for Chinese readability (matches web upgrade)
  static const String fontFamily = 'NotoSansSC';

  // Typography styles with proper hierarchy and character
  static const TextStyle headline1 = TextStyle(
    fontSize: 32.0,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.02,
    color: gray800,
    fontFamily: fontFamily,
  );

  static const TextStyle headline2 = TextStyle(
    fontSize: 24.0,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.015,
    color: gray800,
    fontFamily: fontFamily,
  );

  static const TextStyle headline3 = TextStyle(
    fontSize: 20.0,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.01,
    color: gray800,
    fontFamily: fontFamily,
  );

  static const TextStyle titleLarge = TextStyle(
    fontSize: 18.0,
    fontWeight: FontWeight.w600,
    color: gray800,
    fontFamily: fontFamily,
  );

  static const TextStyle titleMedium = TextStyle(
    fontSize: 16.0,
    fontWeight: FontWeight.w600,
    color: gray700,
    fontFamily: fontFamily,
  );

  static const TextStyle titleSmall = TextStyle(
    fontSize: 14.0,
    fontWeight: FontWeight.w600,
    color: gray600,
    letterSpacing: 0.02,
    fontFamily: fontFamily,
  );

  static const TextStyle bodyLarge = TextStyle(
    fontSize: 16.0,
    fontWeight: FontWeight.w400,
    color: gray700,
    lineHeight: 1.6,
    fontFamily: fontFamily,
  );

  static const TextStyle bodyMedium = TextStyle(
    fontSize: 15.0,
    fontWeight: FontWeight.w400,
    color: gray600,
    lineHeight: 1.6,
    fontFamily: fontFamily,
  );

  static const TextStyle bodySmall = TextStyle(
    fontSize: 13.0,
    fontWeight: FontWeight.w400,
    color: gray500,
    lineHeight: 1.5,
    fontFamily: fontFamily,
  );

  static const TextStyle labelMedium = TextStyle(
    fontSize: 14.0,
    fontWeight: FontWeight.w500,
    color: gray600,
    letterSpacing: 0.01,
    fontFamily: fontFamily,
  );

  static const TextStyle labelSmall = TextStyle(
    fontSize: 12.0,
    fontWeight: FontWeight.w500,
    color: gray500,
    letterSpacing: 0.04,
    textTransform: TextTransform.uppercase,
    fontFamily: fontFamily,
  );

  // Elevation / shadow scale — warmer shadows
  static const boxShadowSm = [
    BoxShadow(
      color: Color(0x33F59E0B), // Amber tint at low opacity
      blurRadius: 4.0,
      offset: Offset(0, 1),
    ),
  ];

  static const boxShadowMd = [
    BoxShadow(
      color: Color(0x4D5552), // Slightly darker neutral shadow
      blurRadius: 8.0,
      offset: Offset(0, 2),
    ),
    BoxShadow(
      color: Color(0x1AF59E0B), // Subtle amber tint
      blurRadius: 4.0,
      offset: Offset(0, 1),
    ),
  ];

  static const boxShadowLg = [
    BoxShadow(
      color: Color(0x265552),
      blurRadius: 12.0,
      offset: Offset(0, 4),
    ),
    BoxShadow(
      color: Color(0x1AF59E0B),
      blurRadius: 8.0,
      offset: Offset(0, 2),
    ),
  ];

  // Interactive states — add subtle scale on press
  static const Animation<double> defaultPressAnimation = Tween<double>(
    begin: 1.0,
    end: 0.98,
  ) .chain(CurveTween(curve: Curves.easeOut));

  // Navigation — tab bar with unified accent
  static const Color tabBarInactive = Color(0x999999);
  static const Color tabBarActive = Color(0xFF07C160); /* Keep green for now - platform best practice */

  // Loading indicator style — amber spinner
  static const Color loadingIndicatorColor = Color(0xFFE8734A); /* Warm amber from original */

  // Card style — consistent radius and elevation
  static const BoxDecoration cardDecoration = BoxDecoration(
    color: bgCard,
    borderRadius: BorderRadius.circular(radiusLarge),
    boxShadow: [boxShadowSm[0]],
  );
}
import 'package:flutter/material.dart';

class AppTheme {
  // Paleta de colores Premium Slate / Sporty Red
  static const Color background = Color(0xFF0F172A); // Slate 900
  static const Color primary = Color(0xFF1E293B);    // Slate 800
  static const Color secondary = Color(0xFF334155);  // Slate 700
  static const Color accent = Color(0xFFF9C70F);     // Yellow Accent (Logo color)
  static const Color textLight = Color(0xFFF8FAFC);  // Slate 50
  static const Color textMuted = Color(0xFF94A3B8);  // Slate 400
  static const Color border = Color(0xFF334155);
  static const Color cardBg = Color(0x991E293B);     // Semitransparente para glassmorphism
  static const Color warning = Color(0xFFF59E0B);    // Amber 500
  static const Color success = Color(0xFF10B981);    // Emerald 500

  static ThemeData get darkTheme {
    return ThemeData(
      brightness: Brightness.dark,
      scaffoldBackgroundColor: background,
      primaryColor: primary,
      colorScheme: const ColorScheme.dark(
        primary: primary,
        secondary: secondary,
        error: accent,
        surface: primary,
        onPrimary: textLight,
        onSecondary: textLight,
        onError: textLight,
      ),
      appBarTheme: const AppBarTheme(
        backgroundColor: Colors.transparent,
        elevation: 0,
        centerTitle: true,
        titleTextStyle: TextStyle(
          color: textLight,
          fontSize: 20,
          fontWeight: FontWeight.bold,
          letterSpacing: 0.5,
        ),
        iconTheme: IconThemeData(color: textLight),
      ),
      cardTheme: CardThemeData(
        color: cardBg,
        elevation: 8,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: border, width: 1),
        ),
      ),
      textTheme: const TextTheme(
        headlineLarge: TextStyle(color: textLight, fontSize: 32, fontWeight: FontWeight.bold),
        headlineMedium: TextStyle(color: textLight, fontSize: 24, fontWeight: FontWeight.bold),
        titleLarge: TextStyle(color: textLight, fontSize: 18, fontWeight: FontWeight.w600),
        bodyLarge: TextStyle(color: textLight, fontSize: 16),
        bodyMedium: TextStyle(color: textMuted, fontSize: 14),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: primary,
        labelStyle: const TextStyle(color: textMuted),
        hintStyle: const TextStyle(color: textMuted),
        contentPadding: const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: border),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: border),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: accent, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: accent),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: accent,
          foregroundColor: const Color(0xFF0F172A), // Dark text on yellow button for readability
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          textStyle: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.5,
          ),
          elevation: 4,
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: textLight,
          side: const BorderSide(color: border),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: accent,
        ),
      ),
      sliderTheme: const SliderThemeData(
        activeTrackColor: accent,
        thumbColor: accent,
        inactiveTrackColor: secondary,
      ),
    );
  }
}

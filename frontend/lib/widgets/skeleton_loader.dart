import 'package:flutter/material.dart';
import '../config/theme.dart';

/// Un contenedor con animación de brillo suave (Shimmer)
class ShimmerEffect extends StatefulWidget {
  final Widget child;

  const ShimmerEffect({super.key, required this.child});

  @override
  State<ShimmerEffect> createState() => _ShimmerEffectState();
}

class _ShimmerEffectState extends State<ShimmerEffect> with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _animation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1400),
    )..repeat();

    _animation = Tween<double>(begin: -2.0, end: 2.0).animate(
      CurvedAnimation(parent: _controller, curve: Curves.easeInOutSine),
    );
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _animation,
      builder: (context, child) {
        return ShaderMask(
          blendMode: BlendMode.srcATop,
          shaderCallback: (bounds) {
            return LinearGradient(
              begin: Alignment(_animation.value - 1, -0.3),
              end: Alignment(_animation.value + 1, 0.3),
              colors: [
                const Color(0xFF1E293B),
                const Color(0xFF334155),
                AppTheme.accent.withValues(alpha: 0.35),
                const Color(0xFF334155),
                const Color(0xFF1E293B),
              ],
              stops: const [0.0, 0.35, 0.5, 0.65, 1.0],
            ).createShader(bounds);
          },
          child: widget.child,
        );
      },
    );
  }
}

/// Bloque individual de esqueleto de carga
class SkeletonBox extends StatelessWidget {
  final double? width;
  final double? height;
  final double borderRadius;
  final Margin? margin;

  const SkeletonBox({
    super.key,
    this.width,
    this.height,
    this.borderRadius = 8.0,
    this.margin,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      margin: margin != null ? EdgeInsets.only(top: margin!.top, bottom: margin!.bottom, left: margin!.left, right: margin!.right) : null,
      decoration: BoxDecoration(
        color: const Color(0xFF1E293B),
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    );
  }
}

class Margin {
  final double top;
  final double bottom;
  final double left;
  final double right;
  const Margin({this.top = 0, this.bottom = 0, this.left = 0, this.right = 0});
}

/// Esqueleto para tarjetas de Citas / Servicios
class SkeletonCard extends StatelessWidget {
  const SkeletonCard({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerEffect(
      child: Container(
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: AppTheme.cardBg,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: AppTheme.border.withValues(alpha: 0.5),
            width: 1,
          ),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: const [
                SkeletonBox(width: 160, height: 20, borderRadius: 6),
                SkeletonBox(width: 70, height: 24, borderRadius: 12),
              ],
            ),
            const SizedBox(height: 12),
            const SkeletonBox(width: 220, height: 14, borderRadius: 4),
            const SizedBox(height: 8),
            const SkeletonBox(width: 180, height: 14, borderRadius: 4),
            const SizedBox(height: 16),
            Row(
              children: const [
                Expanded(child: SkeletonBox(height: 40, borderRadius: 10)),
                SizedBox(width: 12),
                SkeletonBox(width: 40, height: 40, borderRadius: 10),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

/// Lista de tarjetas de esqueleto de carga
class SkeletonList extends StatelessWidget {
  final int count;
  final EdgeInsetsGeometry padding;

  const SkeletonList({
    super.key,
    this.count = 4,
    this.padding = const EdgeInsets.all(16),
  });

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      padding: padding,
      itemCount: count,
      physics: const NeverScrollableScrollPhysics(),
      shrinkWrap: true,
      itemBuilder: (context, index) => const SkeletonCard(),
    );
  }
}

/// Esqueleto para el Dashboard de Administración
class SkeletonDashboard extends StatelessWidget {
  const SkeletonDashboard({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerEffect(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SkeletonBox(width: 240, height: 28, borderRadius: 6),
            const SizedBox(height: 20),
            // Tarjetas métricas
            Row(
              children: const [
                Expanded(child: SkeletonBox(height: 90, borderRadius: 14)),
                SizedBox(width: 12),
                Expanded(child: SkeletonBox(height: 90, borderRadius: 14)),
              ],
            ),
            const SizedBox(height: 12),
            Row(
              children: const [
                Expanded(child: SkeletonBox(height: 90, borderRadius: 14)),
                SizedBox(width: 12),
                Expanded(child: SkeletonBox(height: 90, borderRadius: 14)),
              ],
            ),
            const SizedBox(height: 28),
            const SkeletonBox(width: 180, height: 22, borderRadius: 6),
            const SizedBox(height: 16),
            // Lista de técnicos
            const SkeletonBox(width: double.infinity, height: 75, borderRadius: 12),
            const SizedBox(height: 12),
            const SkeletonBox(width: double.infinity, height: 75, borderRadius: 12),
            const SizedBox(height: 12),
            const SkeletonBox(width: double.infinity, height: 75, borderRadius: 12),
          ],
        ),
      ),
    );
  }
}

/// Esqueleto para la vista detallada de una cita
class SkeletonDetail extends StatelessWidget {
  const SkeletonDetail({super.key});

  @override
  Widget build(BuildContext context) {
    return ShimmerEffect(
      child: SingleChildScrollView(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: AppTheme.cardBg,
                borderRadius: BorderRadius.circular(16),
                border: Border.all(color: AppTheme.border.withValues(alpha: 0.5)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: const [
                  SkeletonBox(width: 200, height: 22, borderRadius: 6),
                  SizedBox(height: 12),
                  SkeletonBox(width: 150, height: 16, borderRadius: 4),
                  SizedBox(height: 8),
                  SkeletonBox(width: 180, height: 16, borderRadius: 4),
                  SizedBox(height: 8),
                  SkeletonBox(width: 140, height: 16, borderRadius: 4),
                ],
              ),
            ),
            const SizedBox(height: 20),
            const SkeletonBox(width: 160, height: 20, borderRadius: 6),
            const SizedBox(height: 12),
            Row(
              children: const [
                Expanded(child: SkeletonBox(height: 48, borderRadius: 10)),
                SizedBox(width: 12),
                Expanded(child: SkeletonBox(height: 48, borderRadius: 10)),
              ],
            ),
            const SizedBox(height: 20),
            const SkeletonBox(width: 140, height: 20, borderRadius: 6),
            const SizedBox(height: 12),
            const SkeletonBox(width: double.infinity, height: 120, borderRadius: 14),
          ],
        ),
      ),
    );
  }
}

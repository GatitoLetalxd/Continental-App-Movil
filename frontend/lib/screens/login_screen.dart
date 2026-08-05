import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../config/theme.dart';
import '../providers/auth_provider.dart';
import '../widgets/glass_card.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _isRegister = false;

  // Login
  final _usuarioController = TextEditingController();
  final _passwordController = TextEditingController();

  // Registro cliente
  final _nombreController = TextEditingController();
  final _pinController = TextEditingController();
  final _contacto1Controller = TextEditingController();
  final _contacto2Controller = TextEditingController();

  @override
  void dispose() {
    _usuarioController.dispose();
    _passwordController.dispose();
    _nombreController.dispose();
    _pinController.dispose();
    _contacto1Controller.dispose();
    _contacto2Controller.dispose();
    super.dispose();
  }

  void _submit() async {
    if (!_formKey.currentState!.validate()) return;

    final auth = Provider.of<AuthProvider>(context, listen: false);
    String? error;

    if (_isRegister) {
      final contactos = <String>[];
      if (_contacto1Controller.text.isNotEmpty) contactos.add(_contacto1Controller.text);
      if (_contacto2Controller.text.isNotEmpty) contactos.add(_contacto2Controller.text);

      error = await auth.registrarCliente(
        _nombreController.text.trim(),
        _pinController.text.trim(),
        contactos,
      );
    } else {
      error = await auth.login(
        _usuarioController.text.trim(),
        _passwordController.text.trim(),
      );
    }

    if (error != null) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              error,
              style: const TextStyle(color: Color(0xFF0F172A), fontWeight: FontWeight.bold),
            ),
            backgroundColor: AppTheme.accent,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final auth = Provider.of<AuthProvider>(context);

    return Scaffold(
      body: Stack(
        children: [
          // Fondo gradiente dinámico (Premium Slate)
          Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  Color(0xFF0F172A),
                  Color(0xFF1E293B),
                  Color(0xFF020617),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
            ),
          ),
          // Decoración visual - Esfera difusa roja en la esquina superior
          Positioned(
            top: -100,
            right: -100,
            child: Container(
              width: 300,
              height: 300,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: AppTheme.accent.withValues(alpha: 0.15),
              ),
            ),
          ),
          // Decoración visual - Esfera difusa azul/slate abajo
          Positioned(
            bottom: -50,
            left: -50,
            child: Container(
              width: 250,
              height: 250,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: AppTheme.secondary.withValues(alpha: 0.2),
              ),
            ),
          ),
          // Formulario
          Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(24),
              child: ConstrainedBox(
                constraints: const BoxConstraints(maxWidth: 450),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Logo taller
                    const LogoAutopartesContinental(width: 280),
                    const SizedBox(height: 12),
                    const Text(
                      'Gestión de Tapicería y Accesorios',
                      style: TextStyle(
                        fontSize: 13,
                        color: AppTheme.textMuted,
                        fontWeight: FontWeight.w600,
                        letterSpacing: 0.5,
                      ),
                    ),
                    const SizedBox(height: 24),
                    // Tarjeta Glassmorphic
                    GlassCard(
                      child: Form(
                        key: _formKey,
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.stretch,
                          children: [
                            Text(
                              _isRegister ? 'Registro de Cliente' : 'Iniciar Sesión',
                              style: Theme.of(context).textTheme.headlineMedium,
                              textAlign: TextAlign.center,
                            ),
                            const SizedBox(height: 24),
                            if (!_isRegister) ...[
                              // Login Inputs
                              TextFormField(
                                controller: _usuarioController,
                                decoration: const InputDecoration(
                                  labelText: 'Usuario o Nombre Completo',
                                  prefixIcon: Icon(Icons.person_outline, color: AppTheme.textMuted),
                                ),
                                validator: (val) => val == null || val.isEmpty ? 'Campo obligatorio' : null,
                              ),
                              const SizedBox(height: 16),
                              TextFormField(
                                controller: _passwordController,
                                obscureText: true,
                                decoration: const InputDecoration(
                                  labelText: 'Contraseña o PIN',
                                  prefixIcon: Icon(Icons.lock_outline, color: AppTheme.textMuted),
                                ),
                                validator: (val) => val == null || val.isEmpty ? 'Campo obligatorio' : null,
                              ),
                            ] else ...[
                              // Registro Inputs
                              TextFormField(
                                controller: _nombreController,
                                decoration: const InputDecoration(
                                  labelText: 'Nombre Completo',
                                  prefixIcon: Icon(Icons.person_outline, color: AppTheme.textMuted),
                                ),
                                validator: (val) => val == null || val.isEmpty ? 'Ingrese su nombre' : null,
                              ),
                              const SizedBox(height: 16),
                              TextFormField(
                                controller: _pinController,
                                keyboardType: TextInputType.number,
                                obscureText: true,
                                decoration: const InputDecoration(
                                  labelText: 'PIN de Seguridad (4-8 dígitos)',
                                  prefixIcon: Icon(Icons.lock_outline, color: AppTheme.textMuted),
                                  helperText: 'Solo números para accesos futuros',
                                ),
                                validator: (val) {
                                  if (val == null || val.isEmpty) return 'Ingrese un PIN';
                                  if (!RegExp(r'^\d{4,8}$').hasMatch(val)) return 'Debe ser numérico de 4 a 8 dígitos';
                                  return null;
                                },
                              ),
                              const SizedBox(height: 16),
                              TextFormField(
                                controller: _contacto1Controller,
                                keyboardType: TextInputType.phone,
                                decoration: const InputDecoration(
                                  labelText: 'Teléfono de Contacto 1 (wa.me)',
                                  prefixIcon: Icon(Icons.phone_outlined, color: AppTheme.textMuted),
                                  hintText: 'Ej: 584121234567',
                                ),
                                validator: (val) => val == null || val.isEmpty ? 'Ingrese al menos un teléfono' : null,
                              ),
                              const SizedBox(height: 16),
                              TextFormField(
                                controller: _contacto2Controller,
                                keyboardType: TextInputType.phone,
                                decoration: const InputDecoration(
                                  labelText: 'Teléfono de Contacto 2 (Opcional)',
                                  prefixIcon: Icon(Icons.phone_outlined, color: AppTheme.textMuted),
                                  hintText: 'Ej: 584129876543',
                                ),
                              ),
                            ],
                            const SizedBox(height: 24),
                            auth.isLoading
                                ? const Center(
                                    child: CircularProgressIndicator(color: AppTheme.accent),
                                  )
                                : ElevatedButton(
                                    onPressed: _submit,
                                    child: Text(_isRegister ? 'REGISTRARSE' : 'ENTRAR'),
                                  ),
                            const SizedBox(height: 16),
                            TextButton(
                              onPressed: () {
                                setState(() {
                                  _isRegister = !_isRegister;
                                });
                              },
                              child: Text(
                                _isRegister ? '¿Ya tienes cuenta? Inicia sesión' : '¿Eres cliente nuevo? Regístrate aquí',
                                style: const TextStyle(fontWeight: FontWeight.w600),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class LogoAutopartesContinental extends StatelessWidget {
  final double width;
  const LogoAutopartesContinental({super.key, this.width = 300});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFFF9C70F).withValues(alpha: 0.4),
            blurRadius: 20,
            spreadRadius: 1,
            offset: const Offset(0, 6),
          ),
        ],
      ),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: Image.asset(
          'assets/images/logo.png',
          width: width,
          fit: BoxFit.contain,
        ),
      ),
    );
  }
}

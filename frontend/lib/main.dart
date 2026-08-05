import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'config/theme.dart';
import 'providers/auth_provider.dart';
import 'providers/taller_provider.dart';
import 'services/socket_service.dart';
import 'screens/login_screen.dart';
import 'screens/client_screen.dart';
import 'screens/technician_screen.dart';
import 'screens/admin_screen.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider(create: (_) => TallerProvider()),
        ChangeNotifierProvider(create: (_) => SocketService()),
      ],
      child: const MyApp(),
    ),
  );
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Taller Continental',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.darkTheme, // Usar tema dark premium por defecto
      home: const AuthGate(),
    );
  }
}

class AuthGate extends StatefulWidget {
  const AuthGate({super.key});

  @override
  State<AuthGate> createState() => _AuthGateState();
}

class _AuthGateState extends State<AuthGate> {
  late Future<bool> _authCheck;

  @override
  void initState() {
    super.initState();
    _authCheck = Provider.of<AuthProvider>(context, listen: false).checkAuth();
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<bool>(
      future: _authCheck,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Scaffold(
            body: Center(
              child: CircularProgressIndicator(color: AppTheme.accent),
            ),
          );
        }

        return Consumer<AuthProvider>(
          builder: (context, auth, _) {
            if (!auth.isAuthenticated) {
              return const LoginScreen();
            }

            // Router por Rol del Sistema
            switch (auth.userRol) {
              case 'ADMIN':
                return const AdminScreen();
              case 'TECNICO':
                return const TechnicianScreen();
              case 'CLIENTE':
                return const ClientScreen();
              default:
                return const LoginScreen();
            }
          },
        );
      },
    );
  }
}

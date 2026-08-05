# 🚗 Continental App Móvil & Backend

Sistema de Gestión de Taller Mecánico, Tapicería y Accesorios - Autopartes Continental.

## 📱 Frontend (Flutter)
- **Framework**: Flutter 3.x (Dart 3.x)
- **Plataformas**: Android (APK / App Bundle), Web

### Compilar APK para Android
```bash
cd frontend
flutter pub get
flutter build apk --release
```
El archivo APK generado estará disponible en:
`frontend/build/app/outputs/flutter-apk/app-release.apk`

---

## 🔧 Backend (Node.js + Prisma)
- **Runtime**: Node.js (Express + Socket.io)
- **Base de Datos**: PostgreSQL
- **ORM**: Prisma

### Configuración e Inicio de Servidor
```bash
cd backend
npm install
cp .env.example .env
npx prisma db push
npm run dev
```

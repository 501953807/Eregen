import 'package:flutter/foundation.dart';

/// Global app state tracking the current logged-in user and selected elderly person.
class AppState extends ChangeNotifier {
  static AppState? _instance;
  factory AppState() => _instance ??= AppState._();

  AppState._();

  String? userId;
  String? elderlyId;
  String? elderlyName;
  bool get isAuthenticated => userId != null && userId!.isNotEmpty;
  bool get hasElderly => elderlyId != null && elderlyId!.isNotEmpty;

  void setAuth({required String userId}) {
    this.userId = userId;
    notifyListeners();
  }

  void selectElderly({required String elderlyId, required String name}) {
    this.elderlyId = elderlyId;
    elderlyName = name;
    notifyListeners();
  }

  void logout() {
    userId = null;
    elderlyId = null;
    elderlyName = null;
    notifyListeners();
  }
}

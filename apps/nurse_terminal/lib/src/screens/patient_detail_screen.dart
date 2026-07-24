import 'package:flutter/material.dart';

/// Placeholder for the patient detail screen.
/// Will be fully implemented in Task 12.
class PatientDetailScreen extends StatelessWidget {
  final Map<String, dynamic> patient;

  const PatientDetailScreen({super.key, required this.patient});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('患者详情')),
      body: const Center(child: Text('待实现 — Task 12')),
    );
  }
}

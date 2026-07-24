import 'dart:async';
import 'dart:convert';

import 'package:flutter_blue_plus/flutter_blue_plus.dart';
import 'package:nurse_terminal/src/models/ble_uuids.dart';
import 'package:nurse_terminal/src/models/medical_models.dart';

/// Medical wristband BLE scanner and GATT client.
class MedicalWristbandService {
  final StreamController<BleEvent> _eventController =
      StreamController<BleEvent>.broadcast();

  /// Stream of BLE events for UI updates.
  Stream<BleEvent> get events => _eventController.stream;

  BluetoothDevice? _connectedDevice;
  BluetoothCharacteristic? _pairingChar;
  BluetoothCharacteristic? _patientInfoChar;
  BluetoothCharacteristic? _verificationChar;
  BluetoothCharacteristic? _statusChar;
  BluetoothCharacteristic? _commandChar;

  bool get isConnected => _connectedDevice != null;

  /// Start scanning for medical wristband devices.
  Future<void> startScan() async {
    FlutterBluePlus.startScan(
      timeout: const Duration(seconds: 15),
      androidUsesFineLocation: true,
    );

    FlutterBluePlus.scanResults.listen((results) {
      for (final r in results) {
        final name = r.device.platformName;
        if (name.contains('Eregen') || name.contains('WB-')) {
          _eventController.add(BleDeviceDiscovered(r.device, r.rssi));
        }
      }
    });
  }

  /// Stop BLE scanning.
  Future<void> stopScan() async {
    await FlutterBluePlus.stopScan();
  }

  /// Connect to a medical wristband device.
  Future<bool> connect(BluetoothDevice device) async {
    try {
      await device.connect(timeout: const Duration(seconds: 10));
      _connectedDevice = device;
      _eventController.add(BleConnected(device));

      await _discoverServices();
      return true;
    } catch (e) {
      _eventController.add(BleError('Connection failed: $e'));
      return false;
    }
  }

  /// Disconnect from current device.
  Future<void> disconnect() async {
    if (_connectedDevice != null) {
      await _connectedDevice!.disconnect();
      _connectedDevice = null;
      _clearCharacteristics();
      _eventController.add(const BleDisconnected());
    }
  }

  /// Discover all GATT services and characteristics.
  Future<void> _discoverServices() async {
    if (_connectedDevice == null) return;

    final services = await _connectedDevice!.discoverServices();
    for (final service in services) {
      if (service.uuid.toString() == BleUuids.service) {
        for (final char in service.characteristics) {
          final uuid = char.uuid.toString();
          switch (uuid) {
            case BleUuids.pairingCode:
              _pairingChar = char;
              break;
            case BleUuids.patientInfo:
              _patientInfoChar = char;
              break;
            case BleUuids.verification:
              _verificationChar = char;
              break;
            case BleUuids.status:
              _statusChar = char;
              break;
            case BleUuids.command:
              _commandChar = char;
              break;
          }
        }
        _eventController.add(const BleServicesDiscovered());
        return;
      }
    }
  }

  void _clearCharacteristics() {
    _pairingChar = null;
    _patientInfoChar = null;
    _verificationChar = null;
    _statusChar = null;
    _commandChar = null;
  }

  /// Read pairing code from wristband.
  Future<String?> readPairingCode() async {
    final char = _pairingChar;
    if (char == null) {
      _eventController.add(const BleError('Pairing characteristic not found'));
      return null;
    }

    try {
      final value = await char.read();
      if (value.isEmpty) return null;
      return utf8.decode(value);
    } catch (e) {
      _eventController.add(BleError('Read pairing code failed: $e'));
      return null;
    }
  }

  /// Write pairing code to wristband for authentication.
  Future<bool> writePairingCode(String code) async {
    final char = _pairingChar;
    if (char == null) {
      _eventController.add(const BleError('Pairing characteristic not found'));
      return false;
    }

    try {
      await char.write(utf8.encode(code));
      _eventController.add(const BlePairingSuccess());
      return true;
    } catch (e) {
      _eventController.add(BleError('Write pairing code failed: $e'));
      return false;
    }
  }

  /// Read patient information from wristband.
  Future<PatientInfo?> readPatientInfo() async {
    final char = _patientInfoChar;
    if (char == null) {
      _eventController.add(const BleError('Patient info characteristic not found'));
      return null;
    }

    try {
      final value = await char.read();
      if (value.isEmpty) return null;

      final json = jsonDecode(utf8.decode(value)) as Map<String, dynamic>;
      return PatientInfo.fromJson(json);
    } catch (e) {
      _eventController.add(BleError('Read patient info failed: $e'));
      return null;
    }
  }

  /// Send verification request to wristband and wait for response.
  Future<VerificationResult?> sendVerificationRequest({
    required String requestId,
    required String scanType,
    required String patientId,
    double lat = 0,
    double lon = 0,
    String notes = '',
  }) async {
    final char = _verificationChar;
    if (char == null) {
      _eventController.add(const BleError('Verification characteristic not found'));
      return null;
    }

    try {
      final payload = {
        'request_id': requestId,
        'scan_type': scanType,
        'patient_id': patientId,
        'lat': lat,
        'lon': lon,
        'notes': notes,
      };
      final data = utf8.encode(jsonEncode(payload));
      await char.write(data);

      final completer = Completer<VerificationResult>();
      late StreamSubscription<List<int>> subscription;

      subscription = char.onValueReceived.listen((value) {
        try {
          final json = jsonDecode(utf8.decode(value)) as Map<String, dynamic>;
          completer.complete(VerificationResult.fromJson(json));
          subscription.cancel();
        } catch (e) {
          completer.completeError(e);
        }
      });

      return await completer.future.timeout(
        const Duration(seconds: 10),
        onTimeout: () {
          subscription.cancel();
          throw TimeoutException('Verification response timeout',
              const Duration(seconds: 10));
        },
      );
    } catch (e) {
      _eventController.add(BleError('Send verification failed: $e'));
      return null;
    }
  }

  /// Read device status from wristband.
  Future<Map<String, dynamic>?> readStatus() async {
    final char = _statusChar;
    if (char == null) {
      _eventController.add(const BleError('Status characteristic not found'));
      return null;
    }

    try {
      final value = await char.read();
      if (value.isEmpty) return null;
      return jsonDecode(utf8.decode(value)) as Map<String, dynamic>;
    } catch (e) {
      _eventController.add(BleError('Read status failed: $e'));
      return null;
    }
  }

  /// Send command to wristband.
  Future<bool> sendCommand({
    required String commandType,
    required String commandId,
  }) async {
    final char = _commandChar;
    if (char == null) {
      _eventController.add(const BleError('Command characteristic not found'));
      return false;
    }

    try {
      final payload = {
        'command_type': commandType,
        'command_id': commandId,
        'timestamp_ms': DateTime.now().millisecondsSinceEpoch,
      };
      final data = utf8.encode(jsonEncode(payload));
      await char.write(data);
      _eventController.add(const BleCommandSent());
      return true;
    } catch (e) {
      _eventController.add(BleError('Send command failed: $e'));
      return false;
    }
  }

  /// Listen for notifications on a characteristic.
  Future<void> enableNotifications(BluetoothCharacteristic char) async {
    await char.setNotifyValue(true);
  }

  /// Dispose BLE service and close streams.
  void dispose() {
    _connectedDevice?.disconnect();
    _clearCharacteristics();
    _eventController.close();
  }
}

/// BLE events emitted by the service.
sealed class BleEvent {
  const BleEvent();
}

class BleDeviceDiscovered extends BleEvent {
  final BluetoothDevice device;
  final int rssi;

  const BleDeviceDiscovered(this.device, this.rssi);
}

class BleConnected extends BleEvent {
  final BluetoothDevice device;

  const BleConnected(this.device);
}

class BleDisconnected extends BleEvent {
  const BleDisconnected();
}

class BleServicesDiscovered extends BleEvent {
  const BleServicesDiscovered();
}

class BlePairingSuccess extends BleEvent {
  const BlePairingSuccess();
}

class BleCommandSent extends BleEvent {
  const BleCommandSent();
}

class BleError extends BleEvent {
  final String message;

  const BleError(this.message);
}

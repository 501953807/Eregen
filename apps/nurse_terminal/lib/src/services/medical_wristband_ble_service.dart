import 'dart:async';
import 'dart:convert';

import 'package:nfc_manager/nfc_manager.dart';
import '../models/nfc_tags.dart';
import '../models/medical_models.dart';

/// Medical wristband NFC reader service.
/// Reads NDEF messages from NFC-enabled medical wristbands (ESP32-S3).
class MedicalWristbandService {
  final StreamController<NfcEvent> _eventController =
      StreamController<NfcEvent>.broadcast();

  Stream<NfcEvent> get events => _eventController.stream;

  bool get isReading => _nfcManager != null;
  NfcManager? _nfcManager;
  bool _disposed = false;

  /// Start NFC scanning for wristband devices.
  /// Returns the NDEF message if a wristband is detected within range.
  Future<NdefMessage?> scanWristband({Duration timeout = const Duration(seconds: 15)}) async {
    if (_disposed) return null;

    _eventController.add(const NfcScanningStarted());
    _eventController.add(const NfcMessageEvent('NFC 扫描中，请将腕带靠近设备...'));

    NdefMessage? result;
    _nfcManager = NfcManager.instance;
    await _nfcManager!.startSession(
      onDiscovered: (NfcTag tag) async {
        final ndef = Ndef.from(tag);
        if (ndef == null) {
          _eventController.add(const NfcMessageEvent('标签不支持 NDEF'));
          return;
        }

        final message = ndef.cachedMessage;
        if (message != null) {
          await _processNdefMessage(message);
          result = message;
        } else {
          try {
            final loaded = await ndef.read();
            await _processNdefMessage(loaded);
            result = loaded;
          } catch (e) {
            _eventController.add(NfcMessageEvent('读取 NDEF 失败: $e'));
          }
        }
        _nfcManager!.stopSession();
        _nfcManager = null;
      },
    );

    if (result == null) {
      _eventController.add(const NfcMessageEvent('未在范围内检测到腕带，请重试'));
    } else {
      _eventController.add(const NfcMessageEvent('腕带读取成功'));
    }

    return result;
  }

  Future<void> _processNdefMessage(NdefMessage message) async {
    for (final record in message.records) {
      final typeNameFormat = record.typeNameFormat;
      final type = String.fromCharCodes(record.type);

      if (typeNameFormat == NdefTypeNameFormat.nfcWellknown && type == 'T') {
        final text = utf8.decode(record.payload.sublist(1));
        _eventController.add(NfcMessageEvent('文本记录: $text'));
      } else if (typeNameFormat == NdefTypeNameFormat.media &&
          type == NfcTagTypes.deviceInfo) {
        final data = jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
        _eventController.add(NfcMessageEvent('设备信息: ${data['dev_id'] ?? 'unknown'}'));
      } else if (typeNameFormat == NdefTypeNameFormat.media &&
          type == NfcTagTypes.patient) {
        final data = jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
        _eventController.add(NfcMessageEvent('患者: ${data['patient_id'] ?? 'unknown'}'));
      } else if (typeNameFormat == NdefTypeNameFormat.media &&
          type == NfcTagTypes.verification) {
        final data = jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
        _eventController.add(NfcMessageEvent('核验请求: ${data['request_id'] ?? 'unknown'}'));
      } else if (typeNameFormat == NdefTypeNameFormat.media &&
          type == NfcTagTypes.status) {
        final data = jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
        _eventController.add(NfcMessageEvent('状态: ${data['status'] ?? 'unknown'}'));
      }
    }
  }

  /// Read patient info from an already-discovered NDEF message.
  PatientInfo? parsePatientInfo(NdefMessage message) {
    for (final record in message.records) {
      if (String.fromCharCodes(record.type) == NfcTagTypes.patient) {
        try {
          final data = jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
          return PatientInfo.fromJson(data);
        } catch (e) {
          _eventController.add(NfcMessageEvent('解析患者信息失败: $e'));
          return null;
        }
      }
    }
    return null;
  }

  /// Read device status from an already-discovered NDEF message.
  Map<String, dynamic>? parseDeviceStatus(NdefMessage message) {
    for (final record in message.records) {
      if (String.fromCharCodes(record.type) == NfcTagTypes.status) {
        try {
          return jsonDecode(utf8.decode(record.payload)) as Map<String, dynamic>;
        } catch (e) {
          _eventController.add(NfcMessageEvent('解析设备状态失败: $e'));
          return null;
        }
      }
    }
    return null;
  }

  void dispose() {
    _disposed = true;
    _nfcManager?.stopSession();
    _nfcManager = null;
    _eventController.close();
  }
}

/// Base class for NFC service events.
sealed class NfcEvent {
  const NfcEvent();
}

/// Fired when NFC scanning starts.
class NfcScanningStarted extends NfcEvent {
  const NfcScanningStarted();
}

/// Fired with a message string.
class NfcMessageEvent extends NfcEvent {
  final String message;
  const NfcMessageEvent(this.message);
}

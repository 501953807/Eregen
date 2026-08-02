/// NFC tag types for medical wristband NDEF records
/// Matching shared/protocol/wb_ble.go NFCAuthPayload and NDEF types
class NfcTagTypes {
  static const String deviceInfo = 'application/vnd.eregen.device-info';
  static const String patient = 'application/vnd.eregen.patient';
  static const String verification = 'application/vnd.eregen.verification';
  static const String status = 'application/vnd.eregen.status';
}

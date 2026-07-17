import 'dart:async';
import 'dart:convert';

import 'package:web_socket_channel/web_socket_channel.dart';

/// WebSocket client for real-time alert events.
/// Connects to `GET /ws/alerts?user_id={id}` on the api-server.
class AlertWebSocket {
  final String wsUrl;
  final void Function(Map<String, dynamic> alert) onAlert;
  final void Function()? onDisconnected;

  WebSocketChannel? _channel;
  Timer? _pingTimer;
  bool _reconnecting = false;

  bool get isConnected => _channel != null;

  AlertWebSocket({
    required this.wsUrl,
    required this.onAlert,
    this.onDisconnected,
  });

  void connect() {
    if (isConnected) return;
    try {
      _channel = WebSocketChannel.connect(Uri.parse(wsUrl));
      _startPing();
      _channel!.stream.listen(
        _handleMessage,
        onError: (_) => _handleDisconnect(),
        onDone: () => _handleDisconnect(),
      );
    } catch (_) {
      _reconnect();
    }
  }

  void disconnect() {
    _pingTimer?.cancel();
    _channel?.sink.close();
    _channel = null;
  }

  void _handleMessage(dynamic data) {
    if (data is! String) return;
    final json = jsonDecode(data) as Map<String, dynamic>;
    if (json['type'] == 'connected') return;
    if (json['type'] != null && json['elderly_id'] != null) {
      onAlert(json);
    }
  }

  void _handleDisconnect() {
    _pingTimer?.cancel();
    _channel = null;
    onDisconnected?.call();
    _reconnect();
  }

  void _reconnect() {
    if (_reconnecting) return;
    _reconnecting = true;
    Future.delayed(const Duration(seconds: 3), () {
      _reconnecting = false;
      connect();
    });
  }

  void _startPing() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(const Duration(seconds: 25), (_) {
      _channel?.sink.add('{"type":"ping"}');
    });
  }
}

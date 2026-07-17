import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:webview_flutter/webview_flutter.dart';
import 'package:geolocator/geolocator.dart';
import '../../common/theme.dart';

/// AMap (高德地图) widget using WebView + JS API v2.0 with GCJ-02 coordinate system.
class MapSection extends StatefulWidget {
  const MapSection({super.key});

  @override
  State<MapSection> createState() => _MapSectionState();
}

class _MapSectionState extends State<MapSection> {
  late final WebViewController _controller;
  String _address = '定位中...';
  String _updateTime = '';
  bool _loading = true;
  Timer? _locationTimer;

  @override
  void initState() {
    super.initState();
    _initWebView();
  }

  Future<void> _initWebView() async {
    final hasPermission = await _requestLocationPermission();
    if (!hasPermission) {
      setState(() {
        _loading = false;
        _address = '位置权限已拒绝';
      });
      return;
    }

    // Load embedded AMap HTML
    final htmlData = await rootBundle.loadString('assets/amap.html');
    // Replace placeholder key with real key when configured
    final html = htmlData.replaceAll('YOUR_AMAP_KEY', String.fromEnvironment('AMAP_KEY', defaultValue: ''));

    _controller = WebViewController()
      ..setJavaScriptMode(JavaScriptMode.unrestricted)
      ..setBackgroundColor(const Color(0xFFF0F2F5))
      ..addJavaScriptChannel('Flutter', onMessageReceived: _handleJsMessage)
      ..loadRequest(Uri.parse('about:blank'));

    // Inject initial map setup
    await _controller.loadHtmlString('''
      $html
      <script>
        initMap(31.2304, 121.4737); // Default: Shanghai
      </script>
    ''');

    // Start periodic location updates
    _startLocationUpdates();
  }

  void _handleJsMessage(JavaScriptMessage message) {
    try {
      final data = jsonDecode(message.message) as Map<String, dynamic>;
      if (data['event'] == 'map_ready') {
        setState(() => _loading = false);
      }
      if (data.containsKey('map_address')) {
        setState(() => _address = data['map_address'] as String);
      }
      if (data.containsKey('map_time')) {
        setState(() => _updateTime = data['map_time'] as String);
      }
    } catch (_) {}
  }

  Future<bool> _requestLocationPermission() async {
    bool serviceEnabled;
    LocationPermission permission;

    serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) return false;

    permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) return false;
    }
    if (permission == LocationPermission.deniedForever) return false;
    return true;
  }

  void _startLocationUpdates() {
    _locationTimer?.cancel();
    _locationTimer = Timer.periodic(const Duration(seconds: 30), (_) async {
      try {
        final position = await Geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.high,
        );
        // Update marker via JavaScript: [lng, lat] for AMap
        _controller.runJavaScript('''
          updateMarker(${position.longitude}, ${position.latitude});
        ''');
      } catch (_) {}
    });
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(left: 20, right: 20, top: 0),
      child: Container(
        height: 200,
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: const Color(0xFF4A90D9).withOpacity(0.15),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        clipBehavior: Clip.antiAlias,
        child: Stack(
          children: [
            _loading
                ? const Center(child: CircularProgressIndicator(strokeWidth: 2))
                : WebViewWidget(controller: _controller),
            Positioned(
              left: 12,
              right: 12,
              bottom: 12,
              child: Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.95),
                  borderRadius: BorderRadius.circular(12),
                  boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.08), blurRadius: 8)],
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(_address, style: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 2),
                    Text(_updateTime.isEmpty ? '更新时间：刚刚' : _updateTime,
                        style: const TextStyle(fontSize: 10, color: Color(0xFF999999))),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  void dispose() {
    _locationTimer?.cancel();
    super.dispose();
  }
}

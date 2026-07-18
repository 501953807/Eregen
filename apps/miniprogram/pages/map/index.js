// pages/map/index.js
const { ApiClient } = require('../../utils/api');

Page({
  data: {
    elderlyName: '加载中...',
    locationTime: '',
    latitude: 31.2304,
    longitude: 121.4737,
    address: '上海市浦东新区',
    showFence: false,
    geofences: [],
    timeOptions: [
      { label: '今日', value: 'today' },
      { label: '本周', value: 'week' },
      { label: '本月', value: 'month' }
    ],
    currentRange: 'today',
    nearbyPlaces: [
      { icon: '🏥', name: '社区医院', distance: 350 },
      { icon: '🏪', name: '药房', distance: 520 },
      { icon: '🌳', name: '公园', distance: 800 },
      { icon: '🏬', name: '超市', distance: 450 },
      { icon: '🍽️', name: '餐厅', distance: 280 }
    ]
  },

  onLoad() {
    const elderlyId = wx.getStorageSync('elderly_id') || '';
    if (elderlyId) {
      this._fetchLocation(elderlyId);
    } else {
      this._loadDemoLocation();
    }
  },

  _fetchLocation(elderlyId) {
    const api = new ApiClient();
    api.get(`/location/latest`, {
      query: { elderly_id: elderlyId }
    }).then(res => {
      if (res.code === 'OK' && res.data) {
        this._updateLocation(res.data);
      }
    }).catch(() => {
      this._loadDemoLocation();
    });
  },

  _updateLocation(loc) {
    const now = new Date();
    const timeStr = `${now.getHours().toString().padStart(2, '0')}:${now.getMinutes().toString().padStart(2, '0')}`;

    this.setData({
      latitude: loc.lat || this.data.latitude,
      longitude: loc.lon || this.data.longitude,
      locationTime: `更新于 ${timeStr}`,
      address: this._formatAddress(loc)
    });

    this._drawTrajectoryChart();
  },

  _loadDemoLocation() {
    this.setData({
      elderlyName: '张爷爷',
      locationTime: '更新于 14:32',
      address: '上海市浦东新区陆家嘴环路1088号',
      latitude: 31.235,
      longitude: 121.500
    });
    this._drawTrajectoryChart();
  },

  _formatAddress(loc) {
    // In production, use reverse geocoding API
    return `纬度 ${loc.lat?.toFixed(4)} / 经度 ${loc.lon?.toFixed(4)}`;
  },

  toggleFence() {
    this.setData({ showFence: !this.data.showFence });
    if (this.data.showFence) {
      this._fetchGeofences();
    }
  },

  _fetchGeofences() {
    // Fetch geofences from API
    this.setData({
      geofences: [
        { id: '1', name: '家', radiusMeters: 500, active: true },
        { id: '2', name: '社区医院', radiusMeters: 200, active: false }
      ]
    });
  },

  selectTimeRange(e) {
    const value = e.currentTarget.dataset.value;
    this.setData({ currentRange: value });
    this._drawTrajectoryChart();
  },

  centerOnElder() {
    const mapCtx = wx.createMapContext('eregenMap');
    mapCtx.moveToLocation({
      latitude: this.data.latitude,
      longitude: this.data.longitude
    });
  },

  callElderly() {
    wx.makePhoneCall({ phoneNumber: '13800138000' });
  },

  shareLocation() {
    wx.showShareMenu({ withShareTicket: true });
  },

  showPlaceDetail(e) {
    const place = e.currentTarget.dataset.place;
    wx.showToast({ title: `导航到 ${place.name}`, icon: 'none' });
  },

  onRegionChange(e) {
    // Handle map region changes
  },

  onMarkerTap(e) {
    // Handle marker taps
  },

  goBack() {
    wx.navigateBack();
  },

  _drawTrajectoryChart() {
    const ctx = wx.createCanvasContext('trajectoryChart', this);
    const w = 700, h = 180;
    const pad = { top: 20, bottom: 30, left: 10, right: 10 };
    const chartW = w - pad.left - pad.right;
    const chartH = h - pad.top - pad.bottom;

    // Background
    ctx.setFillStyle('#FFFFFF');
    ctx.fillRect(0, 0, w, h);

    // Grid
    ctx.setStrokeStyle('#F0F0F5');
    ctx.setLineWidth(1);
    for (let i = 0; i <= 3; i++) {
      const y = pad.top + (chartH / 3) * i;
      ctx.beginPath();
      ctx.moveTo(pad.left, y);
      ctx.lineTo(w - pad.right, y);
      ctx.stroke();
    }

    // Simulated trajectory points
    const points = [0.6, 0.5, 0.7, 0.4, 0.8, 0.6, 0.9, 0.7, 0.5, 0.8, 0.6, 0.4];
    const stepX = chartW / (points.length - 1);

    // Area fill
    ctx.beginPath();
    ctx.moveTo(pad.left, pad.top + chartH);
    for (let i = 0; i < points.length; i++) {
      const x = pad.left + stepX * i;
      const y = pad.top + chartH - points[i] * chartH;
      if (i === 0) ctx.lineTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.lineTo(pad.left + stepX * (points.length - 1), pad.top + chartH);
    ctx.closePath();
    ctx.setFillStyle('rgba(74, 144, 217, 0.1)');
    ctx.fill();

    // Line
    ctx.beginPath();
    ctx.setStrokeStyle('#4A90D9');
    ctx.setLineWidth(3);
    for (let i = 0; i < points.length; i++) {
      const x = pad.left + stepX * i;
      const y = pad.top + chartH - points[i] * chartH;
      if (i === 0) ctx.moveTo(x, y);
      else ctx.lineTo(x, y);
    }
    ctx.stroke();

    // Dots
    for (let i = 0; i < points.length; i++) {
      const x = pad.left + stepX * i;
      const y = pad.top + chartH - points[i] * chartH;
      ctx.beginPath();
      ctx.arc(x, y, 4, 0, Math.PI * 2);
      ctx.setFillStyle('#4A90D9');
      ctx.fill();
    }

    // X labels
    ctx.setFontSize(18);
    ctx.setFillStyle('#999999');
    const hours = ['6时', '8时', '10时', '12时', '14时', '16时', '18时', '20时'];
    for (let i = 0; i < hours.length; i++) {
      const x = pad.left + (chartW / (hours.length - 1)) * i;
      ctx.fillText(hours[i], x - 14, h - 8);
    }

    ctx.draw();
  }
});

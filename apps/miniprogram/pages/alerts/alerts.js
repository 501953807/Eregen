const api = require('../../utils/api.js');

Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒'],
    alerts: [],
    loading: true,
  },

  /**
   * Page load — fetch alerts from backend API.
   */
  async onLoad() {
    this.setData({ loading: true });
    try {
      // Fetch unread alerts first; fall back to all if endpoint not ready
      const resp = await api.request('/api/v1/alerts', { method: 'GET', query: { unread: 'true' } });
      this.setData({ alerts: resp.data || [], loading: false });
    } catch (e) {
      // If alert endpoint is not available yet, show empty state with warning
      console.warn('Alert endpoint not ready:', e);
      this.setData({ alerts: [], loading: false });
    }
  },

  /**
   * Switch filter tab.
   */
  switchFilter(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ filterTab: index });
  },

  /**
   * Acknowledge a single alert (mark as read/processed).
   * @param {string} alertId
   */
  async handleAcknowledge(alertId) {
    try {
      await api.request(`/api/v1/alerts/${alertId}/acknowledge`, { method: 'POST' });
      // Remove from local list or update status
      const alerts = this.data.alerts.map(a => a.id === alertId ? { ...a, status: 'read', acknowledgedAt: new Date().toISOString() } : a);
      this.setData({ alerts });
      wx.showToast({ title: '已标记为已处理', icon: 'success', duration: 1500 });
    } catch (e) {
      wx.showToast({ title: '操作失败', icon: 'none' });
      console.error('Acknowledge alert failed:', e);
    }
  },

  /**
   * Open system map at the alert location.
   * @param {string} alertId
   */
  async handleOpenLocation(alertId) {
    const alert = this.data.alerts.find(a => a.id === alertId);
    if (!alert) return;

    // In production, get lat/lon from alert data and call openLocation
    wx.openLocation({
      latitude: alert.lat || 0,
      longitude: alert.lon || 0,
      name: alert.title || '告警位置',
      scale: 15,
    });
  },

  /**
   * Format timestamp for display (HH:mm).
   * @param {string|number} ts
   */
  formatTime(ts) {
    const date = typeof ts === 'string' ? new Date(ts) : new Date(ts * 1000);
    return date.toLocaleString('en-US', { hour: '2-digit', minute: '2-digit' });
  },
})


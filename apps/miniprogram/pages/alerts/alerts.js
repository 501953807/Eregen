// pages/alerts/alerts.js
const api = require('../../utils/api.js');

Page({
  data: {
    filterTab: 0,
    filters: ['全部', '未处理', 'SOS', '跌倒'],
    alerts: [],
    loading: true,
  },

  async onLoad() {
    this.setData({ loading: true });
    try {
      const resp = await api.request('/alerts?unread=true', {}, 'GET');
      this.setData({ alerts: resp?.data || [], loading: false });
    } catch (e) {
      console.warn('Alert endpoint not ready:', e);
      this.setData({ alerts: [], loading: false });
    }
  },

  switchFilter(e) {
    const index = e.currentTarget.dataset.index;
    this.setData({ filterTab: index });
  },

  async handleAcknowledge(alertId) {
    try {
      await api.request(`/alerts/${alertId}/acknowledge`, {}, 'POST');
      const alerts = this.data.alerts.map(a =>
        a.id === alertId ? { ...a, status: 'read', acknowledgedAt: new Date().toISOString() } : a
      );
      this.setData({ alerts });
      wx.showToast({ title: '已标记为已处理', icon: 'success', duration: 1500 });
    } catch (e) {
      wx.showToast({ title: '操作失败', icon: 'none' });
      console.error('Acknowledge alert failed:', e);
    }
  },

  async handleOpenLocation(alertId) {
    const alert = this.data.alerts.find(a => a.id === alertId);
    if (!alert) return;
    wx.openLocation({
      latitude: alert.lat || 0,
      longitude: alert.lon || 0,
      name: alert.title || '告警位置',
      scale: 15,
    });
  },

  formatTime(ts) {
    const date = typeof ts === 'string' ? new Date(ts) : new Date(ts * 1000);
    return date.toLocaleString('en-US', { hour: '2-digit', minute: '2-digit' });
  },
})
